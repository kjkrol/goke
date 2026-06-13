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

<p align="center">
    <a href="#features">Features</a>
    &nbsp;&bull;&nbsp;
    <a href="#installation">Installation</a> 
    &nbsp;&bull;&nbsp;
    <a href="BENCHMARKS.md">Benchmarks</a>
    &nbsp;&bull;&nbsp;
    <a href="#performance">Performance</a>
    &nbsp;&bull;&nbsp; 
    <a href="#example">Example</a>
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

- Abstractions that reflect ideas, not implementation details
- Predictable performance with no hidden costs
- Cache-friendly data layouts
- Zero-allocation hot paths
- Type-safe APIs without reflection
- Native Go development without CGO dependencies

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

| Capability | How |
|:---|:---|
| **Zero-allocation hot paths** | Chunk-based SoA layout with direct pointer arithmetic — no GC pressure during iteration or component access |
| **Predictable iteration speed** | Linear SoA memory access — cache-friendly, branch-free inner loops; sub-nanosecond per entity at scale |
| **O(1) component lookup** | Entity-to-storage is a direct array index, not a hash map — constant time at any world size |
| **Safe entity recycling** | 64-bit generational IDs detect stale references after deletion, preventing ABA bugs |
| **Cache-friendly storage** | Contiguous SoA chunks; growth appends new chunks, removal uses swap-and-pop — no fragmentation |
| **Batch entity creation** | Blueprint templates copy components in bulk — no per-entity allocation |
| **Type-safe component API** | Fully generic — no reflection, no interface boxing, no runtime type assertions |
| **Built-in scheduler** | Declarative `Plan` wires systems into an execution graph — a full ECS runtime, not just a component store |
| **Command Buffer** | Structural changes during iteration are queued and flushed at explicit `Sync()` points — enables safe `RunParallel` |

> 💡 **See the Performance & Scalability section below for benchmark results across worlds ranging from 2¹⁰ to 2²⁰ entities.**

<a id="performance"></a>
# ⏱️ Performance & Scalability
GOKe is designed for predictable performance at scale. By utilizing a **Centralized Record System** (dense array lookup) instead of traditional hash maps, structural operations and query execution remain effectively independent of the total entity count ($N$).

## 📊 Cross-framework comparison
Benchmarks against other Go ECS libraries (Arche, Donburi, Ento, etc.) are maintained in a dedicated project — [**go-ecs-benchmarks**](https://github.com/mlange-42/go-ecs-benchmarks) by [@mlange-42](https://github.com/mlange-42). 

⚠️ Before drawing conclusions, verify which GOKe version (tag) is used in the comparison, as published results may lag behind the latest release.

## Scalability Validation
Benchmark results were validated across worlds ranging from **2¹⁰ (1,024)** to **2²⁰ (1,048,576)** entities on an **Apple M1 Max**. Per-entity costs remained nearly constant throughout this range, demonstrating the scale-independent behavior of GOKe's core operations.

| Category | Operation | Observed Cost (2¹⁰–2²⁰ Entities) | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Throughput** | **Iteration (View.All)** | **0.35 - 2.39 ns/ent** | **0** | Linear SoA (0-10 components) |
| **Subset Query** | **Filter (per-entity)** | **3.09 - 10.85 ns/ent** | **0** | Per-entity record lookup + pointer math |
| **Structural** | **Batch Create** | **8 - 21 ns/ent** | 4-5 | Blueprint-based chunks |
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
        <video src="https://github.com/user-attachments/assets/2b921500-eb3e-49bf-98ee-ac741746e64d" width="400" autoplay loop muted playsinline></video>
        <br>
          <sub><strong>Stats:</strong> 2306 colliding AABBs | 120 TPS | 50 collisions/tick</sub>
      </th>
      <th style="text-align: left; vertical-align: top; width: 400px;">
        <video src="https://github.com/user-attachments/assets/50695c5a-4f77-4352-87da-1fa13168415b" width="400" autoplay loop muted playsinline></video>
        <br>
        <sub><strong>Stats:</strong> 524 colliding AABBs | 120 TPS | 15 collisions/tick</sub>
      </th>
    </tr>
  </thead>
</table>

> Source code: [examples/ebiten-demo](examples/ebiten-demo/main.go)

<a id="example"></a>
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
	posMeta := goke.RegCompType[Pos](ecs)
	_ = goke.RegCompType[Vel](ecs)
	_ = goke.RegCompType[Acc](ecs)

	// --- Type-Safe Entity Template (Blueprint) ---
	// Blueprints place entities into the correct archetype immediately and
	// reserve memory for all components in a single batch operation.
	// Each yielded chunk exposes typed slices for direct, in-place initialization.
	blueprint := goke.NewBlueprint3[Pos, Vel, Acc](ecs)

	var entityID goke.EntityID
	for chunk := range blueprint.Create(1) {
		entityID = chunk.Entity[0]
		chunk.Comp1[0] = Pos{X: 0, Y: 0}
		chunk.Comp2[0] = Vel{X: 1, Y: 1}
		chunk.Comp3[0] = Acc{X: 0.1, Y: 0.1}
	}

	// Initialize view for Pos, Vel, and Acc components
	view := goke.NewView3[Pos, Vel, Acc](ecs)

	// Define the movement system using the functional registration pattern
	movementSystem := goke.RegSysFn(ecs, func(cb *goke.CmdBuf, d time.Duration) {
		// SoA (Structure of Arrays) layout ensures CPU cache friendliness.
		// View.All yields chunk-shaped slices over native memory — the inner
		// loop is on the caller side for aggressive compiler inlining.
		for chunk := range view.All() {
			for i := range chunk.Entity {
				pos, vel, acc := &chunk.Comp1[i], &chunk.Comp2[i], &chunk.Comp3[i]

				vel.X += acc.X
				vel.Y += acc.Y
				pos.X += vel.X
				pos.Y += vel.Y
			}
		}
	})

	// Configure the ECS's execution workflow and synchronization points
	goke.SetPlan(ecs, func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync() // Ensure all component updates are flushed and views are consistent
	})

	// Execute a single simulation step (standard 120 TPS)
	goke.Tick(ecs, time.Second/120)

	p := goke.GetComp[Pos](ecs, entityID, posMeta)
	fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
}
```

Check the [**examples/**](./examples) directory for complete, ready-to-run projects.

<a id="architecture"></a>
# Architecture

GOKe is an archetype-based ECS built around data-oriented design principles. The internal packages each own a single, well-defined responsibility:

| Package | Responsibility |
|:---|:---|
| [`github.com/kjkrol/uid`](https://pkg.go.dev/github.com/kjkrol/uid) | 64-bit generational entity identifiers — safe index recycling, ABA prevention |
| [`internal/comp`](internal/comp/doc.go) | Shared component primitives used across all internal packages — type registration, metadata, and blueprint definitions |
| [`internal/soa`](internal/soa/doc.go) | Low-level Structure-of-Arrays memory layout — L1-cache-sized fixed slabs, column offset calculation, position tracking within a growing slab collection |
| [`internal/colstore`](internal/colstore/doc.go) | Column-oriented storage for a single archetype — manages component columns over SoA slabs, resolves component IDs to memory locations in O(1) |
| [`internal/arch`](internal/arch/doc.go) | Archetype identity, archetype graph, and SoA table storage — creates archetypes on demand and caches structural transitions as graph edges |
| [`internal/entity`](internal/entity/doc.go) | Entity-to-storage mapping: `Index` maps each entity ID to its current archetype and position (`EntityLocation`) in O(1) |
| [`internal/query`](internal/query/doc.go) | Query layer: bakes component masks into precomputed per-archetype offsets, enabling zero-allocation iteration and O(1) per-entity lookup |
| [`internal/orch`](internal/orch/doc.go) | Plan-based task orchestrator: sequential/parallel execution, deferred mutations via command buffers |
| [`internal/reg`](internal/reg/doc.go) | Top-level world registry — wires together all subsystems and exposes the unified API for entity and component management |

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

* **Maximum component types: 128 by default.** The archetype system uses a fixed-size bitmask (`[2]uint64`) for fast component membership checks. Projects requiring more component types can increase this limit by modifying `MaskSize` in `internal/comp` (e.g. `MaskSize = 4` gives 256 component types) and recompiling GOKe — `MaxComponents` is derived automatically as `64 * MaskSize`. This is a compile-time configuration, not a runtime setting.

# License
GOKe is licensed under the MIT License. See the LICENSE [file](./LICENSE) for more details.

<a id="documentation"></a>
# 📖 Documentation
* **API Reference**: Detailed documentation and examples are available on [**pkg.go.dev**](https://pkg.go.dev/github.com/kjkrol/goke).
* **Wiki & Guides**: For a step-by-step deep dive into building your first simulation, check the [**Getting Started with GOKe**](https://github.com/kjkrol/goke/wiki/Getting-Started-with-GOKe) guide.
* **Internal Mechanics**: For a technical breakdown of the engine's core, check the `doc.go` files within the `ecs` packages.
