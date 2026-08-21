package orch

import (
	"time"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
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

// RunnableFunc adapts a plain function to Runnable, the way
// http.HandlerFunc adapts a function to http.Handler.
type RunnableFunc func(*CmdBuf, time.Duration)

func (f RunnableFunc) Update(cb *CmdBuf, d time.Duration) { f(cb, d) }
