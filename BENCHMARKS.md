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

| Operation | Performance | Memory | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Create with 1 Comp** | **38.59 ns/op** | 82 B/op | **0** | Blueprint-based archetype slotting |
| **Create with 2 Comps** | **42.04 ns/op** | 144 B/op | **0** | Blueprint-based archetype slotting |
| **Create with 3 Comps** | **41.26 ns/op** | 66 B/op | **0** | Blueprint-based archetype slotting |
| **Add Next Component** | **79.06 ns/op** | 203 B/op | **0** | Archetype migration (+ data move + data insert) |
| **Add Tag** | **39.59 ns/op** | 119 B/op | **0** | Archetype migration (Metadata only) |
| **Remove Component** | **11.09 ns/op** | 0 B/op | **0** | Archetype migration (Swap-and-pop) |
| **Remove Entity** | **18.88 ns/op** | 0 B/op | **0** | Index recycling & record invalidation |
| **Get Entity Component** | **5.45 ns/op** | 0 B/op | **0** | Direct record lookup with generation check |
| **Structural Stability** | **49.02 ns/op** | 67 B/op | **0** | Stress test of add/remove cycles |



### Query Benchmarks (Apple M1 Max)
The following benchmarks demonstrate the efficiency of SoA (Structure of Arrays) and the $O(1)$ nature of our record-based filtering.

#### 1. Scalability Overview: $O(N)$ vs $O(1)$
| Registry Size ($N$) | Operation | Entities Processed | Total Time | Per Entity | Mechanism |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1,000** | **View3.All** | 1,000 | 639.5 ns | 0.64 ns | Linear SoA |
| **10,000** | **View3.All** | 10,000 | 7,268 ns | 0.73 ns | Linear SoA |
| **100,000** | **View3.All** | 100,000 | 79,472 ns | 0.79 ns | Linear SoA |
| **1,000,000** | **View3.All** | 1,000,000 | 782,517 ns | 0.78 ns | Linear SoA |
| **10,000,000** | **View3.All** | 10,000,000 | 7,781,371 ns | **0.78 ns** | Linear SoA |
| **10,000,000** | **View3.Filter** | **100** | **232.0 ns** | **2.32 ns** | **O(1) Record Lookup** |

#### 2. Complexity Scaling (10M Entities Stress Test)
| View Complexity | All Iterator | Values Iterator (No ID) | Performance Gain |
| :--- | :--- | :--- | :--- |
| **View0 (Entity Only)** | 3.15 ms | - | - |
| **View1 (1 Comp)** | 4.73 ms | 3.18 ms | **+32.8%** |
| **View3 (3 Comps)** | 7.86 ms | 7.34 ms | **+6.6%** |
| **View8 (8 Comps)** | 7.97 ms | 6.71 ms | **+15.8%** |


### Key Technical Takeaways
* **Near-Zero Latency:** Targeted queries (Filter) remain at **~232 ns** regardless of whether the world has 1k or 10M entities.
* **Instruction Efficiency:** Even with 8 components, the engine processes entities bellow **~0.8 ns/entity**, fitting a 10M entity update within a **10ms** window.
* **Values Iteration Gain:** Bypassing Entity ID generation provides up to **33%** additional throughput for heavy computational systems.


## Benchmark Comparison: GOKe vs. [Arche](https://github.com/mlange-42/arche) (RAW Data)

It is important to note that while **GOKe** achieves industry-leading raw iteration speeds, **Arche** is a more established framework providing a broader feature set. The performance trade-offs in **Arche** often stem from supporting complex functionalities that **GOKe** intentionally omits to maintain its lean profile, such as:

* **Entity Relations:** Native support for parent-child hierarchies and linked entities.
* **Batch Operations:** Highly optimized mass entity spawning and destruction.
* **Event Listeners:** Comprehensive system for monitoring entity and component lifecycles (in **goke/ecs**, this includes Cached Views that listen for newly created archetypes to dynamically attach them to the view).

## Benchmark Comparison: GOKe vs. Arche
**Environment:** Apple M1 Max (ARM64)  
**Units:** Nanoseconds per operation (ns/op)

| Operation | GOKe | Arche | Winner | Notes |
| :--- | :--- | :--- | :--- | :--- |
| **Iteration (1 Comp)** | **0.44 ns** | 0.55 ns | **GOKe (+20.0%)** | Superior cache locality |
| **Iteration (2 Comp)** | **0.48 ns** | 1.37 ns | **GOKe (+65.0%)** | Efficient memory layout |
| **Iteration (3 Comp)** | **0.64 ns** | 1.79 ns | **GOKe (+64.2%)** | Minimal per-entity overhead |
| **Create Entity (1 comp)** | 38.59 ns | **20.60 ns** | **Arche (+46%)** | Arche's simpler indexing edge |
| **Add First Component*** | 38.59 ns | **29.30 ns** | **Arche (+24%)** | GOKe uses atomic Blueprint |
| **Add Next Component** | 79.06 ns | **--** | **Arche** | Fast graph traversal |
| **Add Tag** | 39.59 ns | **--** | **Arche** | Metadata-only migration |
| **Remove Component** | **11.09 ns** | **--** | **GOKe/Arche** | Highly optimized swap-and-pop |
| **Remove Entity** | 18.88 ns | **--** | **Arche** | Lightning fast cleanup |

## How to benchmark
Results may vary based on CPU architecture and cache sizes. You can run the suite using:
```bash
go test -bench=. ./... -benchmem
```