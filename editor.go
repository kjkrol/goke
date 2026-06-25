package goke

import "github.com/kjkrol/goke/v2/internal/comp"

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

// Del returns an EditOpt that removes component T from an entity.
func Del[T any]() EditOpt { return comp.Del[T]() }
