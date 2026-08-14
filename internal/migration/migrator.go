package migration

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/addr"
	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/colstore"
	"github.com/kjkrol/goke/v2/internal/comp"
)

// unresolvedID marks a dst entry whose destination archetype has not been
// resolved yet; it can never collide with a real ID (all are < arch.MaxID).
const unresolvedID = arch.MaxID

// Migrator applies a fixed set of component adds/removes to whole batches of
// entities. Unlike ent.Editor (one entity at a time, immediate swap-and-pop),
// it defers compaction until the full batch is known, enabling block column
// copies and a single address-index pass per batch.
type Migrator struct {
	spec        comp.EditSpec
	addrBook    *addr.Book
	archCatalog *arch.Catalog

	// dst memoizes srcArch → dstArch; NullID = unlink (no components left).
	// Resolved lazily on first use — once per source archetype, amortized
	// over whole chunk batches.
	dst [arch.MaxID]arch.ID

	defrag colstore.Defragmenter // by value: shares Migrator's heap allocation

	// scratch is reused across Migrate calls; grows lazily, never shrinks.
	scratch struct {
		ids      []uid.UID64
		slotRefs []colstore.SlotRef
		dstRuns  []dstChunkRun
	}
}

// dstChunkRun records where one destination chunk's batch landed, letting the
// address-book pass recompute every position arithmetically — no per-entity buffer.
type dstChunkRun struct {
	ptr       unsafe.Pointer
	startSlot colstore.Slot
	n         int
}

func New(book *addr.Book, catalog *arch.Catalog, spec comp.EditSpec) *Migrator {
	m := &Migrator{spec: spec, addrBook: book, archCatalog: catalog}
	for i := range m.dst {
		m.dst[i] = unresolvedID
	}
	return m
}

// resolve computes and memoizes the destination archetype for srcArchID by
// jumping straight to the target composition — no edge-graph waypoints, so
// coexisting wide migrators cannot cross-multiply intermediate archetypes.
func (m *Migrator) resolve(srcArchID arch.ID) arch.ID {
	set := m.archCatalog.Archetypes[srcArchID].Composition()
	for i := range m.spec.AddDefs {
		d := m.spec.AddDefs[i]
		if !set.Mask.IsSet(d.ID) {
			set = set.With(d)
		}
	}
	for i := range m.spec.DelDefs {
		set = set.Without(m.spec.DelDefs[i].ID)
	}
	target := arch.NullID
	if !set.Mask.IsEmpty() {
		target = m.archCatalog.Upsert(set)
	}
	m.dst[srcArchID] = target
	return target
}

// Migrate moves ids out of the single source chunk described by snap.
func (m *Migrator) Migrate(snap bulk.ChunkSnapshot, ids []uid.UID64) {
	if len(ids) == 0 {
		return
	}
	srcTable := &m.archCatalog.Archetypes[snap.ArchID].Table
	validIDs, slotRefs := resolveSlotRefs(m.addrBook, srcTable, snap, ids, &m.scratch.ids, &m.scratch.slotRefs)
	if len(validIDs) == 0 {
		return
	}
	m.applyGroup(snap.ArchID, validIDs, slotRefs)
}

// applyGroup migrates ids to the pre-resolved destination in two phases —
// bytes first (block copies, then source compaction), address book second in
// one homogeneous pass — so streaming copies and scattered index writes never
// interleave. Destination positions are final at copy time, so the second
// phase replays them arithmetically from dstRuns; source relocations come
// from the compaction's SlotMove list.
func (m *Migrator) applyGroup(srcArchID arch.ID, ids []uid.UID64, slotRefs []colstore.SlotRef) {
	n := len(ids)
	srcTable := &m.archCatalog.Archetypes[srcArchID].Table

	dstArchID := m.dst[srcArchID]
	if dstArchID == unresolvedID {
		dstArchID = m.resolve(srcArchID)
	}
	if dstArchID == arch.NullID {
		removeBatch(m.addrBook, &m.defrag, srcArchID, srcTable, ids, slotRefs)
		return
	}
	if dstArchID == srcArchID {
		// Every AddDef was already present and no DelDef matched anything —
		// net composition is unchanged. Entities stay exactly where they
		// are: nothing to copy, no address-book update, no compaction.
		return
	}

	dstTable := &m.archCatalog.Archetypes[dstArchID].Table

	firstIdx, firstAvailable, chunkCap := dstTable.ReserveSlots(n)

	m.scratch.dstRuns = m.scratch.dstRuns[:0]
	dstChunkIdx := firstIdx
	available := firstAvailable
	remaining := n
	ei := 0

	for remaining > 0 {
		batchN := min(remaining, available)
		dstPtr, startSlot := dstTable.AllocSlots(dstChunkIdx, batchN)
		m.scratch.dstRuns = append(m.scratch.dstRuns, dstChunkRun{ptr: dstPtr, startSlot: startSlot, n: batchN})

		// Ids arrive in chunk order, so slotRefs form long ascending runs —
		// each run costs one CopyMemory per column instead of per slot.
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

			i += run
			ei += run
		}

		remaining -= batchN
		dstChunkIdx++
		available = chunkCap
	}

	dstTable.ReleaseSlots()

	srcMoves := m.defrag.Compact(srcTable, slotRefs)

	ei = 0
	for _, r := range m.scratch.dstRuns {
		for k := range r.n {
			m.addrBook.MoveUnchecked(ids[ei+k], dstArchID, r.ptr, r.startSlot+colstore.Slot(k))
		}
		ei += r.n
	}
	for _, sm := range srcMoves {
		m.addrBook.MoveUnchecked(sm.ID, srcArchID, sm.NewPtr, sm.NewSlot)
	}
}
