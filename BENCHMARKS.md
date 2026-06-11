# ⏱️ GOKe Benchmarks

[← Back to README](./README.md)

> Detailed performance analysis and hardware specifications for the GOKe ECS engine.

## Environment
- **CPU:** Apple M1 Max (10-core)
- **Memory:** 64 GB RAM
- **Go Version:** 1.25.x
- **OS:** macOS

## Performance Metrics

### Structural Operations
Every operation below results in **0 heap allocations** on the hot path.

| Operation | Performance | Allocs | Technical Mechanism |
| :--- | :--- | :--- | :--- |
| **Add Next Component** | **36.08 ns/op** | **0** | Archetype migration (move + insert) |
| **Add Tag** | **33.96 ns/op** | **0** | Archetype migration (metadata only) |
| **Remove Component** | **7.58 ns/op** | **0** | Archetype migration (swap-and-pop) |
| **Remove Entity (clean)** | **2.83 ns/op** | **0** | Index recycling |
| **Add/Remove Stability** | **92.82 ns/op** | 4 | Stress test (forced archetype churn) |
| **Get Component** | **4.67 ns/op** | **0** | Inlined record lookup + generation check |
| **Get Component (Safe)** | **8.19 ns/op** | **0** | Adds reflection-based type validation |
| **Get Component via View.Filter** | **4.24 ns/op** | **0** | Single-entity Filter for type-safe access |

### Batch Entity Creation

| Components | 2¹⁰ ns/entity | 2¹⁰ Allocs | 2²⁰ ns/entity | 2²⁰ Allocs |
| :--- | ---: | ---: | ---: | ---: |
| **1 comp** | 12.87 | 4 | 18.86 | 5 |
| **2 comp** | 9.98 | 4 | 12.78 | 5 |
| **3 comp** | 9.54 | 4 | 14.03 | 5 |
| **4 comp** | 8.93 | 4 | 12.09 | 5 |
| **5 comp** | 10.05 | 4 | 12.21 | 5 |
| **6 comp** | 9.59 | 4 | 19.32 | 5 |
| **7 comp** | 10.71 | 4 | 18.48 | 5 |
| **8 comp** | 12.08 | 4 | 17.80 | 5 |
| **9 comp** | 13.02 | 4 | 22.45 | 5 |
| **10 comp** | 14.24 | 4 | 20.42 | 5 |

The additional allocation at 2²⁰ entities comes from page growth during large batch construction. Despite the 1024× increase in entity count, creation remains O(n) with a per-entity cost in the low tens of nanoseconds.

### View.All (full archetype scan, SoA pages)

| Components | 2¹⁰ (1,024) ns/entity | 2²⁰ (1,048,576) ns/entity |
| :--- | ---: | ---: |
| **View0** (Entity only) | **0.354** | **0.341** |
| **View1** | **0.358** | **0.344** |
| **View2** | **0.505** | **0.506** |
| **View3** | **0.703** | **0.662** |
| **View4** | **0.839** | **0.842** |
| **View5** | **1.017** | **1.050** |
| **View6** | **1.165** | **1.242** |
| **View7** | **1.420** | **1.550** |
| **View8** | **1.654** | **1.833** |
| **View9** | **1.831** | **2.178** |
| **View10** | **2.029** | **2.530** |

### View.Filter (per-entity subset iteration, N=100)

Yields `iter.Seq2[int, struct{Entity, Comp1, ...}]` — the index is the position in the input `selected` slice, so callers can correlate results and detect skipped entities (not matching the view or already removed).

| Components | 2¹⁰ sorted | 2¹⁰ shuffled | 2²⁰ sorted | 2²⁰ shuffled |
| :--- | ---: | ---: | ---: | ---: |
| **View0** | 3.120 ns/ent | 3.109 ns/ent | 3.151 ns/ent | 3.124 ns/ent |
| **View1** | 4.237 ns/ent | 4.234 ns/ent | 4.251 ns/ent | 4.304 ns/ent |
| **View2** | 4.726 ns/ent | 4.740 ns/ent | 4.774 ns/ent | 4.682 ns/ent |
| **View3** | 5.444 ns/ent | 5.476 ns/ent | 5.584 ns/ent | 5.596 ns/ent |
| **View4** | 6.736 ns/ent | 6.230 ns/ent | 6.320 ns/ent | 6.496 ns/ent |
| **View5** | 7.257 ns/ent | 7.253 ns/ent | 7.538 ns/ent | 7.431 ns/ent |
| **View6** | 7.900 ns/ent | 8.022 ns/ent | 8.015 ns/ent | 8.244 ns/ent |
| **View7** | 9.038 ns/ent | 8.969 ns/ent | 9.298 ns/ent | 9.045 ns/ent |
| **View8** | 9.924 ns/ent | 9.452 ns/ent | 9.966 ns/ent | 9.883 ns/ent |
| **View9** | 10.63 ns/ent | 10.63 ns/ent | 10.80 ns/ent | 10.54 ns/ent |
| **View10** | 11.04 ns/ent | 11.18 ns/ent | 11.46 ns/ent | 11.40 ns/ent |

### Key Technical Takeaways
* **Scale-independent queries:** `Filter` cost remains statistically unchanged from **2¹⁰** to **2²⁰** entities. Query cost depends primarily on the number of requested components (~3.1–11.5 ns/entity), not on total entity count.
* **Linear iteration scaling:** `View.All` scales approximately linearly with component count, ranging from **0.34 ns/entity** (0 components) to **2.5 ns/entity** (10 components), reflecting predictable SoA traversal costs.
* **Sorted vs shuffled:** Filter results show only minor variance between sorted and shuffled entity orders, indicating that the hot path is dominated by direct record lookup and pointer arithmetic rather than access locality.
* **Scale-independent structural operations:** Component migration (~37 ns), tag insertion (~34 ns), component removal (~9 ns), entity destruction (~3 ns), and component access (~4.6 ns) remain effectively constant from **2¹⁰** to **2²⁰** entities.
* **Zero allocations on hot paths:** All query, iteration, lookup, and migration operations report **0 B/op** and **0 allocs/op** (`-benchmem`).

## Benchmark Comparison with Other ECS Libraries

Cross-framework benchmarks are maintained in a dedicated project:
**[go-ecs-benchmarks](https://github.com/mlange-42/go-ecs-benchmarks)** by [@mlange-42](https://github.com/mlange-42).

This project compares GOKe against other Go ECS implementations (Arche, Donburi, Ento, etc.) on a unified workload, eliminating bias from differently-shaped local benchmarks.

> ⚠️ **Check the compared versions.** The results published in go-ecs-benchmarks may lag behind GOKe's main branch. Before drawing conclusions, verify which GOKe version (tag) is used in the comparison and re-run the suite against the version you care about.

## How to benchmark
```bash
go test -bench=. ./... -benchmem
```
