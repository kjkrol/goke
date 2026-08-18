package bench_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
)

// enqueueRemoveSubset returns a system that iterates q chunk by chunk and
// registers one removal command per chunk until limit entities are
// covered. Registration only — removal itself runs at the plan's Sync point.
func enqueueRemoveSubset(q *goke.Query, limit int) goke.SystemFn {
	var remover *goke.Remover
	return goke.SystemFn{
		OnInit: func(si *goke.SysInit) { remover = si.Remover() },
		OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
			q.All()
			taken := 0
			for taken < limit && q.Next() {
				ids := q.Cursor().IDs
				if remaining := limit - taken; len(ids) > remaining {
					ids = ids[:remaining]
				}
				buf := q.BeginMigrate(cb)
				for _, id := range ids {
					buf.Add(id)
				}
				buf.Commit(remover)
				taken += len(ids)
			}
		},
	}
}

// enqueueRemoveScattered returns a system that registers removal commands
// for a randomly scattered subset: pick[i] decides whether the i-th matched
// entity (in chunk iteration order) is included. Scattered picks from one
// chunk still share that chunk's ChunkSnapshot, so each chunk contributes one
// command — but the resulting holes are non-contiguous, exercising the
// slot-level compaction path.
func enqueueRemoveScattered(q *goke.Query, pick []bool) goke.SystemFn {
	var remover *goke.Remover
	return goke.SystemFn{
		OnInit: func(si *goke.SysInit) { remover = si.Remover() },
		OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
			q.All()
			pos := 0
			for q.Next() {
				ids := q.Cursor().IDs
				buf := q.BeginMigrate(cb)
				for _, id := range ids {
					if pos < len(pick) && pick[pos] {
						buf.Add(id)
					}
					pos++
				}
				buf.Commit(remover)
			}
		},
	}
}

// timedRemovalPlan runs remSys and times only its Sync — the per-chunk
// Remove work. restore re-populates the removed entities via Factory,
// entirely outside the timer (removed entities cannot be migrated back).
func timedRemovalPlan(b *testing.B, remSys goke.Runnable, restore func()) goke.Plan {
	return func(ctx goke.RunCtx, d time.Duration) {
		b.StopTimer()
		ctx.Run(remSys, d)
		b.StartTimer()
		_ = ctx.Sync()
		b.StopTimer()
		restore()
		b.StartTimer()
	}
}

// runRemoverLeaf runs one Remove benchmark leaf in both entity orders:
// sorted enqueues a contiguous prefix of the matched chunks, random a
// scattered pick (skipped when subset == pop, where the two are identical:
// the whole population is removed either way). setup (re)populates the world
// after ecs.Reset and returns the timed forward query plus the untimed
// restore closure; only the Sync executing the removal is timed.
func runRemoverLeaf(b *testing.B, ecs *goke.ECS, name string, subset int,
	setup func() (removeQ *goke.Query, restore func())) {

	run := func(order string, mkSys func(*goke.Query) goke.SystemFn) {
		b.Run(name+"/"+order, func(b *testing.B) {
			ecs.Reset()
			removeQ, restore := setup()
			remSys := ecs.RegSys(mkSys(removeQ))
			ecs.SetPlan(timedRemovalPlan(b, remSys, restore))
			measurePerEntity(b, subset, func() {
				for b.Loop() {
					ecs.Tick(0)
				}
			})
		})
	}

	run("sorted", func(q *goke.Query) goke.SystemFn {
		return enqueueRemoveSubset(q, subset)
	})
	if subset < entitiesNumber {
		pick := randPickMask(entitiesNumber, subset, 42)
		run("random", func(q *goke.Query) goke.SystemFn {
			return enqueueRemoveScattered(q, pick)
		})
	}
}
