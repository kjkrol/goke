/*
This test suite validates the "Parallel Execution" capabilities of the engine.

It specifically focuses on the "Disjoint Component Set" rule:
Multiple systems can safely execute in parallel on the same entities, provided
they operate on non-overlapping sets of components.

The test verifies that:
 1. The Scheduler correctly orchestrates concurrent execution using RunParallel.
 2. Data integrity is maintained across different components of the same entity
    when accessed by multiple threads simultaneously.
 3. Post-parallel synchronization (Sync) correctly stabilizes the state for
    subsequent read operations.
*/
package ecs_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/kjkrol/goke/pkg/ecs"
)

// --- Components (Disjoint sets) ---
type Position struct{ X, Y float32 }
type Velocity struct{ VX, VY float32 }
type Health struct{ Current, Max float32 }

// --- PhysicsSystem: Operates only on Motion data ---
type PhysicsSystem struct {
	query *ecs.View2[Position, Velocity]
}

func (s *PhysicsSystem) Init(eng *ecs.Engine) {
	s.query = ecs.NewView2[Position, Velocity](eng)
}
func (s *PhysicsSystem) Update(reg ecs.Lookup, cb *ecs.Commands, d time.Duration) {
	for head := range s.query.All() {
		head.V1.X += head.V2.VX * float32(d.Seconds())
		head.V1.Y += head.V2.VY * float32(d.Seconds())
	}
}

// --- HealthSystem: Operates only on Health data ---
type HealthSystem struct {
	query *ecs.View1[Health]
}

func (s *HealthSystem) Init(eng *ecs.Engine) {
	s.query = ecs.NewView1[Health](eng)
}
func (s *HealthSystem) Update(reg ecs.Lookup, cb *ecs.Commands, d time.Duration) {
	for head := range s.query.All() {
		if head.V1.Current < head.V1.Max {
			head.V1.Current += 1.0 // Simple regen
		}
	}
}

// TestECS_ParallelExecution_Disjoint simulates a high-load scenario where physics
// calculations and health regeneration occur simultaneously.
// It ensures that even if an entity possesses both sets of components,
// the engine can process them in parallel without race conditions because
// the systems access separate memory regions (columns).
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
		pos, _ := ecs.AllocateComponentByInfo[Position](eng, e, posInfo)
		*pos = Position{0, 0}
		vel, _ := ecs.AllocateComponentByInfo[Velocity](eng, e, velInfo)
		*vel = Velocity{10, 10}
		hel, _ := ecs.AllocateComponentByInfo[Health](eng, e, healthInfo)
		*hel = Health{50, 100}
	}

	// 2. Execution Plan: Run Physics and Health in parallel
	eng.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		ctx.RunParallel(d, phys, heal)
		ctx.Sync()
	})

	// 3. Tick engine
	eng.Tick(time.Second) // Simulate 1 second

	// 4. Verification
	query := ecs.NewView2[Position, Health](eng)
	count := 0
	for head := range query.All() {
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
