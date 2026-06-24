package goke

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
	return b.ecs.registry.AddMatcher(b.opts...)
}

// EditorBuilder assembles an Editor's structural-change options. Start with
// NewEditorBuilder, optionally chain Delete, and finish with Build.
type EditorBuilder struct {
	ecs  *ECS
	opts []EditOpt
}

// NewEditorBuilder starts an EditorBuilder, adding the given components
// (equivalent to Add[T] for each).
func (ecs *ECS) NewEditorBuilder(comps ...Addable) *EditorBuilder {
	b := &EditorBuilder{ecs: ecs, opts: make([]EditOpt, 0, len(comps))}
	for _, c := range comps {
		b.opts = append(b.opts, c.asAdd())
	}
	return b
}

// Delete adds component types to remove, built via Del[T]().
func (b *EditorBuilder) Delete(opts ...EditOpt) *EditorBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build creates the Editor from the accumulated options.
func (b *EditorBuilder) Build() *Editor {
	return b.ecs.registry.CreateEditor(b.opts...)
}
