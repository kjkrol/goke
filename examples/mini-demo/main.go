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
	_ = goke.RegComp[Pos](ecs)
	_ = goke.RegComp[Vel](ecs)
	_ = goke.RegComp[Acc](ecs)

	var pos goke.Comp[Pos]
	var vel goke.Comp[Vel]
	var acc goke.Comp[Acc]
	factory := ecs.NewFactory(&pos, &vel, &acc)

	var entityID uid.UID64
	factory.Create(1)
	factory.Next()
	entityID = factory.IDs[0]
	fc := &factory.Cursor
	pos.Slice(fc)[0] = Pos{X: 0, Y: 0}
	vel.Slice(fc)[0] = Vel{X: 1, Y: 1}
	acc.Slice(fc)[0] = Acc{X: 0.1, Y: 0.1}

	// Initialize matcher for Pos, Vel, and Acc components
	query := ecs.NewQueryBuilder(&pos, &vel, &acc).Build()

	// Define the movement system using the functional registration pattern
	cursor := &query.Cursor
	movementSystem := ecs.RegSysFn(func(schedule *goke.CmdBuf, d time.Duration) {
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

	matcher := ecs.NewQueryBuilder(&pos).Build()
	if matcher.Seek(entityID) {
		p := pos.At(&matcher.Cursor)
		fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
	}
}
