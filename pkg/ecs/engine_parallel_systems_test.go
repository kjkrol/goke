package ecs_test

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/pkg/ecs"
)

// --- Components (Disjoint sets) ---
type Position struct{ X, Y float32 }
type Velocity struct{ VX, VY float32 }
type Health struct{ Current, Max float32 }

// --- PhysicsSystem: Operates only on Motion data ---
type PhysicsSystem struct {
	query *ecs.Query2[Position, Velocity]
}

func (s *PhysicsSystem) Init(reg *ecs.Registry) {
	s.query = ecs.NewQuery2[Position, Velocity](reg)
}
func (s *PhysicsSystem) Update(reg ecs.ReadOnlyRegistry, cb *ecs.SystemCommandBuffer, d time.Duration) {
	for head := range s.query.All2() {
		head.V1.X += head.V2.VX * float32(d.Seconds())
		head.V1.Y += head.V2.VY * float32(d.Seconds())
	}
}

// --- HealthSystem: Operates only on Health data ---
type HealthSystem struct {
	query *ecs.Query1[Health]
}

func (s *HealthSystem) Init(reg *ecs.Registry) {
	s.query = ecs.NewQuery1[Health](reg)
}
func (s *HealthSystem) Update(reg ecs.ReadOnlyRegistry, cb *ecs.SystemCommandBuffer, d time.Duration) {
	for head := range s.query.All1() {
		if head.V1.Current < head.V1.Max {
			head.V1.Current += 1.0 // Simple regen
		}
	}
}

func TestECS_ParallelExecution_Disjoint(t *testing.T) {
	eng := ecs.NewEngine()

	// 1. Setup
	posInfo := eng.RegisterComponentType(reflect.TypeFor[Position]())
	velInfo := eng.RegisterComponentType(reflect.TypeFor[Velocity]())
	healthInfo := eng.RegisterComponentType(reflect.TypeFor[Health]())

	phys := &PhysicsSystem{}
	heal := &HealthSystem{}
	eng.RegisterSystem(phys)
	eng.RegisterSystem(heal)

	// Create entities with ALL components
	for range 1000 {
		e := eng.CreateEntity()
		eng.AssignByID(e, posInfo.ID, unsafe.Pointer(&Position{0, 0}))
		eng.AssignByID(e, velInfo.ID, unsafe.Pointer(&Velocity{10, 10}))
		eng.AssignByID(e, healthInfo.ID, unsafe.Pointer(&Health{50, 100}))
	}

	// 2. Execution Plan: Run Physics and Health in parallel
	eng.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		ctx.RunParallel(d, phys, heal)
		ctx.Sync()
	})

	// 3. Run
	eng.UpdateSystems(time.Second) // Simulate 1 second

	// 4. Verification
	query := ecs.NewQuery2[Position, Health](eng.Registry)
	count := 0
	for head := range query.All2() {
		count++
		// Check Physics result: 0 + 10*1s = 10
		if head.V1.X != 10 {
			t.Errorf("Physics failed: expected X=10, got %f", head.V1.X)
		}
		// Check Health result: 50 + 1 = 51
		if head.V2.Current != 51 {
			t.Errorf("Health failed: expected HP=51, got %f", head.V2.Current)
		}
	}

	if count != 1000 {
		t.Errorf("Expected 1000 entities, found %d", count)
	}
}
