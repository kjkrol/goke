package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

type (
	Order struct {
		ID    string
		Total float64
	}
	Discount  struct{ Percentage float64 }
	Processed struct{}
)

var processedID, orderID, discountID goke.CompID

func main() {
	ecs := goke.New()
	processedID = goke.RegComp[Processed](ecs)
	_ = goke.RegComp[Order](ecs)
	discountID = goke.RegComp[Discount](ecs)

	// Initialize an entity with Order and Discount component data
	var order goke.Col[Order]
	var discount goke.Col[Discount]
	factory := goke.CreateFactory(ecs, goke.Track(&order), goke.Track(&discount))
	var entityID uid.UID64
	factory.Create(1)
	factory.Next()
	entityID = factory.IDs[0]
	fc := &factory.Cursor
	order.Slice(fc)[0] = Order{ID: "ORD-99", Total: 100.0}
	discount.Slice(fc)[0] = Discount{Percentage: 20.0}

	// Define the Billing System to calculate discounted totals for unprocessed orders
	query := goke.CreateView(ecs, goke.Track(&order), goke.Track(&discount), goke.Exclude[Processed]())
	cursor := &query.Cursor
	billing := goke.RegSysFn(ecs, func(schedule *goke.CmdBuf, d time.Duration) {
		query.All()
		for query.Next() {
			orders := order.Slice(cursor)
			discounts := discount.Slice(cursor)
			for i, entityID := range query.Cursor.IDs {
				orders[i].Total = orders[i].Total * (1 - discounts[i].Percentage/100)
				// Defer the assignment of the Processed tag to the next synchronization point
				goke.CmdBufAddComp(schedule, entityID, processedID, Processed{})
			}
		}
	})

	// Define the Teardown System to monitor simulation exit conditions
	close := false
	query2 := goke.CreateView(ecs, goke.Include[Processed]())
	teardownSystem := goke.RegSysFn(ecs, func(cb *goke.CmdBuf, d time.Duration) {
		query2.Filter([]uid.UID64{entityID})
		if query2.Next() {
			close = true
		}
	})

	// Configure the execution plan and define system dependencies
	goke.SetPlan(ecs, func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(billing, d)
		ctx.Sync()
		ctx.Run(teardownSystem, d)
		ctx.Sync()
	})

	// Log the initial state before simulation begins
	lookup := goke.CreateLookup(ecs, goke.Track(&order))
	if lookup.Seek(entityID) {
		orderResult := order.At(&lookup.Cursor)
		fmt.Printf("Order id: %v value: %v\n", orderResult.ID, orderResult.Total)
	}

	// Run the main simulation loop until the exit signal is received
	for !close {
		goke.Tick(ecs, time.Second)
		if lookup.Seek(entityID) {
			orderResult := order.At(&lookup.Cursor)
			fmt.Printf("Order id: %v value with discount: %v\n", orderResult.ID, orderResult.Total)
		}
	}
}
