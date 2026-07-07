package bench_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// populatePos creates count entities carrying only {Pos}.
func populatePos(ecs *goke.ECS, count int) []uid.UID64 {
	factory := ecs.NewFactory(new(goke.Comp[Pos]))
	var ids []uid.UID64
	factory.Create(count)
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}
	return ids
}

// populatePosTag creates count entities carrying {Pos, Tag}.
// Tag is zero-size so it cannot be a factory column; entities are created with
// {Pos} first and then promoted to {Pos, Tag} via an Editor (setup only).
func populatePosTag(ecs *goke.ECS, count int) []uid.UID64 {
	ids := populatePos(ecs, count)
	addTag := ecs.NewEditorBuilder(new(goke.Comp[Tag])).Build()
	for _, id := range ids {
		addTag.Update(id)
	}
	return ids
}

// enqueueSubset returns a system that iterates q chunk by chunk and registers
// one MassMigrate command per chunk until limit entities are covered.
// Registration only — the migration itself runs at the plan's Sync point.
func enqueueSubset(q *goke.Query, mig *goke.Migrator, limit int) goke.SystemFn {
	return func(cb *goke.CmdBuf, _ time.Duration) {
		q.All()
		taken := 0
		for taken < limit && q.Next() {
			ids := q.Cursor().IDs
			if rem := limit - taken; len(ids) > rem {
				ids = ids[:rem]
			}
			goke.CmdBufMassMigrate(cb, mig, q.ChunkCtx(), ids)
			taken += len(ids)
		}
	}
}

// enqueueAll returns a system that registers one MassMigrate command per
// chunk for every entity matched by q.
func enqueueAll(q *goke.Query, mig *goke.Migrator) goke.SystemFn {
	return func(cb *goke.CmdBuf, _ time.Duration) {
		q.All()
		for q.Next() {
			goke.CmdBufMassMigrate(cb, mig, q.ChunkCtx(), q.Cursor().IDs)
		}
	}
}

// randPickMask returns a mask of length pop with exactly k true entries at
// uniformly random positions (seeded — reproducible across runs).
func randPickMask(pop, k int, seed uint64) []bool {
	pick := make([]bool, pop)
	for i := range k {
		pick[i] = true
	}
	r := rand.New(rand.NewPCG(seed, seed))
	r.Shuffle(pop, func(i, j int) { pick[i], pick[j] = pick[j], pick[i] })
	return pick
}

// enqueueScattered returns a system that registers MassMigrate commands for a
// randomly scattered subset: pick[i] decides whether the i-th matched entity
// (in chunk iteration order) is included. Scattered picks from one chunk still
// share that chunk's ChunkCtx, so each chunk contributes one command — but the
// resulting holes are non-contiguous, exercising the slot-level compaction path.
func enqueueScattered(q *goke.Query, mig *goke.Migrator, pick []bool) goke.SystemFn {
	buf := make([]uid.UID64, 0, len(pick))
	return func(cb *goke.CmdBuf, _ time.Duration) {
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
			goke.CmdBufMassMigrate(cb, mig, q.ChunkCtx(), buf)
		}
	}
}

// timedMigrationPlan runs migSys and times only its Sync — the per-chunk
// ApplyChunk work. The restoreSys tick (counter-migration back to the initial
// archetype) runs entirely outside the timer.
func timedMigrationPlan(b *testing.B, migSys, restoreSys goke.System) goke.Plan {
	return func(ctx goke.RunCtx, d time.Duration) {
		b.StopTimer()
		ctx.Run(migSys, d)
		b.StartTimer()
		_ = ctx.Sync()
		b.StopTimer()
		ctx.Run(restoreSys, d)
		_ = ctx.Sync()
		b.StartTimer()
	}
}
