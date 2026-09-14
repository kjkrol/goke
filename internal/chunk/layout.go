package chunk

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/comp"
)

type Layout struct {
	ChunkCap   uint32
	ChunkBytes uintptr
	Offsets    []uintptr
	NeedsScan  bool

	// ChunkType describes one chunk as a Go type: an array of entity ids
	// followed by one array per component column. Set only when NeedsScan.
	ChunkType reflect.Type
}

func (l *Layout) Init(compDefs []comp.Def) {
	entityStride := unsafe.Sizeof(uid.UID64(0))
	totalStride := entityStride
	needsScan := false
	for _, compDef := range compDefs {
		totalStride += compDef.Size
		if compDef.NeedsScan {
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

		chunkBytes := alignUp(currentOffset, unsafe.Sizeof(unsafe.Pointer(nil)))

		if capacity == 1 || (chunkBytes <= L1DataCacheSize && !hasCacheSetConflict(offsets)) {
			l.ChunkCap = uint32(capacity)
			l.ChunkBytes = chunkBytes
			l.Offsets = offsets
			if needsScan {
				l.ChunkType = buildChunkType(compDefs, capacity, l)
			}
			return
		}

		capacity--
	}

	panic("unreachable")
}

func alignUp(ptr, align uintptr) uintptr {
	return (ptr + align - 1) & ^(align - 1)
}

// buildChunkType renders the layout as a Go type, asserting that its offsets
// and size match the ones Init computed.
func buildChunkType(compDefs []comp.Def, capacity uintptr, l *Layout) reflect.Type {
	fields := make([]reflect.StructField, 0, len(compDefs)+1)
	fields = append(fields, reflect.StructField{
		Name: "IDs",
		Type: reflect.ArrayOf(int(capacity), reflect.TypeFor[uid.UID64]()),
	})
	for i, compDef := range compDefs {
		fields = append(fields, reflect.StructField{
			Name: "F" + strconv.Itoa(i),
			Type: reflect.ArrayOf(int(capacity), columnElem(compDef)),
		})
	}

	t := reflect.StructOf(fields)

	for i := range t.NumField() {
		if got, want := t.Field(i).Offset, l.Offsets[i]; got != want {
			panic(fmt.Sprintf(
				"chunk: column %d of the generated chunk type sits at offset %d but the layout computed %d — "+
					"the two must agree or every column access is misaddressed", i, got, want))
		}
	}
	if t.Size() != l.ChunkBytes {
		panic(fmt.Sprintf(
			"chunk: generated chunk type is %d bytes but the layout reserved %d — the stride between chunks "+
				"must match the type exactly or every chunk past the first is misaddressed",
			t.Size(), l.ChunkBytes))
	}
	return t
}

// columnElem is the element type for one column's array — the component's own
// type when known, otherwise an opaque stand-in of the same size and alignment.
func columnElem(d comp.Def) reflect.Type {
	if d.Type != nil {
		return d.Type
	}
	if d.NeedsScan {
		panic(fmt.Sprintf(
			"chunk: component %d is marked NeedsScan but carries no reflect.Type — its pointer bitmap cannot be derived", d.ID))
	}
	return opaqueType(d.Size, d.Align)
}

// opaqueType builds a pointer-free type of the given size and alignment.
func opaqueType(size, align uintptr) reflect.Type {
	elem := reflect.TypeFor[byte]()
	switch {
	case align >= 8 && size%8 == 0:
		elem = reflect.TypeFor[uint64]()
	case align >= 4 && size%4 == 0:
		elem = reflect.TypeFor[uint32]()
	case align >= 2 && size%2 == 0:
		elem = reflect.TypeFor[uint16]()
	}
	return reflect.ArrayOf(int(size/elem.Size()), elem)
}
