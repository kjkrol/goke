package goke

import (
	"time"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/query"
	"github.com/kjkrol/goke/v3/iter"
)

// Query matches entities by component mask and provides three access
// patterns: All (chunk-level iteration), Pick (per-entity iteration over a
// given entity subset), and Seek (direct positioning on a single known
// entity). Call All() or Pick() to set the iteration mode and loop with
// Next(), or call Seek() directly for single-entity access.
type Query struct {
	w   query.Walker
	ecs *ECS
}

// All prepares the Query for full chunk iteration and returns q.
// Call Next() to advance through matched entity chunks; read component
// slices with Comp[T].Slice. Do not call All concurrently on the same Query.
func (q *Query) All() *Query { q.w.All(); return q }

// Pick prepares the Query to iterate over the given entities and returns q.
// Call Next() to advance; read component pointers with Comp[T].At. Entities
// that do not match the Query's mask are skipped. Do not call Pick
// concurrently on the same Query.
func (q *Query) Pick(selected []uid.UID64) *Query { q.w.Pick(selected); return q }

// Next advances the iterator one step. Returns false when exhausted.
// The current mode (set by All or Pick) determines which iteration path runs.
func (q *Query) Next() bool { return q.w.Next() }

// Paced starts walking, as All does, the blocks whose turn it is this run by the query's pace
// (PacedBy), d a step; without a pace every block, as All.
func (q *Query) Paced(d time.Duration) *Query {
	q.w.Paced(d)
	return q
}

// Dt is the step of the block a Paced walk stands on: d times its period.
func (q *Query) Dt() time.Duration { return q.w.Dt() }

// Due reports whether the block a walk stands on has its turn this run by the query's pace.
func (q *Query) Due() bool { return q.w.Due() }

// Seek positions the Cursor at entID's storage slot, independent of the
// Query's include/exclude mask — the caller is trusted to know the entity
// carries the tracked components. Returns false if the entity does not
// exist or has been recycled.
func (q *Query) Seek(entID uid.UID64) bool { return q.w.Seek(entID) }

// SeekH positions q's Cursor at entID's storage slot, assuming entID is
// alive and shares the archetype already cached by a prior Seek call on q —
// call Seek once, then SeekH for the rest of a batch from that archetype.
// Returns false if the archetype differs (Cursor is then unusable — call
// Seek instead); behavior is undefined if entID is not alive.
func (q *Query) SeekH(entID uid.UID64) bool { return q.w.SeekH(entID) }

// Clear ends q's walk: Next reports nothing until All, Pick or Paced starts another; what q
// matches stays as it is.
func (q *Query) Clear() { q.w.Clear() }

// Cursor returns the Query's current cursor, populated by Next or Seek.
// Pass it to Comp[T].Slice (All-mode) or Comp[T].At (Pick/Seek-mode).
func (q *Query) Cursor() *iter.Cursor { return &q.w.Cursor }

// ChunkSnapshot captures the chunk most recently advanced to by Next() in
// All-mode — used internally by BeginMigrate. Valid only between a Next()
// that returned true and the following Next(); undefined in Pick/Seek mode.
func (q *Query) ChunkSnapshot() ChunkSnapshot { return q.w.ChunkSnapshot() }

// Entity returns the current entity in Pick-mode iteration.
func (q *Query) Entity() uid.UID64 { return q.w.Entity }

// Idx returns the current index into the slice passed to Pick.
func (q *Query) Idx() int { return q.w.Idx }

// ----------------- BUILDER -----------------

// QueryBuilder assembles a Query's access options. Start with
// NewQueryBuilder, optionally chain Include/Exclude, and finish with Build.
type QueryBuilder struct {
	ecs  *ECS
	opts []Opt
}

// Include adds required (filter-only, no data access) component types,
// built via Include[T]().
func (b *QueryBuilder) Include(opts ...Opt) *QueryBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Exclude adds excluded component types, built via Exclude[T]().
func (b *QueryBuilder) Exclude(opts ...Opt) *QueryBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Optional adds tracked data columns (pass &OptComp[T]) that never gate the match.
func (b *QueryBuilder) Optional(comps ...OptTrackable) *QueryBuilder {
	for _, c := range comps {
		b.opts = append(b.opts, c.asOptTrack())
	}
	return b
}

// Build creates the Query from the accumulated options.
func (b *QueryBuilder) Build() *Query {
	q := &Query{ecs: b.ecs}
	q.w.Init(b.ecs.registry.AddMatcher(b.opts...))
	return q
}

// NewEditorBuilder starts an EditorBuilder, adding the given components
// (equivalent to Add[T] for each). The resulting Editor applies to entities
// matched by q.
func (q *Query) NewEditorBuilder(comps ...Addable) *EditorBuilder {
	b := &EditorBuilder{ecs: q.ecs, opts: make([]EditOpt, 0, len(comps))}
	for _, c := range comps {
		b.opts = append(b.opts, c.asAdd())
	}
	return b
}

// NewValueEditorBuilder starts a ValueEditorBuilder for a single added
// component. The resulting ValueEditor applies to entities matched by q.
func (q *Query) NewValueEditorBuilder(addable Addable) *ValueEditorBuilder {
	return &ValueEditorBuilder{ecs: q.ecs, opts: []EditOpt{addable.asAdd()}}
}

// MigrateBuf stages ids for a bulk structural change. Obtained from
// Query.BeginMigrate; call Add per matched entity, then Commit once.
type MigrateBuf struct {
	cb   *CmdBuf
	snap ChunkSnapshot
	ids  []uid.UID64
}

// BeginMigrate starts a MigrateBuf for the current chunk. Call it once per
// destination op; safe to call multiple times per chunk for a mixed batch.
// Valid only until the next Next() call.
func (q *Query) BeginMigrate(cb *CmdBuf) *MigrateBuf {
	n := len(q.Cursor().IDs)
	return &MigrateBuf{cb: cb, snap: q.ChunkSnapshot(), ids: cb.raw.ReserveIDs(n)}
}

// Add stages id for the eventual Commit call.
func (b *MigrateBuf) Add(id uid.UID64) {
	b.ids = append(b.ids, id)
}

// Commit queues op against every id staged since BeginMigrate. For
// whole-entity removal, pass a Remover (SysInit.Remover) — it satisfies
// bulk.Migrator the same way Editor/ValueEditor do.
func (b *MigrateBuf) Commit(op bulk.Migrator) {
	b.cb.raw.CommitReserved(op, b.snap, b.ids)
}

// Include adds a required component type T to the Query's filter.
// Only entities that possess this component will be matched.
func Include[T any]() Opt { return comp.Include[T]() }

// With adds opts to the query: filters, columns, a pace.
func (b *QueryBuilder) With(opts ...Opt) *QueryBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// PacedBy is the pace of a query walked with Paced: the runs numbered by step, a block carrying a
// ruled component going as its rule says, every other block every run.
func PacedBy(step func() uint64, rules ...PaceRule) Opt { return comp.PacedBy(step, rules...) }

// Every is the pace of the blocks carrying T: the runs where step % n == n-1, over n steps each;
// a block meeting several rules goes by the slowest.
func Every[T any](n uint64) PaceRule { return comp.Every[T](n) }

// Never is the pace of the blocks carrying T: no run, whatever else they carry.
func Never[T any]() PaceRule { return comp.Never[T]() }

// Exclude adds an exclusion for component type T to the Query's filter.
// Entities that possess this component will not be matched.
func Exclude[T any]() Opt { return comp.Exclude[T]() }
