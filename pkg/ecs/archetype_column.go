package ecs

import (
	"reflect"
	"unsafe"
)

type column struct {
	data     unsafe.Pointer
	rawSlice reflect.Value // prevent GC from garbage collecting
	dataType reflect.Type
	len      int
	cap      int
	itemSize uintptr
}

func (c *column) GetElement(row ArchRow) unsafe.Pointer {
	return unsafe.Add(c.data, uintptr(row)*c.itemSize)
}

func (c *column) growTo(newCap int) {
	newSlice := reflect.MakeSlice(reflect.SliceOf(c.dataType), newCap, newCap)
	newPtr := newSlice.UnsafePointer()

	if c.len > 0 {
		copyMemory(newPtr, c.data, uintptr(c.len)*c.itemSize)
	}

	c.data = newPtr
	c.rawSlice = newSlice
	c.cap = newCap
}

func (c *column) zeroData(row ArchRow) {
	ptr := c.GetElement(row)
	zeroMemory(ptr, c.itemSize)

	if int(row) >= c.len {
		c.len = int(row + 1)
	}
}

func (c *column) copyData(dstIdx, srcIdx ArchRow) {
	src := c.GetElement(srcIdx)
	dst := c.GetElement(dstIdx)
	copyMemory(dst, src, c.itemSize)
}

func (c *column) setData(row ArchRow, src unsafe.Pointer) {
	dest := unsafe.Add(c.data, uintptr(row)*c.itemSize)
	memmove(dest, src, c.itemSize)
}

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func zeroMemory(ptr unsafe.Pointer, size uintptr) {
	clear(unsafe.Slice((*byte)(ptr), size))
}

//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, n uintptr)
