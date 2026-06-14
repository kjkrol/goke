/*
This test suite validates the "Parallel Execution" capabilities of the engine.

It specifically focuses on the "Disjoint Component Set" rule:
Multiple systems can safely execute in parallel on the same entities, provided
they operate on non-overlapping sets of components.

The test verifies that:
 1. The Scheduler correctly orchestrates concurrent execution using RunParallel.
 2. Data integrity is maintained across different components of the same entityID
    when accessed by multiple threads simultaneously.
 3. Post-parallel synchronization (Sync) correctly stabilizes the state for
    subsequent read operations.
*/
package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke"
)

// --- Components (Disjoint sets) ---
type Position struct{ X, Y float32 }
type Velocity struct{ VX, VY float32 }
type Health struct{ Current, Max float32 }

// --- PhysicsSystem: Operates only on Motion data ---
type PhysicsSystem struct {
	query *goke.View2[Position, Velocity]
}

func (s *PhysicsSystem) Init(ecs *goke.ECS) {
	s.query = goke.NewView2[Position, Velocity](ecs)
}
func (s *PhysicsSystem) Update(lookup goke.Lookup, schedule *goke.CmdBuf, d time.Duration) {
	for chunk := range s.query.All() {
		for i, _ := range chunk.Entity {
			v1, v2 := &chunk.Comp1[i], &chunk.Comp2[i]
			v1.X += v2.VX * float32(d.Seconds())
			v1.Y += v2.VY * float32(d.Seconds())
		}
	}
}

// --- HealthSystem: Operates only on Health data ---
type HealthSystem struct {
	query *goke.View1[Health]
}

func (s *HealthSystem) Init(eng *goke.ECS) {
	s.query = goke.NewView1[Health](eng)
}
func (s *HealthSystem) Update(lookup goke.Lookup, schedule *goke.CmdBuf, d time.Duration) {
	for chunk := range s.query.All() {
		for i, _ := range chunk.Entity {
			health := &chunk.Comp1[i]
			if health.Current < health.Max {
				health.Current += 1.0
			}
		}
	}
}

// TestECS_ParallelExecution_Disjoint simulates a high-load scenario where physics
// calculations and health regeneration occur simultaneously.
// It ensures that even if an entity possesses both sets of components,
// the engine can process them in parallel without race conditions because
// the systems access separate memory regions (columns).
func TestECS_ParallelExecution_Disjoint(t *testing.T) {
	ecs := goke.New()

	// 1. Setup
	_ = goke.RegCompType[Position](ecs)
	_ = goke.RegCompType[Velocity](ecs)
	_ = goke.RegCompType[Health](ecs)

	phys := &PhysicsSystem{}
	heal := &HealthSystem{}
	goke.RegSys(ecs, phys)
	goke.RegSys(ecs, heal)

	// Create entities with ALL components
	blueprint := goke.NewBlueprint3[Position, Velocity, Health](ecs)
	for chunk := range blueprint.Create(1000) {
		for i, _ := range chunk.Entity {
			chunk.Comp1[i] = Position{0, 0}
			chunk.Comp2[i] = Velocity{10, 10}
			chunk.Comp3[i] = Health{50, 100}
		}
	}

	// 2. Execution Plan: Run Physics and Health in parallel
	goke.SetPlan(ecs, func(ctx goke.RunCtx, d time.Duration) {
		ctx.RunParallel(d, phys, heal)
		ctx.Sync()
	})

	// 3. Tick ecs
	goke.Tick(ecs, time.Second) // Simulate 1 second

	// 4. Verification
	query := goke.NewView2[Position, Health](ecs)
	count := 0
	for chunk := range query.All() {
		for i, _ := range chunk.Entity {
			v1, v2 := &chunk.Comp1[i], &chunk.Comp2[i]
			count++
			// Check Physics result: 0 + 10*1s = 10
			if v1.X != 10 {
				t.Errorf("Physics failed: expected X=10, got %f", v1.X)
			}
			// Check Health result: 50 + 1 = 51
			if v2.Current != 51 {
				t.Errorf("Health failed: expected HP=51, got %f", v2.Current)
			}
		}
	}

	if count != 1000 {
		t.Errorf("Expected 1000 entities, found %d", count)
	}
}
