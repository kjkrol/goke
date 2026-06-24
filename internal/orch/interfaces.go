package orch

import (
	"time"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/comp"
)

type Mutator interface {
	UpsertComp(uid.UID64, comp.ID) (unsafe.Pointer, error)
	RemoveComp(uid.UID64, comp.ID) error
	Remove(uid.UID64) bool
}

type Runnable interface {
	Update(*CmdBuf, time.Duration)
}
