<p align="center">
  <img src="assets/logo.png" alt="GOKE Logo" width="300">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/kjkrol/goke"><img src="https://img.shields.io/badge/GoDoc-Reference-007d9c?style=flat-square&logo=go" alt="GoDoc"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License"></a>
</p>

# GOKE

*aka "Golang kjkrol ECS"*

**GOKE** is an ultra-lightweight, high-performance, and type-safe Entity Component System (ECS) for Go. It is engineered for maximum data throughput, leveraging modern **Go 1.23+ Iterators** and a Data-Oriented Design (DOD) architecture.

## Why GOKE is Blazing Fast

Unlike many ECS implementations in Go that rely on component maps or reflection during the update loop, GOKE shifts the heavy lifting to the initialization phase using:

* **Archetype-Based Storage (SoA)**: Entities with the same component composition are stored in contiguous memory blocks (columns). This layout is **L1/L2 Cache friendly**, allowing the CPU to utilize hardware prefetching.
* **Flat Cache View**: Views pre-calculate direct pointers to component columns within archetypes. This eliminates map lookups and pointer chasing inside the hot loop.
* **Zero-Overhead Iteration**: Powered by native `for range` over functions (`iter.Seq`), enabling the Go compiler to perform aggressive loop inlining.
* **Deterministic $O(1)$ Filtering**: Querying 100 specific entities takes the same time (~462 ns) whether the registry contains 1,000 or 100,000 entities by bypassing hash map probing.

## Key Features

GOKE provides a professional-grade toolkit for high-performance simulation and game development:

* **Type-Safe Generics**: Built-in support for `NewView1[A]` up to `NewViewy8[A..H]`. No type assertions or `interface{}` overhead in the hot loop.
* **Go 1.23+ Range Iterators**: Seamless integration with native `for range` loops via `iter.Seq`, allowing the compiler to perform aggressive loop inlining.
* **Command Buffer**: Thread-safe structural changes (Create/Remove/Add/Unassign) are buffered and applied during synchronization points to maintain state consistency.
* **Flexible Systems**: Supporting both stateless **Functional Systems** (closures) and **Stateful Systems** (structs) with full `Init/Update` lifecycles.
* **Advanced Execution Planning**: Deterministic scheduling with `RunParallel` support to utilize multi-core processors while maintaining control via explicit `Sync()` points.
* **Low-Level Access**: Use `AllocateComponent` for direct `*T` pointers to archetype storage or `AllocateComponentByInfo` to bypass reflection entirely.

> 💡 **See it in action**: Check the `cmd` directory for practical, ready-to-run examples, including a concurrent dice game simulation demonstrating parallel systems and state management.


## Core Architecture & "Mechanical Sympathy"

GOKE is designed with an understanding of modern CPU constraints:

### Memory Layout
1.  **Generation-based Recycling**: Entities are 64-bit IDs (32-bit Index / 32-bit Generation). This solves the ABA problem while allowing dense packing of internal storage.
2.  **Hardware Prefetching Optimization**: All view structures (Head/Tail) are strictly limited to a maximum of 4 pointer fields. Beyond this, CPU prefetching efficiency degrades. GOKE adheres to this limit to maintain maximal throughput.
3.  **Archetype Masks**: Supports up to **256 unique component types** using a fast, constant-time bitwise operation (4x64-bit fields).

### Execution Plan
* **Deferred Commands**: State consistency is maintained via `Commands`. Changes (add/remove) are buffered and applied during explicit `Sync()` points.
* **Power to the Programmer**: Support for `RunParallel` execution. GOKE provides the tools for concurrency, assuming the developer ensures disjoint component sets to avoid races.

## Installation

GOKE requires **Go 1.23** or newer.

## Usage Example

```go
package main

import (
    "fmt"
    "time"
    "[github.com/kjkrol/goke/pkg/ecs](https://github.com/kjkrol/goke/pkg/ecs)"
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
| **Create Entity** | **25.2 ns/op** | 91 B/op | **0** | Pre-allocated archetype slotting |
| **Add First Component** | **89.8 ns/op** | 75 B/op | **0** | Archetype migration (+ no data move + data insert) |
| **Add Next Component** | **145.8 ns/op** | 196 B/op | **0** | Archetype migration (+ data move + data insert) |
| **Add Tag** | **86.2 ns/op** | 17 B/op | **0** | Archetype migration (+ data move + no data insert) |
| **Remove Component** | **47.2 ns/op** | 0 B/op | **0** | Archetype migration (Swap-and-pop) |
| **Remove Entity** | **42.0 ns/op** | 0 B/op | **0** | Index recycling & record invalidation |
| **Structural Stability** | **133.0 ns/op** | 94 B/op | **0** | Stress test of add/remove cycles |

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
| View Complexity | Standard Iterator | Pure Iterator (No ID) | Performance Gain |
| :--- | :--- | :--- | :--- |
| **View0.All (Entity Only)** | 4.19 ms | - | - |
| **View1.All (1 Comp)** | 6.22 ms | 4.42 ms | **+29.0%** |
| **View3.All (3 Comps)** | 9.85 ms | 8.02 ms | **+18.5%** |
| **View8.All (8 Comps)** | 10.08 ms | 8.14 ms | **+19.2%** |

#### 3. Full Benchmark Logs
<details>
<summary>Click to view all scenarios (View 0-8, Pure, Filtered)</summary>

| Registry Size ($N$) | Operation | Total Time | Per Entity | Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **1,000** | View0.All | 435.9 ns | 0.43 ns | Linear SoA |
| **1,000** | View3.All | 822.0 ns | 0.82 ns | Linear SoA |
| **1,000** | View8.All | 809.4 ns | 0.80 ns | Linear SoA |
| **1,000** | View3.Filter 100 | 432.1 ns | 4.32 ns | O(1) Lookup |
| **1,000** | View3.Values | 644.6 ns | 0.64 ns | Pure Scan |
| **10,000** | View0.All | 4,208 ns | 0.42 ns | Linear SoA |
| **10,000** | View3.All | 10,306 ns | 1.03 ns | Linear SoA |
| **10,000** | View8.All | 9,920 ns | 0.99 ns | Linear SoA |
| **10,000** | View3.Filter 100 | 438.0 ns | 4.38 ns | O(1) Lookup |
| **10,000** | View3.Values | 7,673 ns | 0.76 ns | Pure Scan |
| **100,000** | View0.All | 41,813 ns | 0.41 ns | Linear SoA |
| **100,000** | View3.All | 97,686 ns | 0.97 ns | Linear SoA |
| **100,000** | View8.All | 97,582 ns | 0.97 ns | Linear SoA |
| **100,000** | View3.Filter 100 | 454.9 ns | 4.54 ns | O(1) Lookup |
| **100,000** | View3.Values | 77,924 ns | 0.77 ns | Pure Scan |
| **1,000,000** | View0.All | 415,515 ns | 0.41 ns | Linear SoA |
| **1,000,000** | View3.All | 989,272 ns | 0.98 ns | Linear SoA |
| **1,000,000** | View8.All | 992,372 ns | 0.99 ns | Linear SoA |
| **1,000,000** | View3.Filter 100 | 444.9 ns | 4.44 ns | O(1) Lookup |
| **1,000,000** | View3.Values | 797,392 ns | 0.79 ns | Pure Scan |
| **10,000,000** | View0.All | 4,195,179 ns | 0.41 ns | Linear SoA |
| **10,000,000** | View1.All | 6,229,583 ns | 0.62 ns | Linear SoA |
| **10,000,000** | View3.All | 9,857,888 ns | 0.98 ns | Linear SoA |
| **10,000,000** | View8.All | 10,089,932 ns | 1.00 ns | Linear SoA |
| **10,000,000** | View3.Filter 100 | 448.3 ns | 4.48 ns | O(1) Lookup |
| **10,000,000** | View1.Values | 4,424,216 ns | 0.44 ns | Pure Scan |
| **10,000,000** | View3.Values | 8,029,977 ns | 0.80 ns | Pure Scan |
| **10,000,000** | View8.Values | 8,140,268 ns | 0.81 ns | Pure Scan |
| **10,000,000** | View3 FilterValues 100 | 432.1 ns | 4.32 ns | Pure Lookup |

</details>

### Key Technical Takeaways
* **Near-Zero Latency:** Targeted queries (Filter) remain at **~440 ns** regardless of whether the world has 1k or 10M entities.
* **Instruction Efficiency:** Even with 8 components, the engine processes entities at **~1 ns/entity**, fitting a 10M entity update within a **10ms** window.
* **Pure Mode Gain:** Bypassing Entity ID generation provides up to **29%** additional throughput for heavy computational systems.


## Benchmark Comparison: goke/ecs vs. Arche (RAW Data)

It is important to note that while **goke/ecs** achieves industry-leading raw iteration speeds, **Arche** is a more established framework providing a broader feature set. The performance trade-offs in **Arche** often stem from supporting complex functionalities that **goke/ecs** intentionally omits to maintain its lean profile, such as:

* **Entity Relations:** Native support for parent-child hierarchies and linked entities.
* **Batch Operations:** Highly optimized mass entity spawning and destruction.
* **Event Listeners:** Comprehensive system for monitoring entity and component lifecycles (in **goke/ecs**, this includes Cached Views that listen for newly created archetypes to dynamically attach them to the view).

## Benchmark Comparison: goke/ecs vs. Arche
**Environment:** Apple M1 Max (ARM64)  
**Units:** Nanoseconds per operation (ns/op)

| Category | goke/ecs | Arche | Winner | Verified |
| :--- | :--- | :--- | :--- | :--- |
| **Iteration (1 Comp)** | **0.41 ns** | 0.55 ns | **goke/ecs (+24%)** | YES |
| **Iteration (2 Comp)** | **0.49 ns** | 1.37 ns | **goke/ecs (+64%)** | YES |
| **Iteration (3 Comp)** | **0.65 ns** | 1.79 ns | **goke/ecs (+63%)** | YES |
| **Create Entity** | 25.20 ns | **20.60 ns** | **Arche (+18%)** | YES |
| **Add First Component** | 89.80 ns | **29.30 ns** | **Arche (+67%)** | YES |
| **Add Next Component** | 145.80 ns | **??** | **Arche (+??%)** | NO |
| **Add Tag** | 86.20 ns | **??** | **Arche (+??%)** | NO |
| **Remove Component** | 47.20 ns | **??** | **Arche (+??%)** | NO |
| **Remove Entity** | 42.00 ns | **??** | **Arche (+??%)** | NO |
| **View Filter** | 4.48 ns | **??** | **Arche (+??%)** | NO |

## Roadmap

* **Batch Operations:** Implementation of high-performance bulk operations for entity creation and destruction, aimed at maximizing overhead reduction during large-scale data processing.

## License

GOKE is licensed under the MIT License. See the LICENSE file for more details.

## Documentation

Detailed API documentation and examples are available on [pkg.go.dev](https://pkg.go.dev/github.com/kjkrol/goke).

For a deep dive into the internal mechanics, check the `doc.go` files within the `ecs` packages.