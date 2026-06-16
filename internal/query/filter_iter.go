package query

import (
	"unsafe"
)

// nextFilter advances the View's Filter-mode iterator to the next matching entity.
// Sets v.Entity and v.Idx for the matched entity. Returns false when exhausted.
func (v *View) nextFilter() bool {
	for v.pos < len(v.selected) {
		e := v.selected[v.pos]
		v.Idx = v.pos
		v.pos++
		link, ok := v.EntityIndex.Get(e)
		if !ok {
			continue
		}
		if link.ArchId != v.lastArchID {
			v.bt = v.Get(link.ArchId)
			v.lastArchID = link.ArchId
		}
		if v.bt == nil {
			continue
		}
		v.Entity = e
		v.ptr = v.bt.Table.Chunks[link.Pos.ChunkIdx].Ptr
		v.slot = uintptr(link.Pos.ChunkSlot)
		return true
	}
	return false
}

// at returns a pointer to the component at compIdx for the current Filter-mode entity.
// The row stride is unsafe.Sizeof(T) — a compile-time constant for concrete T —
// so the multiply folds into addressing with no runtime size table.
func at[T any](v *View, compIdx int) *T {
	var zero T
	offset := v.bt.CompOffsets[compIdx] + v.slot*unsafe.Sizeof(zero)
	return (*T)(unsafe.Add(v.ptr, offset))
}
