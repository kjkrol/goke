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
	engine := goke.NewEngine()
	posType := goke.ComponentRegister[Pos](engine)
	velType := goke.ComponentRegister[Vel](engine)
	accType := goke.ComponentRegister[Acc](engine)

	entity := goke.EntityCreate(engine)

	// --- Direct Access (Optimized/Fastest) ---
	// EntityEnsureComponent returns a direct pointer to the storage location.
	// Assigning via pointer dereference avoids extra copies and redundant lookups.
	if ptr, _ := goke.EntityEnsureComponent[Pos](engine, entity, posType); ptr != nil {
		*ptr = Pos{X: 0, Y: 0}
	}
	if ptr, _ := goke.EntityEnsureComponent[Vel](engine, entity, velType); ptr != nil {
		*ptr = Vel{X: 1, Y: 1}
	}
	if ptr, _ := goke.EntityEnsureComponent[Acc](engine, entity, accType); ptr != nil {
		*ptr = Acc{X: 0.1, Y: 0.1}
	}

	// Initialize view for Pos, Vel, and Acc components
	view := goke.NewView3[Pos, Vel, Acc](engine)

	// Define the movement system using the functional registration pattern
	movementSystem := goke.SystemFuncRegister(engine, func(cb *goke.Schedule, d time.Duration) {
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
	engine.SetExecutionPlan(func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync() // Ensure all component updates are flushed and views are consistent
	})

	// Execute a single simulation step (standard 60 FPS tick)
	engine.Tick(time.Millisecond * 16)

	p, _ := goke.EntityGetComponent[Pos](engine, entity, posType)
	fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
}
