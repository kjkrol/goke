# GOKE

*aka "Golang kjkrol ECS"*

**GOKE** is an ultra-lightweight, high-performance, and type-safe Entity Component System (ECS) for Go. It is engineered for maximum data throughput, leveraging modern **Go 1.23+ Iterators** and a Data-Oriented Design (DOD) architecture.

## Why GOKE is Blazing Fast

Unlike many ECS implementations in Go that rely on component maps or reflection during the update loop, GOKE shifts the heavy lifting to the initialization phase using:

* **Archetype-Based Storage (SoA)**: Entities with the same component composition (Archetypes) are stored in contiguous memory blocks. This layout is **L1/L2 Cache friendly**, allowing the CPU to utilize hardware prefetching to stream data directly into registers.
* **Flat Cache View**: Views pre-calculate direct pointers to component columns within archetypes during indexing. This eliminates map lookups and pointer chasing inside the hot loop.
* **Zero-Overhead Iteration**: By using a specialized code generator and fixed-size arrays, the iteration loop performance is comparable to raw C++ array processing.
* **World-Class Benchmarks**: In "Simple Update" tests (Pos+Vel+Acc), GOKE achieves processing times as low as **~0.54 ns per entity**, reaching the physical limits of modern CPU instruction pipelining.



## Key Features

* **Zero Reflection in Update**: All pointer arithmetic and bitmask calculations are resolved during `Reindex`, not every frame.
* **Modern Iterators**: Powered by native `for range` over functions (`iter.Seq2`), enabling the Go compiler to perform aggressive loop inlining.
* **True Type Safety**: Fully powered by Go Generics. No more `interface{}` casting or type assertions inside your systems.
* **No Bounds Checks**: The generated code uses fixed-size array metadata, allowing the compiler to elide bounds check instructions in the hottest parts of the engine.

## Installation

GOKE requires **Go 1.23** or newer.

```bash
go get github.com/kjkrol/goke
```

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

type MovementSystem struct {
    query *ecs.Query3[Pos, Vel, Acc]
}

func (s *MovementSystem) Init(reg *ecs.Registry) {
    // View automatically finds matching archetypes and caches data pointers
    s.query = ecs.NewQuery3[Pos, Vel, Acc](reg)
}

func (s *MovementSystem) Update(
    reg ecs.ReadOnlyRegistry,
    cb *ecs.SystemCommandBuffer,
    d time.Duration,
) {
    // Ultra-fast iteration over contiguous memory blocks
    for head := range s.query.All3() {
        pos, vel, acc := head.V1, head.V2, head.V3
        
        vel.X += acc.X
        vel.Y += acc.Y
        pos.X += vel.X
        pos.Y += vel.Y
    }
}

func (s *MovementSystem) ShouldSync() bool { return false }

func main() {
    engine := ecs.NewEngine()

    e := engine.CreateEntity()
    ecs.Assign(engine, e, Pos{X: 0, Y: 0})
    ecs.Assign(engine, e, Vel{X: 1, Y: 1})
    ecs.Assign(engine, e, Acc{X: 0.1, Y: 0.1})

    movementSystem := &MovementSystem{}
    engine.RegisterSystem(movementSystem)
    
    engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		ctx.RunSystem(movementSystem, d)
		ctx.Sync()
	})
    
    // Run the system update loop
    engine.Run(time.Millisecond * 16)
}
```

## Performance & Scalability

The engine is engineered for extreme scalability and deterministic performance. By utilizing a **Centralized Record System** (dense array lookup) instead of traditional hash maps, we have effectively decoupled query performance from the total entity count ($N$).

### Benchmarks (Apple M1 Max)
The following benchmarks demonstrate the efficiency of SoA (Structure of Arrays) and the $O(1)$ nature of our record-based filtering.

| Registry Size ($N$) | Operation | Entities Processed ($k$) | Total Time | Per Entity | Mechanism |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 100,000 | **View1 All** | 100,000 | ~43,462 ns | **0.43 ns** | Linear SoA Access |
| 100,000 | **View3 All** | 100,000 | ~73,639 ns | **0.73 ns** | Multi-column SoA |
| **1,000** | **View3 Filtered** | 100 | **~455 ns** | 4.55 ns | O(1) Record Lookup |
| **10,000** | **View3 Filtered** | 100 | **~463 ns** | 4.63 ns | O(1) Record Lookup |
| **100,000** | **View3 Filtered** | 100 | **~462 ns** | 4.62 ns | O(1) Record Lookup |

### Benchmarks (AMD Ryzen 7 5825U)
The following benchmarks demonstrate the efficiency of SoA (Structure of Arrays) and the $O(1)$ nature of our record-based filtering on x86_64 architecture.

| Registry Size ($N$) | Operation | Entities Processed ($k$) | Total Time | Per Entity | Mechanism |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 100,000 | **View1 All** | 100,000 | ~47,292 ns | **0.47 ns** | Linear SoA Access |
| 100,000 | **View3 All** | 100,000 | ~56,188 ns | **0.56 ns** | Multi-column SoA |
| **1,000** | **View3 Filtered** | 100 | **~507 ns** | 5.07 ns | O(1) Record Lookup |
| **10,000** | **View3 Filtered** | 100 | **~575 ns** | 5.75 ns | O(1) Record Lookup |
| **100,000** | **View3 Filtered** | 100 | **~516 ns** | 5.16 ns | O(1) Record Lookup |

### Key Technical Takeaways

* **Deterministic $O(1)$ Filtering:** As shown in the benchmarks, querying 100 specific entities takes exactly the same time (~462 ns) whether the registry contains 1,000 or 100,000 entities. This is achieved by bypassing hash map probing entirely.
* **Extreme Data Locality:** With a processing speed of **~0.43 ns per entity** in linear scans, the engine operates at near-memory-bandwidth limits, fully utilizing CPU prefetching and L1/L2 caches.
* **Hybrid Iterator Strategy:** Our generated `Filtered` queries implement a "Last Archetype Cache". If the filtered list contains entities from the same archetype, the engine reuses column pointers, significantly reducing pointer arithmetic and memory-to-register traffic.
* **Zero-Map Overhead:** By moving the "Entity to Index" mapping to a centralized, dense array of `Records`, we eliminated the cache-miss-heavy map lookups that typically slow down ECS engines as they scale.

## Memory Architecture (L1/L2 Cache Optimization)

Traditional ECS designs often suffer from "Cache Misses" because components for a single entity are scattered across the heap. GOKE solves this by:
1. Grouping identical entity types into **Archetypes**.
2. Storing component data in **Structure of Arrays (SoA)** format.
3. Using **Flat Cache Views** to provide the CPU with a predictable, linear stream of data.

This approach ensures that when the CPU fetches a component for one entity, the hardware prefetcher automatically pulls the data for the next several entities into the **L1/L2 cache** before the code even asks for them.



## License

GOKE is licensed under the MIT License. See the LICENSE file for more details.