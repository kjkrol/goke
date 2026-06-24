package comp_test

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

func TestAccessOpt_Include(t *testing.T) {
	var mi comp.DefIndex
	mi.Init()
	var s comp.AccessSpec

	if err := comp.Include[position]()(&s, &mi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.TagIDs) != 1 {
		t.Errorf("expected Include to add a TagID, got %v", s.TagIDs)
	}
	if len(s.CompInfos) != 0 {
		t.Errorf("expected Include to not track a data column, got %v", s.CompInfos)
	}
}

func TestAccessOpt_Track(t *testing.T) {
	var mi comp.DefIndex
	mi.Init()
	var s comp.AccessSpec
	var col1 iter.ArrayRef[position]
	var col2 iter.ArrayRef[velocity]

	if err := comp.Track(&col1)(&s, &mi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col1.Idx != 0 {
		t.Errorf("expected first Track to set Idx 0, got %d", col1.Idx)
	}

	if err := comp.Track(&col2)(&s, &mi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col2.Idx != 1 {
		t.Errorf("expected second Track to set Idx 1, got %d", col2.Idx)
	}
	if len(s.CompInfos) != 2 {
		t.Errorf("expected 2 tracked data columns, got %v", s.CompInfos)
	}
}

func TestAccessOpt_Exclude(t *testing.T) {
	var mi comp.DefIndex
	mi.Init()
	var s comp.AccessSpec

	if err := comp.Exclude[position]()(&s, &mi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.ExCompIDs) != 1 {
		t.Errorf("expected Exclude to add an ExCompID, got %v", s.ExCompIDs)
	}
}
