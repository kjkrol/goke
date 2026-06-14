package comp

import (
	"reflect"
)

type ID uint8

const (
	MaskSize         = 2
	MaxComponents    = 64 * MaskSize
	EntityID      ID = ^ID(0) // sentinel — max uint8, outside the valid component ID range (0..MaxComponents-1)
)

type Meta struct {
	ID    ID
	Size  uintptr
	Align uintptr
	Type  reflect.Type
}
