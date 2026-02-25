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

GOKe is primarily a **high-performance ECS for game development**, designed to manage massive entity counts while keeping the Go Garbage Collector (GC) completely silent. However, its core architecture a **data-oriented orchestrator** — makes it suitable for any scenario requiring cache-friendly iteration over millions of objects.

## 🎮 Gaming (Ebitengine & Frameworks)

GOKe is the perfect companion for **Ebitengine** or purely server-side game loops. Managing thousands of active objects (bullets, particles, NPCs) often hits CPU bottlenecks due to pointer chasing and GC pressure. GOKe solves this via:

* **Zero-Alloc Updates**: Update thousands of entities in a single tick without triggering the GC.
* **Decoupled Logic**: Keep your rendering logic in Ebitengine and your game state in GOKe's optimized archetypes, utilizing structures like **[Bucket Grid](https://github.com/kjkrol/gokg)**.
* **Deterministic Physics**: Run complex collision detection systems across all entities using `RunParallel`.

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

## 🧬 Simulations & High-Throughput Data
Beyond gaming, GOKe shines in any domain where latency consistency is critical and object counts are in the millions.

* **Agent-Based Simulations**: Crowd dynamics, epidemiological models, or particle physics where $O(N)$ iteration speed is the bottleneck.
* **Real-time Telemetry**: Processing high-frequency data streams (e.g., IoT sensor fusion) where predictable memory access patterns prevent latency spikes.
* **Heavy Compute Pipelines**: Logic that requires transforming large datasets every frame (e.g., 16ms window) without allocation overhead.

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

Core capabilities designed for predictable performance, cache locality, and zero-allocation cycles:

* **Type-Safe Generics**: Views (`NewView1[A]` ... `NewView8`) use Go generics to eliminate interface overhead, boxing, and runtime type assertions in the hot loop.
* **Go 1.23+ Range Iterators**: Uses native `iter.Seq` for standard `for range` loops. This allows the compiler to inline iteration logic directly, avoiding callback overhead.
* **Deferred Mutations**: Structural changes (Create/Remove/Add components) are buffered via a **Command Buffer** and applied at synchronization points to ensure thread safety without heavy locking.
* **Parallel Execution**: `RunParallel` distributes system execution across available CPU cores with deterministic synchronization, scaling linearly with hardware resources.
* **Zero-Alloc Hot Loop**: The architecture guarantees zero heap allocations during the update cycle (tick), preventing GC pauses during simulation.
* **Entity Blueprints**: Fast, template-based instantiation. Allows creating thousands of entities with identical component layouts using optimal memory copy operations.

> 💡 **See it in action**: Check the `cmd` directory for the concurrent dice game simulation demonstrating parallel systems and state management.

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
	posDesc := goke.RegisterComponent[Pos](ecs)
	_ = goke.RegisterComponent[Vel](ecs)
	_ = goke.RegisterComponent[Acc](ecs)

	// --- Type-Safe Entity Template (Blueprint) ---
	// Blueprints place the entity into the correct archetype immediately and
	// reserve memory for all components in a single atomic operation.
	// This returns typed pointers for direct, in-place initialization.
	blueprint := goke.NewBlueprint3[Pos, Vel, Acc](ecs)

	// Create the entity and get direct access to its memory slots.
	entity, pos, vel, acc := blueprint.Create()
	*pos = Pos{X: 0, Y: 0}
	*vel = Vel{X: 1, Y: 1}
	*acc = Acc{X: 0.1, Y: 0.1}

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

	p, _ := goke.GetComponent[Pos](ecs, entity, posDesc)
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

GOKe is an archetype-based ECS designed for deterministic performance. It shifts structural overhead (like offset calculations) to the initialization phase and uses a chunked SoA layout to maintain consistent throughput regardless of scale.

## Data-Oriented Memory Design
The storage layer is engineered to maximize cache hits and eliminate allocation spikes ("GC jitter").

* **Chunked SoA (Structure of Arrays)**: Instead of monolithic slices that require costly resizing (copying millions of elements) when capacity is exceeded, GOKe manages data in **fixed-size Memory Pages** (aligned to L1 Cache, e.g., 96KB).
    * **Stable Growth**: Memory allocation is linear and "stepless". Adding the 1,000,001st entity simply allocates one small memory chunk, avoiding the massive latency spike of doubling a large array.
    * **Cache Locality**: Inside each chunk, components are packed in a tight SoA layout (`[IDs...][CompA...][CompB...]`), ensuring high-efficiency hardware prefetching.
* **Generation-based Recycling**: Entities are tracked via 64-bit IDs (32-bit Index / 32-bit Generation). This prevents **entity aliasing**—stale references to destroyed entities are instantly recognized as invalid when their memory slot is reused.
* **Archetype Masks**: Supports rapid composition checks using fast, constant-time bitwise operations. This allows for complex queries over component types without iterating over unrelated data.

## High-Throughput Access & Iteration
GOKe bypasses traditional bottlenecks like reflection and map lookups in the execution phase.

* **Flat Cache View**: Views pre-calculate direct pointers to component columns within active chunks. This **eliminates map lookups** and pointer chasing inside the hot loop.
* **Zero-Overhead Iteration**: Powered by native `for range` over functions (`iter.Seq`), allowing the Go compiler to perform aggressive loop inlining directly over the memory pages.
* **Deterministic $O(1)$ Filter**: Querying specific entities via the **Centralized Record System** takes constant time regardless of the total entity count ($N$) by mapping IDs directly to `(ChunkIndex, RowIndex)` coordinates.
* **Hardware Prefetching Optimization**: View structures are optimized to keep the prefetcher strictly focused on the data stream, minimizing cache pollution during iteration.


## Execution Planning & Consistency
* **Deferred Commands**: State consistency is maintained via `Commands`. Structural changes (add/remove) are buffered and applied during explicit `Sync()` points to ensure memory safety and cache integrity.
* **Thread-Safe Concurrency**: Native support for `RunParallel` execution. GOKe provides the infrastructure for multi-core scaling, assuming the developer ensures disjoint component sets to avoid race conditions.

<a id="performance"></a>
# ⏱️ Performance & Scalability
> The engine is engineered for extreme scalability and deterministic performance. By utilizing a **Centralized Record System** (dense array lookup) instead of traditional hash maps, we have effectively decoupled both structural changes and query performance from the total entity count ($N$).

GOKe delivers near-metal speeds by eliminating heap allocations and leveraging L1/L2 cache locality.

| Category | Operation | Performance (1k Baseline) | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Throughput** | **Iteration** | **0.36 – 0.66 ns/ent** | **0** | Linear SoA (1-8 components) |
| **Scalability** | **$O(1)$ Filter** | **1.39 – 5.38 ns/ent** | **0** | Centralized Record Lookup |
| **Structural** | **Create Entity** | **21.31 - 26.84 ns/op** | **0** | Blueprint-based (1-4 comps) |
| **Structural** | **Migrate Component** | **37.36 ns/op** | **0** | Archetype Move (Insert) |
| **Structural** | **Remove Entity** | **17.95 ns/op** | **0** | Index Recycling |
| **Access** | **Get Component** | **4.49 ns/op** | **0** | Direct Generation Check |

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
