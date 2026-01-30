package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/ecs"
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
	engine := ecs.NewEngine()
	processedInfo := ecs.RegisterComponent[Processed](engine)

	// Initialize an entity with Order and Discount component data
	entity := engine.CreateEntity()
	ecs.SetComponent(engine, entity, Order{ID: "ORD-99", Total: 200.0})
	ecs.SetComponent(engine, entity, Discount{Percentage: 20.0})

	// Define the Billing System to calculate discounted totals for unprocessed orders
	view := ecs.NewView2[Order, Discount](engine, ecs.Without[Processed]())
	billing := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for head := range view.All() {
			ord, disc := head.V1, head.V2
			ord.Total = ord.Total * (1 - disc.Percentage/100)

			// Defer the assignment of the Processed tag to the next synchronization point
			ecs.AssignComponent(cb, head.Entity, processedInfo, Processed{})
		}
	})

	// Define the Teardown System to monitor simulation exit conditions
	close := false
	view2 := ecs.NewView0(engine, ecs.WithTag[Processed]())
	teardownSystem := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for e := range view2.Filter([]ecs.Entity{entity}) {
			_ = e
			close = true
			break
		}
	})

	// Configure the execution plan and define system dependencies
	engine.RegisterSystem(billing)
	engine.RegisterSystem(teardownSystem)
	engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		ctx.Run(billing, d)
		ctx.Sync()
		ctx.Run(teardownSystem, d)
		ctx.Sync()
	})

	// Log the initial state before simulation begins
	orderResult, _ := ecs.GetComponent[Order](engine, entity)
	fmt.Printf("Order id: %v value: %v\n", orderResult.ID, orderResult.Total)

	// Run the main simulation loop until the exit signal is received
	for !close {
		engine.Tick(time.Duration(time.Second))
		fmt.Printf("Order id: %v value with discount: %v\n", orderResult.ID, orderResult.Total)
	}
}
