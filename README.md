# GOKe

<p align="center">
  <img src=".github/docs/img/logo.png" alt="GOKe Logo" width="300">
  <br>
  <a href="https://go.dev">
    <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  </a>
  <a href="https://pkg.go.dev/github.com/kjkrol/goke">
    <img src="https://img.shields.io/badge/GoDoc-Reference-007d9c?style=flat-square&logo=go" alt="GoDoc">
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License">
  </a>
  <a href="https://goreportcard.com/report/github.com/kjkrol/goke">
    <img src="https://goreportcard.com/badge/github.com/kjkrol/goke" alt="Go Report Card">
  </a>
  <a href="https://app.codecov.io/gh/kjkrol/goke">
    <img src="https://img.shields.io/codecov/c/github/kjkrol/goke?style=flat-square&logo=codecov" alt="Codecov Coverage">
  </a>
  <a href="https://github.com/kjkrol/goke/actions">
    <img src="https://github.com/kjkrol/goke/actions/workflows/go.yml/badge.svg" alt="Go Quality Check">
  </a>
  <a href="https://github.com/avelino/awesome-go">
    <img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go">
  </a>
</p>

**GOKe** is a type-safe, archetype-based [Entity Component System](https://en.wikipedia.org/wiki/Entity_component_system) (ECS) for [Go](https://go.dev/). It uses a **Structure of Arrays (SoA)** storage model and Data-Oriented Design principles to enable cache-friendly iteration and efficient processing of large numbers of entities.

While primarily designed for game development, its archetype-based SoA architecture also makes it well suited for simulations, AI agents, real-time analytics, and other performance-critical workloads.

<p align="center">
    <a href="#features">Features</a>
    &nbsp;&bull;&nbsp;
    <a href="#installation">Installation</a> 
    &nbsp;&bull;&nbsp;
    <a href="BENCHMARKS.md">Benchmarks</a>
    &nbsp;&bull;&nbsp;
    <a href="#performance">Performance</a>
    &nbsp;&bull;&nbsp; 
    <a href="#usage">Usage</a>
    &nbsp;&bull;&nbsp; 
    <a href="#architecture">Architecture</a>
    &nbsp;&bull;&nbsp; 
    <a href="#roadmap">Roadmap</a>
    &nbsp;&bull;&nbsp; 
    <a href="#documentation">Documentation</a>
</p>

# Design Goals

GOKe is primarily an ECS for game development, but its archetype-based
SoA architecture also makes it well suited for simulations, AI agents,
real-time analytics, and other performance-critical workloads.

The project is built around a few core principles:

- Predictable performance over clever abstractions
- Cache-friendly data layouts
- Zero-allocation hot paths
- Type-safe APIs without reflection
- Native Go development without CGO dependencies
- Explicit execution over automatic scheduling

While native C and Rust ECS frameworks may achieve higher peak throughput,
GOKe is designed to maximize performance within the Go ecosystem. For many
projects, avoiding CGO boundaries, external dependencies, and cross-language
integration costs can outweigh the gains of a faster foreign implementation.

<a id="installation"></a>
# 📦 Installation

GOKe requires **Go 1.26** or newer.

```bash
go get github.com/kjkrol/goke
```

<a id="features"></a>

# ✨ Key Features

## ✨ Key Features

* **Strictly Zero-Allocation API** All runtime hot paths—including iteration, component access, and view filtering—execute without heap allocations. Memory is allocated only during structural changes (entity creation, component addition, or storage growth), eliminating garbage collector pressure during normal update loops.
* **Cache-Friendly Paged Iteration** Entity traversal operates directly on contiguous memory pages. This layout maximizes CPU cache locality and enables highly efficient iteration throughput.
* **Memory-Conscious Storage (Chunked SoA)** Data is stored in chunked Structure-of-Arrays (SoA) pages. Capacity growth allocates new pages instead of triggering large slice reallocations. Deleted entities are removed using a swap-and-pop operation, keeping storage densely packed and avoiding fragmentation.
* **Generational O(1) Lookups** Direct entity-to-storage mapping provides constant-time component access. Backed by 64-bit identifiers (`uid.UID64`), it enables safe entity recycling while preventing ABA issues and stale entity references.
* **Blueprint-Based Mass Spawning** High-throughput template instantiation optimized for creating large batches of entities. Blueprints leverage contiguous memory copying and integrate naturally with high-performance object pooling workflows.
* **Safe Parallel Execution & Scheduling** An explicit scheduling model provides deterministic execution order with built-in parallel execution (`RunParallel`). Structural changes requested during parallel execution are captured by Command Buffers and applied at explicit `Sync()` points, eliminating data races without hidden locking overhead. Runtime entity creation inside systems is intentionally unsupported; use blueprints or object pools instead.
* **Type-Safe API** Fully generic component access without reflection, runtime type assertions, or interface-based component storage.
> 💡 **See the Performance & Scalability section below for benchmark results across worlds ranging from 2¹⁰ to 2²⁰ entities.**

<a id="performance"></a>
# ⏱️ Performance & Scalability
GOKe is designed for predictable performance at scale. By utilizing a **Centralized Record System** (dense array lookup) instead of traditional hash maps, structural operations and query execution remain effectively independent of the total entity count ($N$).

## 📊 Cross-framework comparison
Benchmarks against other Go ECS libraries (Arche, Donburi, Ento, etc.) are maintained in a dedicated project — [**go-ecs-benchmarks**](https://github.com/mlange-42/go-ecs-benchmarks) by [@mlange-42](https://github.com/mlange-42). 

⚠️ Before drawing conclusions, verify which GOKe version (tag) is used in the comparison, as published results may lag behind the latest release.

## Scalability Validation
Benchmark results were validated across worlds ranging from **2¹⁰ (1,024)** to **2²⁰ (1,048,576)** entities. Per-entity costs remained nearly constant throughout this range, demonstrating the scale-independent behavior of GOKe's core operations.

| Category | Operation | Observed Cost (2¹⁰–2²⁰ Entities) | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Throughput** | **Iteration (View.All)** | **0.35 - 2.39 ns/ent** | **0** | Linear SoA (0-10 components) |
| **Subset Query** | **Filter (per-entity)** | **3.09 - 10.85 ns/ent** | **0** | Per-entity record lookup + pointer math |
| **Structural** | **Batch Create** | **8 - 21 ns/ent** | 4-5 | Blueprint-based pages |
| **Structural** | **Migrate Component** | **35 ns/op** | **0** | Archetype Move (Insert) |
| **Structural** | **Add Tag** | **33 ns/op** | **0** | Archetype Move (Metadata) |
| **Structural** | **Remove Component** | **8 ns/op** | **0** | Swap-and-pop |
| **Structural** | **Remove Entity** | **3 ns/op** | **0** | Index Recycling |
| **Access** | **Get Component** | **4.5 ns/op** | **0** | Inlined Record Lookup |

> **Deep Dive**: For a complete breakdown of benchmark methodology, hardware specifications, scaling tests, and performance charts, see [**BENCHMARKS.md**](./BENCHMARKS.md).

### Reproducing Results

Run the benchmark suite on your own hardware:

```bash
make bench
```

# Real-World Example

The following demo showcases a simple collision simulation built with GOKe and Ebitengine.

It simulates thousands of moving AABBs while maintaining a fixed 120 TPS update loop using archetype-based storage, cache-friendly iteration, and parallel systems.

<table>
  <thead>
    <tr>
      <th style="text-align: left; vertical-align: top; width: 400px;">
        <video src="https://github.com/user-attachments/assets/fa8d1aca-2060-466d-8204-9d6a7443d580" width="400" autoplay loop muted playsinline></video>
        <br>
          <sub><strong>Stats:</strong> 2048 colliding AABBs | 120 TPS | 0.1-1 collisions/tick</sub>
      </th>
      <th style="text-align: left; vertical-align: top; width: 400px;">
        <video src="https://github.com/user-attachments/assets/f1ef2c0b-fb7b-478a-bc88-77faa48c0623" width="400" autoplay loop muted playsinline></video>
        <br>
        <sub><strong>Stats:</strong> 64 colliding AABBs | 120 TPS | 0.1-1 collisions/tick</sub>
      </th>
    </tr>
  </thead>
</table>

> Source code: [examples/ebiten-demo](examples/ebiten-demo/main.go)

<a id="usage"></a>
# Example
> **New to ECS?** Check out the [**Getting Started with GOKe**](https://github.com/kjkrol/goke/wiki/Getting-Started-with-GOKe) guide for a step-by-step deep dive into building your first simulation.

```go
package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke"
)

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }

func main() {
	// Initialize the ECS world.
	// The ECS instance acts as the central coordinator for entities and systems.
	ecs := goke.New()

	// Define component metadata.
	// This binds Go types to internal descriptors, allowing the engine to
	// pre-calculate memory layouts and manage data in contiguous arrays.
	posDesc := goke.RegisterComponent[Pos](ecs)
	_ = goke.RegisterComponent[Vel](ecs)
	_ = goke.RegisterComponent[Acc](ecs)

	// --- Type-Safe Entity Template (Blueprint) ---
	// Blueprints place entities into the correct archetype immediately and
	// reserve memory for all components in a single batch operation.
	// Each yielded page exposes typed slices for direct, in-place initialization.
	blueprint := goke.NewBlueprint3[Pos, Vel, Acc](ecs)

	var entity goke.Entity
	for page := range blueprint.Create(1) {
		entity = page.Entity[0]
		page.Comp1[0] = Pos{X: 0, Y: 0}
		page.Comp2[0] = Vel{X: 1, Y: 1}
		page.Comp3[0] = Acc{X: 0.1, Y: 0.1}
	}

	// Initialize view for Pos, Vel, and Acc components
	view := goke.NewView3[Pos, Vel, Acc](ecs)

	// Define the movement system using the functional registration pattern
	movementSystem := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
		// SoA (Structure of Arrays) layout ensures CPU cache friendliness.
		// View.All yields page-shaped slices over native memory — the inner
		// loop is on the caller side for aggressive compiler inlining.
		for page := range view.All() {
			for i := range page.Entity {
				pos, vel, acc := &page.Comp1[i], &page.Comp2[i], &page.Comp3[i]

				vel.X += acc.X
				vel.Y += acc.Y
				pos.X += vel.X
				pos.Y += vel.Y
			}
		}
	})

	// Configure the ECS's execution workflow and synchronization points
	goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync() // Ensure all component updates are flushed and views are consistent
	})

	// Execute a single simulation step (standard 120 TPS)
	goke.Tick(ecs, time.Second/120)

	p := goke.GetComponent[Pos](ecs, entity, posDesc)
	fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
}
```

### Explore Examples
Check the [**examples/**](./examples) directory for complete, ready-to-run projects.

> ⚠️ **IMPORTANT**:
> **Setup Required**: To keep the core ECS engine lightweight and free of GUI dependencies, examples are managed as isolated modules. Before running them, you must initialize the workspace:
> ```bash
> make setup
> ```

* [**Mini Demo**](./examples/mini-demo/main.go) – The minimalist starter.
* [**Simple Demo**](./examples/simple-demo/main.go) – A slightly more advanced introduction to the ECS lifecycle.
* [**Parallel Demo**](./examples/parallel-demo/main.go) – **Advanced showcase**:
  * Coordination of multiple systems.
  * Concurrent execution using `RunParallel`.
  * Handling structural changes via **Command Buffer** and explicit **Sync points**.
* [**Ebiten Demo**](./examples/ebiten-demo/main.go) – **Graphics Integration & Spatial Physics**:
  * Real-time rendering using [Ebitengine](https://github.com/kjkrol/gokg).
  * High-performance spatial management using [GOKg](https://github.com/kjkrol/gokg).
  * Custom physics pipeline: **Velocity Inversion** is processed strictly before **Position Compensation** to ensure boundary stability.
  * **Note**: Run `make` inside the example directory to fetch dependencies and start the demo.

<a id="architecture"></a>
# Core Architecture

GOKe is an archetype-based ECS built around data-oriented design principles. The storage layer is designed for predictable performance, efficient memory usage, and cache-friendly iteration while maintaining a fully type-safe API.

## Type-Safe API

All component access is resolved at compile time using Go's type system. GOKe avoids reflection, dynamic type assertions, and string-based component lookups in performance-critical code paths.

## Archetype-Based Storage

Entities are grouped into archetypes according to their component composition.

Each archetype is identified by a fixed-size **128-bit component mask**, enabling fast composition checks through bitwise operations. This allows queries to quickly determine whether an archetype matches a required component set without inspecting individual entities.

The maximum number of component types is currently 128 and can be adjusted if needed.

## Paged SoA Memory Layout

Component data is stored using a paged **Structure of Arrays (SoA)** architecture.

Each archetype owns one or more memory pages. From the system's perspective, a page exposes typed component columns:

```text
Page

┌─────────────┬─────────────┬─────────────┬─────────────┐
│ []Entity    │ []CompA     │ []CompB     │ []CompC     │
├─────────────┼─────────────┼─────────────┼─────────────┤
│ e0          │ a0          │ b0          │ c0          │
│ e1          │ a1          │ b1          │ c1          │
│ e2          │ a2          │ b2          │ c2          │
│ ...         │ ...         │ ...         │ ...         │
└─────────────┴─────────────┴─────────────┴─────────────┘
```

This layout allows systems to iterate over large numbers of entities with predictable memory access patterns and excellent cache locality.

Internally, these columns are not separate allocations. Each page is backed by a single contiguous memory block, while typed slices are reconstructed from precomputed layout offsets:

```text
Page (single allocation)

┌────────────────────────────────────────────────────────────┐
│ data []byte                                                │
├────────────────────────────────────────────────────────────┤
│ Entity Column │ CompA Column │ CompB Column │ ...          │
└────────────────────────────────────────────────────────────┘

Layout Offsets

Entity Column ─► offset 0
CompA Column  ─► offset A
CompB Column  ─► offset B
CompC Column  ─► offset C
```

This design provides:

- cache-friendly iteration
- predictable memory growth
- efficient memory utilization
- minimal allocation pressure

Unlike monolithic arrays, growing an archetype only requires allocating an additional page rather than resizing and copying an entire component store.

## Entity Lookup

Every entity is backed by a centralized record that maps directly to its storage location.

```text
EntityID
    ↓
Record
    ↓
(Archetype, Page, Slot)
```

This enables direct access to entity data without hash maps or archetype scans.

As a result, operations such as component access, component insertion, component removal, and entity destruction maintain stable costs regardless of total world size.

## Generational Entity IDs

Entities are represented by `uid.UID64`, a compact generational identifier designed for safe index recycling and stale reference detection.

Destroyed entities can safely reuse storage slots without risking accidental access through outdated references.

For implementation details, see the [uid](https://github.com/kjkrol/uid) package.

## Dense Storage

When an entity is removed, the final slot of the page is immediately moved into the freed slot.

This swap-and-pop strategy:

- avoids holes in memory
- maintains dense storage
- preserves iteration performance
- eliminates the need for background defragmentation

As a result, archetype pages remain compact throughout their lifetime.

# Scheduler

GOKe includes a lightweight scheduler responsible for system execution, synchronization, and deferred structural changes.

The scheduler is intentionally explicit. It does not perform dependency analysis, automatic system ordering, or conflict detection. Instead, developers define execution flow directly through an `ExecutionPlan`.

## Systems

A system is any type implementing the `System` interface:

```go
type System interface {
    Update(
        ReadOnlyRegistry,
        *SystemCommandBuffer,
        time.Duration,
    )
}
```

Each system receives:

- read-only access to the registry
- its own command buffer
- frame delta time

Systems can safely inspect the world state while deferring structural modifications through the command buffer.

## Execution Plans

Execution order is defined explicitly through an `ExecutionPlan`.

```go
goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
    ctx.RunParallel(d, rollSys, betSys)
    ctx.Sync()

    ctx.Run(judgeSys, d)
    ctx.Sync()

    ctx.Run(displayWinnerSys, d)
})
```

Execution plans provide:

- deterministic execution order
- explicit synchronization points
- optional parallel execution

This makes update flow easy to reason about and avoids hidden scheduling behavior.

## Parallel Execution

Independent systems can be executed concurrently using `RunParallel()`.

```go
ctx.RunParallel(d,
    physicsSystem,
    aiSystem,
    animationSystem,
)
```

GOKe does not perform automatic dependency analysis.

Developers are responsible for ensuring that parallel systems do not introduce conflicting writes or race conditions.

## Command Buffers

Each system owns a dedicated command buffer.

Structural operations are queued during execution and applied later during `Sync()`:

- Add Component
- Remove Component
- Remove Entity

This allows systems to safely modify archetype composition while iterating over entities.

## Synchronization

Queued commands become visible only after a synchronization point.

```text
System A
System B
    ↓
   Sync
    ↓
Changes become visible
```

This guarantees predictable behavior and prevents structural changes from invalidating active iterations.

<a id="roadmap"></a>
# 🗺️ Roadmap
Current development focus and planned improvements:

* **Ebitengine Integration:** Dedicated helpers for seamless state synchronization between GOKe systems and Ebitengine's loop — partially prototyped in the [ebiten-demo](./examples/ebiten-demo/main.go), with the goal of extracting it into a separate companion repository.
* **Entity Relations via Tags:** Extend the Tag system to model relationships between entities (parent-child, links, ownership, ...) — adding relational semantics on top of the existing archetype-mask machinery, without sacrificing the zero-allocation hot loop.

> **Live Feature Tracker**
> We manage our long-term goals through GitHub Issues. View all planned core engine expansions and functional capabilities here:
> [**Explore all Pending Features ↗**](https://github.com/kjkrol/goke/issues?q=state%3Aopen%20label%3Afeature)

# When NOT to Use GOKe
GOKe is optimized for large-scale, data-oriented workloads. It may not be the best fit for every project.

* **Small Data Sets** — For a few hundred objects, plain Go structs and slices are often simpler and sufficiently fast.
* **Deep Hierarchies** — ECS excels at flat data layouts. Tree-oriented domains such as UI systems or DOM-like structures may be better served by traditional object graphs.
* **High Structural Churn** — Archetype migration is efficient, but workloads that continuously add and remove components from large numbers of entities every frame may reduce the benefits of archetype-based storage.
* **Behavior-Centric Designs** — If your application is primarily organized around objects and methods rather than data transformations, an ECS may introduce unnecessary complexity.

# Limitations

* **Maximum component types: 128 by default.** The archetype system uses a fixed-size bitmask (`[2]uint64`) for fast component membership checks. Projects requiring more component types can increase this limit by modifying `MaskSize` and `MaxComponents` and recompiling GOKe. This is a compile-time configuration, not a runtime setting.

# License
GOKe is licensed under the MIT License. See the LICENSE [file](./LICENSE) for more details.

<a id="documentation"></a>
# 📖 Documentation
* **API Reference**: Detailed documentation and examples are available on [**pkg.go.dev**](https://pkg.go.dev/github.com/kjkrol/goke).
* **Wiki & Guides**: For a step-by-step deep dive into building your first simulation, check the [**Getting Started with GOKe**](https://github.com/kjkrol/goke/wiki/Getting-Started-with-GOKe) guide.
* **Internal Mechanics**: For a technical breakdown of the engine's core, check the `doc.go` files within the `ecs` packages.
