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

	"github.com/kjkrol/goke/v2"
)

// --- Components (Disjoint sets) ---
type Health struct{ Current, Max float32 }

// --- PhysicsSystem: Operates only on Motion data ---
type PhysicsSystem struct {
	query *goke.Query
	pos   goke.Comp[Position]
	vel   goke.Comp[Velocity]
}

func (s *PhysicsSystem) Init(ecs *goke.ECS) {
	s.query = ecs.NewQueryBuilder(&s.pos, &s.vel).Build()
}
func (s *PhysicsSystem) Update(schedule *goke.CmdBuf, d time.Duration) {
	cursor := s.query.Cursor()
	s.query.All()
	for s.query.Next() {
		pos := s.pos.Slice(cursor)
		vel := s.vel.Slice(cursor)
		for i := range cursor.IDs {
			pos[i].X += vel[i].VX * float32(d.Seconds())
			pos[i].Y += vel[i].VY * float32(d.Seconds())
		}
	}
}

// --- HealthSystem: Operates only on Health data ---
type HealthSystem struct {
	query  *goke.Query
	health goke.Comp[Health]
}

func (s *HealthSystem) Init(eng *goke.ECS) {
	s.query = eng.NewQueryBuilder(&s.health).Build()
}
func (s *HealthSystem) Update(schedule *goke.CmdBuf, d time.Duration) {
	s.query.All()
	for s.query.Next() {
		cursor := s.query.Cursor()
		health := s.health.Slice(cursor)
		for i := range cursor.IDs {
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
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)
	_ = goke.RegComp[Health](ecs)

	phys := &PhysicsSystem{}
	heal := &HealthSystem{}
	ecs.RegSys(phys)
	ecs.RegSys(heal)

	// Create entities with ALL components
	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var health goke.Comp[Health]
	factory := ecs.NewFactory(&pos, &vel, &health)
	fc := &factory.Cursor
	factory.Create(1000)
	for factory.Next() {
		positions := pos.Slice(fc)
		velocities := vel.Slice(fc)
		healths := health.Slice(fc)
		for i := range factory.IDs {
			positions[i] = Position{0, 0}
			velocities[i] = Velocity{10, 10}
			healths[i] = Health{50, 100}
		}
	}

	// 2. Execution Plan: Run Physics and Health in parallel
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.RunParallel(d, phys, heal)
		ctx.Sync()
	})

	// 3. Tick ecs
	ecs.Tick(time.Second) // Simulate 1 second

	// 4. Verification
	query := ecs.NewQueryBuilder(&pos, &health).Build()
	cursor := query.Cursor()
	count := 0
	query.All()
	for query.Next() {
		positions := pos.Slice(cursor)
		healths := health.Slice(cursor)
		for i := range cursor.IDs {
			count++
			if positions[i].X != 10 {
				t.Errorf("Physics failed: expected X=10, got %f", positions[i].X)
			}
			if healths[i].Current != 51 {
				t.Errorf("Health failed: expected HP=51, got %f", healths[i].Current)
			}
		}
	}

	if count != 1000 {
		t.Errorf("Expected 1000 entities, found %d", count)
	}
}
