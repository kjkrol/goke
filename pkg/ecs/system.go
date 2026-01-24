package ecs

import (
	"time"
	"unsafe"
)

type ReadOnlyRegistry interface {
	GetComponent(e Entity, compID ComponentID) (unsafe.Pointer, error)
	// HasComponent(e Entity, compID ComponentID) bool
}

type System interface {
	Init(*Registry)
	Update(ReadOnlyRegistry, *SystemCommandBuffer, time.Duration)
	ShouldSync() bool
}
