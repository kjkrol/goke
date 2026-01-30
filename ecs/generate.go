//go:generate go run internal/gen/views/main.go
package ecs

// This file triggers the generation of type-safe ECS views.
// The generator located in internal/gen/views creates view_gen_N.go files,
// providing optimized structures for different component counts.
