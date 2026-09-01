package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type (
	Order struct {
		ID    int
		Total float64
	}
	Discount  struct{ Percentage float64 }
	Processed struct{}
)

var orderID, discountID goke.CompID

func main() {
	ecs := goke.New()
	_ = ecs.RegComp[Processed]()
	_ = ecs.RegComp[Order]()
	discountID = ecs.RegComp[Discount]()

	var order goke.Comp[Order]
	var discount goke.Comp[Discount]
	var entityID uid.UID64

	// Setup
	spawnSystem := goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			factory := si.NewFactory(&order, &discount)
			factory.Create(1)
			factory.Next()
			entityID = factory.IDs[0]
			fc := &factory.Cursor
			order.Slice(fc)[0] = Order{ID: 99, Total: 100.0}
			discount.Slice(fc)[0] = Discount{Percentage: 20.0}
		},
	}
	var logQuery *goke.Query
	logSystem := goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			logQuery = si.NewQueryBuilder(&order).Build()
		},
		OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) {
			if logQuery.Seek(entityID) {
				o := order.At(logQuery.Cursor())
				fmt.Printf("Order id: %v value: %v\n", o.ID, o.Total)
			}
		},
	}
	ecs.Setup(spawnSystem, logSystem)

	// Configure the execution plan and define system dependencies
	var processed goke.Comp[Processed]
	var query *goke.Query
	var markProcessed *goke.Editor
	billingSystem := ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			query = si.NewQueryBuilder(&order, &discount).Exclude(goke.Exclude[Processed]()).Build()
			markProcessed = query.NewEditorBuilder(&processed).Build() // create editor that adds processed tag
		},
		OnUpdate: func(cb *goke.CmdBuf, d time.Duration) {
			cursor := query.Cursor()
			query.All()
			for query.Next() {
				orders := order.Slice(cursor)
				discounts := discount.Slice(cursor)
				mb := query.BeginMigrate(cb) // create migration buffer
				for i, entityID := range cursor.IDs {
					orders[i].Total = orders[i].Total * (1 - discounts[i].Percentage/100)
					mb.Add(entityID) // add entity to migration buffer
				}
				mb.Commit(markProcessed) // commit migration buffer using Editor
			}
		},
	})

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

	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(billingSystem, d)
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
