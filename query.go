package goke

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/internal/query"
	"github.com/kjkrol/goke/v2/iter"
)

// Query matches entities by component mask and provides three access
// patterns: All (chunk-level iteration), Pick (per-entity iteration over a
// given entity subset), and Seek (direct positioning on a single known
// entity). Call All() or Pick() to set the iteration mode and loop with
// Next(), or call Seek() directly for single-entity access.
type Query struct {
	m *query.Matcher
}

// All prepares the Query for full chunk iteration and returns q.
// Call Next() to advance through matched entity chunks; read component
// slices with Comp[T].Slice. Do not call All concurrently on the same Query.
func (q *Query) All() *Query { q.m.All(); return q }

// Pick prepares the Query to iterate over the given entities and returns q.
// Call Next() to advance; read component pointers with Comp[T].At. Entities
// that do not match the Query's mask are skipped. Do not call Pick
// concurrently on the same Query.
func (q *Query) Pick(selected []uid.UID64) *Query { q.m.Pick(selected); return q }

// Next advances the iterator one step. Returns false when exhausted.
// The current mode (set by All or Pick) determines which iteration path runs.
func (q *Query) Next() bool { return q.m.Next() }

// Seek positions the Cursor at entID's storage slot, independent of the
// Query's include/exclude mask — the caller is trusted to know the entity
// carries the tracked components. Returns false if the entity does not
// exist or has been recycled.
func (q *Query) Seek(entID uid.UID64) bool { return q.m.Seek(entID) }

// SeekH (Seek homogeneous) positions q's Cursor at entID's storage slot. It
// assumes entID is alive and shares the archetype already cached by a prior
// Seek call on q — call q.Seek once first to establish it, then call SeekH
// for the rest of a batch known to be alive and come from that one
// archetype.
//
// Returns false if entID's archetype turns out to differ from the cached
// one — the Cursor must not be used in that case; call q.Seek(entID)
// instead. Behavior is undefined if entID is not alive.
func (q *Query) SeekH(entID uid.UID64) bool { return q.m.SeekH(entID) }

// Clear resets the Query to its zero state. Called internally when a Query
// is released back to its catalog.
func (q *Query) Clear() { q.m.Clear() }

// Cursor returns the Query's current cursor, populated by Next or Seek.
// Pass it to Comp[T].Slice (All-mode) or Comp[T].At (Pick/Seek-mode).
func (q *Query) Cursor() *iter.Cursor { return &q.m.Cursor }

// Entity returns the current entity in Pick-mode iteration.
func (q *Query) Entity() uid.UID64 { return q.m.Entity }

// Idx returns the current index into the slice passed to Pick.
func (q *Query) Idx() int { return q.m.Idx }

// ----------------- BUILDER -----------------

// QueryBuilder assembles a Query's access options. Start with
// NewQueryBuilder, optionally chain Include/Exclude, and finish with Build.
type QueryBuilder struct {
	ecs  *ECS
	opts []Opt
}

// NewQueryBuilder starts a QueryBuilder, tracking the given components as
// data columns (equivalent to Track[T] for each).
func (ecs *ECS) NewQueryBuilder(comps ...Trackable) *QueryBuilder {
	b := &QueryBuilder{ecs: ecs, opts: make([]Opt, 0, len(comps))}
	for _, c := range comps {
		b.opts = append(b.opts, c.asTrack())
	}
	return b
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

// Build creates the Query from the accumulated options.
func (b *QueryBuilder) Build() *Query {
	return &Query{m: b.ecs.registry.AddMatcher(b.opts...)}
}

// Include adds a required component type T to the Query's filter.
// Only entities that possess this component will be matched.
func Include[T any]() Opt { return comp.Include[T]() }

// Exclude adds an exclusion for component type T to the Query's filter.
// Entities that possess this component will not be matched.
func Exclude[T any]() Opt { return comp.Exclude[T]() }
