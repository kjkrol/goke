package ent

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/colstore"
	"github.com/kjkrol/goke/v3/internal/comp"
)

// ValueEditor is Editor's value-carrying counterpart: it adds exactly one
// component and writes a caller-supplied value into it per entity, alongside
// any number of removals.
type ValueEditor struct {
	spec        comp.EditSpec // AddDefs has exactly one entry
	addrBook    *addr.Book
	archCatalog *arch.Catalog

	// dst memoizes srcArch → dstArch; NullID = unlink (no components left).
	dst [arch.MaxID]arch.ID

	defrag colstore.Defragmenter // by value: shares ValueEditor's heap allocation

	// scratch is reused across MigrateWithValue calls; grows lazily, never shrinks.
	scratch struct {
		ids      []uid.UID64
		slotRefs []colstore.SlotRef
		origIdx  []int
		dstRuns  []dstChunkRun
		payload  []byte // compacted-payload staging, used only when the slow path drops ids
	}
}

func NewValueEditor(book *addr.Book, catalog *arch.Catalog, spec comp.EditSpec) *ValueEditor {
	vm := &ValueEditor{spec: spec, addrBook: book, archCatalog: catalog}
	for i := range vm.dst {
		vm.dst[i] = unresolvedID
	}
	return vm
}

// ValueType returns the reflect.Type of the one component this ValueEditor
// adds, for callers to validate a Comp[T] against before staging a
// mismatched-type payload — same-sized-but-different types (e.g. two structs
// each holding a pair of float32s) would otherwise defeat a size-only check.
func (vm *ValueEditor) ValueType() reflect.Type { return vm.spec.AddDefs[0].Type }

// resolve computes and memoizes the destination archetype for srcArchID.
func (vm *ValueEditor) resolve(srcArchID arch.ID) arch.ID {
	target := resolveDst(vm.archCatalog, vm.spec, srcArchID)
	vm.dst[srcArchID] = target
	return target
}

// MigrateWithValue satisfies bulk.ValueMigrator. payload holds len(ids)
// contiguous values (the added component's type, elemSize each) in the same
// order as ids.
func (vm *ValueEditor) MigrateWithValue(snap bulk.ChunkSnapshot, ids []uid.UID64, payload unsafe.Pointer) {
	if len(ids) == 0 {
		return
	}
	srcTable := &vm.archCatalog.Archetypes[snap.ArchID].Table
	validIDs, slotRefs := resolveSlotRefs(vm.addrBook, srcTable, snap, ids, &vm.scratch.ids, &vm.scratch.slotRefs, &vm.scratch.origIdx)
	if len(validIDs) == 0 {
		return
	}

	elemSize := vm.spec.AddDefs[0].Size
	if len(validIDs) < len(ids) && elemSize > 0 {
		// Slow path dropped some ids — payload was built against the
		// original ids order, so compact it in lockstep via origIdx (the
		// slow path never reorders, only skips: origIdx[j] is the original
		// index of the j-th survivor).
		origIdx := vm.scratch.origIdx[:len(validIDs)]
		vm.scratch.payload = growTo(vm.scratch.payload, len(validIDs)*int(elemSize))
		compacted := unsafe.Pointer(unsafe.SliceData(vm.scratch.payload))
		for j, i := range origIdx {
			copyMemory(elemAt(compacted, j, elemSize), elemAt(payload, i, elemSize), elemSize)
		}
		payload = compacted
	}

	vm.applyGroup(snap.ArchID, validIDs, slotRefs, payload)
}

// applyGroup mirrors Editor.applyGroup, additionally writing payload into
// the added column alongside the block copy.
func (vm *ValueEditor) applyGroup(srcArchID arch.ID, ids []uid.UID64, slotRefs []colstore.SlotRef, payload unsafe.Pointer) {
	n := len(ids)
	srcTable := &vm.archCatalog.Archetypes[srcArchID].Table

	dstArchID := vm.dst[srcArchID]
	if dstArchID == unresolvedID {
		dstArchID = vm.resolve(srcArchID)
	}
	if dstArchID == arch.NullID {
		removeBatch(vm.addrBook, &vm.defrag, srcArchID, srcTable, ids, slotRefs)
		return
	}

	addDef := vm.spec.AddDefs[0]
	elemSize := addDef.Size

	if dstArchID == srcArchID {
		// dstArchID == srcArchID has two distinct causes: the added component
		// was already present (write in place, below), or Add and Del both
		// targeted it and canceled out in resolveDst — in which case the
		// column never existed in srcTable and ComponentAt returns nil.
		if elemSize == 0 {
			return
		}
		for i, ref := range slotRefs {
			dst := srcTable.ComponentAt(ref.Ptr, ref.Slot, addDef.ID)
			if dst == nil {
				return
			}
			copyMemory(dst, elemAt(payload, i, elemSize), elemSize)
		}
		return
	}

	dstTable := &vm.archCatalog.Archetypes[dstArchID].Table

	firstIdx, firstAvailable, chunkCap := dstTable.ReserveSlots(n)

	vm.scratch.dstRuns = vm.scratch.dstRuns[:0]
	dstChunkIdx := firstIdx
	available := firstAvailable
	remaining := n
	ei := 0

	for remaining > 0 {
		batchN := min(remaining, available)
		dstPtr, startSlot := dstTable.AllocSlots(dstChunkIdx, batchN)
		vm.scratch.dstRuns = append(vm.scratch.dstRuns, dstChunkRun{ptr: dstPtr, startSlot: startSlot, n: batchN})

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
			if elemSize > 0 {
				// nil when Add and Del both targeted this component and
				// canceled it back out of dst's composition — see the
				// dstArchID == srcArchID branch above for the same case.
				if colDst := dstTable.ComponentAt(dstPtr, dstSlot, addDef.ID); colDst != nil {
					copyMemory(colDst, elemAt(payload, ei, elemSize), uintptr(run)*elemSize)
				}
			}

			i += run
			ei += run
		}

		remaining -= batchN
		dstChunkIdx++
		available = chunkCap
	}

	dstTable.ReleaseSlots()

	srcMoves := vm.defrag.Compact(srcTable, slotRefs)

	ei = 0
	for _, r := range vm.scratch.dstRuns {
		for k := range r.n {
			vm.addrBook.MoveUnchecked(ids[ei+k], dstArchID, r.ptr, r.startSlot+colstore.Slot(k))
		}
		ei += r.n
	}
	for _, sm := range srcMoves {
		vm.addrBook.MoveUnchecked(sm.ID, srcArchID, sm.NewPtr, sm.NewSlot)
	}
}

// copyMemory copies size bytes from src to dst. Duplicated locally rather
// than imported — internal/ent's allow-list doesn't include
// internal/chunk (whose CopyMemory does the same thing), matching the
// existing precedent of internal/orch/scheduler.go's own private copy.
func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

// elemAt returns a pointer to the i-th elemSize-wide element of base.
func elemAt(base unsafe.Pointer, i int, elemSize uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(base) + uintptr(i)*elemSize)
}
