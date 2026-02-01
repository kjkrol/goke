# GOKe

<p align="center">
  <img src="assets/logo.png" alt="GOKe Logo" width="300">
  <br>
    <a href="https://go.dev">
    <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  </a>
  <a href="https://pkg.go.dev/github.com/kjkrol/goke">
    <img src="https://img.shields.io/badge/GoDoc-Reference-007d9c?style=flat-square&logo=go" alt="GoDoc">
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License">
  </a>
  <a href="https://github.com/kjkrol/goke/actions">
    <img src="https://github.com/kjkrol/goke/actions/workflows/go.yml/badge.svg" alt="Go Quality Check">
  </a>
</p>

**GOKe** is an ultra-lightweight, high-performance, and type-safe [Entity Component System](https://en.wikipedia.org/wiki/Entity_component_system) (aka ECS) for [Go](https://go.dev/). It is engineered for maximum data throughput, leveraging modern **Go 1.23+ Iterators** and a Data-Oriented Design (DOD) architecture.

<p align="center">
  <a href="#installation">Installation</a> 
  &nbsp;&bull;&nbsp; 
  <a href="#usage">Usage</a>
  &nbsp;&bull;&nbsp; 
  <a href="#architecture">Architecture</a>
  &nbsp;&bull;&nbsp; 
  <a href="#performance">Performance</a>
  &nbsp;&bull;&nbsp; 
  <a href="#features">Features</a>
  &nbsp;&bull;&nbsp;
  <a href="#roadmap">Roadmap</a>
  &nbsp;&bull;&nbsp; 
  <a href="BENCHMARKS.md">Benchmarks</a>
</p>

# 🚀 Use Cases: Why GOKe?

GOKe is not just a game engine component; it is a **high-performance data orchestrator**. It excels in scenarios where you need to manage a massive number of objects with high-frequency updates while keeping the Go Garbage Collector (GC) quiet.


## 🎮 Gaming (Ebitengine)
GOKe is a perfect companion for **Ebitengine** and other Go game frameworks. In game development, managing thousands of active objects (bullets, particles, NPCs) can quickly hit a CPU bottleneck due to pointer chasing and GC pressure. 

By using GOKe with Ebitengine:
* **Massive Sprite Batches**: You can update and filter thousands of game entities in a single tick and send them to the GPU buffer with minimal overhead.
* **Decoupled Logic**: Keep your rendering logic in Ebitengine and your game state in GOKe's optimized archetypes.
* **Deterministic Physics**: Run complex collision detection systems across all entities using `RunParallel` without worrying about state inconsistency.

## 🧬 High-Mass Simulations
If your project involves millions of "agents" (e.g., crowd simulation, epidemiological models, or particle physics), GOKe’s **Linear SoA (Structure of Arrays)** layout is essential. It ensures that data is packed tightly in memory, allowing the CPU to process entities at sub-nanosecond speeds by minimizing cache misses.

## ⚡ Latency-Critical Data Pipelines
In fields like **FinTech** or **Real-time Telemetry**, unpredictable GC pauses can be a dealbreaker. Because GOKe achieves **zero allocations in the hot path**, it provides the deterministic performance required for processing streams of data (like order matching or sensor fusion) without sudden latency spikes.

## 🤖 Complex State Management (Digital Twins)
When building digital replicas of complex systems (factories, power grids, or IoT networks), you often have thousands of sensors with overlapping sets of properties. GOKe’s **Archetype-based filtering** allows you to query and update specific subsets of these sensors with $O(1)$ or near-$O(1)$ complexity.

## 📐 Real-time Tooling & CAD
For applications that require high-speed manipulation of large geometric datasets or scene graphs, GOKe offers a way to decouple data from logic. It allows you to run heavy analytical "Systems" over your data blocks at near-metal speeds.

## ⚖️ When NOT to use GOKe
To ensure GOKe is the right tool for your project, consider these trade-offs:
* **Small Data Sets:** If you only manage a few hundred objects, a simple slice of structs will be easier to maintain and fast enough.
* **Deep Hierarchies:** ECS is designed for flat, high-speed iteration. If your data is naturally a deep tree (like a UI DOM), a classic tree structure might be more intuitive.
* **High Structural Churn:** If you are adding/removing components from thousands of entities *every single frame*, the overhead of archetype migration might offset the iteration gains.

<a id="installation"></a>
# 📦 Installation

GOKe requires **Go 1.23** or newer.

```bash
go get github.com/kjkrol/goke
```

<a id="features"></a>
# ✨ Key Features

GOKe provides a professional-grade toolkit for high-performance simulation and game development:

* **Type-Safe Generics**: Built-in support for `NewView1[A]` up to `NewViewy8[A..H]`. No type assertions or `interface{}` overhead in the hot loop.
* **Go 1.23+ Range Iterators**: Seamless integration with native `for range` loops via `iter.Seq`, allowing the compiler to perform aggressive loop inlining.
* **Command Buffer**: Thread-safe structural changes (Create/Remove/Add/Unassign) are buffered and applied during synchronization points to maintain state consistency.
* **Flexible Systems**: Supporting both stateless **Functional Systems** (closures) and **Stateful Systems** (structs) with full `Init/Update` lifecycles.
* **Advanced Execution Planning**: Deterministic scheduling with `RunParallel` support to utilize multi-core processors while maintaining control via explicit `Sync()` points.
* **Low-Level Access**: Use `AllocateComponent` for direct `*T` pointers to archetype storage or `AllocateComponentByInfo` to bypass reflection entirely.

> 💡 **See it in action**: Check the `cmd` directory for practical, ready-to-run examples, including a concurrent dice game simulation demonstrating parallel systems and state management.

<a id="usage"></a>
# 💻 Usage Example
> 📘 **New to ECS?** Check out the [**Getting Started with GOKe**](https://github.com/kjkrol/goke/wiki/Getting-Started-with-GOKe) guide for a step-by-step deep dive into building your first simulation.

```go
package main

import (
    "fmt"
    "time"

    "github.com/kjkrol/goke/pkg/ecs"
)

type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }
type Acc struct{ X, Y float32 }

func main() {
    engine := ecs.NewEngine()
    entity := engine.CreateEntity()

    // 1. Setup Components (Standard or Fast Allocation)
    ecs.SetComponent(engine, entity, Pos{X: 0, Y: 0})
    
    // Direct Access (Fastest): Returns a direct pointer to storage
    acc, _ := ecs.AllocateComponent[Acc](engine, entity)
    *acc = Acc{X: 0.1, Y: 0.1}

    // 2. Define View
    view := ecs.NewView3[Pos, Vel, Acc](engine)

    // 3. Register Functional Systems
    moveSys := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
        for head := range view.Values() {
            p, v, a := head.V1, head.V2, head.V3
            v.X += a.X; v.Y += a.Y
            p.X += v.X; p.Y += v.Y
        }
    })

    // 4. Define Execution Plan
    engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
        ctx.Run(moveSys, d)
        ctx.Sync()
    })

    // 5. Run simulation
    engine.Tick(time.Millisecond * 16)
    
    res, _ := ecs.GetComponent[Pos](engine, entity)
    fmt.Printf("Final Position: %+v\n", res)
}
```

<a id="architecture"></a>
# 🏗️ Core Architecture & "Mechanical Sympathy"
GOKe is designed with a deep understanding of modern CPU constraints. By shifting heavy computation to the initialization phase and aligning memory with hardware prefetching, the engine achieves deterministic, near-metal performance.

## Data-Oriented Memory Design
The storage layer is engineered to maximize cache hits and minimize the work of the Go Garbage Collector.

* **Archetype-Based Storage (SoA)**: Entities with the same component composition are stored in contiguous memory blocks (columns). This **Structure of Arrays** layout is L1/L2 Cache friendly, enabling hardware prefetching.
* **Generation-based Recycling**: Entities are 64-bit IDs (32-bit Index / 32-bit Generation). This solves the **ABA problem** while allowing dense packing of internal storage.
* **Archetype Masks**: Supports up to **128 unique component types** using fast, constant-time bitwise operations (2x64-bit bitsets).

## High-Throughput Access & Iteration
GOKe bypasses traditional bottlenecks like reflection and map lookups in the execution phase.

* **Flat Cache View**: Views pre-calculate direct pointers to component columns within archetypes during the initialization/warm-up phase. This **eliminates map lookups** and pointer chasing inside the hot loop.
* **Zero-Overhead Iteration**: Powered by native `for range` over functions (`iter.Seq`), allowing the Go compiler to perform aggressive loop inlining.
* **Deterministic $O(1)$ Filtering**: Querying specific entities via the **Centralized Record System** takes constant time (~462 ns) regardless of the total entity count ($N$) by bypassing hash map probing.
* **Hardware Prefetching Optimization**: View structures (Head/Tail) are strictly limited to a **maximum of 4 pointer fields**. Beyond this, CPU prefetching efficiency degrades; GOKe adheres to this limit to maintain maximal throughput.


## Execution Planning & Consistency
* **Deferred Commands**: State consistency is maintained via `Commands`. Structural changes (add/remove) are buffered and applied during explicit `Sync()` points to ensure memory safety and cache integrity.
* **Thread-Safe Concurrency**: Native support for `RunParallel` execution. GOKe provides the infrastructure for multi-core scaling, assuming the developer ensures disjoint component sets to avoid race conditions.

<a id="performance"></a>
# ⏱️ Performance & Scalability
> The engine is engineered for extreme scalability and deterministic performance. By utilizing a **Centralized Record System** (dense array lookup) instead of traditional hash maps, we have effectively decoupled both structural changes and query performance from the total entity count ($N$).

For detailed performance metrics, hardware specifications, and comparison with other ECS engines, see [BENCHMARKS.md](./BENCHMARKS.md).

## Key Performance Benchmarks
* **Deterministic $O(1)$ Filtering**: Querying specific entities takes constant time (**~460ns**) regardless of whether the world contains 1,000 or 1,000,000 entities.
* **Massive Iteration Throughput**: Thanks to SoA data locality, the engine achieves processing speeds of **~0.43 ns per entity**, operating at the limits of modern CPU cache efficiency.
* **Zero-Allocation Hot Path**: After the initial pre-allocation, all core operations (creation, removal, iteration) incur **0 heap allocations**, eliminating Go GC stuttering.
* **High-Speed Structural Changes**: Archetype migration uses direct `memmove` operations, allowing component additions in **~160ns** and tag assignments in **~86ns**.

### Reproducing Results
Run the suite on your own hardware:
```bash
go test -bench=. ./... -benchmem
```

<a id="roadmap"></a>
# 🗺️ Roadmap
Current development focus and planned improvements:

* **Batch Operations:** High-performance bulk operations for entity creation/destruction to maximize overhead reduction during large-scale processing.
* **Multi-component Operations:** Variadic Archetype Transitions allowing multiple component changes in a single atomic operation.
* **Ebitengine Integration:** Dedicated helpers for seamless state synchronization between GOKe systems and Ebitengine's loop.

> 🛠️ **Live Feature Tracker**
> We manage our long-term goals through GitHub Issues. View all planned core engine expansions and functional capabilities here:
> [**Explore all Pending Features ↗**](https://github.com/kjkrol/goke/issues?q=state%3Aopen%20label%3Afeature)

# License

GOKe is licensed under the MIT License. See the LICENSE [file](./LICENSE) for more details.

# Documentation

Detailed API documentation and examples are available on [pkg.go.dev](https://pkg.go.dev/github.com/kjkrol/goke).

For a deep dive into the internal mechanics, check the `doc.go` files within the `ecs` packages.
