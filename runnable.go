package goke

import (
	"time"

	"github.com/kjkrol/goke/v2/internal/orch"
)

// Runnable is an opaque handle returned by RegSys/RegSysFn — pass it to
// RunCtx.Run/RunParallel inside a Plan. It is not meant to be called
// directly; Update is driven by the scheduler.
type Runnable = orch.Runnable

// runnableAdapter bridges System (goke.CmdBuf) to orch.Runnable
// (orch.CmdBuf) — built once at registration, not per tick, so it adds no
// hot-path allocation.
type runnableAdapter struct {
	sys     System
	wrapped *CmdBuf
}

func (a *runnableAdapter) Update(_ *orch.CmdBuf, d time.Duration) {
	a.sys.Update(a.wrapped, d)
}
