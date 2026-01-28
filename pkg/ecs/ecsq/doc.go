// Package ecsq provides high-performance, type-safe query structures for the Goke ECS engine.
//
// The package name 'ecsq' stands for ECS Query. It contains generated QueryN and HeadN
// types and is intended to be used primarily through ecs.Engine.QueryN factory methods
// to maintain a clean API surface.
//
// Performance and Cache Locality:
// All types in this package are optimized for zero-allocation iteration and direct
// memory access to component columns. The engine stores components in contiguous
// memory blocks (archetype columns). If the requested components are sized to fit
// within the CPU's L1/L2 cache lines, queries will benefit from maximal hardware
// prefetching and minimal cache misses.
//
// Query Capabilities and Hardware Constraints:
// The package provides optimized Queries that can return up to 8 selected components
// simultaneously. For more complex filtering, an unlimited number of additional
// components can be defined using WithTag/Without constraints.
//
// Iteration results are returned through specialized structures: HeadN, TailN, PHeadN,
// and PTailN. This design is deliberate: each structure is strictly limited to
// a maximum of 4 pointer fields. Beyond this threshold, CPU prefetching efficiency
// significantly degrades. By adhering to this limit, the engine ensures that the
// Go compiler and hardware prefetchers can maintain optimal throughput during iteration.
//
// Code Generation:
// This package is primarily composed of generated code. Manual changes to
// query_gen_*.go files will be overwritten during the next generation cycle.
package ecsq
