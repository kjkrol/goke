package comp_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
)

type (
	far  struct{}
	gone struct{}
)

func TestPacedBy_ResolvesItsRules(t *testing.T) {
	var mi comp.DefIndex
	mi.Init()
	var s comp.AccessSpec
	step := func() uint64 { return 0 }
	if err := comp.PacedBy(step, comp.Every[far](3), comp.Never[gone]())(&s, &mi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []comp.PaceRule{
		{ID: mi.Intern(reflect.TypeFor[far]()).ID, Period: 3},
		{ID: mi.Intern(reflect.TypeFor[gone]()).ID},
	}
	if s.Pace == nil || len(s.Pace.Rules) != 2 || s.Pace.Rules[0] != want[0] ||
		s.Pace.Rules[1] != want[1] {
		t.Errorf("pace %+v, want the rules %+v", s.Pace, want)
	}
}

func TestPacedBy_RefusesASecondPaceNoStepAndEveryZero(t *testing.T) {
	step := func() uint64 { return 0 }
	for name, opts := range map[string][]comp.AccessOpt{
		"twice":   {comp.PacedBy(step), comp.PacedBy(step)},
		"no step": {comp.PacedBy(nil)},
		"every 0": {comp.PacedBy(step, comp.Every[far](0))},
	} {
		var mi comp.DefIndex
		mi.Init()
		var s comp.AccessSpec
		var err error
		for _, opt := range opts {
			if err = opt(&s, &mi); err != nil {
				break
			}
		}
		if err == nil {
			t.Errorf("%s: the pace was taken", name)
		}
	}
}

func TestPace_PeriodIsTheSlowestRuleMetNeverWinning(t *testing.T) {
	var mi comp.DefIndex
	mi.Init()
	farID := mi.Intern(reflect.TypeFor[far]()).ID
	goneID := mi.Intern(reflect.TypeFor[gone]()).ID
	slowID := mi.Intern(reflect.TypeFor[position]()).ID
	p := &comp.Pace{Rules: []comp.PaceRule{
		{ID: farID, Period: 2},
		{ID: slowID, Period: 5},
		{ID: goneID},
	}}
	var none, farOnly, both, never comp.Mask
	farOnly = farOnly.Set(farID)
	both = both.Set(farID).Set(slowID)
	never = never.Set(farID).Set(goneID)
	for _, c := range []struct {
		name string
		mask comp.Mask
		want uint64
	}{
		{"none", none, 1},
		{"far", farOnly, 2},
		{"far and slow", both, 5},
		{"far and gone", never, 0},
	} {
		if got := p.Period(c.mask); got != c.want {
			t.Errorf("%s: period %d, want %d", c.name, got, c.want)
		}
	}
	var unpaced *comp.Pace
	if got := unpaced.Period(never); got != 1 {
		t.Errorf("no pace: period %d, want every run", got)
	}
}
