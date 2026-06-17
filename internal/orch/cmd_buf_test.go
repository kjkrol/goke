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
	compA  comp.Meta
	compB  comp.Meta
	target uid.UID64
}

func (s *modifyTestSystem) Update(_ Lookup, cb *CmdBuf, d time.Duration) {
	cb.RemoveComp(s.target, s.compA)
	AddComp(cb, s.target, s.compB, mockCompB{Msg: "added"})
}

func TestCmdBuf_ComponentCmds(t *testing.T) {
	var registry reg.Registry
	registry.Init(reg.Config{
		Entity: ent.Config{Cap: 100, FreeCap: 100},
		View:   query.Config{Cap: 10},
	})

	compA := registry.RegCompType(reflect.TypeFor[mockCompA]())
	compB := registry.RegCompType(reflect.TypeFor[mockCompB]())

	sched := NewScheduler(&registry, &registry)

	var colA iter.Col[mockCompA]
	f := registry.CreateFactory(comp.Track(&colA), comp.Track(new(iter.Col[mockCompB])))
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

	_, err = registry.GetComp(e, compA.ID)
	if err == nil {
		t.Errorf("Expected compA to be removed")
	}

	ptrB, err := registry.GetComp(e, compB.ID)
	if err != nil {
		t.Fatalf("Expected compB to be present")
	}
	if (*mockCompB)(ptrB).Msg != "added" {
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
