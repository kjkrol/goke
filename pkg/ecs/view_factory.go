package ecs

type ViewOption func(*ViewBuilder)

// WithTag adds a component to the query's inclusion filter without fetching its data.
//
// It can be used with:
//  1. "True Tags": Zero-sized structs (struct{}) used purely as markers.
//  2. "Data Components": Regular structs where you only care about their presence
//     for filtering, but don't need to access their internal fields during iteration.
//
// This component will be added to the ArchetypeMask to filter entities, but will
// not be included in the returned data heads (e.g., Head1, Head2).
func WithTag[T any]() ViewOption {
	return func(b *ViewBuilder) {
		OnTagType[T](b)
	}
}

func Without[T any]() ViewOption {
	return func(b *ViewBuilder) {
		OnCompExcludeType[T](b)
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
