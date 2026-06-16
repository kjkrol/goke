package goke_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
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

	orderDesc := goke.RegCompType[Order](ecs)
	_ = goke.RegCompType[Discount](ecs)
	processedDesc := goke.RegCompType[Processed](ecs)

	blueprint1 := goke.NewFactory2[Order, Discount](ecs)

	var eA, eB uid.UID64
	for chunk := range blueprint1.Create(1) {
		eA = chunk.Entity[0]
		chunk.Comp1[0] = Order{ID: "ORD-001", Total: 100.0}
		chunk.Comp2[0] = Discount{Percentage: 10.0}
	}

	blueprint2 := goke.NewFactory1[Order](ecs)
	for chunk := range blueprint2.Create(1) {
		eB = chunk.Entity[0]
		chunk.Comp1[0] = Order{ID: "ORD-002", Total: 50.0}
	}

	var colOrder goke.Col[Order]
	var colDiscount goke.Col[Discount]
	query1 := goke.NewView(ecs, colOrder.Track(), colDiscount.Track())
	processedCount := 0

	billingSystem := goke.RegSysFn(ecs, func(cb *goke.CmdBuf, d time.Duration) {
		query1.All()
		for query1.Next() {
			orders := colOrder.Slice(query1)
			discounts := colDiscount.Slice(query1)
			for i, entityID := range query1.EntSlice {
				processedCount++
				orders[i].Total *= (1 - discounts[i].Percentage/100)
				goke.CmdBufAddComp(cb, entityID, processedDesc, Processed{})
			}
		}
	})
	query2 := goke.NewView(ecs, goke.Include[Processed](), goke.Include[Order](), goke.Include[Discount]())
	cleanerSystem := goke.RegSysFn(ecs, func(schedule *goke.CmdBuf, d time.Duration) {
		query2.All()
		for query2.Next() {
			for _, entityID := range query2.EntSlice {
				schedule.RemoveEntity(entityID)
			}
		}
	})

	goke.SetPlan(ecs, func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(billingSystem, d)

		// test this stage
		order, _ := goke.SafeGetComp[Order](ecs, eA, orderDesc)
		if order.Total != 90.0 {
			t.Errorf("Discount has not been applied, Total: %v", order.Total)
		}

		ctx.Sync()
		query2.All()
		for query2.Next() {
			for _, entityID := range query2.EntSlice {
				_ = entityID
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
	_, err := goke.SafeGetComp[Order](ecs, eA, orderDesc)
	if err == nil {
		t.Error("Entity eA should have been removed from the registry")
	}

	// Entity B should still exist
	_, errB := goke.SafeGetComp[Order](ecs, eB, orderDesc)
	if errB != nil {
		t.Error("Entity eB should not have been removed")
	}
}

func TestECS_SafeGetComp_TypeSafety(t *testing.T) {
	ecs := goke.New()

	posDesc := goke.RegCompType[Position](ecs)
	_ = goke.RegCompType[Velocity](ecs)

	var entityID uid.UID64
	blueprint := goke.NewFactory1[Position](ecs)
	for chunk := range blueprint.Create(1) {
		entityID = chunk.Entity[0]
		chunk.Comp1[0] = Position{X: 10, Y: 20}
	}

	t.Run("Should fail when requesting wrong type for valid ID", func(t *testing.T) {
		_, err := goke.SafeGetComp[Velocity](ecs, entityID, posDesc)

		if err == nil {
			t.Fatal("Expected error due to type mismatch, but got nil")
		}

		expectedMsg := "type mismatch"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
		}
	})

	t.Run("Should succeed when type matches descriptor", func(t *testing.T) {
		p, err := goke.SafeGetComp[Position](ecs, entityID, posDesc)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if p.X != 10 || p.Y != 20 {
			t.Errorf("Data corruption: expected {10, 20}, got %+v", p)
		}
	})
}

func TestECS_GetComp(t *testing.T) {
	ecs := goke.New()
	posDesc := goke.RegCompType[Position](ecs)

	var entityID uid.UID64
	blueprint := goke.NewFactory1[Position](ecs)
	for chunk := range blueprint.Create(1) {
		entityID = chunk.Entity[0]
		chunk.Comp1[0] = Position{X: 10, Y: 20}
	}

	ptr := goke.GetComp[Position](ecs, entityID, posDesc)
	if ptr == nil {
		t.Fatalf("expected component")
	}
	if ptr.X != 10 {
		t.Errorf("wrong value")
	}

	fakeEntity := uid.UID64(999)
	ptrFake := goke.GetComp[Position](ecs, fakeEntity, posDesc)
	if ptrFake != nil {
		t.Errorf("expected nil")
	}
}

func TestECS_RemoveComp(t *testing.T) {
	ecs := goke.New()
	posDesc := goke.RegCompType[Position](ecs)

	var entityID uid.UID64
	blueprint := goke.NewFactory1[Position](ecs)
	for chunk := range blueprint.Create(1) {
		entityID = chunk.Entity[0]
	}

	err := goke.RemoveComp(ecs, entityID, posDesc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ptr := goke.GetComp[Position](ecs, entityID, posDesc)
	if ptr != nil {
		t.Errorf("expected component to be removed")
	}
}

func TestECS_RemoveEntity(t *testing.T) {
	ecs := goke.New()
	posDesc := goke.RegCompType[Position](ecs)
	_ = posDesc // to avoid unused variable error if any

	var entityID uid.UID64
	blueprint := goke.NewFactory1[Position](ecs)
	for chunk := range blueprint.Create(1) {
		entityID = chunk.Entity[0]
	}

	ok := goke.RemoveEntity(ecs, entityID)
	if !ok {
		t.Errorf("expected entity to be removed")
	}

	ok = goke.RemoveEntity(ecs, entityID)
	if ok {
		t.Errorf("expected false for already removed entity")
	}
}

func TestECS_UpsertComp(t *testing.T) {
	ecs := goke.New()
	posDesc := goke.RegCompType[Position](ecs)
	_ = goke.RegCompType[Velocity](ecs)

	var entityID uid.UID64
	blueprint := goke.NewFactory1[Position](ecs)
	for chunk := range blueprint.Create(1) {
		entityID = chunk.Entity[0]
	}

	_, err := goke.SafeUpsertComp[Velocity](ecs, entityID, posDesc)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}

	ptr := goke.UpsertComp[Position](ecs, entityID, posDesc)
	if ptr == nil {
		t.Fatalf("expected valid pointer")
	}
	ptr.X = 55

	val := goke.GetComp[Position](ecs, entityID, posDesc)
	if val.X != 55 {
		t.Errorf("expected 55, got %v", val.X)
	}

	fakeEntity := uid.UID64(999)
	_, err = goke.SafeUpsertComp[Position](ecs, fakeEntity, posDesc)
	if err == nil {
		t.Errorf("expected invalid entity error")
	}
}

func TestECS_UpsertComp_Panic(t *testing.T) {
	ecs := goke.New()
	posDesc := goke.RegCompType[Position](ecs)

	fakeEntity := uid.UID64(999)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected UpsertComp to panic on invalid entity")
		}
	}()

	goke.UpsertComp[Position](ecs, fakeEntity, posDesc)
}

func TestECS_Reset(t *testing.T) {
	ecs := goke.New()
	posDesc := goke.RegCompType[Position](ecs)

	var entityID uid.UID64
	blueprint := goke.NewFactory1[Position](ecs)
	for chunk := range blueprint.Create(1) {
		entityID = chunk.Entity[0]
	}

	goke.Reset(ecs)

	ptr := goke.GetComp[Position](ecs, entityID, posDesc)
	if ptr != nil {
		t.Errorf("expected entity to be reset/removed")
	}
}

func TestECS_NewWithOptions(t *testing.T) {
	ecs := goke.New(func(c *goke.Config) {
		c.Entity.FreeCap = 500
	})
	if ecs == nil {
		t.Fatal("expected ecs")
	}
}
