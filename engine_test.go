package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke"
)

// components

type Order struct {
	ID    string
	Total float64
}

type Discount struct {
	Percentage float64
}

type Processed struct{}

func TestECS_UseCase(t *testing.T) {
	engine := goke.NewEngine()

	orderType := goke.ComponentRegister[Order](engine)
	discountType := goke.ComponentRegister[Discount](engine)
	processedType := goke.ComponentRegister[Processed](engine)

	eA := goke.EntityCreate(engine)
	order, _ := goke.EntityEnsureComponent[Order](engine, eA, orderType)
	*order = Order{ID: "ORD-001", Total: 100.0}
	discount, _ := goke.EntityEnsureComponent[Discount](engine, eA, discountType)
	*discount = Discount{Percentage: 10.0}

	eB := goke.EntityCreate(engine)
	order2, _ := goke.EntityEnsureComponent[Order](engine, eB, orderType)
	*order2 = Order{ID: "ORD-002", Total: 50.0}

	query1 := goke.NewView2[Order, Discount](engine)
	processedCount := 0

	billingSystem := goke.SystemFuncRegister(engine, func(cb *goke.Schedule, d time.Duration) {
		for head := range query1.All() {
			processedCount++
			ord, disc := head.V1, head.V2
			ord.Total = ord.Total * (1 - disc.Percentage/100)
			goke.ScheduleAddComponent(cb, head.Entity, processedType, Processed{})
		}
	})
	query2 := goke.NewView0(engine, goke.WithTag[Processed](), goke.WithTag[Order](), goke.WithTag[Discount]())
	cleanerSystem := goke.SystemFuncRegister(engine, func(schedule *goke.Schedule, d time.Duration) {
		for entity := range query2.All() {
			goke.ScheduleRemoveEntity(schedule, entity)
		}
	})

	engine.SetExecutionPlan(func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(billingSystem, d)

		// test this stage
		order, _ := goke.EntityGetComponent[Order](engine, eA, orderType)
		if order.Total != 90.0 {
			t.Errorf("Discount has not been applied, Total: %v", order.Total)
		}

		ctx.Sync()
		ctx.Run(cleanerSystem, d)
		ctx.Sync()
	})
	engine.Tick(time.Duration(time.Second))

	// Final Assertions
	if processedCount != 1 {
		t.Errorf("Expected 1 processed order, got %d", processedCount)
	}

	// Entity A should be removed from Registry
	_, err := goke.EntityGetComponent[Order](engine, eA, orderType)
	if err == nil {
		t.Error("Entity eA should have been removed from the registry")
	}

	// Entity B should still exist
	_, errB := goke.EntityGetComponent[Order](engine, eB, orderType)
	if errB != nil {
		t.Error("Entity eB should not have been removed")
	}
}
