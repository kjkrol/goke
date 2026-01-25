package ecs

import (
	"time"
	"unsafe"
)

type ReadOnlyRegistry interface {
	GetComponent(e Entity, compID ComponentID) (unsafe.Pointer, error)
}

type System interface {
	Init(*Registry)
	Update(ReadOnlyRegistry, *SystemCommandBuffer, time.Duration)
}
