package chunk

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
)

type Layout struct {
	ChunkCap   uint32
	ChunkBytes uintptr
	Offsets    []uintptr
}

func (l *Layout) Init(compDefs []comp.Def) {
	entityStride := unsafe.Sizeof(uid.UID64(0))
	totalStride := entityStride
	for _, compDef := range compDefs {
		totalStride += compDef.Size
	}

	capacity := uintptr(L1DataCacheSize) / totalStride
	if capacity == 0 {
		capacity = 1
	}

	for capacity >= 1 {
		offsets := make([]uintptr, len(compDefs)+1)
		currentOffset := uintptr(0)

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
