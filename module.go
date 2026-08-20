package goke

import "time"

// Module bundles a set of related systems and the components they own,
// letting a self-contained package (e.g. a physics or rendering module)
// expose one value that a game registers instead of wiring up each system
// individually.
type Module interface {
	// RegSystems registers this module's systems via ecs.RegSys, storing
	// the resulting Runnable handles for RunPlan to use. Called once by
	// [ECS.RegModule].
	RegSystems(ecs *ECS)

	// RunPlan runs this module's systems in the order and with the Sync
	// points the module requires. Call it from inside your own SetPlan
	// closure, alongside any other systems or modules.
	RunPlan(ctx RunCtx, d time.Duration)

	CompProvider
}

// RegModule registers a Module's systems — equivalent to calling RegSys on
// each of them, except the module keeps the resulting handles to itself
// and replays them, in the required order, from RunPlan.
func (ecs *ECS) RegModule(m Module) {
	m.RegSystems(ecs)
}
