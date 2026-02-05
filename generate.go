//go:generate go run internal/gen/views/main.go
//go:generate go run internal/gen/blueprints/main.go
package goke

// This file triggers the generation of type-safe ECS views and blueprints.
// The generators create optimized structures for different component counts,
// ensuring type safety without the overhead of reflection at runtime.
