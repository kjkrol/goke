package chunk

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/comp"
)

type Layout struct {
	ChunkCap   uint32
	ChunkBytes uintptr
	Offsets    []uintptr
	NeedsScan  bool
}

func (l *Layout) Init(compDefs []comp.Def) {
	entityStride := unsafe.Sizeof(uid.UID64(0))
	totalStride := entityStride
	needsScan := false
	for _, compDef := range compDefs {
		totalStride += compDef.Size
		if len(comp.OffChunkFields(compDef.Type)) > 0 {
			needsScan = true
		}
	}
	l.NeedsScan = needsScan

	capacity := uintptr(L1DataCacheSize) / totalStride
	if capacity == 0 {
		capacity = 1
	}

	for capacity >= 1 {
		offsets := make([]uintptr, len(compDefs)+1)
		currentOffset := uintptr(0)

		// The entity ID column always sits at chunk offset 0 — orch.CmdBuf's
		// MassMigrate relies on this to detect positional id windows by
		// comparing the ids slice base pointer with the chunk pointer.
		entityAlign := unsafe.Alignof(uid.UID64(0))
		currentOffset = alignUp(currentOffset, entityAlign)
		offsets[0] = currentOffset
		currentOffset += entityStride * capacity

		for i, compDef := range compDefs {
			currentOffset = alignUp(currentOffset, compDef.Align)
			offsets[i+1] = currentOffset
			currentOffset += compDef.Size * capacity
		}

		if capacity == 1 || (currentOffset <= L1DataCacheSize && !hasCacheSetConflict(offsets)) {
			l.ChunkCap = uint32(capacity)
			if needsScan {
				currentOffset = alignUp(currentOffset, unsafe.Sizeof(unsafe.Pointer(nil)))
			}
			l.ChunkBytes = currentOffset
			l.Offsets = offsets
			return
		}

		capacity--
	}

	panic("unreachable")
}

func alignUp(ptr, align uintptr) uintptr {
	return (ptr + align - 1) & ^(align - 1)
}
