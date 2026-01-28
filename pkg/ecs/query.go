package ecs

import (
	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/goke/pkg/ecs/ecsq"
)

// ViewOption defines criteria for filtering entities in a query beyond
// the required components.
type ViewOption = core.ViewOption

// WithTag restricts the query to entities that possess the specified component T,
// even if T is not requested in the query's returned data.
func WithTag[T any]() ViewOption {
	return core.WithTag[T]()
}

// Without restricts the query to entities that do not possess the specified component T.
func Without[T any]() ViewOption {
	return core.Without[T]()
}

// NewQuery0 creates a query for entities based solely on ViewOptions,
// without retrieving any component data.
func NewQuery0(eng *Engine, options ...ViewOption) *ecsq.Query0 {
	return ecsq.NewQuery0(eng.registry, options...)
}

// NewQuery1 creates a type-safe query to retrieve 1 component type
// from entities matching the criteria.
func NewQuery1[T1 any](eng *Engine, options ...ViewOption) *ecsq.Query1[T1] {
	return ecsq.NewQuery1[T1](eng.registry, options...)
}

// NewQuery2 creates a type-safe query to retrieve 2 component types
// from entities matching the criteria.
func NewQuery2[T1, T2 any](eng *Engine, options ...ViewOption) *ecsq.Query2[T1, T2] {
	return ecsq.NewQuery2[T1, T2](eng.registry, options...)
}

// NewQuery3 creates a type-safe query to retrieve 3 component types.
func NewQuery3[T1, T2, T3 any](eng *Engine, options ...ViewOption) *ecsq.Query3[T1, T2, T3] {
	return ecsq.NewQuery3[T1, T2, T3](eng.registry, options...)
}

// NewQuery4 creates a type-safe query to retrieve 4 component types.
// This is the maximum number of components returned in a single Head structure
// for optimal CPU cache prefetching.
func NewQuery4[T1, T2, T3, T4 any](eng *Engine, options ...ViewOption) *ecsq.Query4[T1, T2, T3, T4] {
	return ecsq.NewQuery4[T1, T2, T3, T4](eng.registry, options...)
}

// NewQuery5 creates a type-safe query to retrieve 5 component types.
// Components are split between Head and Tail structures to maintain performance.
func NewQuery5[T1, T2, T3, T4, T5 any](eng *Engine, options ...ViewOption) *ecsq.Query5[T1, T2, T3, T4, T5] {
	return ecsq.NewQuery5[T1, T2, T3, T4, T5](eng.registry, options...)
}

// NewQuery6 creates a type-safe query to retrieve 6 component types.
func NewQuery6[T1, T2, T3, T4, T5, T6 any](eng *Engine, options ...ViewOption) *ecsq.Query6[T1, T2, T3, T4, T5, T6] {
	return ecsq.NewQuery6[T1, T2, T3, T4, T5, T6](eng.registry, options...)
}

// NewQuery7 creates a type-safe query to retrieve 7 component types.
func NewQuery7[T1, T2, T3, T4, T5, T6, T7 any](eng *Engine, options ...ViewOption) *ecsq.Query7[T1, T2, T3, T4, T5, T6, T7] {
	return ecsq.NewQuery7[T1, T2, T3, T4, T5, T6, T7](eng.registry, options...)
}

// NewQuery8 creates a type-safe query to retrieve 8 component types.
func NewQuery8[T1, T2, T3, T4, T5, T6, T7, T8 any](eng *Engine, options ...ViewOption) *ecsq.Query8[T1, T2, T3, T4, T5, T6, T7, T8] {
	return ecsq.NewQuery8[T1, T2, T3, T4, T5, T6, T7, T8](eng.registry, options...)
}
