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

### Batch Entity Creation (1024 entities per call)

| Components | Total ns/op | ns/entity | Allocs |
| :--- | :--- | :--- | :--- |
| **1 comp** | 10,972 | **10.71** | 4 |
| **2 comp** | 11,567 | **11.30** | 4 |
| **3 comp** | 9,103 | **8.89** | 4 |
| **4 comp** | 8,572 | **8.37** | 4 |
| **5 comp** | 7,982 | **7.80** | 4 |
| **6 comp** | 9,216 | **9.00** | 4 |
| **7 comp** | 11,721 | **11.45** | 4 |
| **8 comp** | 12,388 | **12.10** | 4 |
| **9 comp** | 20,738 | **20.25** | 4 |
| **10 comp** | 15,771 | **15.40** | 4 |

### View.All (full archetype scan, SoA pages, 1024 entities)

| Components | Total ns/op | ns/entity |
| :--- | :--- | :--- |
| **View0** (Entity only) | 346.6 | **0.35** |
| **View1** | 358.9 | **0.36** |
| **View2** | 515.5 | **0.52** |
| **View3** | 679.5 | **0.68** |
| **View4** | 854.5 | **0.85** |
| **View5** | 1,020 | **1.02** |
| **View6** | 1,176 | **1.18** |
| **View7** | 1,436 | **1.44** |
| **View8** | 1,637 | **1.64** |
| **View9** | 1,831 | **1.83** |
| **View10** | 2,058 | **2.06** |

### View.Filter (per-entity subset iteration, N=100 from 1024 pool)
Yields `iter.Seq2[int, struct{Entity, Comp1, ...}]` — the index is the position in the input `selected` slice, so callers can correlate results and detect skipped entities (not matching the view or already removed).

| Components | sorted | shuffled |
| :--- | :--- | :--- |
| **View0** | 3.20 ns/ent | 3.15 ns/ent |
| **View1** | 4.31 ns/ent | 4.32 ns/ent |
| **View2** | 4.75 ns/ent | 4.79 ns/ent |
| **View3** | 5.58 ns/ent | 5.50 ns/ent |
| **View4** | 6.45 ns/ent | 6.40 ns/ent |
| **View5** | 7.42 ns/ent | 7.46 ns/ent |
| **View6** | 8.10 ns/ent | 8.17 ns/ent |
| **View7** | 9.27 ns/ent | 9.09 ns/ent |
| **View8** | 9.79 ns/ent | 9.86 ns/ent |
| **View9** | 10.78 ns/ent | 10.84 ns/ent |
| **View10** | 11.27 ns/ent | 11.33 ns/ent |

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
