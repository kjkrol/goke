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

var processedDesc, orderDesc, discountDesc goke.ComponentDesc

func main() {
	ecs := goke.New()
	processedDesc = goke.RegisterComponent[Processed](ecs)
	orderDesc = goke.RegisterComponent[Order](ecs)
	discountDesc = goke.RegisterComponent[Discount](ecs)

	// Initialize an entity with Order and Discount component data
	blueprint := goke.NewBlueprint2[Order, Discount](ecs)
	var entity goke.Entity
	for page := range blueprint.Create(1) {
		entity = page.Entity[0]
		page.Comp1[0] = Order{ID: "ORD-99", Total: 100.0}
		page.Comp2[0] = Discount{Percentage: 20.0}
	}

	// Define the Billing System to calculate discounted totals for unprocessed orders
	view := goke.NewView2[Order, Discount](ecs, goke.Exclude[Processed]())
	billing := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
		for page := range view.All() {
			for i, entity := range page.Entity {
				ord, disc := &page.Comp1[i], &page.Comp2[i]
				ord.Total = ord.Total * (1 - disc.Percentage/100)

				// Defer the assignment of the Processed tag to the next synchronization point
				goke.ScheduleAddComponent(schedule, entity, processedDesc, Processed{})
			}
		}
	})

	// Define the Teardown System to monitor simulation exit conditions
	close := false
	view2 := goke.NewView0(ecs, goke.Include[Processed]())
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
	orderResult, _ := goke.SafeGetComponent[Order](ecs, entity, orderDesc)
	fmt.Printf("Order id: %v value: %v\n", orderResult.ID, orderResult.Total)

	// Run the main simulation loop until the exit signal is received
	for !close {
		goke.Tick(ecs, time.Second)
		orderResult, _ := goke.SafeGetComponent[Order](ecs, entity, orderDesc)
		fmt.Printf("Order id: %v value with discount: %v\n", orderResult.ID, orderResult.Total)
	}
}
