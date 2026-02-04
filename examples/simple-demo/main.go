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
	engine := goke.NewEngine()
	processedType := goke.ComponentRegister[Processed](engine)
	orderType := goke.ComponentRegister[Order](engine)
	discountType := goke.ComponentRegister[Discount](engine)

	// Initialize an entity with Order and Discount component data
	entity := goke.EntityCreate(engine)
	if ptr, _ := goke.EntityEnsureComponent[Order](engine, entity, orderType); ptr != nil {
		*ptr = Order{ID: "ORD-99", Total: 200.0}
	}

	if ptr, _ := goke.EntityEnsureComponent[Discount](engine, entity, discountType); ptr != nil {
		*ptr = Discount{Percentage: 20.0}
	}

	// Define the Billing System to calculate discounted totals for unprocessed orders
	view := goke.NewView2[Order, Discount](engine, goke.Without[Processed]())
	billing := goke.SystemFuncRegister(engine, func(schedule *goke.Schedule, d time.Duration) {
		for head := range view.All() {
			ord, disc := head.V1, head.V2
			ord.Total = ord.Total * (1 - disc.Percentage/100)

			// Defer the assignment of the Processed tag to the next synchronization point
			goke.ScheduleAddComponent(schedule, head.Entity, processedType, Processed{})
		}
	})

	// Define the Teardown System to monitor simulation exit conditions
	close := false
	view2 := goke.NewView0(engine, goke.WithTag[Processed]())
	teardownSystem := goke.SystemFuncRegister(engine, func(cb *goke.Schedule, d time.Duration) {
		for e := range view2.Filter([]goke.Entity{entity}) {
			_ = e
			close = true
			break
		}
	})

	// Configure the execution plan and define system dependencies
	goke.SystemRegister(engine, billing)
	goke.SystemRegister(engine, teardownSystem)
	engine.SetExecutionPlan(func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(billing, d)
		ctx.Sync()
		ctx.Run(teardownSystem, d)
		ctx.Sync()
	})

	// Log the initial state before simulation begins
	orderResult, _ := goke.EntityGetComponent[Order](engine, entity, orderType)
	fmt.Printf("Order id: %v value: %v\n", orderResult.ID, orderResult.Total)

	// Run the main simulation loop until the exit signal is received
	for !close {
		engine.Tick(time.Duration(time.Second))
		fmt.Printf("Order id: %v value with discount: %v\n", orderResult.ID, orderResult.Total)
	}
}
