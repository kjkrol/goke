# GOKe

<p align="center">
  <img src=".github/docs/img/logo.png" alt="GOKe Logo" width="300">
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
  &nbsp;&bull;&nbsp; 
  <a href="#documentation">Documentation</a>
</p>

# 🚀 Use Cases: Why GOKe?

GOKe is not just a game engine component; it is a **high-performance data orchestrator**. It excels in scenarios where you need to manage a massive number of objects with high-frequency updates while keeping the Go Garbage Collector (GC) quiet.

## 🎮 Gaming (Ebitengine)

GOKe is a perfect companion for **Ebitengine** and other Go game frameworks. In game development, managing thousands of active objects (bullets, particles, NPCs) can quickly hit a CPU bottleneck due to pointer chasing and GC pressure.

By using GOKe with Ebitengine:
* **Massive Sprite Batches**: You can update and filter thousands of game entities in a single tick and send them to the GPU buffer with minimal overhead.
* **Decoupled Logic**: Keep your rendering logic in Ebitengine and your game state in GOKe's optimized archetypes, utilizing structures like **[Bucket Grid](https://github.com/kjkrol/gokg)**.
* **Deterministic Physics**: Run complex collision detection systems across all entities using `RunParallel`.

<table align="center">
  <thead>
    <tr>
      <th style="text-align: left; vertical-align: top; width: 400px;">
        <video src="https://github.com/user-attachments/assets/fa8d1aca-2060-466d-8204-9d6a7443d580" width="400" autoplay loop muted playsinline></video>
        <br>
        <p align="center">
          <sub><strong>Stats:</strong> 2048 colliding AABBs | 120 TPS | 0.1-1 collisions/tick</sub>
        </p>
      </th>
      <th style="text-align: left; vertical-align: top; width: 400px;">
        <video src="https://github.com/user-attachments/assets/f1ef2c0b-fb7b-478a-bc88-77faa48c0623" width="400" autoplay loop muted playsinline></video>
        <br>
        <sub>
          <p align="center">
            <strong>Stats:</strong> 64 colliding AABBs | 120 TPS | 0.1-1 collisions/tick
          </p>
        </sub>
      </th>
    </tr>
  </thead>
<tbody>
    <tr>
      <td colspan="2" align="center">
        <sub>
          <strong>Check out the <a href="examples/ebiten-demo/main.go">full source code</a></strong>
        </sub>
      </td>
    </tr>
  </tbody>
</table>

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
  posType := goke.RegisterComponentType[Pos](ecs)
  velType := goke.RegisterComponentType[Vel](ecs)
  accType := goke.RegisterComponentType[Acc](ecs)

  entity := goke.CreateEntity(ecs)

  // --- Component Assignment with In-Place Memory Access ---
  // EnsureComponent acts as a high-performance Upsert. By returning a direct
  // pointer to the component's slot in the archetype storage, it allows for
  // in-place modification (*ptr = T{...}).
  *goke.EnsureComponent[Pos](ecs, entity, posType) = Pos{X: 0, Y: 0}
  *goke.EnsureComponent[Vel](ecs, entity, velType) = Vel{X: 1, Y: 1}
  *goke.EnsureComponent[Acc](ecs, entity, accType) = Acc{X: 0.1, Y: 0.1}

  // Initialize view for Pos, Vel, and Acc components
  view := goke.NewView3[Pos, Vel, Acc](ecs)

  // Define the movement system using the functional registration pattern
  movementSystem := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
    // SoA (Structure of Arrays) layout ensures CPU Cache friendliness.
    for head := range view.Values() {
      pos, vel, acc := head.V1, head.V2, head.V3

      vel.X += acc.X
      vel.Y += acc.Y
      pos.X += vel.X
      pos.Y += vel.Y
    }
  })

  // Configure the ECS's execution workflow and synchronization points
  goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
    ctx.Run(movementSystem, d)
    ctx.Sync() // Ensure all component updates are flushed and views are consistent
  })

  // Execute a single simulation step (standard 120 TPS)
  goke.Tick(ecs, time.Second/120)

  p, _ := goke.GetComponent[Pos](ecs, entity, posType)
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
# 🏗️ Core Architecture & "Mechanical Sympathy"
GOKe is designed with a deep understanding of modern CPU constraints. By shifting heavy computation to the initialization phase and aligning memory with hardware prefetching, the engine achieves deterministic, near-metal performance.

## Data-Oriented Memory Design
The storage layer is engineered to maximize cache hits and minimize the work of the Go Garbage Collector.

* **Archetype-Based Storage (SoA)**: Entities with the same component composition are stored in contiguous memory blocks (columns). This **Structure of Arrays** layout is L1/L2 Cache friendly, enabling hardware prefetching.
* **Generation-based Recycling**: Entities are 64-bit IDs (32-bit Index / 32-bit Generation). This prevents **entity aliasing** by ensuring that any stale references (handles) to a destroyed entity are correctly identified as invalid when the index is reused for a new entity.
* **Archetype Masks**: Supports up to **128 unique component types** by default using fast, constant-time bitwise operations (2x64-bit bitsets). The mask size—and thus the component limit—can be increased at compile-time via build flags to suit larger projects.

## High-Throughput Access & Iteration
GOKe bypasses traditional bottlenecks like reflection and map lookups in the execution phase.

* **Flat Cache View**: Views pre-calculate direct pointers to component columns within archetypes during the initialization/warm-up phase. This **eliminates map lookups** and pointer chasing inside the hot loop.
* **Zero-Overhead Iteration**: Powered by native `for range` over functions (`iter.Seq`), allowing the Go compiler to perform aggressive loop inlining.
* **Deterministic $O(1)$ Filtering**: Querying specific entities via the **Centralized Record System** takes constant time regardless of the total entity count ($N$) by bypassing hash map probing and utilizing direct index-to-record mapping.
* **Hardware Prefetching Optimization**: View structures (Head/Tail) are strictly limited to a **maximum of 4 pointer fields**. Beyond this, CPU prefetching efficiency degrades; GOKe adheres to this limit to maintain maximal throughput.


## Execution Planning & Consistency
* **Deferred Commands**: State consistency is maintained via `Commands`. Structural changes (add/remove) are buffered and applied during explicit `Sync()` points to ensure memory safety and cache integrity.
* **Thread-Safe Concurrency**: Native support for `RunParallel` execution. GOKe provides the infrastructure for multi-core scaling, assuming the developer ensures disjoint component sets to avoid race conditions.

<a id="performance"></a>
# ⏱️ Performance & Scalability
> The engine is engineered for extreme scalability and deterministic performance. By utilizing a **Centralized Record System** (dense array lookup) instead of traditional hash maps, we have effectively decoupled both structural changes and query performance from the total entity count ($N$).

GOKe delivers near-metal speeds by eliminating heap allocations and leveraging L1/L2 cache locality.

| Category | Operation | Performance | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Throughput** | **View Iteration** | **0.48 – 0.98 ns/ent** | **0** | Linear SoA (1-8 components) |
| **Scalability** | **$O(1)$ Filter** | **4.48 ns/ent** | **0** | Centralized Record Lookup |
| **Structural** | **Add Component** | **28.30 ns/op** | **0** | Archetype Migration (Insert) |
| **Structural** | **Migrate Component** | **43.27 ns/op** | **0** | Archetype Move + Insert |
| **Structural** | **Remove Entity** | **14.44 ns/op** | **0** | Index Recycling |
| **Access** | **Get Component** | **3.46 ns/op** | **0** | Direct Generation Check |


> 📊 **Deep Dive**: For a full breakdown of hardware specs, stress tests, and $O(N)$ vs $O(1)$ scaling charts, see [**BENCHMARKS.md**](./BENCHMARKS.md).

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

<a id="documentation"></a>
# 📖 Documentation
* **API Reference**: Detailed documentation and examples are available on [**pkg.go.dev**](https://pkg.go.dev/github.com/kjkrol/goke).
* **Wiki & Guides**: For a step-by-step deep dive into building your first simulation, check the [**Getting Started with GOKe**](https://github.com/kjkrol/goke/wiki/Getting-Started-with-GOKe) guide.
* **Internal Mechanics**: For a technical breakdown of the engine's core, check the `doc.go` files within the `ecs` packages.
