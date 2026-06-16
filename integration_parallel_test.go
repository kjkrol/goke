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
	query *goke.View
	pos   goke.Col[Position]
	vel   goke.Col[Velocity]
}

func (s *PhysicsSystem) Init(ecs *goke.ECS) {
	s.query = goke.NewView(ecs, s.pos.Track(), s.vel.Track())
}
func (s *PhysicsSystem) Update(lookup goke.Lookup, schedule *goke.CmdBuf, d time.Duration) {
	s.query.All()
	for s.query.Next() {
		pos := s.pos.Slice(s.query)
		vel := s.vel.Slice(s.query)
		for i := range s.query.EntSlice {
			pos[i].X += vel[i].VX * float32(d.Seconds())
			pos[i].Y += vel[i].VY * float32(d.Seconds())
		}
	}
}

// --- HealthSystem: Operates only on Health data ---
type HealthSystem struct {
	query  *goke.View
	health goke.Col[Health]
}

func (s *HealthSystem) Init(eng *goke.ECS) {
	s.query = goke.NewView(eng, s.health.Track())
}
func (s *HealthSystem) Update(lookup goke.Lookup, schedule *goke.CmdBuf, d time.Duration) {
	s.query.All()
	for s.query.Next() {
		health := s.health.Slice(s.query)
		for i := range s.query.EntSlice {
			if health[i].Current < health[i].Max {
				health[i].Current += 1.0
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
	blueprint := goke.NewFactory3[Position, Velocity, Health](ecs)
	for chunk := range blueprint.Create(1000) {
		for i := range chunk.Entity {
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
	var vpos goke.Col[Position]
	var vhp goke.Col[Health]
	query := goke.NewView(ecs, vpos.Track(), vhp.Track())
	count := 0
	query.All()
	for query.Next() {
		pos := vpos.Slice(query)
		hp := vhp.Slice(query)
		for i := range query.EntSlice {
			count++
			if pos[i].X != 10 {
				t.Errorf("Physics failed: expected X=10, got %f", pos[i].X)
			}
			if hp[i].Current != 51 {
				t.Errorf("Health failed: expected HP=51, got %f", hp[i].Current)
			}
		}
	}

	if count != 1000 {
		t.Errorf("Expected 1000 entities, found %d", count)
	}
}
