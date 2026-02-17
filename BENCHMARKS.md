# ⏱️ GOKe Benchmarks

[← Back to README](./README.md)

> Detailed performance analysis and hardware specifications for the GOKe ECS engine.

## Environment
- **CPU:** Apple M1 Max (10-core)
- **Memory:** 64 GB RAM
- **Go Version:** 1.25.x
- **OS:** macOS

## Performance Metrics

### Structural Operations (Apple M1 Max)
These benchmarks highlight the efficiency of our archetype-based memory management. Every operation below results in **0 heap allocations**, ensuring no GC pressure during the simulation loop.

| Operation | Performance | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- |
| **Create with 1 Comp** | **22.46 ns/op ±21%** | **0** | Blueprint-based archetype slotting |
| **Create with 2 Comps** | **24.33 ns/op ±2%** | **0** | Blueprint-based archetype slotting |
| **Create with 3 Comps** | **27.12 ns/op ± 3%** | **0** | Blueprint-based archetype slotting |
| **Create with 4 Comps** | **28.72 ns/op ± 1%** | **0** | Blueprint-based archetype slotting |
| **Add Next Component** | **37.36 ns/op ±7%** | **0** | Archetype migration (+ data move + data insert) |
| **Add Tag** | **34.36 ns/op ± 3%** | **0** | Archetype migration (Metadata only) |
| **Remove Component** | **9.64 ns/op ±42%** | **0** | Archetype migration (Swap-and-pop) |
| **Remove Entity** | **17.95 ns/op ±2%** | **0** | Index recycling & record invalidation |
| **Get Entity Component** | **4.49 ns/op ±3%** | **0** | Direct record lookup with generation check |
| **Structural Stability** | **30.17 ns/op ±17%** | **0** | Stress test of add/remove cycles |



### Query Benchmarks (Apple M1 Max)
The following benchmarks demonstrate the efficiency of SoA (Structure of Arrays) and the $O(1)$ nature of our record-based filtering.

#### 1. Scalability Overview: $O(N)$ vs $O(1)$
| Registry Size ($N$) | Operation | Entities Processed | Total Time | Per Entity | Mechanism |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1,000** | **View3.All** | 1,000 | 538.4 ns | 0.54 ns | Linear SoA |
| **10,000** | **View3.All** | 10,000 | 6,942 ns | 0.69 ns | Linear SoA |
| **100,000** | **View3.All** | 100,000 | 68,571 ns | 0.69 ns | Linear SoA |
| **1,000,000** | **View3.All** | 1,000,000 | 1,397,688 ns | 1.34 ns | Linear SoA |
| **10,000,000** | **View3.All** | 10,000,000 | 20,132,807 ns | **2.01 ns** | Linear SoA |
| **10,000,000** | **View3.Filter** | **100** | **288.0 ns** | **2.88 ns** | **O(1) Record Lookup** |

#### 2. Complexity Scaling (10M Entities Stress Test)
| View Complexity | All Iterator | Values Iterator (No ID) |
| :--- | :--- | :--- | 
| **View0 (Entity Only)** | 0.34 ns | - |
| **View1 (1 Comp)** | 1.20 ns | 1.20 ns |
| **View3 (3 Comps)** | 2.18 ns | 2.18 ns |
| **View8 (8 Comps)** | 2.20 ns | 2.20 ns |


### Key Technical Takeaways
* **Near-Zero Latency:** Targeted queries (Filter) remain at **~2.88 ns** regardless of whether the world has 1k or 10M entities.
* **Instruction Efficiency:** Even with 8 components, the engine processes entities bellow **~2.2 ns/entity**, fitting a 10M entity update within a **22ms** window.
* **Values Iteration:** At 10M entities, performance saturates at the same level as standard iteration, indicating memory bandwidth limits outweigh ID generation overhead.

## Benchmark Comparison: GOKe vs. [Arche](https://github.com/mlange-42/arche) (RAW Data)

It is important to note that while **GOKe** achieves industry-leading raw iteration speeds, **Arche** is a more established framework providing a broader feature set. The performance trade-offs in **Arche** often stem from supporting complex functionalities that **GOKe** intentionally omits to maintain its lean profile, such as:

* **Entity Relations:** Native support for parent-child hierarchies and linked entities.
* **Batch Operations:** Highly optimized mass entity spawning and destruction.
* **Event Listeners:** Comprehensive system for monitoring entity and component lifecycles (in **goke/ecs**, this includes Cached Views that listen for newly created archetypes to dynamically attach them to the view).

## Benchmark Comparison: GOKe vs. Arche
**Environment:** Apple M1 Max (ARM64)  
**Units:** Nanoseconds per operation (**ns/op**)

| Operation | GOKe | Arche | Winner | Notes |
| :--- | :--- | :--- | :--- | :--- |
| **Iteration (1 Comp)** | **0.36** | 0.55 | **GOKe (+52.8%)** | Superior cache locality |
| **Iteration (2 Comp)** | **0.39** | 1.37 | **GOKe (+251.3%)** | Efficient memory layout |
| **Iteration (3 Comp)** | **0.53** | 1.79 | **GOKe (+237.7%)** | Minimal per-entity overhead |
| **Create Entity (1 comp)** | 22.46 | **20.60** | **Arche (+9.0%)** | Draw |
| **Add Next Component** | 37.36 | **--** | **Arche** | Fast graph traversal |
| **Remove Component** | **9.64** | **--** | **GOKe/Arche** | Highly optimized swap-and-pop |
| **Remove Entity** | 17.95 | **--** | **Arche** | Fast fast cleanup |

## How to benchmark
Results may vary based on CPU architecture and cache sizes. You can run the suite using:
```bash
go test -bench=. ./... -benchmem
```
