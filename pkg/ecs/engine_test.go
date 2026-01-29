package ecs_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/kjkrol/goke/pkg/ecs"
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
	engine := ecs.NewEngine()

	processedTypeInfo := engine.RegisterComponentType(reflect.TypeFor[Processed]())

	eA := engine.CreateEntity()
	order, _ := ecs.AllocateComponent[Order](engine, eA)
	*order = Order{ID: "ORD-001", Total: 100.0}
	discount, _ := ecs.AllocateComponent[Discount](engine, eA)
	*discount = Discount{Percentage: 10.0}

	eB := engine.CreateEntity()
	order2, _ := ecs.AllocateComponent[Order](engine, eB)
	*order2 = Order{ID: "ORD-002", Total: 50.0}

	query1 := ecs.NewQuery2[Order, Discount](engine)
	processedCount := 0

	billingSystem := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for head := range query1.All2() {
			processedCount++
			ord, disc := head.V1, head.V2
			ord.Total = ord.Total * (1 - disc.Percentage/100)
			ecs.AssignComponent(cb, head.Entity, processedTypeInfo, Processed{})
		}
	})
	query2 := ecs.NewQuery0(engine, ecs.WithTag[Processed](), ecs.WithTag[Order](), ecs.WithTag[Discount]())
	cleanerSystem := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for entity := range query2.All() {
			cb.RemoveEntity(entity)
		}
	})

	engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		ctx.Run(billingSystem, d)

		// test this stage
		order, _ := ecs.GetComponent[Order](engine, eA)
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
	_, err := ecs.GetComponent[Order](engine, eA)
	if err == nil {
		t.Error("Entity eA should have been removed from the registry")
	}

	// Entity B should still exist
	_, errB := ecs.GetComponent[Order](engine, eB)
	if errB != nil {
		t.Error("Entity eB should not have been removed")
	}
}
