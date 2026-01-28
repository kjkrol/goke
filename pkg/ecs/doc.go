// Package ecs provides a high-performance Entity Component System (ECS) engine
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
//     Components are user-defined structs registered within the ComponentsRegistry.
//     The engine treats them as contiguous blocks of memory. By registering a
//     component type, the engine gains metadata (Size and reflect.Type) used to
//     build Archetype Columns. This registry-based approach enables zero-allocation
//     component access and ensures that data of the same type is perfectly aligned
//     for SIMD-like processing speeds.
//
//  3. Systems & Execution Plans:
//     Logic is decoupled into Systems. The engine supports both interface-based
//     systems (System interface) and lightweight functional systems (SystemFunc).
//     The order and concurrency of execution are defined via an ExecutionPlan.
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
//     the SystemCommandBuffer and applied during explicit synchronization points (Sync).
//
//  6. Type-Safe Queries:
//     Data retrieval is handled through highly optimized, generated queries
//     (provided by the ecsq package). These queries provide type safety without
//     the overhead of reflection during the main loop.
//
// # Hardware Constraints & Limits
//
// To maintain extreme performance, the engine operates with certain fixed limits:
//
//   - Component Types: The engine supports up to 256 unique component types per registry.
//     This is determined by the ArchetypeMask (4x64-bit fields), ensuring that
//     archetype matching remains a fast, constant-time bitwise operation.
//
//   - Memory Pre-allocation: Archetypes are initialized with a default capacity
//     (1024 entities). While storage grows dynamically, pre-allocating large
//     batches of entities can further minimize memory fragmentation and GC pressure.
//
//   - Entity Indexing: Entities are 64-bit identifiers, allowing for a virtually
//     unlimited number of entities, constrained only by the available system RAM.
package ecs
