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

// seekComp reads a single entity's component via Matcher.Seek.
// Returns nil if the entity does not exist.
func seekComp[T any](ecs *goke.ECS, e uid.UID64) *T {
	var col goke.Col[T]
	m := ecs.CreateMatcher(goke.Track(&col))
	if !m.Seek(e) {
		return nil
	}
	return col.At(&m.Cursor)
}

func TestECS_UseCase(t *testing.T) {
	ecs := goke.New()

	_ = goke.RegComp[Order](ecs)
	_ = goke.RegComp[Discount](ecs)
	processedID := goke.RegComp[Processed](ecs)

	var order goke.Col[Order]
	var discount goke.Col[Discount]
	factory1 := ecs.CreateFactory(goke.Add(&order), goke.Add(&discount))

	var eA, eB uid.UID64
	factory1.Create(1)
	factory1.Next()
	eA = factory1.IDs[0]
	fc1 := &factory1.Cursor
	order.Slice(fc1)[0] = Order{ID: "ORD-001", Total: 100.0}
	discount.Slice(fc1)[0] = Discount{Percentage: 10.0}

	factory2 := ecs.CreateFactory(goke.Add(&order))
	factory2.Create(1)
	factory2.Next()
	eB = factory2.IDs[0]
	order.Slice(&factory2.Cursor)[0] = Order{ID: "ORD-002", Total: 50.0}

	query1 := ecs.CreateMatcher(goke.Track(&order), goke.Track(&discount))
	cursor1 := &query1.Cursor
	processedCount := 0

	billingSystem := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		query1.All()
		for query1.Next() {
			orders := order.Slice(cursor1)
			discounts := discount.Slice(cursor1)
			for i, entityID := range query1.Cursor.IDs {
				processedCount++
				orders[i].Total *= (1 - discounts[i].Percentage/100)
				goke.CmdBufAddComp(cb, entityID, processedID, Processed{})
			}
		}
	})
	query2 := ecs.CreateMatcher(goke.Include[Processed](), goke.Include[Order](), goke.Include[Discount]())
	cleanerSystem := ecs.RegSysFn(func(schedule *goke.CmdBuf, d time.Duration) {
		query2.All()
		for query2.Next() {
			for _, entityID := range query2.Cursor.IDs {
				schedule.RemoveEntity(entityID)
			}
		}
	})

	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(billingSystem, d)

		// test this stage
		result := seekComp[Order](ecs, eA)
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
	ecs.Tick(time.Duration(time.Second))

	// Final Assertions
	if processedCount != 1 {
		t.Errorf("Expected 1 processed order, got %d", processedCount)
	}

	// Entity A should be removed from Registry
	if seekComp[Order](ecs, eA) != nil {
		t.Error("Entity eA should have been removed from the registry")
	}

	// Entity B should still exist
	if seekComp[Order](ecs, eB) == nil {
		t.Error("Entity eB should not have been removed")
	}
}

func TestECS_Seek(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Col[Position]
	factory := ecs.CreateFactory(goke.Add(&pos))
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]
	pos.Slice(&factory.Cursor)[0] = Position{X: 10, Y: 20}

	matcher := ecs.CreateMatcher(goke.Track(&pos))

	if !matcher.Seek(entityID) {
		t.Fatalf("expected Seek to find the entity")
	}
	if got := pos.At(&matcher.Cursor); got.X != 10 {
		t.Errorf("wrong value: got X=%v, want 10", got.X)
	}

	fakeEntity := uid.UID64(999)
	if matcher.Seek(fakeEntity) {
		t.Errorf("expected Seek to fail for a nonexistent entity")
	}
}

// TestECS_Seek_AcrossArchetypes exercises the per-archetype cache: alternating
// Seeks between two archetypes must re-resolve the table and offsets each switch,
// never returning stale ones.
func TestECS_Seek_AcrossArchetypes(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	// Entity A: {Position}
	var posA goke.Col[Position]
	fa := ecs.CreateFactory(goke.Add(&posA))
	fa.Create(1)
	fa.Next()
	eA := fa.IDs[0]
	posA.Slice(&fa.Cursor)[0] = Position{X: 1}

	// Entity B: {Position, Velocity} — a different archetype
	var posB goke.Col[Position]
	var velB goke.Col[Velocity]
	fb := ecs.CreateFactory(goke.Add(&posB), goke.Add(&velB))
	fb.Create(1)
	fb.Next()
	eB := fb.IDs[0]
	posB.Slice(&fb.Cursor)[0] = Position{X: 2}

	var pos goke.Col[Position]
	matcher := ecs.CreateMatcher(goke.Track(&pos))

	for i := 0; i < 3; i++ {
		if !matcher.Seek(eA) || pos.At(&matcher.Cursor).X != 1 {
			t.Fatalf("iter %d: expected eA.X == 1", i)
		}
		if !matcher.Seek(eB) || pos.At(&matcher.Cursor).X != 2 {
			t.Fatalf("iter %d: expected eB.X == 2", i)
		}
	}
}

func TestECS_RemoveComp(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var entityID uid.UID64
	factory := ecs.CreateFactory(goke.Add(new(goke.Col[Position])))
	factory.Create(1)
	factory.Next()
	entityID = factory.IDs[0]

	editor := ecs.CreateEditor(goke.Del[Position]())
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}

	// Position was the entity's only component, so removing it unlinks the entity.
	ptr := seekComp[Position](ecs, entityID)
	if ptr != nil {
		t.Errorf("expected component to be removed")
	}
}

func TestECS_RemoveEntity(t *testing.T) {
	ecs := goke.New()
	posID := goke.RegComp[Position](ecs)
	_ = posID // to avoid unused variable error if any

	var entityID uid.UID64
	fcPosOpt := goke.Add(new(goke.Col[Position]))
	factory := ecs.CreateFactory(fcPosOpt)
	factory.Create(1)
	factory.Next()
	entityID = factory.IDs[0]

	ok := ecs.RemoveEnt(entityID)
	if !ok {
		t.Errorf("expected entity to be removed")
	}

	ok = ecs.RemoveEnt(entityID)
	if ok {
		t.Errorf("expected false for already removed entity")
	}
}

func TestECS_Editor_AddComp(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	// Entity starts with only Velocity.
	var vel goke.Col[Velocity]
	factory := ecs.CreateFactory(goke.Add(&vel))
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	// Add Position and write its value through the editor's cursor.
	var pos goke.Col[Position]
	editor := ecs.CreateEditor(goke.Add(&pos))
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}
	pos.At(&editor.Cursor).X = 55

	val := seekComp[Position](ecs, entityID)
	if val == nil || val.X != 55 {
		t.Errorf("expected Position.X == 55, got %v", val)
	}
}

func TestECS_Factory_DelPanics(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	defer func() {
		if recover() == nil {
			t.Errorf("expected CreateFactory to panic when given a Del opt")
		}
	}()

	ecs.CreateFactory(goke.Del[Position]())
}

func TestECS_Editor_InvalidEntity(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Col[Position]
	editor := ecs.CreateEditor(goke.Add(&pos))

	if editor.Update(uid.UID64(999)) {
		t.Errorf("expected Update to return false for a nonexistent entity")
	}
}

func TestECS_Reset(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var entityID uid.UID64
	factory := ecs.CreateFactory(goke.Add(new(goke.Col[Position])))
	factory.Create(1)
	factory.Next()
	entityID = factory.IDs[0]

	ecs.Reset()

	ptr := seekComp[Position](ecs, entityID)
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
