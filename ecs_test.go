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
	ecs := goke.New()

	orderType := goke.RegisterComponentType[Order](ecs)
	discountType := goke.RegisterComponentType[Discount](ecs)
	processedType := goke.RegisterComponentType[Processed](ecs)

	eA := goke.CreateEntity(ecs)
	*goke.EnsureComponent[Order](ecs, eA, orderType) = Order{ID: "ORD-001", Total: 100.0}
	*goke.EnsureComponent[Discount](ecs, eA, discountType) = Discount{Percentage: 10.0}

	eB := goke.CreateEntity(ecs)
	*goke.EnsureComponent[Order](ecs, eB, orderType) = Order{ID: "ORD-002", Total: 50.0}

	query1 := goke.NewView2[Order, Discount](ecs)
	processedCount := 0

	billingSystem := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		for head := range query1.All() {
			processedCount++
			ord, disc := head.V1, head.V2
			ord.Total = ord.Total * (1 - disc.Percentage/100)
			goke.ScheduleAddComponent(cb, head.Entity, processedType, Processed{})
		}
	})
	query2 := goke.NewView0(ecs, goke.WithTag[Processed](), goke.WithTag[Order](), goke.WithTag[Discount]())
	cleanerSystem := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
		for entity := range query2.All() {
			goke.ScheduleRemoveEntity(schedule, entity)
		}
	})

	goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(billingSystem, d)

		// test this stage
		order, _ := goke.GetComponent[Order](ecs, eA, orderType)
		if order.Total != 90.0 {
			t.Errorf("Discount has not been applied, Total: %v", order.Total)
		}

		ctx.Sync()
		ctx.Run(cleanerSystem, d)
		ctx.Sync()
	})
	goke.Tick(ecs, time.Duration(time.Second))

	// Final Assertions
	if processedCount != 1 {
		t.Errorf("Expected 1 processed order, got %d", processedCount)
	}

	// Entity A should be removed from Registry
	_, err := goke.GetComponent[Order](ecs, eA, orderType)
	if err == nil {
		t.Error("Entity eA should have been removed from the registry")
	}

	// Entity B should still exist
	_, errB := goke.GetComponent[Order](ecs, eB, orderType)
	if errB != nil {
		t.Error("Entity eB should not have been removed")
	}
}
