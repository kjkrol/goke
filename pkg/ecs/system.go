package ecs

import (
	"time"

	"github.com/kjkrol/goke/internal/core"
)

type (
	// Lookup provides a thread-safe, read-only view of the components.
	// It is used by systems to inspect the state of entities without the risk
	// of concurrent modification.
	Lookup = core.ReadOnlyRegistry
	// Commands buffers structural changes (like adding/removing components
	// or entities) that occur during a system update. These changes are applied
	// atomically during a synchronization point.
	Commands = core.SystemCommandBuffer

	// System is the interface for logic units that process entity data.
	//
	// Update is called by the scheduler and receives a ReadOnlyRegistry to inspect
	// data and a SystemCommandBuffer to request changes.
	//
	// Init is called once when the system is registered, providing access
	// to the Engine instance for setup (e.g., pre-registering components or queries).
	System interface {
		Update(reg Lookup, cb *Commands, d time.Duration)
		Init(*Engine)
	}
)

func AssignComponent[T any](cb *Commands, e Entity, info core.ComponentInfo, value T) {
	core.AssignComponent(cb, e, info, value)
}

// SystemFunc defines a function signature for stateless logic units.
// It allows for quick system implementation without defining a dedicated struct.
type SystemFunc func(cb *Commands, d time.Duration)

type functionalSystem struct {
	updateFn SystemFunc
}

func (f *functionalSystem) Init(reg *Engine) {}

func (f *functionalSystem) Update(reg Lookup, cb *Commands, d time.Duration) {
	f.updateFn(cb, d)
}

var _ System = (*functionalSystem)(nil)
