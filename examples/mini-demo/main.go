package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
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
	var entityID uid.UID64

	// One-time world seeding: Query/Editor/Factory can only be constructed
	// inside a system's Init, so even top-level setup runs as a system.
	ecs.Setup(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			factory := si.NewFactory(&pos, &vel, &acc)
			cursor := &factory.Cursor

			factory.Create(1)
			factory.Next()
			entityID = factory.IDs[0]
			pos.Slice(cursor)[0] = Pos{X: 0, Y: 0}
			vel.Slice(cursor)[0] = Vel{X: 1, Y: 1}
			acc.Slice(cursor)[0] = Acc{X: 0.1, Y: 0.1}
		},
	})

	// Define the movement system: builds its own Query in Init, iterates in Update.
	var query *goke.Query
	movementSystem := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			query = si.NewQueryBuilder(&pos, &vel, &acc).Build()
		},
		OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) {
			cursor := query.Cursor()
			query.All()
			for query.Next() {
				posSlice := pos.Slice(cursor)
				velSlice := vel.Slice(cursor)
				accSlice := acc.Slice(cursor)
				for i := range cursor.IDs {
					velSlice[i].X += accSlice[i].X
					velSlice[i].Y += accSlice[i].Y
					posSlice[i].X += velSlice[i].X
					posSlice[i].Y += velSlice[i].Y
				}
			}
		},
	})

	// Define the report system: builds its own Query in Init, reads and
	// prints in Update — the same pattern as movementSystem, just for a
	// read instead of a write. Reads flow through systems too, not raw
	// lookups pulled out of the ECS.
	var reportQuery *goke.Query
	reportSystem := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			reportQuery = si.NewQueryBuilder(&pos).Build()
		},
		OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) {
			if reportQuery.Seek(entityID) {
				p := pos.At(reportQuery.Cursor())
				fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
			}
		},
	})

	// Configure the ECS's execution workflow and synchronization points
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync() // Ensure all component updates are flushed and matchers are consistent
		ctx.Run(reportSystem, d)
	})

	// Execute a single simulation step (standard 120 TPS)
	ecs.Tick(time.Second / 120)
}
