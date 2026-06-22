// Package goke provides a high-performance Entity Component System (ECS) engine
// designed for data-oriented programming and mechanical sympathy.
//
// # Core Philosophy
//
// The engine is built on an Archetype-based storage model using Structure of Arrays (SoA).
// By storing components of the same type in contiguous memory columns, the engine
// ensures that entities live "densely" in memory. This layout allows the CPU to
// leverage hardware prefetching and linear cache access, drastically reducing
// cache misses compared to traditional Object-Oriented patterns.
//
// # Key Concepts
//
//  1. Entities & Generation-based Recycling:
//     Entities are 64-bit identifiers consisting of an Index (32-bit) and a
//     Generation (32-bit). When an entity is removed, its index returns to a pool
//     and its generation is incremented. This ensures that stale references to
//     deleted entities (ABA problem) are easily detected, while allowing the
//     internal storage to reuse memory slots for dense packing and high cache hit rates.
//
//  2. Components as Data Columns:
//     Components are user-defined structs registered within the Registry.
//     The engine treats them as contiguous blocks of memory. By registering a
//     component type, the engine gains metadata (Size and reflect.Type) used to
//     build Archetype Columns. This registry-based approach enables zero-allocation
//     component access and ensures that data of the same type is perfectly aligned
//     for SIMD-like processing speeds.
//
//  3. Systems & Execution Plans:
//     Logic is decoupled into Systems. The engine supports both interface-based
//     systems (System interface) and lightweight functional systems (SystemFn).
//     The order and concurrency of execution are defined via a Plan.
//
//  4. Thread Safety & Parallelism:
//     The engine allows for synchronous or parallel system execution. While the engine
//     provides the tools for high-performance concurrent processing (RunParallel),
//     it follows a "Power to the Programmer" philosophy: it is the developer's
//     responsibility to ensure that systems running in parallel operate on disjoint
//     component sets to avoid data races.
//
//  5. Deferred Commands:
//     To maintain state consistency during system updates, modifications to the
//     world (like adding components or removing entities) are buffered via
//     the CmdBuf and applied during explicit synchronization points (Sync).
//
//  6. Type-Safe Views & Cache-Optimized Queries:
//     Data retrieval is handled through [View] obtained via [CreateView].
//     Component columns are declared with [Col][T] and accessed via
//     [Col.Slice] (bulk) or [Col.At] (per-entity). Bulk iteration via
//     View.All yields SoA chunks (Go slices over native memory), while
//     subset queries via View.Filter yield per-entity component pointers
//     resolved via the entity-to-storage index. All access is zero-allocation and reflection-free.
//
// # Hardware Constraints & Limits
//
// To maintain extreme performance, the engine operates with certain fixed limits:
//
//   - Component Types: The engine supports up to 128 unique component types per registry.
//     This is determined by the Mask (2x64-bit fields), ensuring that
//     archetype matching remains a fast, constant-time bitwise operation.
//
//   - Memory Pre-allocation: Archetypes and internal structures are initialized
//     with predefined capacities (configurable via ECSOption). This reduces
//     early memory fragmentation and minimizes GC pressure during the initial
//     entity burst.
//
//   - Entity Indexing: Entities are 64-bit identifiers, allowing for a virtually
//     unlimited number of entities, constrained only by the available system RAM.
//
//   - View Complexity: A single [View] can track any number of component columns
//     declared via [Col][T]. Additional types can be used as filter-only
//     constraints via Include/Exclude opts without occupying tracked columns.
//
// # Internal Package Dependencies
//
// The internal packages form a strict acyclic dependency graph. Each layer
// may only import packages from layers below it:
//
//	Layer 0   iter     — column-access primitives: Cursor, Col[T]
//	Layer 1   comp     — shared primitives: ID, Def, Mask, Blueprint, DefIndex  (→ iter)
//	Layer 2   mem      — cache-aligned chunked memory layout     (→ comp)
//	Layer 2   orch     — scheduler, plans, command buffers       (→ comp)
//	Layer 3   colstore — column-oriented storage                 (→ comp, mem)
//	Layer 4   arch     — archetype ID, Mask, graph               (→ comp, mem, colstore)
//	Layer 5   addr     — entity address book: Entry, Index, Book (→ arch, mem)
//	Layer 6   ent      — entity lifecycle, Manager, Factory      (→ addr, arch, colstore, comp, mem, iter)
//	Layer 7   query    — query layer, view baking                (→ addr, arch, colstore, comp, iter)
//	Layer 8   reg      — top-level Registry                      (→ ent, arch, comp, query)
//
// Expressed as a directed graph (arrow = "is imported by"):
//
//	iter ──► comp ──► mem ──► colstore ──► arch ──► addr ──► ent ──► reg
//	           └──► orch                              │               ▲
//	                                                  └──► query ─────┘
//
// [github.com/kjkrol/uid] is an external module used across layers (mem, orch, colstore,
// arch, addr, ent, query, reg) for 64-bit generational entity identifiers.
//
// orch and reg are fully independent of each other. The top-level goke
// package is the only place that wires them together, passing a pointer to
// the embedded reg.Registry to orch.NewScheduler as an orch.Mutator.
package goke
