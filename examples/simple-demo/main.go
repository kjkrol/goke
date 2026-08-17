package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/v2"
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
	var order goke.Comp[Order]
	var discount goke.Comp[Discount]
	var entityID uid.UID64

	// spawnSystem seeds the world. logInitialSystem reads what spawnSystem
	// just created — ecs.Setup runs each given system fully (Init, Update,
	// Sync) before the next one starts, so logInitialSystem's Init already
	// sees the spawned entity.
	spawnSystem := goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			factory := si.NewFactory(&order, &discount)
			factory.Create(1)
			factory.Next()
			entityID = factory.IDs[0]
			fc := &factory.Cursor
			order.Slice(fc)[0] = Order{ID: "ORD-99", Total: 100.0}
			discount.Slice(fc)[0] = Discount{Percentage: 20.0}
		},
	}
	var initialQuery *goke.Query
	logInitialSystem := goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			initialQuery = si.NewQueryBuilder(&order).Build()
		},
		OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) {
			if initialQuery.Seek(entityID) {
				o := order.At(initialQuery.Cursor())
				fmt.Printf("Order id: %v value: %v\n", o.ID, o.Total)
			}
		},
	}
	ecs.Setup(spawnSystem, logInitialSystem)

	// Define the Billing System to calculate discounted totals for unprocessed orders
	var query *goke.Query
	billing := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			query = si.NewQueryBuilder(&order, &discount).Exclude(goke.Exclude[Processed]()).Build()
		},
		OnUpdate: func(schedule *goke.CmdBuf, d time.Duration) {
			cursor := query.Cursor()
			query.All()
			for query.Next() {
				orders := order.Slice(cursor)
				discounts := discount.Slice(cursor)
				for i, entityID := range cursor.IDs {
					orders[i].Total = orders[i].Total * (1 - discounts[i].Percentage/100)
					// Defer the assignment of the Processed tag to the next synchronization point
					goke.AddOne(schedule, entityID, processedID, Processed{})
				}
			}
		},
	})

	// Define the Teardown System to monitor simulation exit conditions
	close := false
	var query2 *goke.Query
	teardownSystem := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			query2 = si.NewQueryBuilder().Include(goke.Include[Processed]()).Build()
		},
		OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
			query2.Pick([]uid.UID64{entityID})
			if query2.Next() {
				close = true
			}
		},
	})

	// Define the Report System to log the order's value every tick — reads
	// flow through systems too, not raw lookups pulled out of the ECS.
	var reportQuery *goke.Query
	reportSystem := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			reportQuery = si.NewQueryBuilder(&order).Build()
		},
		OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) {
			if reportQuery.Seek(entityID) {
				o := order.At(reportQuery.Cursor())
				fmt.Printf("Order id: %v value with discount: %v\n", o.ID, o.Total)
			}
		},
	})

	// Configure the execution plan and define system dependencies
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(billing, d)
		ctx.Sync()
		ctx.Run(teardownSystem, d)
		ctx.Sync()
		ctx.Run(reportSystem, d)
	})

	// Run the main simulation loop until the exit signal is received
	for !close {
		ecs.Tick(time.Second)
	}
}
