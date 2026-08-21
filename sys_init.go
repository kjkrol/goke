package goke

// SysInit is the capability handle passed to System.Init and used by
// ecs.Setup — the only way to construct Query and Factory builders. Editor
// and ValueEditor builders come from an already-built Query instead (see
// Query.NewEditorBuilder / Query.NewValueEditorBuilder).
type SysInit struct {
	ecs *ECS
}

// NewQueryBuilder starts a QueryBuilder, tracking the given components as
// data columns (equivalent to Track[T] for each).
func (s *SysInit) NewQueryBuilder(comps ...Trackable) *QueryBuilder {
	b := &QueryBuilder{ecs: s.ecs, opts: make([]Opt, 0, len(comps))}
	for _, c := range comps {
		b.opts = append(b.opts, c.asTrack())
	}
	return b
}

// NewFactory resolves or creates the archetype from the given components and
// returns a reusable Factory ready for repeated Create/Next cycles. Each
// component behaves like Add[T] — pass &comp directly.
func (s *SysInit) NewFactory(comps ...Addable) *Factory {
	opts := make([]EditOpt, len(comps))
	for i, c := range comps {
		opts[i] = c.asAdd()
	}
	return s.ecs.registry.CreateFactory(opts...)
}

// RegComp registers component type T (idempotent, safe to call lazily)
// from within Init.
func (s *SysInit) RegComp[T any]() CompID { return s.ecs.RegComp[T]() }

// Remover returns the shared Remover for whole-entity removal — pass it to
// MigrateBuf.Commit. It carries no per-call configuration (removes any
// entity regardless of archetype), so every call returns the same instance.
func (s *SysInit) Remover() *Remover { return s.ecs.registry.CreateRemover() }
