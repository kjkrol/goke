package core

import (
	"time"
	"unsafe"

	"github.com/kjkrol/uid"
)

type ReadOnlyRegistry interface {
	ComponentGet(e uid.UID64, compID ComponentID) (unsafe.Pointer, error)
}

type System interface {
	Update(ReadOnlyRegistry, *SystemCommandBuffer, time.Duration)
}
