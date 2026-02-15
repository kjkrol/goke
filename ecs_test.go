package goke_test

import (
	"fmt"
	"strings"
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

	orderDesc := goke.RegisterComponent[Order](ecs)
	_ = goke.RegisterComponent[Discount](ecs)
	processedDesc := goke.RegisterComponent[Processed](ecs)

	blueprint1 := goke.NewBlueprint2[Order, Discount](ecs)

	eA, order, discount := blueprint1.Create()
	*order = Order{ID: "ORD-001", Total: 100.0}
	*discount = Discount{Percentage: 10.0}

	blueprint2 := goke.NewBlueprint1[Order](ecs)
	eB, order := blueprint2.Create()
	*order = Order{ID: "ORD-002", Total: 50.0}

	fmt.Printf("eB= %d\n", eB.Index())

	query1 := goke.NewView2[Order, Discount](ecs)
	processedCount := 0

	billingSystem := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		for head := range query1.All() {
			processedCount++
			ord, disc := head.V1, head.V2
			ord.Total = ord.Total * (1 - disc.Percentage/100)
			goke.ScheduleAddComponent(cb, head.Entity, processedDesc, Processed{})
		}
	})
	query2 := goke.NewView0(ecs, goke.Include[Processed](), goke.Include[Order](), goke.Include[Discount]())
	cleanerSystem := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
		for head := range query2.All() {
			goke.ScheduleRemoveEntity(schedule, head.Entity)
		}
	})

	goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(billingSystem, d)

		// test this stage
		order, _ := goke.SafeGetComponent[Order](ecs, eA, orderDesc)
		if order.Total != 90.0 {
			t.Errorf("Discount has not been applied, Total: %v", order.Total)
		}

		ctx.Sync()
		for head := range query2.All() {
			fmt.Printf("entity %d\n", head.Entity.Index())
		}
		ctx.Run(cleanerSystem, d)
		ctx.Sync()
	})
	goke.Tick(ecs, time.Duration(time.Second))

	// Final Assertions
	if processedCount != 1 {
		t.Errorf("Expected 1 processed order, got %d", processedCount)
	}

	// Entity A should be removed from Registry
	_, err := goke.SafeGetComponent[Order](ecs, eA, orderDesc)
	if err == nil {
		t.Error("Entity eA should have been removed from the registry")
	}

	// Entity B should still exist
	_, errB := goke.SafeGetComponent[Order](ecs, eB, orderDesc)
	if errB != nil {
		t.Error("Entity eB should not have been removed")
	}
}

func TestECS_GetComponent_TypeSafety(t *testing.T) {
	ecs := goke.New()

	posDesc := goke.RegisterComponent[Position](ecs)
	_ = goke.RegisterComponent[Velocity](ecs)

	blueprint := goke.NewBlueprint1[Position](ecs)
	e, pos := blueprint.Create()
	*pos = Position{X: 10, Y: 20}

	t.Run("Should fail when requesting wrong type for valid ID", func(t *testing.T) {
		_, err := goke.SafeGetComponent[Velocity](ecs, e, posDesc)

		if err == nil {
			t.Fatal("Expected error due to type mismatch, but got nil")
		}

		expectedMsg := "type mismatch"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
		}
	})

	t.Run("Should succeed when type matches descriptor", func(t *testing.T) {
		p, err := goke.SafeGetComponent[Position](ecs, e, posDesc)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if p.X != 10 || p.Y != 20 {
			t.Errorf("Data corruption: expected {10, 20}, got %+v", p)
		}
	})
}
