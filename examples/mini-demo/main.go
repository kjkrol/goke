package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/v2"
)

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }

func main() {
	// Initialize the ECS world.
	// The ECS instance acts as the central coordinator for entities and systems.
	ecs := goke.New()

	var pos goke.Comp[Pos]
	var vel goke.Comp[Vel]
	var acc goke.Comp[Acc]

	factory := ecs.NewFactory(&pos, &vel, &acc)
	cursor := &factory.Cursor

	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]
	pos.Slice(cursor)[0] = Pos{X: 0, Y: 0}
	vel.Slice(cursor)[0] = Vel{X: 1, Y: 1}
	acc.Slice(cursor)[0] = Acc{X: 0.1, Y: 0.1}

	// Initialize matcher for Pos, Vel, and Acc components
	query := ecs.NewQueryBuilder(&pos, &vel, &acc).Build()

	// Define the movement system using the functional registration pattern
	cursor = &query.Cursor
	movementSystem := ecs.RegSysFn(func(_ *goke.CmdBuf, _ time.Duration) {
		query.All()
		for query.Next() {
			posSlice := pos.Slice(cursor)
			velSlice := vel.Slice(cursor)
			accSlice := acc.Slice(cursor)
			for i := range query.Cursor.IDs {
				velSlice[i].X += accSlice[i].X
				velSlice[i].Y += accSlice[i].Y
				posSlice[i].X += velSlice[i].X
				posSlice[i].Y += velSlice[i].Y
			}
		}
	})

	// Configure the ECS's execution workflow and synchronization points
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync() // Ensure all component updates are flushed and matchers are consistent
	})

	// Execute a single simulation step (standard 120 TPS)
	ecs.Tick(time.Second / 120)

	lookup := ecs.NewQueryBuilder(&pos).Build()
	if lookup.Seek(entityID) {
		p := pos.At(&lookup.Cursor)
		fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
	}
}
