package arch

import (
	"strings"
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
)

func newCheckedPositions(t *testing.T) (*Catalog, *Archetype, *position, comp.Def) {
	t.Helper()
	cat := newTestCatalog()
	pos, _ := testMetas()
	a := &cat.Archetypes[cat.Upsert(comp.Composition{}.With(pos))]
	p := spawnEntity(t, a, 1)
	spawnEntity(t, a, 2)
	v := (*position)(a.Table.ComponentAt(a.Table.ChunkPtrAt(p.Idx), p.Slot, pos.ID))
	return cat, a, v, pos
}

// a chunk whose bytes changed while neither its counter nor its table's version moved is named
func TestMutCheck_PanicsOnAWritePastTheCounter(t *testing.T) {
	cat, _, v, _ := newCheckedPositions(t)
	m := NewMutCheck()
	m.Check(cat)
	m.Check(cat)

	v.X = 3
	defer func() {
		msg, _ := recover().(string)
		if !strings.Contains(msg, "[arch.position]") {
			t.Errorf("panic %q does not name the archetype [arch.position]", msg)
		}
	}()
	m.Check(cat)
	t.Fatal("expected a write past the counter to panic")
}

// a counted write, a moved table version, a new chunk or a reset check is no write past the
// counter
func TestMutCheck_QuietWhenTheChangeIsTold(t *testing.T) {
	cat, a, v, pos := newCheckedPositions(t)
	m := NewMutCheck()
	m.Check(cat)

	v.X = 1
	*a.Table.VerAt(a.Table.ChunkPtrAt(0)) += 1
	m.Check(cat)

	v.X = 2
	a.Table.RemoveAt(a.Table.ChunkPtrAt(0), 1)
	m.Check(cat)

	_, _, chunkCap := a.Table.ReserveSlots(1)
	a.Table.ReserveSlots(chunkCap)
	a.Table.ReleaseSlots()
	m.Check(cat)
	a.Table.Purge()
	a.Table.ReserveSlots(2 * chunkCap)
	a.Table.ReleaseSlots()
	(*position)(a.Table.ComponentAt(a.Table.ChunkPtrAt(1), 0, pos.ID)).Y = 5
	m.Check(cat)

	other := &cat.Archetypes[cat.Upsert(comp.Composition{}.With(pos).With(testVel()))]
	spawnEntity(t, other, 3)
	m.Check(cat)

	v.X = 4
	m.Reset()
	m.Check(cat)
}

func testVel() comp.Def {
	_, vel := testMetas()
	return vel
}
