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

// AddOne queues adding a component value to a single entity reached
// outside a Query loop (an external event, a saved id) — not for entities
// already being visited in a loop, where Query.BeginMigrate + Editor/
// ValueEditor batches the change instead. Overwrites existing data.
func (cb *CmdBuf) AddOne[T any](e uid.UID64, compID CompID, value T) {
	orch.AddOne(cb.raw, e, compID, value)
}

// RemoveCompOne queues removing a component from a single entity reached
// outside a Query loop — not for entities already being visited in a loop,
// where Query.BeginMigrate + Editor batches the change instead. No-op if
// the entity doesn't have it.
func (cb *CmdBuf) RemoveCompOne(id uid.UID64, compID CompID) { cb.raw.RemoveCompOne(id, compID) }

// RemoveOne queues removing a single entity reached outside a Query loop —
// not for entities already being visited in a loop, where Query.BeginMigrate
// + Remover batches the change instead.
func (cb *CmdBuf) RemoveOne(id uid.UID64) { cb.raw.RemoveOne(id) }
