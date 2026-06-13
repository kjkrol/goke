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
* [**Ebiten Demo**](./ebiten-demo/main.go) – **Graphics Integration & Spatial Physics**:
  * Real-time rendering using [Ebitengine](https://github.com/kjkrol/gokg).
  * High-performance spatial management using [GOKg](https://github.com/kjkrol/gokg).
  * Custom physics pipeline: **Velocity Inversion** is processed strictly before **Position Compensation** to ensure boundary stability.
  * **Note**: Run `make` inside the example directory to fetch dependencies and start the demo.
