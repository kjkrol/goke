package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke"
)

type (
	Order struct {
		ID    string
		Total float64
	}
	Discount  struct{ Percentage float64 }
	Processed struct{}
)

func main() {
	ecs := goke.New()
	processedType := goke.RegisterComponentType[Processed](ecs)
	orderType := goke.RegisterComponentType[Order](ecs)
	discountType := goke.RegisterComponentType[Discount](ecs)

	// Initialize an entity with Order and Discount component data
	entity := goke.CreateEntity(ecs)
	*goke.EnsureComponent[Order](ecs, entity, orderType) = Order{ID: "ORD-99", Total: 200.0}
	*goke.EnsureComponent[Discount](ecs, entity, discountType) = Discount{Percentage: 20.0}
	
	// Define the Billing System to calculate discounted totals for unprocessed orders
	view := goke.NewView2[Order, Discount](ecs, goke.Without[Processed]())
	billing := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
		for head := range view.All() {
			ord, disc := head.V1, head.V2
			ord.Total = ord.Total * (1 - disc.Percentage/100)

			// Defer the assignment of the Processed tag to the next synchronization point
			goke.ScheduleAddComponent(schedule, head.Entity, processedType, Processed{})
		}
	})

	// Define the Teardown System to monitor simulation exit conditions
	close := false
	view2 := goke.NewView0(ecs, goke.WithTag[Processed]())
	teardownSystem := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		for e := range view2.Filter([]goke.Entity{entity}) {
			_ = e
			close = true
			break
		}
	})

	// Configure the execution plan and define system dependencies
	goke.RegisterSystem(ecs, billing)
	goke.RegisterSystem(ecs, teardownSystem)
	goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(billing, d)
		ctx.Sync()
		ctx.Run(teardownSystem, d)
		ctx.Sync()
	})

	// Log the initial state before simulation begins
	orderResult, _ := goke.GetComponent[Order](ecs, entity, orderType)
	fmt.Printf("Order id: %v value: %v\n", orderResult.ID, orderResult.Total)

	// Run the main simulation loop until the exit signal is received
	for !close {
		goke.Tick(ecs, time.Second)
		fmt.Printf("Order id: %v value with discount: %v\n", orderResult.ID, orderResult.Total)
	}
}
