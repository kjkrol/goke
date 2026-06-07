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
| **Add Next Component** | **35.89 ns/op** | **0** | Archetype migration (move + insert) |
| **Add Tag** | **34.11 ns/op** | **0** | Archetype migration (metadata only) |
| **Remove Component** | **7.48 ns/op** | **0** | Archetype migration (swap-and-pop) |
| **Remove Entity (clean)** | **2.67 ns/op** | **0** | Index recycling |
| **Add/Remove Stability** | **86.44 ns/op** | 4 | Stress test (forced archetype churn) |
| **Get Component** | **4.70 ns/op** | **0** | Inlined record lookup + generation check |
| **Get Component (Safe)** | **8.72 ns/op** | **0** | Adds reflection-based type validation |
| **Get Component via View.Filter** | **4.61 ns/op** | **0** | Single-entity Filter for type-safe access |

### Batch Entity Creation (1024 entities per call)

| Components | Total ns/op | ns/entity | Allocs |
| :--- | :--- | :--- | :--- |
| **1 comp** | 13,483 | **13.17** | 4 |
| **2 comp** | 10,381 | **10.14** | 4 |
| **3 comp** | 12,533 | **12.24** | 4 |
| **4 comp** | 14,439 | **14.10** | 4 |
| **5 comp** | 15,529 | **15.16** | 4 |
| **6 comp** | 16,449 | **16.06** | 4 |
| **7 comp** | 17,984 | **17.56** | 5 |
| **8 comp** | 21,240 | **20.74** | 5 |
| **9 comp** | 23,909 | **23.35** | 5 |
| **10 comp** | 25,044 | **24.46** | 5 |

### View.All (full archetype scan, SoA pages, 1k entities)

| Components | Total ns/op | ns/entity |
| :--- | :--- | :--- |
| **View0** (Entity only) | 336.8 | **0.34** |
| **View1** | 345.6 | **0.35** |
| **View2** | 491.0 | **0.49** |
| **View3** | 653.6 | **0.65** |
| **View4** | 812.4 | **0.81** |
| **View5** | 971.0 | **0.97** |
| **View6** | 1,143 | **1.14** |
| **View7** | 1,374 | **1.37** |
| **View8** | 1,573 | **1.57** |
| **View9** | 1,768 | **1.77** |
| **View10** | 1,960 | **1.96** |

### View.Filter (per-entity subset iteration, N=100 from 1k pool)
Yields `iter.Seq2[int, struct{Entity, Comp1, ...}]` — the index is the position in the input `selected` slice, so callers can correlate results and detect skipped entities (not matching the view or already removed).

| Components | sorted | shuffled |
| :--- | :--- | :--- |
| **View1** | 4.22 ns/ent | 4.23 ns/ent |
| **View2** | 4.74 ns/ent | 4.68 ns/ent |
| **View3** | 5.53 ns/ent | 5.58 ns/ent |
| **View4** | 6.37 ns/ent | 6.38 ns/ent |
| **View5** | 7.39 ns/ent | 7.38 ns/ent |
| **View6** | 8.03 ns/ent | 8.05 ns/ent |
| **View7** | 9.03 ns/ent | 8.90 ns/ent |
| **View8** | 9.34 ns/ent | 9.55 ns/ent |
| **View9** | 10.54 ns/ent | 10.61 ns/ent |
| **View10** | 10.90 ns/ent | 10.93 ns/ent |

### Key Technical Takeaways
* **Constant per-entity lookup:** `Filter` cost stays around 4.2 ns/entity (1 component) regardless of registry size — the index lookup is direct via record array; only the requested subset size matters.
* **Linear iteration scaling:** `View.All` scales linearly with the component count (~0.15 ns/entity per added component), bounded by memory bandwidth.
* **Sorted vs shuffled:** Filter results show <2% variance between sorted and shuffled access patterns, confirming that the per-entity path is dominated by record lookup, not cache locality.
* **Zero allocations:** All hot-path operations report 0 B/op and 0 allocs/op (verified by `-benchmem`).

## Benchmark Comparison with Other ECS Libraries

Cross-framework benchmarks are maintained in a dedicated project:
**[go-ecs-benchmarks](https://github.com/mlange-42/go-ecs-benchmarks)** by [@mlange-42](https://github.com/mlange-42).

This project compares GOKe against other Go ECS implementations (Arche, Donburi, Ento, etc.) on a unified workload, eliminating bias from differently-shaped local benchmarks.

> ⚠️ **Check the compared versions.** The results published in go-ecs-benchmarks may lag behind GOKe's main branch. Before drawing conclusions, verify which GOKe version (tag) is used in the comparison and re-run the suite against the version you care about.

## How to benchmark
```bash
go test -bench=. ./... -benchmem
```
