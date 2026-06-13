package goke

import (
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/orch"
	"github.com/kjkrol/goke/internal/reg"
	"github.com/kjkrol/uid"
)

type (
	// EntityID is a 64-bit unique identifier for an entity in the ECS world.
	EntityID = uid.UID64
	// CompID is the unique integer identifier for a registered component type.
	CompID = comp.ID
	// CompMeta holds type metadata for a registered component (ID, size, alignment, reflect.Type).
	CompMeta = comp.Meta

	// Config holds initialization parameters for the ECS.
	Config = reg.Config
	// RunCtx provides methods to schedule systems sequentially or in parallel within a Plan.
	RunCtx = orch.RunCtx
	// Plan defines the execution order and concurrency of systems each tick.
	Plan = orch.Plan

	// Lookup provides a read-only view of component data for use inside system updates.
	Lookup = orch.Lookup
	// CmdBuf queues structural changes (add/remove component, destroy entity) during a tick.
	// Changes are applied at the next synchronization point, keeping iterators valid.
	CmdBuf = orch.CmdBuf
)
