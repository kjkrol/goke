package query

import (
	"reflect"
	"testing"
	"time"

	"github.com/kjkrol/goke/v3/internal/comp"
)

type (
	pacedPos  struct{ X float32 }
	pacedFar  struct{}
	pacedGone struct{}
)

// three archetypes of one entity each: near, far (every second run), gone (never); far is made
// after the matcher, so its period is set as it is announced
func pacedMatcher(t *testing.T, now *uint64) (*Matcher, map[uint64]int) {
	t.Helper()
	cat, cc, em := newQueryCatalog()
	pos := cc.Intern(reflect.TypeFor[pacedPos]())
	farID := cc.Intern(reflect.TypeFor[pacedFar]()).ID
	goneID := cc.Intern(reflect.TypeFor[pacedGone]()).ID
	spawn := func(tags ...comp.ID) {
		var s comp.AccessSpec
		s.Comp(pos)
		for _, tag := range tags {
			s.Tag(tag)
		}
		f := em.CreateFactory(s)
		f.Create(1)
		f.Next()
	}
	spawn()
	spawn(goneID)
	var spec comp.AccessSpec
	spec.Comp(pos)
	spec.Pace = &comp.Pace{
		Step: func() uint64 { return *now },
		Rules: []comp.PaceRule{
			{ID: farID, Period: 2},
			{ID: goneID},
		},
	}
	m := cat.AddMatcher(&spec)
	spawn(farID)
	periods := map[uint64]int{}
	for i, bt := range m.BakedTables {
		periods[bt.Period] = i
	}
	if len(periods) != 3 {
		t.Fatalf("periods %v, want 1, 2 and 0", periods)
	}
	return m, periods
}

func TestMatcher_PacedWalksTheTablesWhoseTurnItIs(t *testing.T) {
	var now uint64
	m, periods := pacedMatcher(t, &now)
	const d = 10 * time.Millisecond
	for _, c := range []struct {
		now  uint64
		want map[uint64]time.Duration // by period, the step given
	}{
		{0, map[uint64]time.Duration{1: d}},
		{1, map[uint64]time.Duration{1: d, 2: 2 * d}},
	} {
		now = c.now
		got := map[uint64]time.Duration{}
		for m.Paced(d); m.Next(); {
			got[m.BakedTables[m.tableIdx].Period] = m.Dt()
		}
		if len(got) != len(c.want) || got[1] != c.want[1] || got[2] != c.want[2] {
			t.Errorf("run %d: walked %v, want %v", c.now, got, c.want)
		}
	}
	due := map[int]bool{}
	for m.All(); m.Next(); {
		due[m.tableIdx] = m.Due()
	}
	if !due[periods[1]] || !due[periods[2]] || due[periods[0]] {
		t.Errorf("run 1, walking all: due %v, want the near and the far, never the gone", due)
	}
}

func TestMatcher_ClearForgetsThePace(t *testing.T) {
	var now uint64
	m, _ := pacedMatcher(t, &now)
	m.Clear()
	if m.pace != nil || m.Next() {
		t.Error("a cleared matcher keeps its pace or walks on")
	}
}
