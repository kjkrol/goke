package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
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
	posDesc := goke.RegCompType[Pos](ecs)
	_ = goke.RegCompType[Vel](ecs)
	_ = goke.RegCompType[Acc](ecs)

	factory := goke.NewFactory3[Pos, Vel, Acc](ecs)

	var entityID uid.UID64
	for chunk := range factory.Create(1) {
		entityID = chunk.Entity[0]
		chunk.Comp1[0] = Pos{X: 0, Y: 0}
		chunk.Comp2[0] = Vel{X: 1, Y: 1}
		chunk.Comp3[0] = Acc{X: 0.1, Y: 0.1}
	}

	// Initialize view for Pos, Vel, and Acc components
	var pos goke.Col[Pos]
	var vel goke.Col[Vel]
	var acc goke.Col[Acc]
	query := goke.NewView(ecs, pos.Track(), vel.Track(), acc.Track())

	// Define the movement system using the functional registration pattern
	movementSystem := goke.RegSysFn(ecs, func(schedule *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			posSlice := pos.Slice(query)
			velSlice := vel.Slice(query)
			accSlice := acc.Slice(query)
			for i := range query.EntSlice {
				velSlice[i].X += accSlice[i].X
				velSlice[i].Y += accSlice[i].Y
				posSlice[i].X += velSlice[i].X
				posSlice[i].Y += velSlice[i].Y
			}
		}
	})

	// Configure the ECS's execution workflow and synchronization points
	goke.SetPlan(ecs, func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync() // Ensure all component updates are flushed and views are consistent
	})

	// Execute a single simulation step (standard 120 TPS)
	goke.Tick(ecs, time.Second/120)

	p, _ := goke.SafeGetComp[Pos](ecs, entityID, posDesc)
	fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
}
