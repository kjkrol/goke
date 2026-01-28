package core

import (
	"time"
	"unsafe"
)

type ReadOnlyRegistry interface {
	GetComponent(e Entity, compID ComponentID) (unsafe.Pointer, error)
}

type System interface {
	Update(ReadOnlyRegistry, *SystemCommandBuffer, time.Duration)
}
