# Examples

> ⚠️ **Setup Required**: Examples are managed as isolated modules to keep the core ECS engine free of GUI dependencies. Before running any example, initialize the workspace:
> ```bash
> make setup
> ```

* [**Mini Demo**](./mini-demo/main.go) – The minimalist starter.
* [**Simple Demo**](./simple-demo/main.go) – A slightly more advanced introduction to the ECS lifecycle.
* [**Single-Entity Demo**](./single-entity-demo/main.go) – When to reach for `CmdBuf.AddOne`:
  entities named directly by an external event, with no `Query` iteration to batch through.
* [**Bulk Edit Demo**](./bulk-edit-demo/main.go) – The complementary case: batching an
  add-component and a remove-entity change across a whole matched population via
  `Editor`/`Remover` + `Query.BeginMigrate`/`Commit`, one migration per chunk instead of one
  call per entity.
* [**Parallel Demo**](./parallel-demo/main.go) – **Advanced showcase**:
  * Coordination of multiple systems.
  * Concurrent execution using `RunParallel`.
  * Handling structural changes via **Command Buffer** and explicit **Sync points**.
* [**Module Demo**](./module-demo/main.go) – Composing self-contained modules:
  * Two independent `Module`s, each owning its own components and systems, wired up without
    knowing each other's internals.
  * `ECS.RegModule` + `Setup`/`SetPlan` — one-time world seeding vs. per-tick simulation.
* [**Persistence Demo**](./persistence-demo/main.go) – Saving and loading a world to/from a file:
  * `Pause`/`Save`/`Load`, with entity IDs preserved across the round trip.

For a full graphics/physics integration example — real-time rendering via
[Ebitengine](https://ebitengine.org/), spatial management via [GOKg](https://github.com/kjkrol/gokg),
collision detection and resolution — see the companion repository
[**gokebiten**](https://github.com/kjkrol/gokebiten) and its
[collision-demo](https://github.com/kjkrol/gokebiten/tree/main/examples/collision-demo).
