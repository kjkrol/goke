package ecs

import (
	"reflect"
	"unsafe"
)

// column to teraz struktura trzymana w tablicy, a nie wskaźnik na stercie
type column struct {
	id       ComponentID    // ID komponentu (niezbędne przy iteracji po slice)
	data     unsafe.Pointer // Wskaźnik na początek bloku pamięci
	dataType reflect.Type
	len      int
	cap      int
	itemSize uintptr
}

// GetElement zwraca wskaźnik do danych konkretnej encji (row)
// inline hint: to powinno być inlinowane przez kompilator
func (c *column) GetElement(index int) unsafe.Pointer {
	return unsafe.Add(c.data, uintptr(index)*c.itemSize)
}

func (c *column) growTo(newCap int) {
	// Tworzymy nowy slice o nowym rozmiarze używając reflect (zachowuje typowanie dla GC)
	newSlice := reflect.MakeSlice(reflect.SliceOf(c.dataType), newCap, newCap)
	newPtr := newSlice.UnsafePointer()

	if c.len > 0 {
		// Kopiujemy stare dane do nowego bloku
		copyMemory(newPtr, c.data, uintptr(c.len)*c.itemSize)
	}

	c.data = newPtr
	c.cap = newCap
}

// zeroData zeruje pamięć usuniętej encji (dla bezpieczeństwa GC, jeśli są tam wskaźniki)
func (c *column) zeroData(index int) {
	ptr := c.GetElement(index)
	zeroMemory(ptr, c.itemSize)
}

// copyData przenosi dane wewnątrz kolumny (Swap & Pop)
func (c *column) copyData(dstIdx, srcIdx int) {
	src := c.GetElement(srcIdx)
	dst := c.GetElement(dstIdx)
	copyMemory(dst, src, c.itemSize)
}

// setData wpisuje nowe dane z zewnątrz (np. przy dodawaniu encji)
func (c *column) setData(rowIdx int, src unsafe.Pointer) {
	dest := c.GetElement(rowIdx)
	memmove(dest, src, c.itemSize)
}

// Low-level memory ops
//
//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, n uintptr)

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	// Użycie unsafe.Slice jest bardzo szybkie w Go 1.17+
	copy(unsafe.Slice((*byte)(dst), int(size)), unsafe.Slice((*byte)(src), int(size)))
}

func zeroMemory(ptr unsafe.Pointer, size uintptr) {
	s := unsafe.Slice((*byte)(ptr), int(size))
	for i := range s {
		s[i] = 0
	}
}
