package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

const population = 2000

type Position struct{ X, Y float32 }
type Stale struct{}

func main() {
	ecs := goke.New()

	var pos goke.Comp[Position]
	var ids []uid.UID64
	var anyQuery *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos)
		factory.Create(population)
		for factory.Next() {
			ids = append(ids, factory.IDs...)
		}
		anyQuery = si.NewQueryBuilder().Build()
	}})

	var stale goke.Comp[Stale]
	var liveQuery *goke.Query
	var tagStale *goke.Editor
	markStale := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			liveQuery = si.NewQueryBuilder(&pos).Exclude(goke.Exclude[Stale]()).Build()
			tagStale = liveQuery.NewEditorBuilder(&stale).Build()
		},
		OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
			liveQuery.All()
			for liveQuery.Next() {
				cursor := liveQuery.Cursor()
				if len(cursor.IDs) == 0 {
					continue
				}
				mb := liveQuery.BeginMigrate(cb)
				for _, id := range cursor.IDs {
					mb.Add(id)
				}
				mb.Commit(tagStale)
			}
		},
	})

	var staleQuery *goke.Query
	var remover *goke.Remover
	sweepStale := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			staleQuery = si.NewQueryBuilder().Include(goke.Include[Stale]()).Build()
			remover = si.Remover()
		},
		OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
			staleQuery.All()
			for staleQuery.Next() {
				cursor := staleQuery.Cursor()
				if len(cursor.IDs) == 0 {
					continue
				}
				mb := staleQuery.BeginMigrate(cb)
				for _, id := range cursor.IDs {
					mb.Add(id)
				}
				mb.Commit(remover)
			}
		},
	})

	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(markStale, d)
		ctx.Sync()
		ctx.Run(sweepStale, d)
		ctx.Sync()
	})

	fmt.Printf("spawned %d entities\n", len(ids))
	ecs.Tick(time.Second / 120)

	remaining := 0
	for _, id := range ids {
		anyQuery.Pick([]uid.UID64{id})
		if anyQuery.Next() {
			remaining++
		}
	}
	fmt.Printf("after one tick (tag + bulk remove, both in one migration pass per chunk): %d entities remain\n", remaining)
}
