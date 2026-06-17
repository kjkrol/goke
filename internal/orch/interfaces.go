package orch

import (
	"time"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
)

type Lookup interface {
	GetComp(e uid.UID64, compID comp.ID) (unsafe.Pointer, error)
}

type Mutator interface {
	UpsertComp(uid.UID64, comp.Meta) (unsafe.Pointer, error)
	RemoveComp(uid.UID64, comp.Meta) error
	Remove(uid.UID64) bool
}

type Runnable interface {
	Update(Lookup, *CmdBuf, time.Duration)
}
