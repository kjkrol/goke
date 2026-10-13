package goke_test

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type tagOnly struct{}

func TestColumn_ReadsWhatCompDoes(t *testing.T) {
	ecs := goke.New()
	var pos goke.Comp[Position]
	posCol := goke.NewColumn(reflect.TypeFor[Position]())
	velCol := goke.NewOptColumn(reflect.TypeFor[Velocity]())
	tagCol := goke.NewColumn(reflect.TypeFor[tagOnly]())
	var all, seek, typed *goke.Query
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&pos, tagCol)
		f.Create(3)
		for f.Next() {
			for i, p := range pos.Slice(&f.Cursor) {
				pos.Slice(&f.Cursor)[i] = Position{X: float32(i), Y: p.Y}
			}
			ids = append(ids, f.IDs...)
			if tagCol.Base(&f.Cursor) != nil || tagCol.At(&f.Cursor) != nil {
				t.Error("a tag's column has memory")
			}
		}
		all = si.NewQueryBuilder(posCol).Optional(velCol).Build()
		seek = si.NewQueryBuilder(posCol).Optional(velCol).Build()
		typed = si.NewQueryBuilder(&pos).Build()
	}})
	if tagCol.Idx != -1 || tagCol.Size() != 0 || tagCol.Type() != reflect.TypeFor[tagOnly]() {
		t.Errorf("a tag's column: Idx %d, Size %d, Type %v", tagCol.Idx, tagCol.Size(), tagCol.Type())
	}
	if velCol.Type() != reflect.TypeFor[Velocity]() || velCol.Size() != unsafe.Sizeof(Velocity{}) {
		t.Errorf("the optional column: Type %v, Size %d", velCol.Type(), velCol.Size())
	}
	seen := 0
	for all.All(); all.Next(); {
		cur := all.Cursor()
		base := posCol.Base(cur)
		if velCol.Present(cur) || velCol.Base(cur) != nil {
			t.Fatal("an absent optional column is present")
		}
		for i := range cur.IDs {
			got := *(*Position)(unsafe.Add(base, uintptr(i)*posCol.Size()))
			if got.X != float32(i) {
				t.Errorf("entity %d read %v", i, got)
			}
			seen++
		}
	}
	if seen != 3 {
		t.Fatalf("walked %d entities, want 3", seen)
	}
	if !seek.Seek(ids[2]) || !typed.Seek(ids[2]) {
		t.Fatal("Seek failed")
	}
	cur, tcur := seek.Cursor(), typed.Cursor()
	if velCol.At(cur) != nil {
		t.Error("an absent optional column has a value at the entity")
	}
	if posCol.At(cur) != unsafe.Pointer(pos.At(tcur)) || pos.At(tcur).X != 2 {
		t.Errorf("Column.At is not where Comp.At is: %v, %v", posCol.At(cur), pos.At(tcur))
	}
	if pos.Index() != 0 || posCol.Idx != 0 {
		t.Errorf("track positions: Comp %d, Column %d; want 0 and 0", pos.Index(), posCol.Idx)
	}
	if tokens := goke.LoadComps(posCol, &pos); tokens[0].Name != tokens[1].Name {
		t.Errorf("a Column loads as %q, the Comp as %q", tokens[0].Name, tokens[1].Name)
	}
}

func TestReserve_StagesWhatAddOneDoes(t *testing.T) {
	ecs := goke.New()
	velID := ecs.RegCompOf(reflect.TypeFor[Velocity]())
	if velID != ecs.RegComp[Velocity]() {
		t.Fatal("RegCompOf and RegComp disagree on the id")
	}
	var pos goke.Comp[Position]
	var vel goke.OptComp[Velocity]
	velCol := goke.NewOptColumn(reflect.TypeFor[Velocity]())
	var q, typed *goke.Query
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		if si.RegCompOf(reflect.TypeFor[Velocity]()) != velID {
			t.Error("SysInit.RegCompOf gives another id")
		}
		f := si.NewFactory(&pos)
		f.Create(2)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
		q = si.NewQueryBuilder(&pos).Optional(velCol).Build()
		typed = si.NewQueryBuilder(&pos).Optional(&vel).Build()
	}})
	if vel.Index() != 0 || velCol.Idx != 0 {
		t.Errorf("optional positions: OptComp %d, OptColumn %d; want 0 and 0", vel.Index(), velCol.Idx)
	}
	var own *goke.CmdBuf
	adder := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
		own = cb
		*(*Velocity)(cb.Reserve(ids[0], velID)) = Velocity{VX: 7}
	}})
	if ecs.CmdBuf(adder) == nil {
		t.Fatal("no CmdBuf for a registered system")
	}
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(adder, d)
		_ = ctx.Sync()
	})
	ecs.Tick(0)
	if own != ecs.CmdBuf(adder) {
		t.Error("CmdBuf(r) is not the buffer Update was handed")
	}
	if !q.Seek(ids[0]) || !typed.Seek(ids[0]) || velCol.At(q.Cursor()) == nil {
		t.Fatal("the reserved component did not land")
	}
	cur, tcur := q.Cursor(), typed.Cursor()
	if v := *(*Velocity)(velCol.At(cur)); v.VX != 7 || velCol.At(cur) != unsafe.Pointer(vel.At(tcur)) {
		t.Errorf("landed %v at %v, OptComp at %v; want VX 7 at one place", v, velCol.At(cur), vel.At(tcur))
	}
	bases := 0
	for q.All(); q.Next(); {
		cur := q.Cursor()
		if !velCol.Present(cur) {
			continue
		}
		base := velCol.Base(cur)
		for typed.All(); typed.Next(); {
			if tcur := typed.Cursor(); vel.Present(tcur) && base == unsafe.Pointer(&vel.Slice(tcur)[0]) {
				bases++
			}
		}
	}
	if bases != 1 {
		t.Errorf("OptColumn.Base matched OptComp.Slice on %d chunks, want the one chunk", bases)
	}
	if q.Seek(ids[1]) && velCol.At(q.Cursor()) != nil {
		t.Error("the other entity got the component too")
	}
}

func TestLoadCompOf_NamesAsLoadCompDoes(t *testing.T) {
	a, b := goke.LoadComp[Position](), goke.LoadCompOf(reflect.TypeFor[Position]())
	if a.Name != b.Name {
		t.Errorf("LoadComp names %q, LoadCompOf %q", a.Name, b.Name)
	}
}
