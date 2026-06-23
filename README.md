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
- Inlining-friendly hot paths
- Type-safe APIs without reflection
- Native Go development without CGO dependencies

While native C and Rust ECS frameworks may achieve higher peak throughput,
GOKe is designed to maximize performance within the Go ecosystem. For many
projects, **avoiding CGO boundaries**, external dependencies, and cross-language
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
| **Predictable iteration cost** | Per-entity overhead stays constant regardless of how much logic runs in the loop body |
| **O(1) component lookup** | Entity-to-storage is a direct array index, not a hash map — constant time at any world size |
| **Safe entity recycling** | 64-bit generational IDs detect stale references after deletion, preventing ABA bugs |
| **Cache-friendly storage** | Contiguous SoA chunks; growth appends new chunks, removal uses swap-and-pop — no fragmentation |
| **Batch entity creation** | `Factory` writes components directly into chunk-shaped batches — no per-entity allocation |
| **Type-safe component API** | Fully generic — no reflection, no interface boxing, no runtime type assertions |
| **Built-in scheduler** | Declarative `Plan` wires systems into an execution graph — a full ECS runtime, not just a component store |
| **Command Buffer** | Structural changes during iteration are queued and flushed at explicit `Sync()` points — enables safe `RunParallel` |

> 💡 **See the Performance & Scalability section below for benchmark results validated from 2¹⁰ to 2²⁰ entities.**

<a id="performance"></a>
# ⏱️ Performance & Scalability
GOKe is designed for predictable performance at scale. By utilizing a **Centralized Record System** (dense array lookup) instead of traditional hash maps, structural operations and query execution remain effectively independent of the total entity count ($N$).

## 📊 Cross-framework comparison
Benchmarks against other Go ECS libraries (Arche, Donburi, Ento, etc.) are maintained in a dedicated project — [**go-ecs-benchmarks**](https://github.com/mlange-42/go-ecs-benchmarks) by [@mlange-42](https://github.com/mlange-42). 

⚠️ Before drawing conclusions, verify which GOKe version (tag) is used in the comparison, as published results may lag behind the latest release.

## Scalability Validation
`Editor`/`Factory.Create`/`Matcher.All` were validated across worlds ranging from **2¹⁰ (1,024)** to **2²⁰ (1,048,576)** entities on an **Apple M1 Max**. Structural per-entity cost stays nearly constant across that range; `Matcher.All`'s per-entity cost roughly doubles at 2²⁰ once the working set outgrows cache (a known trade-off of the chunked SoA layout). `Pick`/`Seek`/`Remove` are currently only benchmarked at a fixed population.

| Category | Operation | Observed Cost (2¹⁰–2²⁰ Entities) | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Throughput** | **Iteration (Matcher.All)** | **0.32 - 3.41 ns/ent** | **0** | Linear SoA (0-10 components) |
| **Subset Query** | **Pick (per-entity, 1,024 only)** | **3.6 - 11.2 ns/ent** | **0** | Per-entity record lookup + pointer math |
| **Direct Access** | **Seek (single entity, 1,024 only)** | **2.9 - 10.7 ns/ent** | **0** | Index lookup, independent of include/exclude mask |
| **Structural** | **Batch Create** | **3.9 - 9.0 ns/ent** | 0-1 | Factory-based chunk writes |
| **Structural** | **Add Component** | **32 - 90 ns/op** | 0 at 2¹⁰; grows with N at 2²⁰ | Archetype migration (1 → 1+N components) |
| **Structural** | **Add Tag** | **30 - 71 ns/op** | 0 at 2¹⁰; grows with N at 2²⁰ | Archetype migration (zero-size component) |
| **Structural** | **Remove Component** | **83 - 110 ns/op** | 0 at 2¹⁰; grows with N at 2²⁰ | Archetype migration (10 → 10-N components) |
| **Structural** | **Remove Entity** | **3.2 ns/op** (population 100,000) | **0** | Swap-and-pop + index recycling |

> **Deep Dive**: For the full per-component-count breakdown, methodology, and reproduction instructions, see [**BENCHMARKS.md**](./BENCHMARKS.md).

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
	"github.com/kjkrol/uid"
)

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }

func main() {
	// Initialize the ECS world.
	ecs := goke.New()

	// Register component types — each Go type is assigned a stable CompID.
	_ = goke.RegComp[Pos](ecs)
	_ = goke.RegComp[Vel](ecs)
	_ = goke.RegComp[Acc](ecs)

	// Col[T] handles typed access to a component column.
	// The same Col[T] can be reused across factory and matcher.
	var pos goke.Col[Pos]
	var vel goke.Col[Vel]
	var acc goke.Col[Acc]

	// Create a factory for bulk entity spawning.
	factory := ecs.CreateFactory(goke.Add(&pos), goke.Add(&vel), goke.Add(&acc))

	var entityID uid.UID64
	factory.Create(1)
	factory.Next()
	entityID = factory.IDs[0]
	fc := &factory.Cursor
	pos.Slice(fc)[0] = Pos{X: 0, Y: 0}
	vel.Slice(fc)[0] = Vel{X: 1, Y: 1}
	acc.Slice(fc)[0] = Acc{X: 0.1, Y: 0.1}

	// Create a matcher — declares which components to iterate.
	query := ecs.CreateMatcher(goke.Track(&pos), goke.Track(&vel), goke.Track(&acc))

	// Register a system using the functional pattern.
	cursor := &query.Cursor
	movementSystem := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		// SoA layout: Matcher.All advances chunk by chunk — the inner loop
		// iterates over contiguous memory for cache-friendly access.
		query.All()
		for query.Next() {
			posSlice := pos.Slice(cursor)
			velSlice := vel.Slice(cursor)
			accSlice := acc.Slice(cursor)
			for i := range cursor.IDs {
				velSlice[i].X += accSlice[i].X
				velSlice[i].Y += accSlice[i].Y
				posSlice[i].X += velSlice[i].X
				posSlice[i].Y += velSlice[i].Y
			}
		}
	})

	// Configure the execution plan and synchronization points.
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(movementSystem, d)
		ctx.Sync()
	})

	// Execute a single simulation step (120 TPS).
	ecs.Tick(time.Second / 120)

	// Read a single entity's component via Seek (cursor-based, typed).
	matcher := ecs.CreateMatcher(goke.Track(&pos))
	if matcher.Seek(entityID) {
		p := pos.At(&matcher.Cursor)
		fmt.Printf("Final Position: {X: %.2f, Y: %.2f}\n", p.X, p.Y)
	}
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
| [`internal/chunk`](internal/chunk/doc.go) | Cache-aligned chunked memory layout — L1-cache-sized fixed slabs, field offset calculation, slot tracking within a growing slab collection; keeps one spare slab on shrink so repeated grow/shrink cycles stay allocation-free |
| [`internal/colstore`](internal/colstore/doc.go) | Column-oriented storage for a single archetype — manages component columns over `chunk.Pack` chunks, resolves component IDs to memory locations in O(1) |
| [`internal/arch`](internal/arch/doc.go) | Archetype identity, archetype graph, and SoA table storage — creates archetypes on demand and caches structural transitions as graph edges |
| [`internal/addr`](internal/addr/doc.go) | Entity address book — manages entity ID lifecycle (uid pool) and maps each ID to its current storage address (`Entry`) via a flat index in O(1); generation check guards against stale references |
| [`internal/ent`](internal/ent/doc.go) | Entity lifecycle — delegates ID allocation and address tracking to `addr.Book`, manages component migration (add/remove moves entity to a new archetype), and batch entity creation via `Factory` |
| [`internal/query`](internal/query/doc.go) | Query layer: `Matcher` bakes component masks into precomputed per-archetype offsets, enabling zero-allocation bulk iteration (`All`), per-entity subset iteration (`Pick`), and O(1) single-entity access (`Seek`) |
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
* **Internal Mechanics**: For a technical breakdown of the engine's core, check the `doc.go` files within the `internal` packages.
