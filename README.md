# GOKE

*aka "Golang kjkrol ECS"*

**GOKE** is a lightweight, high-performance, and type-safe Entity Component System (ECS) for Go. It leverages modern **Go 1.23+ Iterators** to provide a clean, idiomatic API while maintaining maximum execution speed by eliminating reflection and type assertions from the main loop.

## Key Features

* **Zero Reflection in Update**: Using `CachedQuery`, component lookups and bitmask calculations happen once during initialization, not every frame.
* **Modern Iterators**: Uses `for range` over functions, allowing the Go compiler to perform loop inlining and optimizations.
* **True Type Safety**: Fully powered by Go Generics. No more casting `interface{}` or `any` inside your systems.
* **Minimalist API**: A focused set of tools (`Registry`, `Engine`, `System`) that stays out of your way.

## Installation

Goke requires **Go 1.23** or newer.

```bash
go get github.com/kjkrol/goke
```

## Performance Design

Goke is designed to be as close to "manual" loop performance as possible:

1. **Bitmask Filtering**: Entities are filtered using ultra-fast bitwise operations on uint64 slices.
2. **Storage Caching**: Pointer maps for components are fetched once per system update, avoiding map-of-interface lookups inside the entity loop.
3. **Heap Escape Prevention**: By using specialized Row structs and iterators, Goke helps the Go compiler keep variables on the stack, reducing Garbage Collector pressure.
4. **No Indirection**: Unlike function-pointer based queries, CachedQueryN uses direct method calls, enabling the Go compiler to inline the iteration logic.

## Full Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/kjkrol/goke/pkg/ecs"
)

type Order struct { ID string; Total float64 }
type Status struct { Processed bool }
type Discount struct { Percentage float64 }

type BillingSystem struct {
    query *ecs.CachedQuery3[Order, Status, Discount]
}

func (s *BillingSystem) Init(reg *ecs.Registry) {
    s.query = ecs.NewQuery3[Order, Status, Discount](reg)
}

func (s *BillingSystem) Update(reg *ecs.Registry, d time.Duration) {
    for _, row := range s.query.All() {
        // V1=Order, V2=Status, V3=Discount
        row.V1.Total *= (1 - row.V3.Percentage/100)
        row.V2.Processed = true
    }
}

func main() {
    engine := ecs.NewEngine()

    // Create entity and assign components
    entity := engine.CreateEntity()
    ecs.Assign(engine, entity, Order{ID: "ORD-99", Total: 200.0})
    ecs.Assign(engine, entity, Status{Processed: false})
    ecs.Assign(engine, entity, Discount{Percentage: 20.0})

    // Register and run systems
    engine.RegisterSystems([]ecs.System{&BillingSystem{}})
    
    fmt.Printf("Initial Total: %v\n", ecs.Get[Order](engine, entity).Total)
    engine.UpdateSystems(time.Second)
    fmt.Printf("Discounted Total: %v\n", ecs.Get[Order](engine, entity).Total)
}
```