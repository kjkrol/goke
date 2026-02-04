package goke

import (
	"time"

	"github.com/kjkrol/goke/internal/core"
)

type (
	// Lookup provides a thread-safe, read-only view of the components.
	// It is used by systems to inspect the state of entities without the risk
	// of concurrent modification.
	Lookup = core.ReadOnlyRegistry

	// Schedule manages structural changes to the world that are deferred until
	// a synchronization point. It allows for safe modification of entities and
	// components during system updates without invalidating current iterators.
	Schedule = core.SystemCommandBuffer

	// System is the interface for logic units that process entity data.
	//
	// Update is called by the scheduler and receives a ReadOnlyRegistry to inspect
	// data and a SystemCommandBuffer to request changes.
	//
	// Init is called once when the system is registered, providing access
	// to the ECS instance for setup (e.g., pre-registering components or views).
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
	core.AddComponent(schedule, e, compDesc, value)
}

// ScheduleRemoveComponent queues the removal of a component from an entity.
// This operation is ignored if the entity does not have the specified component.
func ScheduleRemoveComponent(schedule *Schedule, e Entity, compDesc ComponentDesc) {
	core.RemoveComponent(schedule, e, compDesc)
}

// ScheduleCreateEntity queues the creation of a new entity.
// It returns a "virtual" Entity handle that can be used within the same
// schedule to add components to the newly created entity before it is
// officially spawned in the world.
func ScheduleCreateEntity(schedule *Schedule) Entity {
	return core.CreateEntity(schedule)
}

// ScheduleRemoveEntity queues the destruction of an entity and all its
// associated components. Any pending operations on this entity in the
// same schedule will be discarded.
func ScheduleRemoveEntity(schedule *Schedule, e Entity) {
	core.RemoveEntity(schedule, e)
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
