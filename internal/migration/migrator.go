package migration

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/addr"
	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/colstore"
	"github.com/kjkrol/goke/v2/internal/comp"
)

// Migrator applies a fixed set of structural changes — adding and removing
// components — to a batch of entities in a single bulk archetype migration.
//
// Unlike ent.Editor, which migrates one entity at a time with immediate
// swap-and-pop compaction, Migrator defers compaction until the full batch
// is known, enabling block-level column copies and direct per-entity
// address-index writes with no intermediate buffering.
type Migrator struct {
	spec        comp.EditSpec
	addrBook    *addr.Book
	archCatalog *arch.Catalog
	dst         [arch.MaxID]arch.ID // srcArch → dstArch; NullID = unlink (no components left)

	// defrag owns the scratch buffers for CompactHoles; embedded by value so
	// its memory lives in the same heap allocation as Migrator.
	defrag colstore.Defragmenter

	// scratch holds pre-allocated working memory reused across ApplyChunk calls.
	// All slices grow lazily and never shrink; no GC pressure after warm-up.
	scratch struct {
		ids      []uid.UID64
		slotRefs []colstore.SlotRef
	}
}

// New creates a Migrator and pre-computes the destination archetype for every
// archetype that already exists in catalog. Future archetypes are registered
// via OnNewArchetype, typically called by migration.Catalog on behalf of
// arch.Catalog's onArchetypeCreated hook.
func New(book *addr.Book, catalog *arch.Catalog, spec comp.EditSpec) *Migrator {
	m := &Migrator{spec: spec, addrBook: book, archCatalog: catalog}
	for archID := arch.RootID; archID < catalog.Len(); archID++ {
		m.OnNewArchetype(archID)
	}
	return m
}

// OnNewArchetype pre-computes and stores the destination archetype for srcArchID.
// Called eagerly when a new archetype is added to the catalog so that ApplyChunk
// needs no graph traversal on the hot path.
func (m *Migrator) OnNewArchetype(srcArchID arch.ID) {
	target := srcArchID
	for i := range m.spec.AddDefs {
		d := m.spec.AddDefs[i]
		if !m.archCatalog.Archetypes[target].Mask().IsSet(d.ID) {
			target = m.archCatalog.EnsureEdgeNext(d, target)
		}
	}
	for i := range m.spec.DelDefs {
		d := m.spec.DelDefs[i]
		if m.archCatalog.Archetypes[target].Mask().IsSet(d.ID) {
			next, shouldUnlink := m.archCatalog.EnsureEdgePrev(d, target)
			if shouldUnlink {
				m.dst[srcArchID] = arch.NullID
				return
			}
			target = next
		}
	}
	m.dst[srcArchID] = target
}

// ApplyChunk migrates ids from the single source chunk described by ctx.
// ctx must come from Matcher.ChunkCtx() — it provides ArchID, Ptr, and Idx
// directly, so addr.Book is consulted only for liveness and Slot per entity.
// This is the hot path: one Compact call per chunk, direct index writes,
// zero per-entity archetype or pointer lookups beyond the Slot.
func (m *Migrator) ApplyChunk(ctx arch.ChunkCtx, ids []uid.UID64) {
	n := len(ids)
	if n == 0 {
		return
	}
	m.scratch.ids = growTo(m.scratch.ids, n)
	m.scratch.slotRefs = growTo(m.scratch.slotRefs, n)

	valid := 0
	for _, id := range ids {
		entry, ok := m.addrBook.Get(id)
		if !ok {
			continue
		}
		m.scratch.ids[valid] = id
		m.scratch.slotRefs[valid] = colstore.SlotRef{Ptr: ctx.Ptr, Idx: ctx.Idx, Slot: entry.Slot}
		valid++
	}
	if valid == 0 {
		return
	}
	m.applyGroup(ctx.ArchID, m.scratch.ids[:valid], m.scratch.slotRefs[:valid])
}

// applyGroup migrates all ids from srcArchID to the resolved destination
// archetype. ids and slotRefs are pre-collected with Idx already set (no
// further addr.Book reads). Index updates are written directly as entities
// are copied — no intermediate buffer: destination slots are final at copy
// time (ReserveSlots pre-allocated them, and compaction only relocates
// survivors inside the source table).
func (m *Migrator) applyGroup(srcArchID arch.ID, ids []uid.UID64, slotRefs []colstore.SlotRef) {
	n := len(ids)
	dstArchID := m.dst[srcArchID]
	if dstArchID == arch.NullID {
		m.removeAll(srcArchID, ids, slotRefs)
		return
	}

	srcTable := &m.archCatalog.Archetypes[srcArchID].Table
	dstTable := &m.archCatalog.Archetypes[dstArchID].Table

	firstIdx, firstAvailable, chunkCap := dstTable.ReserveSlots(n)

	dstChunkIdx := firstIdx
	available := firstAvailable
	remaining := n
	ei := 0

	for remaining > 0 {
		batchN := min(remaining, available)
		dstPtr, startSlot := dstTable.AllocSlots(dstChunkIdx, batchN)

		// Copy in maximal contiguous runs: ids captured from Query.All() come in
		// chunk order, so consecutive slotRefs usually form one ascending slot
		// range — each run costs one CopyMemory per column instead of per slot.
		i := 0
		for i < batchN {
			base := slotRefs[ei]
			run := 1
			for i+run < batchN &&
				slotRefs[ei+run].Ptr == base.Ptr &&
				slotRefs[ei+run].Slot == base.Slot+colstore.Slot(run) {
				run++
			}

			dstSlot := startSlot + colstore.Slot(i)
			dstTable.CopyRangeFrom(srcTable, base.Ptr, base.Slot, dstPtr, dstSlot, run)
			dstTable.SetEntityRange(dstPtr, dstSlot, ids[ei:ei+run])
			for k := range run {
				m.addrBook.MoveUnchecked(ids[ei+k], dstArchID, dstPtr, dstSlot+colstore.Slot(k))
			}

			i += run
			ei += run
		}

		remaining -= batchN
		dstChunkIdx++
		available = chunkCap
	}

	dstTable.ReleaseSlots()

	srcMoves := m.defrag.Compact(srcTable, slotRefs)
	for _, sm := range srcMoves {
		m.addrBook.MoveUnchecked(sm.ID, srcArchID, sm.NewPtr, sm.NewSlot)
	}
}

// removeAll removes all entities from their source archetype and recycles IDs.
// Used when the migration spec leaves entities with no components.
func (m *Migrator) removeAll(srcArchID arch.ID, ids []uid.UID64, srcPositions []colstore.SlotRef) {
	srcTable := &m.archCatalog.Archetypes[srcArchID].Table
	for i, id := range ids {
		src := srcPositions[i]
		swapped, didSwap := srcTable.RemoveAt(src.Ptr, src.Slot)
		if didSwap {
			m.addrBook.Move(swapped, srcArchID, src.Ptr, src.Slot)
		}
		m.addrBook.Delete(id)
	}
}

// growTo returns s resized to exactly n elements, reallocating only when
// the current capacity is insufficient. The slice never shrinks.
func growTo[T any](s []T, n int) []T {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]T, n)
}
