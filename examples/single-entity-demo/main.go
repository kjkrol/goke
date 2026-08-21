package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type Health struct{ Value int }
type Buff struct{ Multiplier float32 }

// pickupEvent stands in for something arriving from outside the ECS
// entirely — a network message, an input callback, a saved reference —
// naming an entity directly, with no Query involved in reaching it.
type pickupEvent struct {
	target     uid.UID64
	multiplier float32
}

func main() {
	ecs := goke.New()
	buffID := ecs.RegComp[Buff]()

	var health goke.Comp[Health]
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&health)
		factory.Create(3)
		for factory.Next() {
			hp := health.Slice(&factory.Cursor)
			for i := range hp {
				hp[i] = Health{Value: 100}
			}
			ids = append(ids, factory.IDs...)
		}
	}})

	// pendingPickups simulates events arriving between ticks, each naming
	// an entity by id alone — nothing here iterates a Query to find them.
	pendingPickups := []pickupEvent{
		{target: ids[0], multiplier: 1.5},
		{target: ids[2], multiplier: 2.0},
	}

	// applyPickups reaches each target directly via CmdBuf.AddOne — the
	// right tool here, since these ids come from outside any Query
	// iteration: there's no cursor to piggyback a batch on, and the ids
	// are scattered across (possibly different) archetypes, so a forced
	// Query.Pick + BeginMigrate per id would pay Pick's lookup cost for
	// no batching benefit. Compare with examples/bulk-edit-demo, where
	// entities are already being visited in a Query loop and batching
	// through Editor/Remover instead pays off.
	applyPickups := ecs.RegSys(goke.SystemFn{
		OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
			for _, ev := range pendingPickups {
				cb.AddOne(ev.target, buffID, Buff{Multiplier: ev.multiplier})
			}
			pendingPickups = pendingPickups[:0]
		},
	})

	var buff goke.Comp[Buff]
	var report *goke.Query
	reportSystem := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			report = si.NewQueryBuilder(&health, &buff).Build()
		},
		OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) {
			for _, id := range ids {
				report.Pick([]uid.UID64{id})
				if report.Next() {
					b := buff.At(report.Cursor())
					fmt.Printf("entity %v: Buff{Multiplier: %.1f}\n", id, b.Multiplier)
				}
			}
		},
	})

	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(applyPickups, d)
		ctx.Sync()
		ctx.Run(reportSystem, d)
	})

	ecs.Tick(time.Second / 120)
}
