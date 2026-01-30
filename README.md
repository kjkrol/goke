<p align="center">
  <img src="assets/logo.png" alt="GOKe Logo" width="300">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/kjkrol/goke"><img src="https://img.shields.io/badge/GoDoc-Reference-007d9c?style=flat-square&logo=go" alt="GoDoc"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License"></a>
</p>

# GOKe

*aka "Golang kjkrol ECS"*

**GOKe** is an ultra-lightweight, high-performance, and type-safe Entity Component System (ECS) for Go. It is engineered for maximum data throughput, leveraging modern **Go 1.23+ Iterators** and a Data-Oriented Design (DOD) architecture.

## Why GOKe is Blazing Fast

Unlike many ECS implementations in Go that rely on component maps or reflection during the update loop, GOKe shifts the heavy lifting to the initialization phase using:

* **Archetype-Based Storage (SoA)**: Entities with the same component composition are stored in contiguous memory blocks (columns). This layout is **L1/L2 Cache friendly**, allowing the CPU to utilize hardware prefetching.
* **Flat Cache View**: Views pre-calculate direct pointers to component columns within archetypes. This eliminates map lookups and pointer chasing inside the hot loop.
* **Zero-Overhead Iteration**: Powered by native `for range` over functions (`iter.Seq`), enabling the Go compiler to perform aggressive loop inlining.
* **Deterministic $O(1)$ Filtering**: Querying 100 specific entities takes the same time (~462 ns) whether the registry contains 1,000 or 100,000 entities by bypassing hash map probing.

## Key Features

GOKe provides a professional-grade toolkit for high-performance simulation and game development:

* **Type-Safe Generics**: Built-in support for `NewView1[A]` up to `NewViewy8[A..H]`. No type assertions or `interface{}` overhead in the hot loop.
* **Go 1.23+ Range Iterators**: Seamless integration with native `for range` loops via `iter.Seq`, allowing the compiler to perform aggressive loop inlining.
* **Command Buffer**: Thread-safe structural changes (Create/Remove/Add/Unassign) are buffered and applied during synchronization points to maintain state consistency.
* **Flexible Systems**: Supporting both stateless **Functional Systems** (closures) and **Stateful Systems** (structs) with full `Init/Update` lifecycles.
* **Advanced Execution Planning**: Deterministic scheduling with `RunParallel` support to utilize multi-core processors while maintaining control via explicit `Sync()` points.
* **Low-Level Access**: Use `AllocateComponent` for direct `*T` pointers to archetype storage or `AllocateComponentByInfo` to bypass reflection entirely.

> 💡 **See it in action**: Check the `cmd` directory for practical, ready-to-run examples, including a concurrent dice game simulation demonstrating parallel systems and state management.


---

## 🚀 Use Cases: Why GOKe?

GOKe is not just a game engine component; it is a **high-performance data orchestrator**. It excels in scenarios where you need to manage a massive number of objects with high-frequency updates while keeping the Go Garbage Collector (GC) quiet.

### 🧬 High-Mass Simulations
If your project involves millions of "agents" (e.g., crowd simulation, epidemiological models, or particle physics), GOKe’s **Linear SoA (Structure of Arrays)** layout is essential. It ensures that data is packed tightly in memory, allowing the CPU to process entities at sub-nanosecond speeds by minimizing cache misses.



### ⚡ Latency-Critical Data Pipelines
In fields like **FinTech** or **Real-time Telemetry**, unpredictable GC pauses can be a dealbreaker. Because GOKe achieves **zero allocations in the hot path**, it provides the deterministic performance required for processing streams of data (like order matching or sensor fusion) without sudden latency spikes.

### 🤖 Complex State Management (Digital Twins)
When building digital replicas of complex systems (factories, power grids, or IoT networks), you often have thousands of sensors with overlapping sets of properties. GOKe’s **Archetype-based filtering** allows you to query and update specific subsets of these sensors with $O(1)$ or near-$O(1)$ complexity.



### 📐 Real-time Tooling & CAD
For applications that require high-speed manipulation of large geometric datasets or scene graphs, GOKe offers a way to decouple data from logic. It allows you to run heavy analytical "Systems" over your data blocks at near-metal speeds.

---

### ⚖️ When NOT to use GOKe (Transparency)
To ensure GOKe is the right tool for your project, consider these trade-offs:
* **Small Data Sets:** If you only manage a few hundred objects, a simple slice of structs will be easier to maintain and fast enough.
* **Deep Hierarchies:** ECS is designed for flat, high-speed iteration. If your data is naturally a deep tree (like a UI DOM), a classic tree structure might be more intuitive.
* **High Structural Churn:** If you are adding/removing components from thousands of entities *every single frame*, the overhead of archetype migration might offset the iteration gains.

---

## Core Architecture & "Mechanical Sympathy"

GOKe is designed with an understanding of modern CPU constraints:

### Memory Layout
1.  **Generation-based Recycling**: Entities are 64-bit IDs (32-bit Index / 32-bit Generation). This solves the ABA problem while allowing dense packing of internal storage.
2.  **Hardware Prefetching Optimization**: All view structures (Head/Tail) are strictly limited to a maximum of 4 pointer fields. Beyond this, CPU prefetching efficiency degrades. GOKe adheres to this limit to maintain maximal throughput.
3.  **Archetype Masks**: Supports up to **128 unique component types** using a fast, constant-time bitwise operation (2x64-bit fields).

### Execution Plan
* **Deferred Commands**: State consistency is maintained via `Commands`. Changes (add/remove) are buffered and applied during explicit `Sync()` points.
* **Power to the Programmer**: Support for `RunParallel` execution. GOKe provides the tools for concurrency, assuming the developer ensures disjoint component sets to avoid races.

## Installation

GOKe requires **Go 1.23** or newer.

## Usage Example

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

## Performance & Scalability

The engine is engineered for extreme scalability and deterministic performance. By utilizing a **Centralized Record System** (dense array lookup) instead of traditional hash maps, we have effectively decoupled both structural changes and query performance from the total entity count ($N$).

### Structural Operations (Apple M1 Max)
These benchmarks highlight the efficiency of our archetype-based memory management. Every operation below results in **0 heap allocations**, ensuring no GC pressure during the simulation loop.

| Operation | Performance | Memory | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Create Entity** | **22.95 ns/op** | 69 B/op | **0** | Pre-allocated archetype slotting |
| **Add First Component** | **28.30 ns/op** | 69 B/op | **0** | Archetype migration (+ no data move + data insert) |
| **Add Next Component** | **43.27 ns/op** | 115 B/op | **0** | Archetype migration (+ data move + data insert) |
| **Add Tag** | **21.52 ns/op** | 16 B/op | **0** | Archetype migration (+ data move + no data insert) |
| **Remove Component** | **15.82 ns/op** | 0 B/op | **0** | Archetype migration (Swap-and-pop) |
| **Remove Entity** | **14.44 ns/op** | 0 B/op | **0** | Index recycling & record invalidation |
| **Structural Stability** | **47.80 ns/op** | 64 B/op | **0** | Stress test of add/remove cycles |
| **Get Entity Component** | **14.44 ns/op** | 0 B/op | **0** | Retrieves a copy of the component data using entity index and generation validation |

### Key Technical Takeaways

* **Zero-Allocation Architecture:** All core operations (adding/removing components, creating entities) are performed with zero heap allocations after the initial pre-allocation phase. This eliminates stuttering caused by the Go Garbage Collector.
* **Archetype Migration Efficiency:** Even when structural changes require moving data between memory blocks, the operation is extremely fast (~160ns for components, ~86ns for tags). We use direct memory block copies (`memmove`), bypassing the overhead of reflection or interfaces.
* **Deterministic O(1) Filtering:** Querying specific entities takes constant time (~460ns) regardless of whether the world contains 1,000 or 100,000 entities, thanks to our dense array Record system.
* **Data Locality:** By storing components in contiguous SoA (Structure of Arrays) layouts, the engine achieves processing speeds of **~0.43 ns per entity**, operating at the limits of modern CPU cache efficiency.

### Query Benchmarks (Apple M1 Max)
The following benchmarks demonstrate the efficiency of SoA (Structure of Arrays) and the $O(1)$ nature of our record-based filtering.

### Query Performance & Scalability (Apple M1 Max)

The engine demonstrates perfect linear scaling for full world iterations and deterministic constant-time performance for filtered lookups across four orders of magnitude.

#### 1. Scalability Overview: $O(N)$ vs $O(1)$
| Registry Size ($N$) | Operation | Entities Processed | Total Time | Per Entity | Mechanism |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1,000** | **View3.All** | 1,000 | 822.0 ns | 0.82 ns | Linear SoA |
| **10,000** | **View3.All** | 10,000 | 10,306 ns | 1.03 ns | Linear SoA |
| **100,000** | **View3.All** | 100,000 | 97,686 ns | 0.97 ns | Linear SoA |
| **1,000,000** | **View3.All** | 1,000,000 | 0.98 ms | 0.98 ns | Linear SoA |
| **10,000,000** | **View3.All** | 10,000,000 | 9.85 ms | **0.98 ns** | Linear SoA |
| **10,000,000** | **View3.Filter** | **100** | **448.3 ns** | **4.48 ns** | **O(1) Record Lookup** |

#### 2. Complexity Scaling (10M Entities Stress Test)
| View Complexity | All Iterator | Values Iterator (No ID) | Performance Gain |
| :--- | :--- | :--- | :--- |
| **View0 (Entity Only)** | 4.19 ms | - | - |
| **View1 (1 Comp)** | 6.22 ms | 4.42 ms | **+29.0%** |
| **View3 (3 Comps)** | 9.85 ms | 8.02 ms | **+18.5%** |
| **View8 (8 Comps)** | 10.08 ms | 8.14 ms | **+19.2%** |

#### 3. Full Benchmark Logs
<details>
<summary>Click to view all scenarios (View 0-8, Values, Filtered)</summary>

| Registry Size ($N$) | Operation | Total Time | Per Entity | Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **1,000** | View0.All | 435.9 ns | 0.43 ns | Linear SoA |
| **1,000** | View3.All | 822.0 ns | 0.82 ns | Linear SoA |
| **1,000** | View8.All | 809.4 ns | 0.80 ns | Linear SoA |
| **1,000** | View3.Filter 100 | 432.1 ns | 4.32 ns | O(1) Lookup |
| **1,000** | View3.Values | 644.6 ns | 0.64 ns | Values Scan |
| **10,000** | View0.All | 4,208 ns | 0.42 ns | Linear SoA |
| **10,000** | View3.All | 10,306 ns | 1.03 ns | Linear SoA |
| **10,000** | View8.All | 9,920 ns | 0.99 ns | Linear SoA |
| **10,000** | View3.Filter 100 | 438.0 ns | 4.38 ns | O(1) Lookup |
| **10,000** | View3.Values | 7,673 ns | 0.76 ns | Values Scan |
| **100,000** | View0.All | 41,813 ns | 0.41 ns | Linear SoA |
| **100,000** | View3.All | 97,686 ns | 0.97 ns | Linear SoA |
| **100,000** | View8.All | 97,582 ns | 0.97 ns | Linear SoA |
| **100,000** | View3.Filter 100 | 454.9 ns | 4.54 ns | O(1) Lookup |
| **100,000** | View3.Values | 77,924 ns | 0.77 ns | Values Scan |
| **1,000,000** | View0.All | 415,515 ns | 0.41 ns | Linear SoA |
| **1,000,000** | View3.All | 989,272 ns | 0.98 ns | Linear SoA |
| **1,000,000** | View8.All | 992,372 ns | 0.99 ns | Linear SoA |
| **1,000,000** | View3.Filter 100 | 444.9 ns | 4.44 ns | O(1) Lookup |
| **1,000,000** | View3.Values | 797,392 ns | 0.79 ns | Values Scan |
| **10,000,000** | View0.All | 4,195,179 ns | 0.41 ns | Linear SoA |
| **10,000,000** | View1.All | 6,229,583 ns | 0.62 ns | Linear SoA |
| **10,000,000** | View3.All | 9,857,888 ns | 0.98 ns | Linear SoA |
| **10,000,000** | View8.All | 10,089,932 ns | 1.00 ns | Linear SoA |
| **10,000,000** | View3.Filter 100 | 448.3 ns | 4.48 ns | O(1) Lookup |
| **10,000,000** | View1.Values | 4,424,216 ns | 0.44 ns | Values Scan |
| **10,000,000** | View3.Values | 8,029,977 ns | 0.80 ns | Values Scan |
| **10,000,000** | View8.Values | 8,140,268 ns | 0.81 ns | Values Scan |
| **10,000,000** | View3 FilterValues 100 | 432.1 ns | 4.32 ns | Values Lookup |

</details>

### Key Technical Takeaways
* **Near-Zero Latency:** Targeted queries (Filter) remain at **~440 ns** regardless of whether the world has 1k or 10M entities.
* **Instruction Efficiency:** Even with 8 components, the engine processes entities at **~1 ns/entity**, fitting a 10M entity update within a **10ms** window.
* **Values Iteration Gain:** Bypassing Entity ID generation provides up to **29%** additional throughput for heavy computational systems.


## Benchmark Comparison: GOKe vs. Arche (RAW Data)

It is important to note that while **GOKe** achieves industry-leading raw iteration speeds, **Arche** is a more established framework providing a broader feature set. The performance trade-offs in **Arche** often stem from supporting complex functionalities that **GOKe** intentionally omits to maintain its lean profile, such as:

* **Entity Relations:** Native support for parent-child hierarchies and linked entities.
* **Batch Operations:** Highly optimized mass entity spawning and destruction.
* **Event Listeners:** Comprehensive system for monitoring entity and component lifecycles (in **goke/ecs**, this includes Cached Views that listen for newly created archetypes to dynamically attach them to the view).

## Benchmark Comparison: GOKe vs. Arche
**Environment:** Apple M1 Max (ARM64)  
**Units:** Nanoseconds per operation (ns/op)

| Operation | GOKe | Arche | Winner | Notes |
| :--- | :--- | :--- | :--- | :--- |
| **Iteration (1 Comp)** | **0.41 ns** | 0.55 ns | **GOKe (+25%)** | Superior cache locality |
| **Iteration (2 Comp)** | **0.49 ns** | 1.37 ns | **GOKe (+64%)** | Efficient memory layout |
| **Iteration (3 Comp)** | **0.65 ns** | 1.79 ns | **GOKe (+63%)** | Minimal per-entity overhead |
| **Create Entity** | 22.95 ns | **20.60 ns** | **Arche (+10%)** | Slight edge in indexing |
| **Add First Component** | **28.30 ns** | 29.30 ns | **GOKe (+3%)** | Optimal transition |
| **Add Next Component** | **43.27 ns** | **--** | **Arche** | Fast graph traversal |
| **Add Tag** | **21.52 ns** | **--** | **Arche** | No data move overhead |
| **Remove Component** | **15.81 ns** | **--** | **Arche** | Accelerated by edgesPrev |
| **Remove Entity** | **14.44 ns** | **--** | **Arche** | Lightning fast cleanup |
| **View Filter** | **4.48 ns** | **--** | **Arche** | 128-bit mask bitwise ops |

## Roadmap

* **Batch Operations:** Implementation of high-performance bulk operations for entity creation and destruction, aimed at maximizing overhead reduction during large-scale data processing.

## License

GOKe is licensed under the MIT License. See the LICENSE file for more details.

## Documentation

Detailed API documentation and examples are available on [pkg.go.dev](https://pkg.go.dev/github.com/kjkrol/goke).

For a deep dive into the internal mechanics, check the `doc.go` files within the `ecs` packages.
