package goke_test

import (
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

	var order goke.Col[Order]
	var discount goke.Col[Discount]
	blueprint1 := goke.CreateEntFactory(ecs, goke.Track(&order), goke.Track(&discount))

	var eA, eB uid.UID64
	blueprint1.Create(1)
	blueprint1.Next()
	eA = blueprint1.IDs[0]
	fc1 := &blueprint1.Cursor
	order.Slice(fc1)[0] = Order{ID: "ORD-001", Total: 100.0}
	discount.Slice(fc1)[0] = Discount{Percentage: 10.0}

	blueprint2 := goke.CreateEntFactory(ecs, goke.Track(&order))
	blueprint2.Create(1)
	blueprint2.Next()
	eB = blueprint2.IDs[0]
	order.Slice(&blueprint2.Cursor)[0] = Order{ID: "ORD-002", Total: 50.0}

	query1 := goke.CreateView(ecs, goke.Track(&order), goke.Track(&discount))
	cursor1 := &query1.Cursor
	processedCount := 0

	billingSystem := goke.RegSysFn(ecs, func(cb *goke.CmdBuf, d time.Duration) {
		query1.All()
		for query1.Next() {
			orders := order.Slice(cursor1)
			discounts := discount.Slice(cursor1)
			for i, entityID := range query1.Cursor.IDs {
				processedCount++
				orders[i].Total *= (1 - discounts[i].Percentage/100)
				goke.CmdBufAddComp(cb, entityID, processedDesc, Processed{})
			}
		}
	})
	query2 := goke.CreateView(ecs, goke.Include[Processed](), goke.Include[Order](), goke.Include[Discount]())
	cleanerSystem := goke.RegSysFn(ecs, func(schedule *goke.CmdBuf, d time.Duration) {
		query2.All()
		for query2.Next() {
			for _, entityID := range query2.Cursor.IDs {
				schedule.RemoveEntity(entityID)
			}
		}
	})

	goke.SetPlan(ecs, func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(billingSystem, d)

		// test this stage
		result := goke.GetComp[Order](ecs, eA, orderDesc)
		if result.Total != 90.0 {
			t.Errorf("Discount has not been applied, Total: %v", result.Total)
		}

		ctx.Sync()
		query2.All()
		for query2.Next() {
			for _, entityID := range query2.Cursor.IDs {
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
	if goke.GetComp[Order](ecs, eA, orderDesc) != nil {
		t.Error("Entity eA should have been removed from the registry")
	}

	// Entity B should still exist
	if goke.GetComp[Order](ecs, eB, orderDesc) == nil {
		t.Error("Entity eB should not have been removed")
	}
}

func TestECS_GetComp(t *testing.T) {
	ecs := goke.New()
	posDesc := goke.RegCompType[Position](ecs)

	var entityID uid.UID64
	var pos goke.Col[Position]
	blueprint := goke.CreateEntFactory(ecs, goke.Track(&pos))
	blueprint.Create(1)
	blueprint.Next()
	entityID = blueprint.IDs[0]
	pos.Slice(&blueprint.Cursor)[0] = Position{X: 10, Y: 20}

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
	fcPosOpt := goke.Track(new(goke.Col[Position]))
	blueprint := goke.CreateEntFactory(ecs, fcPosOpt)
	blueprint.Create(1)
	blueprint.Next()
	entityID = blueprint.IDs[0]

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
	fcPosOpt := goke.Track(new(goke.Col[Position]))
	blueprint := goke.CreateEntFactory(ecs, fcPosOpt)
	blueprint.Create(1)
	blueprint.Next()
	entityID = blueprint.IDs[0]

	ok := goke.RemoveEnt(ecs, entityID)
	if !ok {
		t.Errorf("expected entity to be removed")
	}

	ok = goke.RemoveEnt(ecs, entityID)
	if ok {
		t.Errorf("expected false for already removed entity")
	}
}

func TestECS_UpsertComp(t *testing.T) {
	ecs := goke.New()
	posDesc := goke.RegCompType[Position](ecs)

	fcPosOpt := goke.Track(new(goke.Col[Position]))
	blueprint := goke.CreateEntFactory(ecs, fcPosOpt)
	blueprint.Create(1)
	blueprint.Next()
	entityID := blueprint.IDs[0]

	ptr := goke.UpsertComp[Position](ecs, entityID, posDesc)
	if ptr == nil {
		t.Fatalf("expected valid pointer")
	}
	ptr.X = 55

	val := goke.GetComp[Position](ecs, entityID, posDesc)
	if val.X != 55 {
		t.Errorf("expected 55, got %v", val.X)
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
	blueprint := goke.CreateEntFactory(ecs, goke.Track(new(goke.Col[Position])))
	blueprint.Create(1)
	blueprint.Next()
	entityID = blueprint.IDs[0]

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
