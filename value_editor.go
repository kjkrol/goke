package goke

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"
)

// ValueEditorBuilder assembles a ValueEditor's structural-change
// options. Start with NewValueEditorBuilder, optionally chain Remove, and
// finish with Build.
type ValueEditorBuilder struct {
	ecs  *ECS
	opts []EditOpt
}

// Remove adds component types to remove, built via Remove[T]().
func (b *ValueEditorBuilder) Remove(opts ...EditOpt) *ValueEditorBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build creates the ValueEditor from the accumulated options.
func (b *ValueEditorBuilder) Build() *ValueEditor {
	return b.ecs.registry.CreateValueEditor(b.opts...)
}

// AddCompValue queues ids for migration via vm and returns an
// uninitialized per-id slice to write the added component's value into
// before Sync. col must match vm's added component (panics otherwise).
func (cb *CmdBuf) AddCompValue[T any](vm *ValueEditor, col *Comp[T], snap ChunkSnapshot, ids []uid.UID64) []T {
	if t := reflect.TypeFor[T](); t != vm.ValueType() {
		panic("goke: CmdBuf.AddCompValue: " + t.String() + " does not match this ValueEditor's added component " + vm.ValueType().String())
	}
	var zero T
	ptr := cb.raw.AddCompValue(vm, snap, ids, unsafe.Sizeof(zero), unsafe.Alignof(zero))
	return unsafe.Slice((*T)(ptr), len(ids))
}
