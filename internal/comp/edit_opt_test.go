package comp_test

import (
	"errors"
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

func TestEditOpt_Add(t *testing.T) {
	var mi comp.DefIndex
	mi.Init()
	var s comp.EditSpec
	var col1 iter.Col[position]
	var col2 iter.Col[velocity]

	if err := comp.Add(&col1)(&s, &mi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col1.Idx != 0 {
		t.Errorf("expected first Add to set Idx 0, got %d", col1.Idx)
	}

	if err := comp.Add(&col2)(&s, &mi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col2.Idx != 1 {
		t.Errorf("expected second Add to set Idx 1, got %d", col2.Idx)
	}
	if len(s.AddDefs) != 2 {
		t.Errorf("expected 2 AddDefs, got %v", s.AddDefs)
	}
}

func TestEditOpt_Del(t *testing.T) {
	var mi comp.DefIndex
	mi.Init()
	var s comp.EditSpec

	if err := comp.Del[position]()(&s, &mi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.DelDefs) != 1 {
		t.Errorf("expected 1 DelDef, got %v", s.DelDefs)
	}
}

func TestEditSpec_Init(t *testing.T) {
	t.Run("applies Add and Del opts", func(t *testing.T) {
		var mi comp.DefIndex
		mi.Init()
		var s comp.EditSpec
		var col iter.Col[position]

		s.Init(&mi, comp.Add(&col), comp.Del[velocity]())

		if len(s.AddDefs) != 1 || len(s.DelDefs) != 1 {
			t.Errorf("expected 1 AddDef and 1 DelDef, got add=%v del=%v", s.AddDefs, s.DelDefs)
		}
	})

	t.Run("panics when an opt returns an error", func(t *testing.T) {
		var mi comp.DefIndex
		mi.Init()
		var s comp.EditSpec
		failingOpt := comp.EditOpt(func(*comp.EditSpec, *comp.DefIndex) error {
			return errors.New("boom")
		})

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Init to panic when an opt returns an error")
			}
		}()
		s.Init(&mi, failingOpt)
	})
}
