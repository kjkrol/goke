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
	ptr, _ := ecs.AddComponent[Order](engine, entity)
	ptr.ID = "ORD-99"
	ptr.Total = 200.0

	// based on unsafe.Pointer -> requires allocation on heap (slower)
	ecs.Assign(engine, entity, Status{Processed: false})
	ecs.Assign(engine, entity, Discount{Percentage: 20.0})

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

	order, _ := ecs.GetComponent[Order](engine, entity)
	fmt.Printf("Order id: %v value: %v\n", order.ID, order.Total)
	engine.Run(time.Duration(time.Second))
	fmt.Printf("Order id: %v value with discount: %v\n", order.ID, order.Total)
}
