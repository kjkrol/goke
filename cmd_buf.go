package goke

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/orch"
)

// CmdBuf queues deferred structural changes during a tick, applied at the
// next synchronization point. The only documented way to queue a bulk
// change is Query.BeginMigrate/Add/Commit — the underlying batch primitives
// are intentionally not exposed here.
type CmdBuf struct {
	raw *orch.CmdBuf
}

// AddOne queues the addition of a component value to an entity, for use
// outside Query iteration (a lone id from an external event, callback, or
// saved reference). If the entity already has this component, its data is
// overwritten on flush.
func AddOne[T any](cb *CmdBuf, e uid.UID64, compID CompID, value T) {
	orch.AddOne(cb.raw, e, compID, value)
}

// RemoveCompOne queues the removal of a single component from an entity,
// for use outside Query iteration. No-op if the entity doesn't have it.
func (cb *CmdBuf) RemoveCompOne(id uid.UID64, compID CompID) { cb.raw.RemoveCompOne(id, compID) }

// RemoveOne queues the removal of a whole entity, for use outside Query
// iteration.
func (cb *CmdBuf) RemoveOne(id uid.UID64) { cb.raw.RemoveOne(id) }
