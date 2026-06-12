package goke

import (
	"time"

	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/internal/exec"
)

type (
	// Lookup provides a read-only view of component data.
	// Used by systems to inspect entity state without the risk of concurrent modification.
	Lookup = core.ComponentReader

	// Schedule collects structural changes (add/remove component, destroy entity)
	// and applies them at the next synchronization point, keeping iterators valid
	// during system updates.
	Schedule = exec.CommandBuf

	// System is the interface for logic units that process entity data each tick.
	//
	// Update receives a read-only Lookup for inspecting state and a Schedule
	// for queuing deferred changes.
	//
	// Init is called once on registration, giving the system access to the ECS
	// for setup such as pre-registering components or views.
	System interface {
		Update(Lookup, *Schedule, time.Duration)
		Init(*ECS)
	}
)

// ScheduleAddComponent queues the addition of a component to an entity.
// The component is initialized with the provided value. If the entity
// already has this component, the existing data will be overwritten
// when the schedule is applied.
func ScheduleAddComponent[T any](schedule *Schedule, e Entity, compDesc ComponentDesc, value T) {
	exec.AddComponent(schedule, e, compDesc, value)
}

// ScheduleRemoveComponent queues the removal of a component from an entity.
// This operation is ignored if the entity does not have the specified component.
func ScheduleRemoveComponent(schedule *Schedule, e Entity, compDesc ComponentDesc) {
	exec.RemoveComponent(schedule, e, compDesc)
}

// ScheduleRemoveEntity queues the destruction of an entity and all its
// associated components. Any pending operations on this entity in the
// same schedule will be discarded.
func ScheduleRemoveEntity(schedule *Schedule, e Entity) {
	exec.RemoveEntity(schedule, e)
}

// SystemFunc defines a function signature for stateless logic units.
// It allows for quick system implementation without defining a dedicated struct.
type SystemFunc func(*Schedule, time.Duration)

// ------------- helper struct -------------

type functionalSystem struct {
	updateFn SystemFunc
}

func (f *functionalSystem) Init(*ECS) {}

func (f *functionalSystem) Update(reg Lookup, cb *Schedule, d time.Duration) {
	f.updateFn(cb, d)
}

var _ System = (*functionalSystem)(nil)
