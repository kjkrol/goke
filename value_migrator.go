package goke

import (
	"reflect"
	"unsafe"

	"github.com/kjkrol/uid"
)

// ValueMigratorBuilder assembles a ValueMigrator's structural-change
// options. Start with NewValueMigratorBuilder, optionally chain Remove, and
// finish with Build.
type ValueMigratorBuilder struct {
	ecs  *ECS
	opts []EditOpt
}

// NewValueMigratorBuilder starts a ValueMigratorBuilder, adding the given
// component (equivalent to Add[T]). Unlike NewMigratorBuilder this takes a
// single Addable, not variadic — ValueMigrator supports exactly one added
// component, the one whose value CmdBufAddCompValue lets you write.
func (ecs *ECS) NewValueMigratorBuilder(addable Addable) *ValueMigratorBuilder {
	return &ValueMigratorBuilder{ecs: ecs, opts: []EditOpt{addable.asAdd()}}
}

// Remove adds component types to remove, built via Remove[T]().
func (b *ValueMigratorBuilder) Remove(opts ...EditOpt) *ValueMigratorBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build creates the ValueMigrator from the accumulated options.
func (b *ValueMigratorBuilder) Build() *ValueMigrator {
	return b.ecs.registry.CreateValueMigrator(b.opts...)
}

// CmdBufAddCompValue records a bulk migration of ids from a single source
// chunk, and additionally reserves space for one value per id in vm's added
// component — the returned slice is uninitialized, write into it before the
// tick's Sync. col must be the same component vm was built with — vm is
// untyped at the Go level (it's built at runtime from an Addable), so this
// is checked here by reflect.Type, not by the compiler; a
// same-sized-but-different T would defeat a size-only check. Applied at the
// next Sync point, same cadence as a Query.BeginMigrate commit.
func CmdBufAddCompValue[T any](cb *CmdBuf, vm *ValueMigrator, col *Comp[T], snap ChunkSnapshot, ids []uid.UID64) []T {
	if t := reflect.TypeFor[T](); t != vm.ValueType() {
		panic("goke: CmdBufAddCompValue: " + t.String() + " does not match this ValueMigrator's added component " + vm.ValueType().String())
	}
	var zero T
	ptr := cb.raw.AddCompValue(vm, snap, ids, unsafe.Sizeof(zero), unsafe.Alignof(zero))
	return unsafe.Slice((*T)(ptr), len(ids))
}
