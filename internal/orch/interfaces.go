package orch

import (
	"time"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/comp"
)

type Mutator interface {
	UpsertComp(uid.UID64, comp.ID) (unsafe.Pointer, error)
	RemoveComp(uid.UID64, comp.ID) error
	Remove(uid.UID64) bool
	// Remover returns a shared bulk.Migrator that removes whole entities,
	// for CmdBuf.Remove to queue against without the caller building one.
	Remover() bulk.Migrator
}

type Runnable interface {
	Update(*CmdBuf, time.Duration)
}
