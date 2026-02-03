package core

import (
	"time"
	"unsafe"
)

type ReadOnlyRegistry interface {
	ComponentGet(e Entity, compID ComponentID) (unsafe.Pointer, error)
}

type System interface {
	Update(ReadOnlyRegistry, *SystemCommandBuffer, time.Duration)
}
