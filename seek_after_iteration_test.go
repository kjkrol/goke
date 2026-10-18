package goke_test

import (
	"testing"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type labelled struct{ N uint64 }
type padding struct{ A, B, C uint64 }

// two archetypes carry labelled at different offsets; an All iteration ends on the second,
// and a Seek into the first must not keep the second's offsets
func TestSeek_AfterAllOverAnotherArchetypeReadsTheRightColumn(t *testing.T) {
	ecs := goke.New()
	var comp goke.Comp[labelled]
	var q *goke.Query
	var plain []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&comp)
		f.Create(3)
		for f.Next() {
			for i, id := range f.Cursor.IDs {
				comp.Slice(&f.Cursor)[i] = labelled{N: uint64(id) + 100}
				plain = append(plain, id)
			}
		}
		var pad goke.Comp[padding]
		g := si.NewFactory(&pad, &comp)
		g.Create(3)
		for g.Next() {
			for i, id := range g.Cursor.IDs {
				comp.Slice(&g.Cursor)[i] = labelled{N: uint64(id) + 200}
			}
		}
		q = si.NewQueryBuilder(&comp).Build()
	}})
	id := plain[0]
	if !q.Seek(id) || comp.At(q.Cursor()).N != uint64(id)+100 {
		t.Fatalf("Seek(%v) reads %v", id, comp.At(q.Cursor()).N)
	}
	for q.All(); q.Next(); {
	}
	if !q.Seek(id) {
		t.Fatal("Seek fails after All")
	}
	if got := comp.At(q.Cursor()).N; got != uint64(id)+100 {
		t.Errorf("Seek(%v) after an All iteration reads %v, want %v", id, got, uint64(id)+100)
	}
	for q.Pick(plain[1:]); q.Next(); {
	}
	if !q.Seek(id) {
		t.Fatal("Seek fails after Pick")
	}
	if got := comp.At(q.Cursor()).N; got != uint64(id)+100 {
		t.Errorf("Seek(%v) after a Pick iteration reads %v, want %v", id, got, uint64(id)+100)
	}
}
