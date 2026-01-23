package ecs

type ViewOption func(*ViewBuilder)

func OnType[T any]() ViewOption {
	return func(b *ViewBuilder) {
		id := ensureComponentRegistered[T](b.reg.componentsRegistry)
		b.OnType(id)
	}
}

// View factory based on Functional Options pattern
func NewView(reg *Registry, options ...ViewOption) *View {
	builder := NewViewBuilder(reg)
	for _, option := range options {
		option(builder)
	}
	return builder.Build()
}
