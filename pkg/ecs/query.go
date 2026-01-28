//go:generate go run ../../internal/gen/queries/main.go
package ecs

import "github.com/kjkrol/goke/internal/core"

// ---------------- Queries ------------------------------- //

type ViewOption = core.ViewOption

func WithTag[T any]() ViewOption {
	return core.WithTag[T]()
}

func Without[T any]() ViewOption {
	return core.Without[T]()
}

func NewQuery0(eng *Engine, options ...ViewOption) *Query0 {
	return newQuery0(eng.registry, options...)
}
func NewQuery1[T1 any](eng *Engine, options ...ViewOption) *Query1[T1] {
	return newQuery1[T1](eng.registry, options...)
}
func NewQuery2[T1, T2 any](eng *Engine, options ...ViewOption) *Query2[T1, T2] {
	return newQuery2[T1, T2](eng.registry, options...)
}
func NewQuery3[T1, T2, T3 any](eng *Engine, options ...ViewOption) *Query3[T1, T2, T3] {
	return newQuery3[T1, T2, T3](eng.registry, options...)
}
func NewQuery4[T1, T2, T3, T4 any](eng *Engine, options ...ViewOption) *Query4[T1, T2, T3, T4] {
	return newQuery4[T1, T2, T3, T4](eng.registry, options...)
}
func NewQuery5[T1, T2, T3, T4, T5 any](eng *Engine, options ...ViewOption) *Query5[T1, T2, T3, T4, T5] {
	return newQuery5[T1, T2, T3, T4, T5](eng.registry, options...)
}
func NewQuery6[T1, T2, T3, T4, T5, T6 any](eng *Engine, options ...ViewOption) *Query6[T1, T2, T3, T4, T5, T6] {
	return newQuery6[T1, T2, T3, T4, T5, T6](eng.registry, options...)
}
func NewQuery7[T1, T2, T3, T4, T5, T6, T7 any](eng *Engine, options ...ViewOption) *Query7[T1, T2, T3, T4, T5, T6, T7] {
	return newQuery7[T1, T2, T3, T4, T5, T6, T7](eng.registry, options...)
}
func NewQuery8[T1, T2, T3, T4, T5, T6, T7, T8 any](eng *Engine, options ...ViewOption) *Query8[T1, T2, T3, T4, T5, T6, T7, T8] {
	return newQuery8[T1, T2, T3, T4, T5, T6, T7, T8](eng.registry, options...)
}
