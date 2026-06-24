package colstore

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
)

func TestColumnIndex_SetAndGet(t *testing.T) {
	var m columnIndex
	m.Reset()

	m.Set(comp.ID(0), columnPos(1))
	m.Set(comp.ID(5), columnPos(3))

	if got := m.Get(comp.ID(0)); got != columnPos(1) {
		t.Errorf("expected local ID 1, got %d", got)
	}
	if got := m.Get(comp.ID(5)); got != columnPos(3) {
		t.Errorf("expected local ID 3, got %d", got)
	}
}

func TestColumnIndex_Reset(t *testing.T) {
	var m columnIndex
	m.Reset()

	m.Set(comp.ID(2), columnPos(7))
	m.Reset()

	if got := m.Get(comp.ID(2)); got != invalidColumnPos {
		t.Errorf("expected invalidColumnPos after reset, got %d", got)
	}
}

func TestColumnIndex_UnsetSlotReturnsInvalidColumnPos(t *testing.T) {
	var m columnIndex
	m.Reset()

	if got := m.Get(comp.ID(10)); got != invalidColumnPos {
		t.Errorf("expected invalidColumnPos for unset slot, got %d", got)
	}
}

func TestColumnIndex_OverwriteSlot(t *testing.T) {
	var m columnIndex
	m.Reset()

	m.Set(comp.ID(1), columnPos(2))
	m.Set(comp.ID(1), columnPos(9))

	if got := m.Get(comp.ID(1)); got != columnPos(9) {
		t.Errorf("expected overwritten value 9, got %d", got)
	}
}
