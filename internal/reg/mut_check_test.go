package reg_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/reg"
	"github.com/kjkrol/goke/v3/iter"
)

// with MutCheck, CheckWrites names the archetype of a value written past its chunk's counter;
// UpsertComp's in-place write is counted, and Reset forgets what was seen
func TestRegistry_CheckWrites(t *testing.T) {
	cfg := reg.DefaultConfig()
	cfg.MutCheck = true
	var r reg.Registry
	r.Init(cfg)
	posID := r.RegComp(reflect.TypeFor[Position]())

	var pos iter.ArrayRef[Position]
	f := r.CreateFactory(comp.Add(&pos))
	f.Create(2)
	f.Next()
	id := f.IDs[0]
	values := pos.Slice(&f.Cursor)
	r.CheckWrites()

	p, err := r.UpsertComp(id, posID)
	if err != nil {
		t.Fatal(err)
	}
	*(*Position)(p) = Position{X: 1}
	r.CheckWrites()

	values[1].X = 2
	r.Reset()
	r.CheckWrites()

	r.RegComp(reflect.TypeFor[Position]())
	f = r.CreateFactory(comp.Add(&pos))
	f.Create(1)
	f.Next()
	values = pos.Slice(&f.Cursor)
	r.CheckWrites()
	values[0].Y = 3
	defer func() {
		if msg, _ := recover().(string); !strings.Contains(msg, "reg_test.Position") {
			t.Errorf("panic %q does not name reg_test.Position", msg)
		}
	}()
	r.CheckWrites()
	t.Fatal("expected a write past the counter to panic")
}

// without MutCheck, CheckWrites does nothing
func TestRegistry_CheckWritesOffByDefault(t *testing.T) {
	r := newRegistry(t)
	r.RegComp(reflect.TypeFor[Position]())
	var pos iter.ArrayRef[Position]
	f := r.CreateFactory(comp.Add(&pos))
	f.Create(1)
	f.Next()
	r.CheckWrites()
	pos.Slice(&f.Cursor)[0].X = 1
	r.CheckWrites()
}
