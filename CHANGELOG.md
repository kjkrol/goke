# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.1.0] - 2026-08-21

### Added ✨
* **`ECS.Pause()`/`Resume()`/`Paused()`** — a stop-the-world toggle for the scheduler. `Tick` panics while paused; `Pause`/`Resume` are idempotent.
* **`ECS.Save(path)`/`Load(path, comps...)`** — round-trip the entire world (component definitions, archetype layout, entity data, and the entity ID pool itself) to a gzip-wrapped file, preserving exact `uid.UID64` values across the trip. `Save` requires the ECS to be paused first.
* **`LoadComp[T]()`/`CompProvider`/`ProvidedComps(...)`** — `Load` matches components against the file by type name, not by registration order, so callers don't need to know a dependency's internal `RegComp` order. `CompProvider` lets a system publish its own components' load tokens (`LoadComps() []CompToken`) so a composing program doesn't need to name them directly either.
* **`Module`/`ECS.RegModule(m)`** — bundles a set of related systems and the components they own behind one value. `RegModule` registers all of a module's systems in one call; `Module.RunPlan(ctx, d)` replays them, from inside your own `SetPlan` closure, in the order and with the `Sync` points the module requires — so composing code doesn't need to know a module's internal system ordering.
* **`examples/module-demo`** — two independent `Module`s (each owning its own components and systems, unaware of each other) composed together via `RegModule` + `Setup` + `SetPlan`.

### Changed
* **`RegComp[T]()` now validates that `T` is encodable at registration time.** A component (recursively, including its fields) must be a bool, numeric kind, string, struct, fixed-size array, or a type implementing both `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` — otherwise `RegComp` panics immediately. This is a general property of the storage model (components with pointers/slices/maps/interfaces/chans/funcs already paid an off-chunk indirection cost during iteration), now enforced up front rather than silently allowed. `RegComp` also logs a one-time warning per type for fields that resolve through `string` or `BinaryMarshaler`, since accessing them means a jump outside the contiguous column chunk.
* **`RegComp[T]()`'s encodability check also rejects unexported struct fields.** Previously these passed registration silently but panicked deep inside `Save`/`Load`'s reflect-based encoder, which cannot read an unexported field regardless of its type — now caught immediately, at registration, with a clear message.
* **`RegComp[T]()` rejects a type whose `BinaryMarshaler`/`BinaryUnmarshaler` comes only from Go's method promotion of an embedded field, when the type also declares other fields of its own.** Go promotes an embedded type's methods to the outer type, so such a type was silently treated as fully covered by the embedded type's marshaler — dropping its own fields on every `Save`. A struct whose only field is the embedded type is unaffected (nothing else to lose).
* **`Module` now embeds `SetupProvider`** — every `Module` implementation must supply `SetupSystems()` (returning `nil` is fine) alongside `RegSystems`/`RunPlan`/`LoadComps`, instead of `SetupProvider` being an easy-to-forget, orthogonal interface. `SetupProvider` remains usable standalone for values with nothing to run every tick.
* **`examples/simple-demo`**: `Order.ID` changed from `string` to `int` — the simplest example in the repo no longer demonstrates the now-documented off-chunk field cost as a default practice.
* **Requires Go 1.27.0** (stable), up from the `1.27rc3` pre-release pin.

### Fixed 🐛
* **Chunk memory holding a `string` or `time.Time`-shaped (`BinaryMarshaler`-backed) component field is now GC-safe.** Such fields used to live in memory the garbage collector permanently classifies as pointer-free (`noscan`), so a pointer written there was invisible to the GC and its target could be collected while still referenced only from chunk memory — reproduced live and deterministically under forced GC pressure. Archetypes containing such a field are now allocated as GC-scanned memory instead, with no change to offset arithmetic or copy logic elsewhere.

### Documentation 📚
* Reworded the README Features table's persistence and module rows to describe what they let you do rather than which methods to call: persistence spells out what "the world's state" is (component types, archetypes, entities with IDs preserved) and how the narrow pointer-backed field case is handled; module composition is framed as packaging components and systems for reuse via `Setup`/`Plan`.
* `examples/persistence-demo` simplified to a plain `Save`/`Load` round-trip with directly registered components — the module-composition pattern it used to demonstrate (`CompProvider`/`ProvidedComps`) now lives in `examples/module-demo` instead, so each example covers one feature.
* Benchmarks are now tracked on an Intel i5-8265U instead of the original Apple M1 Max. `BENCHMARKS.md` and the README's Scalability Validation section were rebuilt around fresh, single-population (1,024 entities) numbers from that machine, replacing the old M1 Max figures outright rather than keeping both around. Also documented three benchmarked operations that existed in `bench/` but were never written up: `Query.SeekH`, `ValueEditor.Add`, and bulk `Remover.Remove`.
* Fixed a noisy `Factory.Create`/`comp=1` benchmark number that was carrying a one-time cold-start allocation cost from a single-sample measurement (was 59.41 ns/ent, 108,212 B/op) — re-measured with `-benchtime=50000x -count=10`, discarding the one cold-start outlier run: 12.71 ns/ent, 16,497 B/op, consistent with the rest of the table.

## [3.0.0] - 2026-08-17

### Breaking Changes ⚠️
* **Module path changed to `github.com/kjkrol/goke/v3`** (required by Go modules for a breaking major release). Update your import: `import "github.com/kjkrol/goke/v3"`, and `go get github.com/kjkrol/goke/v3`.
* **`Query`, `Editor`, `ValueEditor`, and `Factory` can no longer be constructed directly on `*ECS`.** `ECS.NewQueryBuilder`/`ECS.NewFactory`/`ECS.NewEditorBuilder`/`ECS.NewValueEditorBuilder` are all removed. Construction now flows exclusively through systems:
  * `SysInit` — passed to `System.Init`, or obtained via the new one-time `ecs.Setup` — is the only way to get `NewQueryBuilder`/`NewFactory`.
  * `Editor`/`ValueEditor` are built from an already-obtained `*Query` instead: `ecs.NewEditorBuilder(comps...)` → `query.NewEditorBuilder(comps...)` (same for `NewValueEditorBuilder`). This reflects that both apply to the entities `query` matches.
  * Using an already-built `Query`/`Factory`/`Editor`/`ValueEditor` (`.All`/`.Seek`/`.Create`/`.Next`/etc.) is unaffected — the restriction is on construction, not use.
* **`System.Init` signature changed**: `Init(*ECS)` → `Init(*SysInit)`.
* **`SystemFn` is no longer a bare function type.** `type SystemFn func(*CmdBuf, time.Duration)` is now a struct — `SystemFn{OnInit func(*SysInit), OnUpdate func(*CmdBuf, time.Duration)}` — set as a composite literal instead of a function value, and implementing `System` directly. **`ECS.RegSysFn` is removed**; register a functional system with `ecs.RegSys(goke.SystemFn{OnUpdate: fn})` instead — one registration entry point for both stateful and functional systems.
* **`Editor`'s semantics changed from immediate, single-entity mutation to deferred, batch mutation.** The `NewEditorBuilder(comps...).Delete(...).Build()` → apply-per-entity pattern from 2.x is gone. `Editor` (built via `query.NewEditorBuilder(comps...).Remove(...).Build()`) now migrates whole chunks captured during `Query.All` iteration in one pass: stage ids via `query.BeginMigrate(cb)`/`MigrateBuf.Add(id)`, then `MigrateBuf.Commit(editor)` — the batch applies at the next `Sync`, with block column copies and deferred hole compaction instead of per-entity moves.
* **`Del[T]()` renamed to `Remove[T]()`** as the verb for structural removal, matching `EditorBuilder.Remove(...)`.
* **`RegComp[T]()`, `AddOne[T]()`, and `CmdBufAddCompValue[T]()` are now generic methods instead of free generic functions**: `goke.RegComp[T](ecs)` → `ecs.RegComp[T]()`; `goke.AddOne[T](cb, ...)` → `cb.AddOne[T](...)`; `goke.CmdBufAddCompValue[T](cb, vm, ...)` → `cb.AddCompValue[T](vm, ...)`. Requires **Go 1.27rc3+** ([golang/go#77273](https://github.com/golang/go/issues/77273), generic methods) — `go.mod`/`go.work` bump to `go 1.27rc3` accordingly. This RC pin will move to the stable `1.27` release once it ships.
* **`ECS.RemoveEnt` removed.** It mutated the registry immediately and synchronously, bypassing the `CmdBuf`/`Sync` pipeline every other structural change goes through — a hazard if called mid-tick (e.g. from a system that stashed `*ECS` via `SysInit.ECS()`), especially under `RunParallel`. Use `cb.RemoveOne(id)` from inside a system's `Update` instead (queued, applied at the next `Sync`); for bulk removal see `Remover` below.
* **`MigrateBuf.CommitRemove()` removed** — construct a `Remover` explicitly (`SysInit.Remover()`) and call the existing `MigrateBuf.Commit(remover)` instead, same as `Editor`/`ValueEditor`. One mechanism for all three.

### Added ✨
* **`ECS.Setup(systems ...System)`** — runs each given system exactly once, in order: `Init`, then `Update`, then `Sync`, fully applying one system's effects (including deferred structural changes) before the next one's `Init` runs. The sanctioned way to do one-time world seeding — a later system in the list can query and edit entities spawned by an earlier one. Callable only once per `ECS` lifetime; `Reset` starts a new one and allows another `Setup` call.
* **`Remover`** — bulk, whole-entity removal sharing `Editor`'s batch mechanism. Construct once via `SysInit.Remover()` (a plain getter, not `New*` — it carries no query/spec, applying uniformly to any archetype, so every call returns the same shared instance), then `query.BeginMigrate(cb)`/`MigrateBuf.Add(id)`/`MigrateBuf.Commit(remover)`.
* **`ValueEditor`** — `Editor`'s value-carrying counterpart: adds exactly one component and lets the caller write a per-entity value for it (`CmdBufAddCompValue`), computed during the same `Query.All` iteration that stages the batch.
* **`Factory.SpawnAll(count) []uid.UID64`** — spawns `count` entities in one call, driving `Create`/`Next` internally, for callers that only need the resulting IDs.
* **`SysInit.ECS()`** — an escape hatch back to `*ECS` from inside `Init`, for calling `RegComp` lazily (idempotent, safe to call repeatedly) where a component type is only known at first use rather than registered upfront.

### Removed
* **`examples/ebiten-demo`** — the Ebitengine integration layer (the `gokebiten` package: `Game`, `Resources`, input handling, kinematics, collision detection/resolution, rendering) and its collision-simulation demo have moved to their own repository, [**gokebiten**](https://github.com/kjkrol/gokebiten) (demo at `examples/collision-demo`), so the integration can evolve and version independently of goke's core. See the Real-World Example section of the README.

### Internal
* `internal/migration` merged into `internal/ent` — `Editor`/`Remover`/`ValueEditor` now live alongside `Manager`/`Factory` in one package. The two were always siblings at the same dependency layer with zero type-level coupling between them, so this is a pure consolidation.

## [2.1.0] - 2026-06-25

### Breaking Changes ⚠️
* **`Query` is no longer a type alias for the internal `Matcher`** — it's now a struct defined in the public `goke` package, wrapping the internal type behind an unexported field. `query.Cursor` (field access) → `query.Cursor()` (method call); same for `query.Entity` → `query.Entity()` and `query.Idx` → `query.Idx()`.
* **`Query.Add`/`Query.Get`/`Query.EntityIndex` removed.** These were never intended as public API — they leaked onto `Query` as a side effect of the old alias promoting an embedded internal type's methods/fields. The new wrapper only forwards the methods documented on `Query`.

### Added ✨
* **`Query.SeekH`** ("Seek homogeneous") — a per-entity fast path for `Seek`, for batches of entities already known to be alive and to share one archetype (established by a prior `Seek` call). Skips the alive check and the archetype-change check on every call. Returns `false` if an entity's archetype turns out to differ from the cached one — the `Cursor` must not be used in that case; call `Seek` for that entity instead.

### Performance 🚀
`Query`'s methods are now defined directly in the public `goke` package instead of relying on a type alias to an `internal/` type — Go never inlines `internal/` package methods into code outside the module that defines them, so under the old alias, every `Query` method call from user code was a real, non-inlined function call, no matter how small the method. Moving the method bodies into `goke` (as thin forwarders to the internal type) lets the whole call chain inline into the caller, even across a module boundary. This is most visible on per-entity hot paths: `Seek`/`SeekH` random single-entity access went from roughly 1.6x slower than a comparable Go ECS library to performance in the same class on a 1M-entity random-access benchmark. Chunk-level iteration (`Query.All`) was already amortizing per-call overhead over hundreds–thousands of entities per chunk and shows no measurable change.

### Internal
* `aliases.go` split by responsibility: `Query` + `QueryBuilder` + `Include`/`Exclude` → `query.go`; `Editor`-related types + `Del` → `editor.go` (renamed from `builders.go`); `Comp[T]`/`Trackable`/`Addable` → `comp.go`.
* `internal/addr.Index.GetUnchecked` and `internal/query.Matcher.SeekH` added as the internal primitives backing `Query.SeekH`.

## [2.0.0] - 2026-06-23

A full rewrite of the public API and the internal engine. The package-level free-function API (`goke.Tick(ecs, d)`, `goke.RegisterSystem(ecs, sys)`, ...) is replaced by methods on `*ECS`; the generated `View0`–`View10`/`Blueprint1`–`Blueprint10` API is replaced by a unified `Query`/`Factory`/`Editor` API built on typed components (`Comp[T]`); and the monolithic internal package was decomposed into focused, independently-testable internal packages.

### Breaking Changes ⚠️
* **Module path changed to `github.com/kjkrol/goke/v2`** (required by Go modules for any major version ≥ 2 — see the [Go modules major version reference](https://go.dev/ref/mod#major-version-suffixes)). Update your import: `import "github.com/kjkrol/goke/v2"`, and `go get github.com/kjkrol/goke/v2`.
* **Free functions replaced by methods on `*ECS`.** `goke.RemoveEntity(ecs, e)` → `ecs.RemoveEnt(e)`; `goke.RegisterSystem(ecs, sys)` → `ecs.RegSys(sys)`; `goke.RegisterSystemFunc(ecs, fn)` → `ecs.RegSysFn(fn)`; `goke.Plan(ecs, plan)` → `ecs.SetPlan(plan)`; `goke.Tick(ecs, d)` → `ecs.Tick(d)`; `goke.Reset(ecs)` → `ecs.Reset()`.
* **`View0`–`View10` and `Blueprint`/`Blueprint1`–`Blueprint10` (generated, per-arity types) removed.** Replaced by `Query` (built via `ECS.NewQueryBuilder(comps ...).Include(...).Exclude(...).Build()`) for reading, and `Factory` (built via `ECS.NewFactory(comps ...)`) for bulk creation. Component access goes through `Comp[T]` — declare one as a variable and pass `&comp` directly to either builder; it binds itself, no separate wrapping call.
* **`EnsureComponent[T]`/`SafeEnsureComponent[T]`/`GetComponent[T]`/`SafeGetComponent[T]`/`RemoveComponent` removed.** Use a `Query` for reads (`Query.Seek` + `Comp[T].At` for single-entity access, `Query.All`/`Pick` for bulk/subset) and an `Editor` (built via `ECS.NewEditorBuilder(comps ...).Delete(...).Build()`) for structural add/remove.
* **`Entity` removed** — use `uid.UID64` directly (it was always a plain alias).
* **`ComponentID`/`ComponentDesc` replaced by `CompID`.** Component lookups no longer pass around a full descriptor struct (`{ID, Type}`) — only the ID, resolved automatically by `Comp[T]` and the generic option functions. `RegisterComponent[T](ecs) ComponentDesc` → `RegComp[T](ecs) CompID`.
* **`System.Update` signature changed**: `Update(Lookup, *Schedule, time.Duration)` → `Update(*CmdBuf, time.Duration)`. The `Lookup` (read-only registry) parameter is gone — a system that needs to read entities builds and holds its own `Query` (typically created once in `Init`), the same way it would for any other read access. `Schedule` is renamed `CmdBuf`.
* **`ScheduleAddComponent[T]` → `CmdBufAddComp[T]`** (same shape, takes a `CompID` resolved via `RegComp[T]`, typically once in `System.Init` — `CmdBuf`-based writes happen outside the registry's live context, so the ID can't be resolved automatically the way `Query`/`Factory`/`Editor` do it). **`ScheduleRemoveComponent`/`ScheduleRemoveEntity` removed as free functions** — call `cb.RemoveComp(e, compID)` / `cb.RemoveEntity(e)` directly as methods on the `*CmdBuf` passed into `Update`.
* **`Include[T]()`/`Exclude[T]()` kept (same names), but their return type and the type they configure changed** — from `BlueprintOption` (configuring a `core.Blueprint`) to `Opt` (configuring a `Query`, via `.Include(...)`/`.Exclude(...)` on `NewQueryBuilder`'s result).
* **`WithInitialEntityCap`/`WithFreeIndicesCap` renamed** to `WithEntityCap`/`WithEntityFreeCap`.
* Internal package layout changed substantially (`internal/arch`, `internal/ent`, `internal/comp`, `internal/colstore`, `internal/chunk`, `internal/addr`, `internal/reg`, `internal/orch`, `internal/query` replace the previous single `internal/core` package) — not part of the public API, but relevant if you vendor or patch internals.

### Why pull, not push
`View.All`/`View.Filter` were **push iterators**: built on `iter.Seq[...]` (Go's range-over-func), the iteration loop lived inside the engine and pushed each chunk/element out to a caller-supplied `yield` callback. Calling through `yield` is an indirect call through a function value, not a static call — the compiler can't always inline it. Worse, when the caller's loop body was non-trivial ("heavy"), the inliner's failure at that one call site tended to cascade, defeating bounds-check elimination and register allocation across the whole iteration — a caller-side change to ordinary loop logic could silently degrade engine throughput, with no obvious cause from the caller's perspective. `Query.All`/`Pick` (via `Next()` + `Cursor`) is a **pull iterator** instead: the loop body lives directly in the caller's own code, calling an ordinary method (`Next()`) with no closure indirection — inlining and loop optimizations no longer depend on how much work the loop body does.

### Added ✨
* **`Editor`** — explicit, reusable structural-edit handle (built via `NewEditorBuilder`), replacing the old `EnsureComponent`/`RemoveComponent` free-function calls with a single batched migration per `Update`.
* **`Query.Seek`** — direct single-entity resolution, independent of the Query's include/exclude mask, with per-archetype table/offset caching for repeated seeks within the same archetype. Replaces `GetComponent`/`SafeGetComponent`.
* **`Trackable`/`Addable`** — sealed interfaces (unexported method, same pattern as `testing.TB`) satisfied only by `*Comp[T]`, letting `NewQueryBuilder`/`NewFactory`/`NewEditorBuilder` accept components directly instead of a separate wrapping call.
* Comprehensive unit test coverage added across `internal/addr`, `internal/arch`, `internal/chunk`, `internal/colstore`, `internal/comp`, `internal/ent`, `internal/orch`, `internal/query`, `internal/reg`, the root `goke` package, and `iter` — all now at 96–100% statement coverage (up from several packages at 0%).
* New benchmark families: `Benchmark_Editor_Mix` (combined add+remove in one migration), `Benchmark_Remove`, `Benchmark_Stability_Grow`, `Benchmark_Matcher_Seek`.

### Note on `RegComp`
`RegComp[T]` is still needed, but only for one case: `CmdBufAddComp[T]` (used inside a `System.Update`, which has no live access to the registry) needs a pre-resolved `CompID`, typically captured once via `RegComp[T]` in `System.Init`. Every other entry point (`NewFactory`, `NewQueryBuilder`, `NewEditorBuilder`, `Include`/`Exclude`/`Del`) resolves component types automatically — no manual registration required.

### Performance 🚀
`chunk.Pack` keeps one spare chunk on hand after a shrink and reuses it on the next growth, instead of always allocating fresh memory. As a result, `Editor.Add`/`Del`/`Mix` (2–10 components, population 1,024) report **0 allocs/op at steady state** — measured directly on this release, not as a delta against v1.3.4, which used a different internal storage implementation. The 1-component cases for `Add`/`Del` still allocate once per call (the migration crosses more than one chunk boundary in a single step, exceeding the single-slot `spare` cache), a known, accepted limit of a one-deep reuse cache.

See [BENCHMARKS.md](./BENCHMARKS.md) for the full current per-component-count numbers (Apple M1 Max) across `Editor.Add/AddTags/Del/Mix`, `Factory.Create`, `Query.All/Pick/Seek`, and entity lifecycle operations — including the new `Mix` and `Seek` sections that didn't exist in earlier versions of that document.

## [1.3.0] - 2026-06-07

### Breaking Changes ⚠️
* **`View.All()` returns SoA pages directly.** Replaced the previous `Values()` / `Head` / `Tail` pattern with a single `iter.Seq[struct{Entity []Entity, Comp1 []T1, ...}]` that yields chunk-shaped slices over native memory. The inner loop is now on the caller side, exposing full SoA layout for SIMD-friendly access patterns and aggressive compiler inlining.
* **`View.Filter(selected)` redesigned**: now returns `iter.Seq2[int, struct{Entity, Comp1, ...}]`. The index is the position in the input `selected` slice — callers can identify which entities were skipped (not matching the view, or already removed) and correlate results with parallel side-tables without maintaining a manual counter.
* **`Blueprint.Create(n)` is now batch-based**: returns `iter.Seq[chunk]` where each yielded chunk exposes typed slices (`Entity[]`, `Comp1[]`, ...) for direct in-place initialization. Replaces the previous single-entity `Create()` returning a struct of pointers.
* **`GetComponent[T]` returns `*T` directly** (without error). Use `SafeGetComponent[T]` for the error-returning variant with reflection-based type validation.

### Added ✨
* **`View0.Filter`**: membership-only iteration (no components) — useful for "does entity belong to this query?" checks against the archetype mask (`Include`/`Exclude`).
* **`View9` and `View10`**: type-safe queries extended from 8 to 10 simultaneous components. Same extension for `NewBlueprint9` / `NewBlueprint10`.

### Performance 🚀
* **`Get Component`**: 5.13 → **4.70 ns/op** (**−8.4%**) via:
    * Sentinel errors (no `fmt.Errorf` allocations on the error path)
    * Inlined `EntityLinkStore.Get`, `Archetype.GetColumn`, `Memo.GetPage`
    * Single `entity.Unpack()` instead of separate `Index()` + `Generation()` calls
    * `len()` instead of `cap()` for slice bounds (helps Go BCE)
* **`View.Filter` baseline**: 4.22 ns/entity for 1-component view, scaling linearly to ~11 ns/entity at 10 components — all zero-allocation.
* **`View.All` baseline**: 0.34 ns/entity for `View0`, up to ~2 ns/entity at 10 components.

## [1.2.3] - 2026-02-17

### Architecture & Memory 🏗️
* **Chunked SoA (Memory Paging):** Transition from monolithic slices to fixed-size memory pages (aligned to L1 Cache).
    * **Stable Growth:** Memory allocation is now linear and "stepless". This eliminates the massive latency spikes caused by copying millions of elements when a slice capacity is exceeded.
    * **Trade-off:** While "Hot Cache" iteration is significantly faster, the new architecture introduces overhead during iteration over massive datasets (>1M entities) due to pointer indirection between memory chunks.

### Performance 🚀
* **Hot Cache Iteration (1k entities):** Significant throughput increase for datasets fitting in L1/L2 cache.
    * **View Iteration:** Performance improved by **~25%**.
    * **Values Iteration:** Consistent **~25%** speedup across component counts.
* **Direct Access:** `Get Component` optimized to **4.49 ns**, providing near-pointer-access speeds.
* **Structural Stability:** Operations like `Create Entity` and `Add Component` remain stable with a slight performance edge (**~3-5%** improvement).

<details>
  <summary>Click to see full benchstat results (Apple M1 Max)</summary>

#### 1. Low Scale (1k Entities - Hot Cache)
| Task | Before (Monolithic) | After (Chunked) | Δ % |
| :--- | :--- | :--- | :--- |
| **View0_All** (Entity Only) | 360.4 ns | 360.8 ns | **~0.0%** |
| **View3_All** (3 Comps) | 726.7 ns | 538.4 ns | **-25.91%** |
| **View8_All** (8 Comps) | 752.8 ns | 538.3 ns | **-28.49%** |
| **View3_Values** | 720.4 ns | 539.8 ns | **-25.07%** |
| **Get Component** | ~5.50 ns | **4.49 ns** | **-18.00%** |
| **Create Entity (4 comps)** | 29.99 ns | 28.72 ns | -4.23% |
| **Query (Filter 100)** | 141.3 ns | 138.7 ns | -1.84% |

#### 2. High Scale Stress Test (10M Entities)
| Metric | Previous (Monolithic) | Current (Chunked) | Impact |
| :--- | :--- | :--- | :--- |
| **Throughput** | ~0.78 ns/entity | ~2.01 ns/entity | Slower (Chunk hopping overhead) |
| **Memory Alloc** | Exponential (Copy spikes) | **Linear (Stepless)** | **Stable Latency** |

</details>

## [1.2.2] - 2026-02-13

### Performance 🚀
* **Structural Improvements:** Znaczące przyspieszenie operacji cyklu życia encji dzięki optymalizacji indeksowania w Arche.
    * **Create Entity:** Tworzenie encji z komponentami jest teraz szybsze o **36% do 46%**.
    * **Migrate Component:** Dodawanie kolejnych komponentów (migracja struktury) przyspieszyło o **38.22%**.
    * **Add Tag:** Operacja dodawania tagów jest teraz o **21.28%** wydajniejsza.
* **Known Regressions:** Zmiany architektoniczne wprowadziły niewielkie opóźnienia w `Remove Component` (+11.04%) oraz `Get Component` (+40.81%).

<details>
  <summary>Click to see full benchstat results (Apple M1 Max)</summary>

| Task | Before | After | Δ % |
| :--- | :--- | :--- | :--- |
| **Create Entity with 1 Component** | 38.42 ns | 21.31 ns | **-44.53%** |
| **Create Entity with 2 Components** | 38.56 ns | 23.29 ns | **-39.60%** |
| **Create Entity with 3 Components** | 46.03 ns | 24.87 ns | **-45.97%** |
| **Create Entity with 4 Components** | 42.01 ns | 26.84 ns | **-36.11%** |
| **Migrate Component** | 60.68 ns | 37.49 ns | **-38.22%** |
| **Add Tag** | 47.60 ns | 37.47 ns | **-21.28%** |
| **Remove Component** | 10.69 ns | 11.87 ns | +11.04% |
| **Get Component** | 5.462 ns | 7.691 ns | +40.81% |

</details>

## [1.2.1] - 2026-02-11

### Performance 🚀
* **Significant performance boost.** The overall execution time (geomean) decreased by **19.22%**.
* **Filtering optimizations:** Operations like `Filter100` are now over **50% faster** (e.g., `View0_Filter100` dropped from ~288ns to ~134ns).
* **View rendering:** Standard view operations (`ViewX_All`) saw performance improvements ranging from **12% to 15%**.
* **Values extraction:** Most `Values` operations improved by roughly **11% to 17%**.

<details>
  <summary>Click to see full benchstat results (Apple M1 Max)</summary>

#### View & Query Operations
| Task | Before | After | Δ % |
| :--- | :--- | :--- | :--- |
| **View0_All** | 346.0 ns | 339.2 ns | -1.95% |
| **View1_All** | 515.0 ns | 439.6 ns | -14.64% |
| **View2_All** | 567.0 ns | 482.9 ns | -14.82% |
| **View3_All** | 728.1 ns | 639.5 ns | -12.17% |
| **View3WithTag_All** | 692.2 ns | 600.7 ns | -13.22% |
| **View4_All** | 728.4 ns | 639.9 ns | -12.14% |
| **View5_All** | 729.9 ns | 638.4 ns | -12.53% |
| **View6_All** | 728.9 ns | 634.3 ns | -12.98% |
| **View7_All** | 727.8 ns | 634.1 ns | -12.88% |
| **View8_All** | 730.4 ns | 634.3 ns | -13.16% |
| **View0_Filter100** | 288.2 ns | 134.2 ns | **-53.44%** |
| **View3_Filter100** | 466.9 ns | 232.7 ns | **-50.17%** |
| **View1_Values** | 400.5 ns | 332.8 ns | -16.92% |
| **View2_Values** | 477.4 ns | 401.2 ns | -15.97% |
| **View3_Values** | 640.2 ns | 555.8 ns | -13.19% |
| **View4_Values** | 634.3 ns | 554.3 ns | -12.61% |
| **View5_Values** | 635.1 ns | 554.6 ns | -12.67% |
| **View6_Values** | 635.5 ns | 553.9 ns | -12.84% |
| **View7_Values** | 634.4 ns | 642.5 ns | ~ |
| **View8_Values** | 635.4 ns | 564.7 ns | -11.13% |
| **View3_FilterValues100** | 471.9 ns | 230.4 ns | **-51.17%** |
| **View2_Filter100** | - | 197.3 ns | n/a |
| **GEOMEAN** | **573.2 ns** | **445.4 ns** | **-19.22%** |

</details>
