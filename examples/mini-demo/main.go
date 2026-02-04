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
	velDesc := goke.RegisterComponent[Vel](ecs)
	accDesc := goke.RegisterComponent[Acc](ecs)

	entity := goke.CreateEntity(ecs)

	// --- Component Assignment with In-Place Memory Access ---
	// EnsureComponent acts as a high-performance Upsert. By returning a direct
	// pointer to the component's slot in the archetype storage, it allows for
	// in-place modification (*ptr = T{...}).
	*goke.EnsureComponent[Pos](ecs, entity, posDesc) = Pos{X: 0, Y: 0}
	*goke.EnsureComponent[Vel](ecs, entity, velDesc) = Vel{X: 1, Y: 1}
	*goke.EnsureComponent[Acc](ecs, entity, accDesc) = Acc{X: 0.1, Y: 0.1}

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

	p, _ := goke.GetComponent[Pos](ecs, entity, posDesc)
	fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
}
