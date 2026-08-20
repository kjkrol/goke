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

	SetupProvider
	CompProvider
}

// SetupProvider contributes one-time, Setup-phase logic — e.g. world
// seeding — separate from Module's per-tick RegSystems/RunPlan. Part of
// Module, but also usable standalone: a value with nothing to run every
// tick (just a one-time spawn, like a populate helper) can implement just
// this, without the rest of Module.
type SetupProvider interface {
	// SetupSystems returns systems to run once, in order, via ECS.Setup —
	// before Plan-based ticking starts.
	SetupSystems() []System
}
