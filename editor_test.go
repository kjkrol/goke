package goke_test

import (
	"testing"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

func TestEditorBuilder_AddComp(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	// Entity starts with only Velocity.
	var vel goke.Comp[Velocity]
	factory := ecs.NewFactory(&vel)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	// Add Position and write its value through the editor's cursor.
	var pos goke.Comp[Position]
	editor := ecs.NewEditorBuilder(&pos).Build()
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}
	pos.At(&editor.Cursor).X = 55

	val := seekComp[Position](ecs, entityID)
	if val == nil || val.X != 55 {
		t.Errorf("expected Position.X == 55, got %v", val)
	}
}

func TestEditorBuilder_InvalidEntity(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)

	var pos goke.Comp[Position]
	editor := ecs.NewEditorBuilder(&pos).Build()

	if editor.Update(uid.UID64(999)) {
		t.Errorf("expected Update to return false for a nonexistent entity")
	}
}

func TestEditorBuilder_Delete(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	factory := ecs.NewFactory(&pos, &vel)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	editor := ecs.NewEditorBuilder().Delete(goke.Del[Velocity]()).Build()
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}

	if hasComp[Velocity](ecs, entityID) {
		t.Errorf("expected Velocity to be removed")
	}
	if p := seekComp[Position](ecs, entityID); p == nil {
		t.Errorf("expected Position to remain")
	}
}

func TestEditorBuilder_AddAndDelete(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)

	var vel goke.Comp[Velocity]
	factory := ecs.NewFactory(&vel)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	// Swap Velocity for Position in a single migration.
	var pos goke.Comp[Position]
	editor := ecs.NewEditorBuilder(&pos).Delete(goke.Del[Velocity]()).Build()
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}
	pos.At(&editor.Cursor).Y = 7

	if hasComp[Velocity](ecs, entityID) {
		t.Errorf("expected Velocity to be removed")
	}
	p := seekComp[Position](ecs, entityID)
	if p == nil || p.Y != 7 {
		t.Errorf("expected Position.Y == 7, got %v", p)
	}
}

func TestEditorBuilder_ChainedDelete(t *testing.T) {
	ecs := goke.New()
	_ = goke.RegComp[Position](ecs)
	_ = goke.RegComp[Velocity](ecs)
	_ = goke.RegComp[Discount](ecs)

	var pos goke.Comp[Position]
	var vel goke.Comp[Velocity]
	var disc goke.Comp[Discount]
	factory := ecs.NewFactory(&pos, &vel, &disc)
	factory.Create(1)
	factory.Next()
	entityID := factory.IDs[0]

	editor := ecs.NewEditorBuilder().
		Delete(goke.Del[Velocity]()).
		Delete(goke.Del[Discount]()).
		Build()
	if !editor.Update(entityID) {
		t.Fatalf("expected Update to succeed")
	}

	if hasComp[Velocity](ecs, entityID) {
		t.Errorf("expected Velocity to be removed")
	}
	if hasComp[Discount](ecs, entityID) {
		t.Errorf("expected Discount to be removed")
	}
	if pos := seekComp[Position](ecs, entityID); pos == nil {
		t.Errorf("expected Position to remain")
	}
}
