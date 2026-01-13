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
    view *ecs.View3[Pos, Vel, Acc]
}

func (s *MovementSystem) Init(reg *ecs.Registry) {
    // View automatically finds matching archetypes and caches data pointers
    s.view = ecs.NewView3[Pos, Vel, Acc](reg)
}

func (s *MovementSystem) Update(reg *ecs.Registry, d time.Duration) {
    // Ultra-fast iteration over contiguous memory blocks
    for _, row := range s.view.All() {
        pos, vel, acc := row.Values() // Zero-allocation destructuring
        
        vel.X += acc.X
        vel.Y += acc.Y
        pos.X += vel.X
        pos.Y += vel.Y
    }
}

func main() {
    engine := ecs.NewEngine()

    e := engine.CreateEntity()
    ecs.Assign(engine, e, Pos{X: 0, Y: 0})
    ecs.Assign(engine, e, Vel{X: 1, Y: 1})
    ecs.Assign(engine, e, Acc{X: 0.1, Y: 0.1})

    engine.RegisterSystems([]ecs.System{&MovementSystem{}})
    
    // Run the system update loop
    engine.UpdateSystems(time.Millisecond * 16)
}
```

## Performance Metrics

GOKE minimizes "Pointer Chasing" by ensuring data is ready in the CPU cache before the loop requests it. By organizing memory into Archetypes (Structure of Arrays), we achieve near-theoretical limits for data throughput on modern hardware.

| Operation (1000 entities) | Total Time | Per Entity | Mechanism |
| :--- | :--- | :--- | :--- |
| **View1 All** | ~502 ns | **0.50 ns** | Linear SoA Access |
| **View3 All** | ~547 ns | **0.54 ns** | Multi-column SoA |
| **Filtered Access** | ~3850 ns | **3.85 ns** | Entity-to-Index Map |

*Benchmarks performed on AMD Ryzen 7 5825U. Performance results demonstrate that GOKE is significantly faster than traditional map-based or reflection-based ECS implementations in Go.*



## Memory Architecture (L1/L2 Cache Optimization)

Traditional ECS designs often suffer from "Cache Misses" because components for a single entity are scattered across the heap. GOKE solves this by:
1. Grouping identical entity types into **Archetypes**.
2. Storing component data in **Structure of Arrays (SoA)** format.
3. Using **Flat Cache Views** to provide the CPU with a predictable, linear stream of data.

This approach ensures that when the CPU fetches a component for one entity, the hardware prefetcher automatically pulls the data for the next several entities into the **L1/L2 cache** before the code even asks for them.



## License

GOKE is licensed under the MIT License. See the LICENSE file for more details.