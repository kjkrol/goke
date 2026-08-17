package goke

import (
	"time"
)

// System is the interface for stateful logic units that process entity data each tick.
// Init is called once on registration; Update is called every tick.
type System interface {
	Update(*CmdBuf, time.Duration)
	Init(*SysInit)
}

// SystemFn is a lightweight System — set OnInit and/or OnUpdate directly as
// a composite literal instead of declaring a named type with methods. Either
// field may be nil.
type SystemFn struct {
	OnInit   func(*SysInit)
	OnUpdate func(*CmdBuf, time.Duration)
}

func (f SystemFn) Init(si *SysInit) {
	if f.OnInit != nil {
		f.OnInit(si)
	}
}

func (f SystemFn) Update(cb *CmdBuf, d time.Duration) {
	if f.OnUpdate != nil {
		f.OnUpdate(cb, d)
	}
}

var _ System = SystemFn{}
