package ecs_test

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/pkg/ecs"
)

// --- Components ---
type Initial struct{}
type ComponentA struct{ Val int }
type ComponentB struct{ Val int }

var (
	initialInfo ecs.ComponentInfo
	compAInfo   ecs.ComponentInfo
	compBInfo   ecs.ComponentInfo
)

// --- Systems ---

type ProducerSystem struct{}

func (s *ProducerSystem) Init(reg *ecs.Registry) {}
func (s *ProducerSystem) Update(reg ecs.ReadOnlyRegistry, cb *ecs.SystemCommandBuffer, d time.Duration) {
	// Create a new entity and mark it with Initial tag
	e := cb.CreateEntity()
	cb.AssignComponent(e, initialInfo, nil)
}

type WorkerASystem struct {
	query *ecs.Query1[Initial]
}

func (s *WorkerASystem) Init(reg *ecs.Registry) { s.query = ecs.NewQuery1[Initial](reg) }
func (s *WorkerASystem) Update(reg ecs.ReadOnlyRegistry, cb *ecs.SystemCommandBuffer, d time.Duration) {
	for head := range s.query.All1() {
		val := ComponentA{Val: 100}
		cb.AssignComponent(head.Entity, compAInfo, unsafe.Pointer(&val))
	}
}

type WorkerBSystem struct {
	query *ecs.Query1[Initial]
}

func (s *WorkerBSystem) Init(reg *ecs.Registry) { s.query = ecs.NewQuery1[Initial](reg) }
func (s *WorkerBSystem) Update(reg ecs.ReadOnlyRegistry, cb *ecs.SystemCommandBuffer, d time.Duration) {
	for head := range s.query.All1() {
		val := ComponentB{Val: 200}
		cb.AssignComponent(head.Entity, compBInfo, unsafe.Pointer(&val))
	}
}

// --- TEST ---

func TestECS_ParallelExecution(t *testing.T) {
	eng := ecs.NewEngine()

	// Setup components
	initialInfo = eng.RegisterComponentType(reflect.TypeFor[Initial]())
	compAInfo = eng.RegisterComponentType(reflect.TypeFor[ComponentA]())
	compBInfo = eng.RegisterComponentType(reflect.TypeFor[ComponentB]())

	// Instantiate systems
	producer := &ProducerSystem{}
	workerA := &WorkerASystem{}
	workerB := &WorkerBSystem{}

	// Register systems to allocate buffers
	eng.RegisterSystem(producer)
	eng.RegisterSystem(workerA)
	eng.RegisterSystem(workerB)

	// Define the execution plan: 1 -> (2 & 3)
	eng.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		// First: Producer creates entities
		ctx.RunSystem(producer, d)
		ctx.Sync()

		// Second: Workers modify entities in parallel
		ctx.RunParallel(d, workerA, workerB)
		ctx.Sync()
	})

	// Run update
	eng.UpdateSystems(time.Millisecond * 16)

	// Verification
	// We should have at least one entity with both ComponentA and ComponentB
	query := ecs.NewQuery2[ComponentA, ComponentB](eng.Registry)
	found := false
	for range query.All2() {
		found = true
	}

	if !found {
		t.Error("Parallel execution failed: Entity should have both ComponentA and ComponentB")
	}
}
