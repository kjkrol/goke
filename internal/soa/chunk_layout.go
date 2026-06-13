package soa

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
)

type ChunkLayout struct {
	ChunkCap   uint32
	ChunkBytes uintptr
	Offsets    []uintptr
}

func CalculateLayout(compMetas []comp.Meta) ChunkLayout {
	totalStride := unsafe.Sizeof(uid.UID64(0))
	for _, compMeta := range compMetas {
		totalStride += compMeta.Size
	}

	capacity := uintptr(L1DataCacheSize) / totalStride
	if capacity == 0 {
		capacity = 1
	}

	for capacity >= 1 {
		offsets := make([]uintptr, len(compMetas)+1)
		currentOffset := uintptr(0)

		entityAlign := unsafe.Alignof(uid.UID64(0))
		currentOffset = alignUp(currentOffset, entityAlign)
		offsets[0] = currentOffset
		currentOffset += unsafe.Sizeof(uid.UID64(0)) * capacity

		for i, compMeta := range compMetas {
			currentOffset = alignUp(currentOffset, compMeta.Align)
			offsets[i+1] = currentOffset
			currentOffset += compMeta.Size * capacity
		}

		if capacity == 1 || currentOffset <= L1DataCacheSize {
			return ChunkLayout{
				ChunkCap:   uint32(capacity),
				ChunkBytes: currentOffset,
				Offsets:    offsets,
			}
		}

		capacity--
	}

	panic("unreachable")
}
