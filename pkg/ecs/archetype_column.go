package ecs

import (
	"reflect"
	"unsafe"
)

type column struct {
	data     unsafe.Pointer
	dataType reflect.Type
	len      int
	cap      int
	itemSize uintptr
}

func (c *column) GetElement(index int) unsafe.Pointer {
	return unsafe.Add(c.data, uintptr(index)*c.itemSize)
}

func (c *column) growTo(newCap int) {
	newSlice := reflect.MakeSlice(reflect.SliceOf(c.dataType), newCap, newCap)
	newPtr := newSlice.UnsafePointer()

	if c.len > 0 {
		copyMemory(newPtr, c.data, uintptr(c.len)*c.itemSize)
	}

	c.data = newPtr
	c.cap = newCap
}

func (c *column) zeroData(index int) {
	ptr := c.GetElement(index)
	zeroMemory(ptr, c.itemSize)

	if index >= c.len {
		c.len = index + 1
	}
}

func (c *column) copyData(dstIdx, srcIdx int) {
	src := c.GetElement(srcIdx)
	dst := c.GetElement(dstIdx)
	copyMemory(dst, src, c.itemSize)
}

func (c *column) setData(rowIdx int, src unsafe.Pointer) {
	dest := unsafe.Add(c.data, uintptr(rowIdx)*c.itemSize)

	memmove(dest, src, c.itemSize)
}

//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, n uintptr)

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func zeroMemory(ptr unsafe.Pointer, size uintptr) {
	s := unsafe.Slice((*byte)(ptr), size)
	for i := range s {
		s[i] = 0
	}
}
