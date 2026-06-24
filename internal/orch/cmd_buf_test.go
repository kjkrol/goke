package orch

import (
	"reflect"
	"testing"
	"time"

	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
	"github.com/kjkrol/goke/internal/query"
	"github.com/kjkrol/goke/internal/reg"
	"github.com/kjkrol/goke/iter"
	"github.com/kjkrol/uid"
)

type mockCompA struct {
	Val int
}

type mockCompB struct {
	Msg string
}

type modifyTestSystem struct {
	compA  comp.ID
	compB  comp.ID
	target uid.UID64
}

func (s *modifyTestSystem) Update(cb *CmdBuf, d time.Duration) {
	cb.RemoveComp(s.target, s.compA)
	AddComp(cb, s.target, s.compB, mockCompB{Msg: "added"})
}

// AddComp must not try to copy any bytes for a zero-size (tag) component —
// it should queue a command with a nil dataPtr instead of dereferencing a
// zero-size value's address.
func TestAddComp_ZeroSizeComponent(t *testing.T) {
	type tag struct{}
	cb := NewCmdBuf()

	AddComp(cb, uid.UID64(1), comp.ID(5), tag{})

	if len(cb.cmds) != 1 {
		t.Fatalf("expected 1 queued command, got %d", len(cb.cmds))
	}
	cmd := cb.cmds[0]
	if cmd.dataPtr != nil {
		t.Errorf("expected nil dataPtr for a zero-size component, got %v", cmd.dataPtr)
	}
	if cmd.size != 0 {
		t.Errorf("expected size 0, got %d", cmd.size)
	}
}

func TestCmdBuf_ComponentCmds(t *testing.T) {
	var registry reg.Registry
	registry.Init(reg.Config{
		Entity:  ent.Config{Cap: 100, FreeCap: 100},
		Matcher: query.Config{Cap: 10},
	})

	compA := registry.RegComp(reflect.TypeFor[mockCompA]())
	compB := registry.RegComp(reflect.TypeFor[mockCompB]())

	sched := NewScheduler(&registry)

	var colA iter.ArrayRef[mockCompA]
	f := registry.CreateFactory(comp.Add(&colA), comp.Add(new(iter.ArrayRef[mockCompB])))
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*colA.At(&f.Cursor) = mockCompA{Val: 100}

	sys := &modifyTestSystem{
		compA:  compA,
		compB:  compB,
		target: e,
	}

	sched.Register(sys)
	sched.Run(sys, 0)

	err := sched.Sync()
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// compA should be gone: a matcher requiring it must not match e
	matcherA := registry.AddMatcher(comp.Include[mockCompA]())
	if matcherA.Pick([]uid.UID64{e}).Next() {
		t.Errorf("Expected compA to be removed")
	}

	// compB should be present with the assigned value
	var colB iter.ArrayRef[mockCompB]
	matcherB := registry.AddMatcher(comp.Track(&colB))
	if !matcherB.Pick([]uid.UID64{e}).Next() {
		t.Fatalf("Expected compB to be present")
	}
	if colB.At(&matcherB.Cursor).Msg != "added" {
		t.Errorf("Expected compB.Msg to be 'added'")
	}
}

func TestCmdBuf_Clear(t *testing.T) {
	cb := NewCmdBuf()

	cb.RemoveEntity(uid.UID64(1))

	if len(cb.cmds) == 0 {
		t.Fatalf("Expected commands to not be empty")
	}

	cb.Clear()

	if len(cb.cmds) != 0 {
		t.Errorf("Expected commands to be empty")
	}
}

func TestCmdBuf_ReserveSpace(t *testing.T) {
	cb := NewCmdBuf()

	p1 := cb.reserveSpace(10, 1)
	if p1 == nil {
		t.Errorf("Expected non-nil pointer")
	}

	p2 := cb.reserveSpace(8192, 1)
	if p2 == nil {
		t.Errorf("Expected non-nil pointer for large alloc")
	}
}

// reserveSpace must replace an existing-but-undersized page when advancing
// onto it, not just when appending a brand new one.
func TestCmdBuf_ReserveSpace_GrowsExistingUndersizedPage(t *testing.T) {
	cb := NewCmdBuf()

	cb.reserveSpace(4000, 1) // fills most of page 0 (4096 bytes)
	cb.reserveSpace(200, 1)  // spills onto a freshly appended page 1 (4096 bytes)
	if len(cb.pages) != 2 {
		t.Fatalf("setup error: expected 2 pages, got %d", len(cb.pages))
	}

	cb.Clear() // pageIdx/offset reset to 0, but both pages (and their sizes) survive

	cb.reserveSpace(4000, 1)      // fills page 0 again
	p := cb.reserveSpace(5000, 1) // page 1 (4096 bytes) is now too small for 5000
	if p == nil {
		t.Fatal("expected a non-nil pointer after growing the undersized page")
	}
	if len(cb.pages[1]) < 5000 {
		t.Errorf("expected page 1 to grow to at least 5000 bytes, got %d", len(cb.pages[1]))
	}
}
