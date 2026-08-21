# ⏱️ GOKe Benchmarks

[← Back to README](./README.md)

> Detailed performance analysis and hardware specifications for the GOKe ECS engine.

## Environment

Earlier published results were captured on an Apple M1 Max. Benchmarks are now run on a
different, weaker machine — all numbers below reflect that:
- **CPU:** Intel(R) Core(TM) i5-8265U CPU @ 1.60GHz
- **Go Version:** 1.27.0
- **OS:** Linux

## Scalability Validation

All operations below are benchmarked at a population of **1,024** entities, except `Remove`
(100,000) and `Factory.Create` (accumulates to 1,024 across the run). A wider-population (2²⁰)
sweep isn't currently run on this machine.

## Performance Metrics

### Structural Operations — Editor

`Editor.Update` migrates an entity to a new archetype in one move. Its cost is driven by the **width of the source and destination archetypes**, not by how many components the edit itself changes (see [`internal/ent/editor.go`](internal/ent/editor.go) for the full explanation) — that's why `Add` (starting from a 1-component anchor) stays cheap longer than `Del` (starting from a 10-component archetype), even though both add/remove the same number of components.

`subset=512 sorted`/`subset=512 random` migrate half the population, in creation order or randomly sampled (cache-unfriendly, jumps between memory locations); `subset=1,024` migrates the whole population in one pass. All entries are 0 allocs/op unless noted.

#### Add (`comp=1`, growing toward `comp=1+N`)
| N added | subset=512 sorted ns/ent | subset=512 random ns/ent | subset=1,024 ns/ent |
| :--- | ---: | ---: | ---: |
| 1 | 14.31 | 36.08 | 5.943 |
| 2 | 15.64 | 37.67 | 6.999 |
| 3 | 16.15 | 40.65 | 7.052 |
| 4 | 15.06 | 42.74 | 7.984 |
| 5 | 26.99 | 78.35 | 12.94 |
| 6 | 29.45 | 69.75 | 9.161 |
| 7 | 21.29 | 64.49 | 8.204 |
| 8 | 16.73 | 52.12 | 15.33 † |
| 9 | 24.68 | 83.11 | 18.69 † |
| 10 | 19.05 | 53.84 | 18.39 † |

† `subset=1,024` at N=8–10 allocates (65,541 B/op, 1 alloc/op) — the archetype's chunk pool has to grow past its pre-warmed size at that width; every other cell above is 0 allocs/op.

#### Add Tags (`comp=1`, growing toward `comp=1+N` zero-size tags)
| N tags | subset=512 sorted ns/ent | subset=512 random ns/ent | subset=1,024 ns/ent |
| :--- | ---: | ---: | ---: |
| 1 | 28.30 | 67.65 | 12.86 |
| 2 | 28.79 | 53.01 | 7.801 |
| 3 | 17.26 | 45.67 | 7.264 |
| 4 | 17.77 | 44.24 | 8.722 |
| 5 | 16.70 | 44.26 | 7.213 |
| 6 | 24.02 | 69.73 | 7.957 |
| 7 | 15.40 | 37.54 | 5.880 |
| 8 | 14.05 | 36.84 | 5.821 |
| 9 | 13.79 | 37.17 | 5.633 |
| 10 | 13.95 | 35.71 | 6.380 |

All cells above are 0 allocs/op — tags are zero-size, so a migration touching only tag bits never has to grow a chunk's byte columns.

#### Del (`comp=10`, shrinking by N)
| N removed | subset=512 sorted ns/ent | subset=512 random ns/ent | subset=1,024 ns/ent |
| :--- | ---: | ---: | ---: |
| 1 | 89.88 | 118.6 | 19.40 † |
| 2 | 98.66 | 121.8 | 15.22 |
| 3 | 89.78 | 109.9 | 14.95 |
| 4 | 88.39 | 115.0 | 16.03 |
| 5 | 88.00 | 108.4 | 16.04 |
| 6 | 89.80 | 112.1 | 15.46 |
| 7 | 94.51 | 106.7 | 14.66 |
| 8 | 90.49 | 101.0 | 14.22 |
| 9 | 86.60 | 90.35 | 14.11 |
| 10 | 91.09 | 86.59 | 14.23 |

† `subset=1,024` at N=1 allocates (32,775 B/op, 1 alloc/op) — the shrunk-to archetype's chunk pool grows past its pre-warmed size; every other cell above is 0 allocs/op.

#### Mix (combined add+remove in one migration, `comp=N+1 → comp=N+1`, swapping N components)
| N swapped | comp (src=dst) | subset=512 sorted ns/ent | subset=512 random ns/ent | subset=1,024 ns/ent |
| :--- | :--- | ---: | ---: | ---: |
| 1 | 2 | 14.84 | 40.07 | 6.537 |
| 2 | 3 | 14.36 | 44.92 | 6.809 |
| 3 | 4 | 59.08 | 59.29 | 10.63 |
| 4 | 5 | 59.60 | 66.87 | 10.94 |
| 5 | 6 | 62.68 | 81.40 | 11.37 |
| 6 | 7 | 72.20 | 80.03 | 9.811 |
| 7 | 8 | 76.50 | 83.47 | 10.79 |
| 8 | 9 | 83.31 | 83.44 | 14.09 |
| 9 | 10 | 90.95 | 100.1 | 14.93 |
| 10 | 11 | 98.17 | 98.59 | 13.54 |

All cells above are 0 allocs/op.

### ValueEditor.Add (add one component, writing a per-entity value)

`ValueEditor` is `Editor`'s value-carrying counterpart: unlike `Editor`, which migrates entities to an archetype and leaves the new component at its zero value, `ValueEditor` writes a caller-supplied value per entity through `CmdBuf.AddCompValue` during the same pass. Because it exists specifically to add one component with a value, there is no `comps=N` sweep to run — this is the value-carrying equivalent of `Editor.Add`'s `N=1` row, isolating the payload mechanism's overhead (arena reservation, per-run copy, slow-path compaction) from everything else.

| subset=512 sorted ns/ent | subset=512 random ns/ent | subset=1,024 ns/ent |
| ---: | ---: | ---: |
| 14.10 | 43.68 | 6.172 |

0 allocs/op across all three.

### Batch Entity Creation (`Factory.Create`)

| Components | ns/ent | B/op | Allocs |
| :--- | ---: | ---: | ---: |
| 1 | 12.71 † | 16,497 | 0 |
| 2 | 7.969 | 24,717 | 0 |
| 3 | 9.063 | 33,142 | 1 |
| 4 | 9.599 | 37,054 | 1 |
| 5 | 11.90 | 41,310 | 1 |
| 6 | 15.55 | 45,340 | 1 |
| 7 | 16.01 | 53,573 | 1 |
| 8 | 17.05 | 61,862 | 1 |
| 9 | 17.83 | 70,165 | 1 |
| 10 | 18.82 | 78,377 | 1 |

† `Factory.Create` never frees what it allocates (entities accumulate for the whole run — there's nothing to reuse), so its very first batch pays a one-time "cold start" cost later batches don't. A single-repetition measurement of `comp=1` carries that cost fully (previously measured at 59.41 ns/ent, 108,212 B/op — nearly 7× the steady-state byte cost below). Measured correctly with `-benchtime=50000x -count=10` (see [How to benchmark](#how-to-benchmark)): 1 of 10 runs was the cold-start outlier (84.05 ns/ent, 89,897 B/op), discarded; the other 9 were tight and consistent on B/op (16,497 B/op every time) with normal OS-jitter spread on ns/ent (10.18–15.67) — mean 12.71 ns/ent, median 13.11 ns/ent. The value above is that mean.

### Query.All (full archetype scan, SoA chunks)

| Components | ns/ent |
| :--- | ---: |
| 0 (entity only) | 0.3897 |
| 1 | 0.7228 |
| 2 | 0.9275 |
| 3 | 1.164 |
| 4 | 1.819 |
| 5 | 2.057 |
| 6 | 2.577 |
| 7 | 2.955 |
| 8 | 3.336 |
| 9 | 3.996 |
| 10 | 4.178 |

Branch-free, zero-allocation, linear in component count — 0 B/op and 0 allocs/op across the board.

### Query.Pick (per-entity subset iteration, 100 of 1,024 entities)

`sorted` = first 100 entities in creation order. `random` = 100 entities randomly sampled from the full population, in random order (cache-unfriendly, jumps between memory locations).

| Components | sorted ns/ent | random ns/ent |
| :--- | ---: | ---: |
| 0 | 5.761 | 6.425 |
| 1 | 9.784 | 9.161 |
| 2 | 9.228 | 8.790 |
| 3 | 11.03 | 11.39 |
| 4 | 10.41 | 11.81 |
| 5 | 11.77 | 12.61 |
| 6 | 12.69 | 13.63 |
| 7 | 14.16 | 14.92 |
| 8 | 14.70 | 16.45 |
| 9 | 15.03 | 18.65 |
| 10 | 15.70 | 19.04 |

### Query.Seek (single-entity access, 512 of 1,024 entities)

`Seek` resolves an entity's address directly through the index (no mask filtering), independent of `Query`'s include/exclude mask. `sorted` = entities in creation order. `random` = randomly sampled, same technique as `Pick`'s random column.

| Components | sorted ns/ent | random ns/ent |
| :--- | ---: | ---: |
| 0 | 4.383 | 4.766 |
| 1 | 6.882 | 7.845 |
| 2 | 7.727 | 8.022 |
| 3 | 9.395 | 10.63 |
| 4 | 9.242 | 10.43 |
| 5 | 11.20 | 12.39 |
| 6 | 11.29 | 13.07 |
| 7 | 12.69 | 14.81 |
| 8 | 13.41 | 15.72 |
| 9 | 15.60 | 18.70 |
| 10 | 15.34 | 18.54 |

### Query.SeekH (homogeneous seek fast path, 512 of 1,024 entities)

`SeekH` skips `Seek`'s generation check and archetype-change check, assuming the entity is alive and shares the archetype cached by a prior `Seek` call — the pattern for a batch of entities already known to belong to one archetype. Each iteration below seeds the cache with one `Seek`, then runs `SeekH` for the rest of the subset.

| Components | sorted ns/ent | random ns/ent |
| :--- | ---: | ---: |
| 0 | 2.272 | 2.247 |
| 1 | 4.198 | 4.528 |
| 2 | 5.041 | 5.017 |
| 3 | 7.400 | 7.771 |
| 4 | 7.598 | 8.349 |
| 5 | 8.553 | 9.905 |
| 6 | 9.184 | 11.38 |
| 7 | 11.06 | 13.43 |
| 8 | 11.34 | 14.07 |
| 9 | 12.60 | 15.98 |
| 10 | 13.60 | 16.18 |

### Entity Lifecycle

| Operation | ns/op | B/op | Technical Mechanism |
| :--- | ---: | ---: | :--- |
| **Remove** (population 100,000) | 101.7 | 0 | Swap-and-pop + index recycling |
| **Stability (interleaved create/destroy under growth)** | 167.3 | 46 | Generation-based ID recycling under churn |

### Bulk Removal (`Remover.Remove`)

The `Remove` row above times `cb.RemoveOne`'s single-entity swap-and-pop at a fixed 100,000-entity population. `Remover.Remove` is the bulk counterpart: entities are removed through the production `CmdBuf` path — a system queues one `cb.Remove` per chunk, and only the `Sync` that executes the shared `Remover`'s migration is timed. `subset=1,024` removes the whole population in one pass; `subset=512` additionally exercises source-chunk compaction.

| subset=512 sorted ns/ent | subset=512 random ns/ent | subset=1,024 ns/ent |
| ---: | ---: | ---: |
| 86.18 | 82.16 | 14.35 |

### Key Technical Takeaways
* **Migration cost scales with archetype width, not edit size.** `Del` (from a 10-component archetype) is consistently more expensive than `Add` (from a 1-component anchor) for a comparable number of changed components — vacating the source row touches every column the source tracks, regardless of how many are being changed. See [`internal/ent/editor.go`](internal/ent/editor.go).
* **Migrating the whole population (`subset=1,024`) is cheaper per entity than migrating half (`subset=512`).** A full-population pass never has to skip over surviving entities during compaction; a partial subset does, and `random` order compounds that with cache-unfriendly access.
* **Sorted vs random access:** `Pick`/`Seek`/`SeekH` show modest, consistent overhead for randomly-sampled entity order versus sequential — real, but the hot path stays dominated by direct record lookup and pointer arithmetic rather than access locality.
* **Zero allocations on every query path.** `Query.All`/`Pick`/`Seek`/`SeekH` report 0 B/op and 0 allocs/op across every component count.

## Benchmark Comparison with Other ECS Libraries

Cross-framework benchmarks are maintained in a dedicated projects:

- **[go-ecs-benchmarks](https://github.com/mlange-42/go-ecs-benchmarks)** by [@mlange-42](https://github.com/mlange-42).
- **[kjkrol/go-ecs-benchmarks](https://github.com/kjkrol/go-ecs-benchmarks)** by [@kjkrol](https://github.com/kjkrol) (should contains latest GOKe results)

These projects compares GOKe against other Go ECS implementations (Arche, Donburi, Ento, etc.) on a unified workload, eliminating bias from differently-shaped local benchmarks.

> ⚠️ **Check the compared versions.** The results published in go-ecs-benchmarks may lag behind GOKe's main branch. Before drawing conclusions, verify which GOKe version (tag) is used in the comparison and re-run the suite against the version you care about.

## How to benchmark
```bash
go test -bench=. ./... -benchmem
```

For `Factory_Create` specifically, the default time-based calibration (`-benchtime=1s`) picks a different iteration count on every run — and since that benchmark never frees what it allocates, its per-op cost actually depends on how many iterations were run. For comparable, reproducible numbers (e.g. across machines or commits), pin the iteration count explicitly and discard the first (cold-start) repeat:

```bash
go test -bench='^Benchmark_Factory_Create$' -benchmem -benchtime=50000x -count=5 ./bench/...
```
