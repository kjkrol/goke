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

	var eA, eB goke.Entity
	for page := range blueprint1.Create(1) {
		eA = page.Entity[0]
		page.Comp1[0] = Order{ID: "ORD-001", Total: 100.0}
		page.Comp2[0] = Discount{Percentage: 10.0}
	}

	blueprint2 := goke.NewBlueprint1[Order](ecs)
	for page := range blueprint2.Create(1) {
		eB = page.Entity[0]
		page.Comp1[0] = Order{ID: "ORD-002", Total: 50.0}
		fmt.Printf("eB= %d\n", page.Entity[0].Index())
	}

	query1 := goke.NewView2[Order, Discount](ecs)
	processedCount := 0

	billingSystem := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		for page := range query1.All() {
			for i, entity := range page.Entity {
				processedCount++
				page.Comp1[i].Total *= (1 - page.Comp2[i].Percentage/100)
				goke.ScheduleAddComponent(cb, entity, processedDesc, Processed{})
			}
		}
	})
	query2 := goke.NewView0(ecs, goke.Include[Processed](), goke.Include[Order](), goke.Include[Discount]())
	cleanerSystem := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
		for page := range query2.All() {
			for _, entity := range page.Entity {
				goke.ScheduleRemoveEntity(schedule, entity)
			}
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
		for page := range query2.All() {
			for _, entity := range page.Entity {
				fmt.Printf("entity %d\n", entity.Index())
			}
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

	var entity goke.Entity
	blueprint := goke.NewBlueprint1[Position](ecs)
	for page := range blueprint.Create(1) {
		entity = page.Entity[0]
		page.Comp1[0] = Position{X: 10, Y: 20}
	}

	t.Run("Should fail when requesting wrong type for valid ID", func(t *testing.T) {
		_, err := goke.SafeGetComponent[Velocity](ecs, entity, posDesc)

		if err == nil {
			t.Fatal("Expected error due to type mismatch, but got nil")
		}

		expectedMsg := "type mismatch"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
		}
	})

	t.Run("Should succeed when type matches descriptor", func(t *testing.T) {
		p, err := goke.SafeGetComponent[Position](ecs, entity, posDesc)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if p.X != 10 || p.Y != 20 {
			t.Errorf("Data corruption: expected {10, 20}, got %+v", p)
		}
	})
}
