package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke"
)

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }

func main() {
	// Initialize the ECS world.
	// The ECS instance acts as the central coordinator for entities and systems.
	ecs := goke.New()

	// Define component metadata.
	// This binds Go types to internal descriptors, allowing the engine to
	// pre-calculate memory layouts and manage data in contiguous arrays.
	posDesc := goke.RegisterComponent[Pos](ecs)
	_ = goke.RegisterComponent[Vel](ecs)
	_ = goke.RegisterComponent[Acc](ecs)

	// --- Type-Safe Entity Template (Blueprint) ---
	// Blueprints place the entity into the correct archetype immediately and
	// reserve memory for all components in a single atomic operation.
	// This returns typed pointers for direct, in-place initialization.
	blueprint := goke.NewBlueprint3[Pos, Vel, Acc](ecs)

	// Create the entity and get direct access to its memory slots.
	entity, pos, vel, acc := blueprint.Create()
	*pos = Pos{X: 0, Y: 0}
	*vel = Vel{X: 1, Y: 1}
	*acc = Acc{X: 0.1, Y: 0.1}

	// Initialize view for Pos, Vel, and Acc components
	view := goke.NewView3[Pos, Vel, Acc](ecs)

	// Define the movement system using the functional registration pattern
	movementSystem := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
		// SoA (Structure of Arrays) layout ensures CPU Cache friendliness.
		for head := range view.Values() {
			pos, vel, acc := head.V1, head.V2, head.V3

			vel.X += acc.X
			vel.Y += acc.Y
			pos.X += vel.X
			pos.Y += vel.Y
		}
	})

	// Configure the ECS's execution workflow and synchronization points
	goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync() // Ensure all component updates are flushed and views are consistent
	})

	// Execute a single simulation step (standard 120 TPS)
	goke.Tick(ecs, time.Second/120)

	p, _ := goke.SafeGetComponent[Pos](ecs, entity, posDesc)
	fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
}
