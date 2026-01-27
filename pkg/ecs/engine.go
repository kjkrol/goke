package ecs

import (
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

type (
	EngineAPI interface {
		CreateEntity() Entity
		RemoveEntity(entity Entity) bool

		RegisterComponentType(reflect.Type) ComponentInfo
		AssignByID(Entity, ComponentInfo, unsafe.Pointer) error
		AllocateByID(Entity, ComponentInfo) (unsafe.Pointer, error)
		UnassignByID(Entity, ComponentInfo) error
		GetComponent(Entity, ComponentID) (unsafe.Pointer, error)

		RegisterSystem(System)
		RegisterSystemFunc(SystemFunc) System
		SetExecutionPlan(ExecutionPlan)
		Run(time.Duration)
	}
	Engine struct {
		*Registry
		scheduler *SystemScheduler
	}
)

var _ EngineAPI = (*Engine)(nil)

func NewEngine() *Engine {
	reg := NewRegistry()
	return &Engine{
		Registry:  reg,
		scheduler: NewScheduler(reg),
	}
}

func (e *Engine) RegisterSystem(system System) {
	system.Init(e)
	e.scheduler.RegisterSystem(system)
}

func (e *Engine) RegisterSystemFunc(fn SystemFunc) System {
	wrapper := &functionalSystem{updateFn: fn}
	wrapper.Init(e)
	e.scheduler.RegisterSystem(wrapper)
	return wrapper
}

func (e *Engine) SetExecutionPlan(plan ExecutionPlan) {
	e.scheduler.SetExecutionPlan(plan)
}

func (e *Engine) Run(duration time.Duration) {
	e.scheduler.Tick(duration)
}

func RegisterComponent[T any](eng *Engine) ComponentInfo {
	return ensureComponentRegistered[T](eng.componentsRegistry)
}

func Assign[T any](eng *Engine, entity Entity, component T) error {
	compInfo := ensureComponentRegistered[T](eng.componentsRegistry)
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compInfo, data)
}

func AssignByID[T any](eng *Engine, entity Entity, compInfo ComponentInfo, component T) error {
	data := unsafe.Pointer(&component)
	return eng.AssignByID(entity, compInfo, data)
}

func Unassign[T any](eng *Engine, entity Entity) error {
	componentType := reflect.TypeFor[T]()
	compInfo, ok := eng.componentsRegistry.Get(componentType)
	if !ok {
		return fmt.Errorf("Component doesn't exist.")
	}

	return eng.UnassignByID(entity, compInfo)
}

func AddComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compInfo := ensureComponentRegistered[T](eng.componentsRegistry)
	ptr, err := eng.AllocateByID(entity, compInfo)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

func GetComponent[T any](eng *Engine, entity Entity) (*T, error) {
	compType := reflect.TypeFor[T]()
	compInfo, ok := eng.componentsRegistry.Get(compType)
	if !ok {
		return nil, fmt.Errorf("Component doesn't exist.")
	}
	data, err := eng.GetComponent(entity, compInfo.ID)
	if err != nil {
		return nil, err
	}

	return (*T)(data), nil
}

func NewQuery0(eng *Engine, options ...ViewOption) *Query0 {
	return newQuery0(eng.Registry, options...)
}
func NewQuery1[T1 any](eng *Engine, options ...ViewOption) *Query1[T1] {
	return newQuery1[T1](eng.Registry, options...)
}
func NewQuery2[T1, T2 any](eng *Engine, options ...ViewOption) *Query2[T1, T2] {
	return newQuery2[T1, T2](eng.Registry, options...)
}
func NewQuery3[T1, T2, T3 any](eng *Engine, options ...ViewOption) *Query3[T1, T2, T3] {
	return newQuery3[T1, T2, T3](eng.Registry, options...)
}
func NewQuery4[T1, T2, T3, T4 any](eng *Engine, options ...ViewOption) *Query4[T1, T2, T3, T4] {
	return newQuery4[T1, T2, T3, T4](eng.Registry, options...)
}
func NewQuery5[T1, T2, T3, T4, T5 any](eng *Engine, options ...ViewOption) *Query5[T1, T2, T3, T4, T5] {
	return newQuery5[T1, T2, T3, T4, T5](eng.Registry, options...)
}
func NewQuery6[T1, T2, T3, T4, T5, T6 any](eng *Engine, options ...ViewOption) *Query6[T1, T2, T3, T4, T5, T6] {
	return newQuery6[T1, T2, T3, T4, T5, T6](eng.Registry, options...)
}
func NewQuery7[T1, T2, T3, T4, T5, T6, T7 any](eng *Engine, options ...ViewOption) *Query7[T1, T2, T3, T4, T5, T6, T7] {
	return newQuery7[T1, T2, T3, T4, T5, T6, T7](eng.Registry, options...)
}
func NewQuery8[T1, T2, T3, T4, T5, T6, T7, T8 any](eng *Engine, options ...ViewOption) *Query8[T1, T2, T3, T4, T5, T6, T7, T8] {
	return newQuery8[T1, T2, T3, T4, T5, T6, T7, T8](eng.Registry, options...)
}
