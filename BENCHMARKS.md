# ⏱️ GOKe Benchmarks

[← Back to README](./README.md)

> Detailed performance analysis and hardware specifications for the GOKe ECS engine.

## Environment
- **CPU:** Apple M1 Max (10-core)
- **Memory:** 64 GB RAM
- **Go Version:** 1.26.4
- **OS:** macOS

## Scalability Validation

`Editor.Add/AddTags/Del/Mix`, `Factory.Create`, and `Query.All` are benchmarked at both **2¹⁰ (1,024)** and **2²⁰ (1,048,576)** entities. Per-entity cost for the structural (`Editor`/`Factory`) operations stays close between the two scales — see the tables below. `Query.All`'s per-entity cost roughly doubles at 2²⁰ due to chunk-hopping overhead once the working set no longer fits in cache (a known, documented trade-off of the chunked SoA layout — see [CHANGELOG](./CHANGELOG.md)).

`Query.Pick`/`Query.Seek` are benchmarked at a fixed population of 1,024 with a 100-entity subset; `Remove` is benchmarked at a fixed population of 100,000. These don't currently have a 2²⁰ sweep.

## Performance Metrics

### Structural Operations — Editor

`Editor.Update` migrates an entity to a new archetype in one move. Its cost is driven by the **width of the source and destination archetypes**, not by how many components the edit itself changes (see [`internal/ent/editor.go`](internal/ent/editor.go) for the full explanation) — that's why `Add` (starting from a 1-component anchor) stays cheap longer than `Del` (starting from a 10-component archetype), even though both add/remove the same number of components.

#### Add (`comp=1`, growing toward `comp=1+N`)
| N added | 2¹⁰ ns/ent | 2²⁰ ns/ent | 2¹⁰ allocs/op | 2²⁰ allocs/op | 2²⁰ B/op |
| :--- | ---: | ---: | ---: | ---: | ---: |
| 1 | 33.07 | 31.79 | 0 | 212 | 20.8 MB |
| 2 | 40.11 | 38.57 | 0 | 297 | 29.2 MB |
| 3 | 47.78 | 45.58 | 0 | 383 | 37.7 MB |
| 4 | 53.87 | 51.82 | 0 | 425 | 41.8 MB |
| 5 | 59.47 | 58.04 | 0 | 468 | 46.0 MB |
| 6 | 66.54 | 64.06 | 0 | 514 | 50.4 MB |
| 7 | 71.52 | 69.73 | 0 | 598 | 58.7 MB |
| 8 | 78.05 | 76.48 | 0 | 683 | 67.1 MB |
| 9 | 84.20 | 82.05 | 0 | 770 | 75.6 MB |
| 10 | 90.11 | 87.34 | 0 | 854 | 83.9 MB |

#### Add Tags (`comp=1`, growing toward `comp=1+N` zero-size tags)
| N tags | 2¹⁰ ns/ent | 2²⁰ ns/ent | 2¹⁰ allocs/op | 2²⁰ allocs/op | 2²⁰ B/op |
| :--- | ---: | ---: | ---: | ---: | ---: |
| 1 | 31.65 | 29.93 | 0 | 127 | 12.49 MB |
| 2 | 35.44 | 34.24 | 0 | 127 | 12.49 MB |
| 3 | 40.52 | 38.46 | 0 | 127 | 12.49 MB |
| 4 | 44.75 | 42.88 | 0 | 127 | 12.49 MB |
| 5 | 49.15 | 47.00 | 0 | 127 | 12.50 MB |
| 6 | 53.62 | 51.02 | 0 | 127 | 12.50 MB |
| 7 | 57.07 | 55.59 | 0 | 128 | 12.50 MB |
| 8 | 61.93 | 61.07 | 0 | 128 | 12.50 MB |
| 9 | 66.94 | 64.52 | 0 | 128 | 12.50 MB |
| 10 | 70.86 | 68.07 | 0 | 128 | 12.51 MB |

Note how `AddTags`'s 2²⁰ byte cost stays flat (~12.5 MB) regardless of how many tags are toggled — since tags are zero-size, the migration's cost is dominated by moving the population through the (fixed-width) archetype once, not by how many tag bits change.

#### Del (`comp=10`, shrinking by N)
| N removed | 2¹⁰ ns/ent | 2²⁰ ns/ent | 2¹⁰ allocs/op | 2²⁰ allocs/op | 2²⁰ B/op |
| :--- | ---: | ---: | ---: | ---: | ---: |
| 1 | 83.08 | 85.74 | 0 | 724 | 71.2 MB |
| 2 | 85.17 | 86.97 | 0 | 639 | 62.8 MB |
| 3 | 87.12 | 87.30 | 0 | 553 | 54.4 MB |
| 4 | 89.05 | 88.80 | 0 | 512 | 50.3 MB |
| 5 | 90.63 | 91.07 | 0 | 469 | 46.0 MB |
| 6 | 93.01 | 93.00 | 0 | 426 | 41.8 MB |
| 7 | 94.54 | 92.57 | 0 | 341 | 33.4 MB |
| 8 | 96.47 | 94.38 | 0 | 256 | 25.1 MB |
| 9 | 98.85 | 95.97 | 0 | 170 | 16.6 MB |
| 10 | 107.7 | 104.3 | 0 | 128 | 12.5 MB |

#### Mix (combined add+remove in one migration, `comp=N+1 → comp=N+1`, swapping N components)
| N swapped | comp (src=dst) | 2¹⁰ ns/ent | 2²⁰ ns/ent | 2¹⁰ allocs/op | 2²⁰ allocs/op | 2²⁰ B/op |
| :--- | :--- | ---: | ---: | ---: | ---: | ---: |
| 1 | 2 | 42.38 | 40.02 | 0 | 169 | 16.6 MB |
| 2 | 3 | 58.17 | 55.20 | 0 | 212 | 20.9 MB |
| 3 | 4 | 73.84 | 71.12 | 0 | 256 | 25.1 MB |
| 4 | 5 | 88.45 | 86.24 | 0 | 298 | 29.2 MB |
| 5 | 6 | 102.7 | 99.73 | 0 | 342 | 33.4 MB |
| 6 | 7 | 116.4 | 113.7 | 0 | 385 | 37.7 MB |
| 7 | 8 | 131.5 | 128.0 | 0 | 428 | 41.9 MB |
| 8 | 9 | 145.6 | 145.4 | 0 | 470 | 46.0 MB |
| 9 | 10 | 156.0 | 156.2 | 0 | 516 | 50.5 MB |
| 10 | 11 | 170.1 | 180.9 | 0 | 556 | 54.4 MB |

At 1,024 entities, all four operations cycle between two already-allocated archetype shapes, so the per-call cost reported is **0 allocs/op** at steady state (the chunk-reuse `spare` slot absorbs it — see [CHANGELOG](./CHANGELOG.md)). At 1,048,576 entities each iteration genuinely grows/shrinks an archetype spanning the whole population, so real allocation work is unavoidable — the `2²⁰ B/op` column above reflects that real cost, not a regression.

### Batch Entity Creation (`Factory.Create`)

| Components | 2¹⁰ ns/ent | 2²⁰ ns/ent | 2¹⁰ allocs/op | 2²⁰ allocs/op | 2²⁰ B/op |
| :--- | ---: | ---: | ---: | ---: | ---: |
| 1 | 7.726 | 8.946 | 0 | 1 | 91.7 MB |
| 2 | 4.382 | 3.910 | 0 | 1 | 25.2 MB |
| 3 | 4.696 | 4.211 | 0 | 1 | 33.6 MB |
| 4 | 4.801 | 4.446 | 0 | 1 | 37.8 MB |
| 5 | 5.162 | 4.715 | 0 | 1 | 42.0 MB |
| 6 | 5.507 | 5.109 | 0 | 1 | 46.2 MB |
| 7 | 6.053 | 6.555 | 0 | 1 | 54.6 MB |
| 8 | 6.965 | 7.893 | 0 | 1 | 63.0 MB |
| 9 | 6.998 | 7.982 | 0 | 1 | 71.5 MB |
| 10 | 7.405 | 8.471 | 0 | 1 | 79.9 MB |

> **Note on `1_comp`:** `Factory.Create` is the one benchmark that never frees what it allocates (entities accumulate for the whole run — there's nothing to reuse). Its very first batch pays a one-time "cold start" cost that later batches don't; with only one measured repetition, `1_comp`'s number above still carries some of that cold-start weight rather than the pure steady-state cost the other rows settle into. For a stable, comparable number, run with a fixed iteration count and discard the first repeat (see [How to benchmark](#how-to-benchmark)).

### Query.All (full archetype scan, SoA chunks)

| Components | 2¹⁰ ns/ent | 2²⁰ ns/ent |
| :--- | ---: | ---: |
| 0 (entity only) | 0.344 | 0.323 |
| 1 | 0.516 | 0.524 |
| 2 | 0.525 | 0.685 |
| 3 | 0.721 | 1.021 |
| 4 | 0.793 | 1.127 |
| 5 | 0.968 | 1.336 |
| 6 | 1.182 | 1.621 |
| 7 | 1.458 | 2.066 |
| 8 | 1.723 | 2.496 |
| 9 | 1.838 | 2.975 |
| 10 | 1.919 | 3.409 |

At higher component counts, 2²⁰ runs roughly 1.4–1.8× slower per entity than 2¹⁰ — the working set (10 components × ~1M rows) no longer fits in cache, so chunk-to-chunk pointer hops stop being free. Both scales remain branch-free, zero-allocation, linear in component count.

### Query.Pick (per-entity subset iteration, 100 of 1,024 entities)

`sorted` = first 100 entities in creation order. `random` = 100 entities randomly sampled from the full population, in random order (cache-unfriendly, jumps between memory locations).

| Components | sorted ns/ent | random ns/ent |
| :--- | ---: | ---: |
| 0 | 3.617 | 3.585 |
| 1 | 4.665 | 4.675 |
| 2 | 4.969 | 4.977 |
| 3 | 6.526 | 6.511 |
| 4 | 6.597 | 6.670 |
| 5 | 7.617 | 7.654 |
| 6 | 8.056 | 8.139 |
| 7 | 9.009 | 9.028 |
| 8 | 9.555 | 9.518 |
| 9 | 10.58 | 10.63 |
| 10 | 11.13 | 11.18 |

### Query.Seek (single-entity access, 100 of 1,024 entities)

New section — `Seek` resolves an entity's address directly through the index (no mask filtering), independent of `Query`'s include/exclude mask. `default` = entities in creation order. `random` = randomly sampled, same technique as `Pick`'s random column.

| Components | default ns/ent | random ns/ent |
| :--- | ---: | ---: |
| 0 | 2.949 | 3.044 |
| 1 | 4.012 | 4.029 |
| 2 | 4.370 | 4.392 |
| 3 | 5.962 | 5.989 |
| 4 | 6.011 | 6.123 |
| 5 | 7.093 | 7.284 |
| 6 | 7.649 | 7.690 |
| 7 | 8.677 | 8.863 |
| 8 | 9.334 | 9.184 |
| 9 | 10.28 | 10.44 |
| 10 | 10.61 | 10.67 |

### Entity Lifecycle

| Operation | ns/op | B/op | Technical Mechanism |
| :--- | ---: | ---: | :--- |
| **Remove** (population 100,000) | 3.214 | 0 | Swap-and-pop + index recycling |
| **Stability (interleaved create/destroy under growth)** | 25.41 | 33 | Generation-based ID recycling under churn |

### Key Technical Takeaways
* **Migration cost scales with archetype width, not edit size.** `Add` (from a 1-component anchor) stays under 100 ns through 10 added components; `Del` (from a 10-component archetype) starts above 80 ns even for removing just one — because vacating the source row touches every column the source tracks, regardless of how many are being changed. See [`internal/ent/editor.go`](internal/ent/editor.go).
* **Structural per-entity cost is scale-independent; allocation behavior is not.** `Add`/`AddTags`/`Del`/`Mix` cost about the same per entity at 1,024 and at 1,048,576 — but the latter scale necessarily performs real allocation work (growing/shrinking an archetype spanning the whole population), while the former cycles between two already-sized archetypes and reports 0 allocs/op thanks to the `chunk.Pack` `spare`-reuse mechanism.
* **`Query.All` chunk-hopping overhead becomes visible past cache size.** Per-entity cost roughly doubles from 1,024 to 1,048,576 entities at higher component counts — a known trade-off of the chunked SoA layout, not a regression.
* **Sorted vs random access:** `Pick`/`Seek` show only minor variance between sequential and randomly-sampled entity order, indicating the hot path is dominated by direct record lookup and pointer arithmetic rather than access locality.
* **Zero allocations on all query paths.** `Query.All`/`Pick`/`Seek` report 0 B/op and 0 allocs/op across every component count and scale tested.

## Benchmark Comparison with Other ECS Libraries

Cross-framework benchmarks are maintained in a dedicated project:
**[go-ecs-benchmarks](https://github.com/mlange-42/go-ecs-benchmarks)** by [@mlange-42](https://github.com/mlange-42).

This project compares GOKe against other Go ECS implementations (Arche, Donburi, Ento, etc.) on a unified workload, eliminating bias from differently-shaped local benchmarks.

> ⚠️ **Check the compared versions.** The results published in go-ecs-benchmarks may lag behind GOKe's main branch. Before drawing conclusions, verify which GOKe version (tag) is used in the comparison and re-run the suite against the version you care about.

## How to benchmark
```bash
go test -bench=. ./... -benchmem
```

For `Factory_Create` specifically, the default time-based calibration (`-benchtime=1s`) picks a different iteration count on every run — and since that benchmark never frees what it allocates, its per-op cost actually depends on how many iterations were run. For comparable, reproducible numbers (e.g. across machines or commits), pin the iteration count explicitly and discard the first (cold-start) repeat:

```bash
go test -bench='^Benchmark_Factory_Create$' -benchmem -benchtime=50000x -count=5 ./bench/...
```
