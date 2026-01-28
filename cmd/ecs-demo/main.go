package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/pkg/ecs"
)

type (
	Order struct {
		ID    string
		Total float64
	}
	Status   struct{ Processed bool }
	Discount struct{ Percentage float64 }
)

func main() {
	engine := ecs.NewEngine()

	entity := engine.CreateEntity()

	// Direct Access approach (fastest)
	order, _ := ecs.AddComponent[Order](engine, entity)
	*order = Order{ID: "ORD-99", Total: 200.0}

	// based on unsafe.Pointer -> requires allocation on heap (slower)
	status, _ := ecs.AddComponent[Status](engine, entity)
	*status = Status{Processed: false}
	discount, _ := ecs.AddComponent[Discount](engine, entity)
	*discount = Discount{Percentage: 20.0}

	query := ecs.NewQuery3[Order, Status, Discount](engine)
	billing := engine.RegisterSystemFunc(func(reg ecs.ReadOnlyRegistry, cb *ecs.SystemCommandBuffer, d time.Duration) {
		for head := range query.All3() {
			ord, st, disc := head.V1, head.V2, head.V3
			ord.Total = ord.Total * (1 - disc.Percentage/100)
			st.Processed = true
		}
	})
	engine.RegisterSystem(billing)
	engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		ctx.Run(billing, d)
		ctx.Sync()
	})

	orderResult, _ := ecs.GetComponent[Order](engine, entity)
	fmt.Printf("Order id: %v value: %v\n", orderResult.ID, orderResult.Total)
	engine.Run(time.Duration(time.Second))
	fmt.Printf("Order id: %v value with discount: %v\n", orderResult.ID, orderResult.Total)
}
