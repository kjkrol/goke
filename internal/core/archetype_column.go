package core

import (
	"reflect"
	"unsafe"
)

type Column struct {
	Data     unsafe.Pointer
	rawSlice reflect.Value // prevent GC from garbage collecting
	dataType reflect.Type
	len      int
	cap      int
	ItemSize uintptr
}

func (c *Column) GetElement(row ArchRow) unsafe.Pointer {
	return unsafe.Add(c.Data, uintptr(row)*c.ItemSize)
}

func (c *Column) growTo(newCap int) {
	newSlice := reflect.MakeSlice(reflect.SliceOf(c.dataType), newCap, newCap)
	newPtr := newSlice.UnsafePointer()

	if c.len > 0 {
		copyMemory(newPtr, c.Data, uintptr(c.len)*c.ItemSize)
	}

	c.Data = newPtr
	c.rawSlice = newSlice
	c.cap = newCap
}

func (c *Column) zeroData(row ArchRow) {
	ptr := c.GetElement(row)
	zeroMemory(ptr, c.ItemSize)

	if int(row) >= c.len {
		c.len = int(row + 1)
	}
}

func (c *Column) copyData(dstIdx, srcIdx ArchRow) {
	src := c.GetElement(srcIdx)
	dst := c.GetElement(dstIdx)
	copyMemory(dst, src, c.ItemSize)
}

func (c *Column) setData(row ArchRow, src unsafe.Pointer) {
	dest := unsafe.Add(c.Data, uintptr(row)*c.ItemSize)
	memmove(dest, src, c.ItemSize)
}

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func zeroMemory(ptr unsafe.Pointer, size uintptr) {
	clear(unsafe.Slice((*byte)(ptr), size))
}

//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, n uintptr)
