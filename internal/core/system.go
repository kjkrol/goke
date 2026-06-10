package core

import "github.com/kjkrol/uid"

import (
	"time"
	"unsafe"
)

type ReadOnlyRegistry interface {
	ComponentGet(e uid.UID64, compID ComponentID) (unsafe.Pointer, error)
}

type System interface {
	Update(ReadOnlyRegistry, *SystemCommandBuffer, time.Duration)
}
