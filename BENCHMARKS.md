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
| **Create Entity** | **22.95 ns/op** | 69 B/op | **0** | Pre-allocated archetype slotting |
| **Add First Component** | **28.30 ns/op** | 69 B/op | **0** | Archetype migration (+ no data move + data insert) |
| **Add Next Component** | **43.27 ns/op** | 115 B/op | **0** | Archetype migration (+ data move + data insert) |
| **Add Tag** | **21.52 ns/op** | 16 B/op | **0** | Archetype migration (+ data move + no data insert) |
| **Remove Component** | **15.82 ns/op** | 0 B/op | **0** | Archetype migration (Swap-and-pop) |
| **Remove Entity** | **14.44 ns/op** | 0 B/op | **0** | Index recycling & record invalidation |
| **Structural Stability** | **47.80 ns/op** | 64 B/op | **0** | Stress test of add/remove cycles |
| **Get Entity Component** | **14.44 ns/op** | 0 B/op | **0** | Retrieves a copy of the component data using entity index and generation validation |



### Query Benchmarks (Apple M1 Max)
The following benchmarks demonstrate the efficiency of SoA (Structure of Arrays) and the $O(1)$ nature of our record-based filtering.

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

## How to benchmark
Results may vary based on CPU architecture and cache sizes. You can run the suite using:
```bash
go test -bench=. ./... -benchmem
```