package comp_test

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func TestComposition_With(t *testing.T) {
	t.Run("data component", func(t *testing.T) {
		var c comp.Composition
		dataDef := comp.Def{ID: 1, Size: 8}

		c = c.With(dataDef)

		if !c.Mask.IsSet(1) {
			t.Error("expected mask bit 1 to be set")
		}
		if len(c.Defs) != 1 || c.Defs[0] != dataDef {
			t.Errorf("expected Defs to contain the new data def, got %v", c.Defs)
		}
	})

	t.Run("tag (size 0) updates only the mask", func(t *testing.T) {
		var c comp.Composition
		tagDef := comp.Def{ID: 2, Size: 0}

		c = c.With(tagDef)

		if !c.Mask.IsSet(2) {
			t.Error("expected mask bit 2 to be set for a tag")
		}
		if len(c.Defs) != 0 {
			t.Errorf("expected Defs to stay empty for a tag, got %v", c.Defs)
		}
	})
}

func TestComposition_Without(t *testing.T) {
	t.Run("removes a present component", func(t *testing.T) {
		var c comp.Composition
		c = c.With(comp.Def{ID: 1, Size: 8}).With(comp.Def{ID: 2, Size: 4})

		result := c.Without(1)

		if result.Mask.IsSet(1) {
			t.Error("expected mask bit 1 to be cleared")
		}
		if !result.Mask.IsSet(2) {
			t.Error("expected mask bit 2 to remain set")
		}
		if len(result.Defs) != 1 || result.Defs[0].ID != 2 {
			t.Errorf("expected only comp 2 left in Defs, got %v", result.Defs)
		}
	})

	t.Run("no-op when the component isn't set", func(t *testing.T) {
		var c comp.Composition
		c = c.With(comp.Def{ID: 1, Size: 8})

		result := c.Without(99)

		if !result.Mask.Equals(c.Mask) {
			t.Errorf("expected unchanged mask, got %v want %v", result.Mask, c.Mask)
		}
		if len(result.Defs) != len(c.Defs) {
			t.Errorf("expected unchanged Defs, got %v want %v", result.Defs, c.Defs)
		}
	})
}
