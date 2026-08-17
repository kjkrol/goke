package bench_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// enqueueSubsetWithValue mirrors enqueueSubset but stages a value per id via
// CmdBufAddCompValue instead of cb.Migrate — one write per entity,
// matching how a real ValueEditor caller computes it during iteration.
func enqueueSubsetWithValue(q *goke.Query, vm *goke.ValueEditor, col *goke.Comp[Pos], limit int) goke.SystemFn {
	return goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
		q.All()
		taken := 0
		for taken < limit && q.Next() {
			ids := q.Cursor().IDs
			if remaining := limit - taken; len(ids) > remaining {
				ids = ids[:remaining]
			}
			vals := goke.CmdBufAddCompValue(cb, vm, col, q.ChunkSnapshot(), ids)
			for i := range vals {
				vals[i] = Pos{X: 1, Y: 1}
			}
			taken += len(ids)
		}
	}}
}

// enqueueScatteredWithValue mirrors enqueueScattered for the value-carrying path.
func enqueueScatteredWithValue(q *goke.Query, vm *goke.ValueEditor, col *goke.Comp[Pos], pick []bool) goke.SystemFn {
	buf := make([]uid.UID64, 0, len(pick))
	return goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
		q.All()
		pos := 0
		for q.Next() {
			ids := q.Cursor().IDs
			buf = buf[:0]
			for _, id := range ids {
				if pos < len(pick) && pick[pos] {
					buf = append(buf, id)
				}
				pos++
			}
			if len(buf) == 0 {
				continue
			}
			vals := goke.CmdBufAddCompValue(cb, vm, col, q.ChunkSnapshot(), buf)
			for i := range vals {
				vals[i] = Pos{X: 1, Y: 1}
			}
		}
	}}
}

// runValueEditorLeaf mirrors runEditorLeaf for ValueEditor: same
// sorted/random dimensions, same timedMigrationPlan (only Sync timed), only
// the forward wiring differs (writes a value per id instead of just moving it).
func runValueEditorLeaf(b *testing.B, ecs *goke.ECS, name string, subset int,
	setup func() (migrateQ *goke.Query, fwd *goke.ValueEditor, col *goke.Comp[Pos], restore goke.Runnable)) {

	run := func(order string, mkSys func(*goke.Query, *goke.ValueEditor, *goke.Comp[Pos]) goke.SystemFn) {
		b.Run(name+"/"+order, func(b *testing.B) {
			ecs.Reset()
			migrateQ, fwd, col, restoreSys := setup()
			migSys := ecs.RegSys(mkSys(migrateQ, fwd, col))
			ecs.SetPlan(timedMigrationPlan(b, migSys, restoreSys))
			measurePerEntity(b, subset, func() {
				for b.Loop() {
					ecs.Tick(0)
				}
			})
		})
	}

	run("sorted", func(q *goke.Query, vm *goke.ValueEditor, col *goke.Comp[Pos]) goke.SystemFn {
		return enqueueSubsetWithValue(q, vm, col, subset)
	})
	if subset < entitiesNumber {
		pick := randPickMask(entitiesNumber, subset, 42)
		run("random", func(q *goke.Query, vm *goke.ValueEditor, col *goke.Comp[Pos]) goke.SystemFn {
			return enqueueScatteredWithValue(q, vm, col, pick)
		})
	}
}
