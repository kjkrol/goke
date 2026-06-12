package core

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"
)

type ComponentReader interface {
	ComponentGet(e uid.UID64, compID ComponentID) (unsafe.Pointer, error)
}

type ComponentID uint8

const EntityID ComponentID = ComponentID(255)

type ComponentInfo struct {
	ID    ComponentID
	Size  uintptr
	Align uintptr
	Type  reflect.Type
}

