package comp_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/v2/internal/comp"
)

func TestAccessSpec_Comp(t *testing.T) {
	t.Run("appends a data component", func(t *testing.T) {
		var s comp.AccessSpec
		def := comp.Def{ID: 1, Size: 8, Type: reflect.TypeFor[position]()}

		if err := s.Comp(def); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(s.CompInfos) != 1 || s.CompInfos[0] != def {
			t.Errorf("expected CompInfos to contain def, got %v", s.CompInfos)
		}
	})

	t.Run("rejects a duplicate component ID", func(t *testing.T) {
		var s comp.AccessSpec
		def := comp.Def{ID: 1, Size: 8, Type: reflect.TypeFor[position]()}
		_ = s.Comp(def)

		if err := s.Comp(def); err == nil {
			t.Error("expected an error when adding the same component ID twice")
		}
	})

	t.Run("rejects a tag (size 0) as a data column", func(t *testing.T) {
		var s comp.AccessSpec
		tagDef := comp.Def{ID: 1, Size: 0, Type: reflect.TypeFor[position]()}

		if err := s.Comp(tagDef); err == nil {
			t.Error("expected an error when adding a zero-size def as a data column")
		}
	})
}

func TestAccessSpec_Tag(t *testing.T) {
	var s comp.AccessSpec

	if err := s.Tag(5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.TagIDs) != 1 || s.TagIDs[0] != 5 {
		t.Errorf("expected TagIDs to contain 5, got %v", s.TagIDs)
	}

	if err := s.Tag(5); err == nil {
		t.Error("expected an error when tagging the same ID twice")
	}
}

func TestAccessSpec_Exclude(t *testing.T) {
	var s comp.AccessSpec

	if err := s.Exclude(7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.ExCompIDs) != 1 || s.ExCompIDs[0] != 7 {
		t.Errorf("expected ExCompIDs to contain 7, got %v", s.ExCompIDs)
	}

	if err := s.Exclude(7); err == nil {
		t.Error("expected an error when excluding the same ID twice")
	}
}

func TestAccessSpec_CompIDs(t *testing.T) {
	var s comp.AccessSpec
	_ = s.Comp(comp.Def{ID: 3, Size: 8})
	_ = s.Comp(comp.Def{ID: 1, Size: 4})

	ids := s.CompIDs()
	want := []comp.ID{3, 1}
	if len(ids) != len(want) || ids[0] != want[0] || ids[1] != want[1] {
		t.Errorf("expected CompIDs %v in track order, got %v", want, ids)
	}
}

func TestAccessSpec_Compose(t *testing.T) {
	var s comp.AccessSpec
	dataDef := comp.Def{ID: 1, Size: 8}
	_ = s.Comp(dataDef)
	_ = s.Tag(2)

	c := s.Compose()

	if !c.Mask.IsSet(1) || !c.Mask.IsSet(2) {
		t.Errorf("expected Compose's mask to set both tracked and tag bits, got %v", c.Mask)
	}
	if len(c.Defs) != 1 || c.Defs[0] != dataDef {
		t.Errorf("expected Compose's Defs to equal CompInfos, got %v", c.Defs)
	}
}

func TestAccessSpec_Init(t *testing.T) {
	t.Run("applies opts in order", func(t *testing.T) {
		var mi comp.DefIndex
		mi.Init()
		var s comp.AccessSpec

		s.Init(&mi, comp.Include[position](), comp.Exclude[velocity]())

		if len(s.TagIDs) != 1 {
			t.Errorf("expected Include to add one TagID, got %v", s.TagIDs)
		}
		if len(s.ExCompIDs) != 1 {
			t.Errorf("expected Exclude to add one ExCompID, got %v", s.ExCompIDs)
		}
	})

	t.Run("panics when an opt returns an error", func(t *testing.T) {
		var mi comp.DefIndex
		mi.Init()
		var s comp.AccessSpec

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Init to panic when an opt's component is included twice")
			}
		}()
		s.Init(&mi, comp.Include[position](), comp.Include[position]())
	})
}
