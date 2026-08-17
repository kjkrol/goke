# Examples

> ⚠️ **Setup Required**: Examples are managed as isolated modules to keep the core ECS engine free of GUI dependencies. Before running any example, initialize the workspace:
> ```bash
> make setup
> ```

* [**Mini Demo**](./mini-demo/main.go) – The minimalist starter.
* [**Simple Demo**](./simple-demo/main.go) – A slightly more advanced introduction to the ECS lifecycle.
* [**Parallel Demo**](./parallel-demo/main.go) – **Advanced showcase**:
  * Coordination of multiple systems.
  * Concurrent execution using `RunParallel`.
  * Handling structural changes via **Command Buffer** and explicit **Sync points**.

For a full graphics/physics integration example — real-time rendering via
[Ebitengine](https://ebitengine.org/), spatial management via [GOKg](https://github.com/kjkrol/gokg),
collision detection and resolution — see the companion repository
[**gokebiten**](https://github.com/kjkrol/gokebiten) and its
[collision-demo](https://github.com/kjkrol/gokebiten/tree/main/examples/collision-demo).
