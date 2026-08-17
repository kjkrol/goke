package goke

// MigratorBuilder assembles a Migrator's structural-change options. Start with
// NewMigratorBuilder, optionally chain Remove, and finish with Build.
type MigratorBuilder struct {
	ecs  *ECS
	opts []EditOpt
}

// NewMigratorBuilder starts a MigratorBuilder, adding the given components
// (equivalent to Add[T] for each).
func (ecs *ECS) NewMigratorBuilder(comps ...Addable) *MigratorBuilder {
	b := &MigratorBuilder{ecs: ecs, opts: make([]EditOpt, 0, len(comps))}
	for _, c := range comps {
		b.opts = append(b.opts, c.asAdd())
	}
	return b
}

// Remove adds component types to remove, built via Remove[T]().
func (b *MigratorBuilder) Remove(opts ...EditOpt) *MigratorBuilder {
	b.opts = append(b.opts, opts...)
	return b
}

// Build creates the Migrator from the accumulated options.
func (b *MigratorBuilder) Build() *Migrator {
	return b.ecs.registry.CreateMigrator(b.opts...)
}
