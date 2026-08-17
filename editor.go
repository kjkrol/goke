package goke

// EditorBuilder assembles an Editor's structural-change options. Start with
// NewEditorBuilder, optionally chain Remove, and finish with Build.
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

// Remove adds component types to remove, built via Remove[T]().
func (b *EditorBuilder) Remove(opts ...EditOpt) *EditorBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build creates the Editor from the accumulated options.
func (b *EditorBuilder) Build() *Editor {
	return b.ecs.registry.CreateEditor(b.opts...)
}
