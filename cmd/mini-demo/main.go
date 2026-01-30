package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/pkg/ecs"
)

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }

func main() {
	engine := ecs.NewEngine()
	entity := engine.CreateEntity()

	// --- Standard Assign (Reflective/Slower) ---
	// SetComponent involves an internal lookup and potential interface wrapping,
	// which can increase overhead when called frequently for many entities.
	ecs.SetComponent(engine, entity, Pos{X: 0, Y: 0})
	ecs.SetComponent(engine, entity, Vel{X: 1, Y: 1})

	// --- Direct Access (Optimized/Fastest) ---
	// AllocateComponent returns a direct pointer to the storage location.
	// Assigning via pointer dereference avoids extra copies and redundant lookups.
	ptr, _ := ecs.AllocateComponent[Acc](engine, entity)
	*ptr = Acc{X: 0.1, Y: 0.1}

	// Initialize view for Pos, Vel, and Acc components
	view := ecs.NewView3[Pos, Vel, Acc](engine)

	// Define the movement system using the functional registration pattern
	movementSystem := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		// High-performance iteration utilizing Data Locality.
		// Component data is processed in contiguous memory blocks (SoA layout).
		for head := range view.Values() {
			pos, vel, acc := head.V1, head.V2, head.V3

			vel.X += acc.X
			vel.Y += acc.Y
			pos.X += vel.X
			pos.Y += vel.Y
		}
	})

	// Configure the engine's execution workflow and synchronization points
	engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync() // Ensure all component updates are flushed and views are consistent
	})

	// Execute a single simulation step (standard 60 FPS tick)
	engine.Tick(time.Millisecond * 16)

	p, _ := ecs.GetComponent[Pos](engine, entity)
	fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
}
