package goke

import (
	"time"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/orch"
)

// System is the interface for stateful logic units that process entity data each tick.
// Init is called once on registration; Update is called every tick.
type System interface {
	Update(Lookup, *CmdBuf, time.Duration)
	Init(*ECS)
}

// SystemFn is a stateless function-based system — a shorthand alternative to implementing System.
type SystemFn func(*CmdBuf, time.Duration)

// CmdBufAddComp queues the addition of a component value to an entity.
// If the entity already has this component, its data is overwritten on flush.
func CmdBufAddComp[T any](cb *CmdBuf, e uid.UID64, compDef CompDef, value T) {
	orch.AddComp(cb, e, compDef, value)
}

type functionalSystem struct {
	updateFn SystemFn
}

func (f *functionalSystem) Init(*ECS) {}

func (f *functionalSystem) Update(_ Lookup, cb *CmdBuf, d time.Duration) {
	f.updateFn(cb, d)
}

var _ System = (*functionalSystem)(nil)
