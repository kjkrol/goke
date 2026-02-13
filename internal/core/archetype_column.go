package core

import (
	"reflect"
	"unsafe"
)

//------------------------------------------------------------------------------
//							Column
//------------------------------------------------------------------------------

// Column represents "Hot Data". It contains only what's necessary for
// high-performance iteration in systems. Fits in ~24 bytes.
type Column struct {
	CompID   ComponentID
	Data     unsafe.Pointer
	ItemSize uintptr
}

func (c *Column) GetElement(row ArchRow) unsafe.Pointer {
	return unsafe.Add(c.Data, uintptr(row)*c.ItemSize)
}

func (c *Column) CopyData(dstRow, srcRow ArchRow) {
	src := c.GetElement(srcRow)
	dst := c.GetElement(dstRow)
	copyMemory(dst, src, c.ItemSize)
}

func (c *Column) ZeroData(row ArchRow) {
	ptr := c.GetElement(row)
	zeroMemory(ptr, c.ItemSize)
}

func (c *Column) SetData(row ArchRow, src unsafe.Pointer) {
	dest := unsafe.Add(c.Data, uintptr(row)*c.ItemSize)
	copyMemory(dest, src, c.ItemSize)
}

func (c *Column) Clear(fullCap uintptr) {
	zeroMemory(c.Data, c.ItemSize*fullCap)
	c.CompID = 0
	c.ItemSize = 0
}

//------------------------------------------------------------------------------
//							columnMeta
//------------------------------------------------------------------------------

// columnMeta represents "Cold Data". Used only during allocation/resize.
type columnMeta struct {
	rawSlice reflect.Value // prevent GC from garbage collecting
	dataType reflect.Type
}

func (c *columnMeta) Clear() {
	c.rawSlice = reflect.Value{}
	c.dataType = nil
}

//------------------------------------------------------------------------------
//							Memory Block
//------------------------------------------------------------------------------

// MemoryBlock holds the physical memory.
// It separates Hot (Columns) from Cold (Meta) data for cache efficiency.
type MemoryBlock struct {
	// Columns[0] to zawsze EntityID
	Columns []Column
	Meta    []columnMeta
	Len     uint32
	Cap     uint32
}

// Init initializes the memory block and performs the first allocation.
// Unlike NewMemoryBlock, this works on an existing struct (embedding).
func (b *MemoryBlock) Init(cap int, colInfos []ComponentInfo) {
	// Count = Entity Column (1) + Component Columns (N)
	count := len(colInfos) + 1

	b.Columns = make([]Column, count)
	b.Meta = make([]columnMeta, count)
	b.Len = 0
	b.Cap = 0 // Will force allocation in Resize

	// 1. Setup Entity Column (Index 0)
	b.Meta[0] = columnMeta{dataType: reflect.TypeFor[Entity]()}
	b.Columns[0] = Column{
		CompID:   0,
		ItemSize: unsafe.Sizeof(Entity(0)),
	}

	// 2. Setup Component Columns (Index 1..N)
	for i, info := range colInfos {
		idx := i + 1
		b.Meta[idx] = columnMeta{dataType: info.Type}
		b.Columns[idx] = Column{
			CompID:   info.ID,
			ItemSize: info.Size,
		}
	}

	// 3. Initial Allocation
	if cap <= 0 {
		cap = 1
	}
	b.resize(uint32(cap))
}

func (b *MemoryBlock) EnsureCapacity(required uint32) {
	if required <= b.Cap {
		return
	}
	newCap := b.Cap * 2
	if newCap < required {
		newCap = required
	}
	b.resize(newCap)
}

func (b *MemoryBlock) resize(newCap uint32) {
	for i := range b.Columns {
		col := &b.Columns[i] // Hot
		meta := &b.Meta[i]   // Cold

		// 1. Allocate new slice using reflection (Cold path)
		newSlice := reflect.MakeSlice(reflect.SliceOf(meta.dataType), int(newCap), int(newCap))
		newPtr := newSlice.UnsafePointer()

		// 2. Copy existing data (if any)
		if b.Len > 0 {
			oldSizeBytes := uintptr(b.Len) * col.ItemSize
			copyMemory(newPtr, col.Data, oldSizeBytes)
		}

		// 3. Update Hot Column & Cold Meta
		col.Data = newPtr
		meta.rawSlice = newSlice // Keep reference for GC
	}
	b.Cap = newCap
}

func (b MemoryBlock) Reset() {
	for i := range b.Columns {
		b.Columns[i].Clear(uintptr(b.Cap))
		b.Meta[i].Clear()
	}
	b.Len = 0
	b.Cap = 0
}

// -----------------------------------------------------------------------------
// Low-Level Helpers
// -----------------------------------------------------------------------------

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func zeroMemory(ptr unsafe.Pointer, size uintptr) {
	clear(unsafe.Slice((*byte)(ptr), size))
}
