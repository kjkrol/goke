//go:generate go run ../../internal/gen/queries/main.go
package ecsq

// This file triggers the generation of type-safe ECS queries.
// The generator located in internal/gen/queries creates query_gen_N.go files,
// providing optimized structures for different component counts.
