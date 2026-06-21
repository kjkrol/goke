package query

import (
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/colstore"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

func setupBakedTable(t *testing.T) (*BakedTable, arch.ID) {
	t.Helper()
	var cc comp.DefIndex
	cc.Init()
	type Pos struct{ X, Y float32 }
	posMeta := cc.Intern(reflect.TypeFor[Pos]())

	var ac arch.Catalog
	ac.Init(func(*arch.Archetype) {})
	archID := ac.Upsert(comp.Composition{}.With(posMeta))
	a := &ac.Archetypes[archID]
	a.Table.SetIDSeeder(func(dst []uid.UID64, _ colstore.Pos) { dst[0] = uid.UID64(1) })
	idx, _, _ := a.Table.ReserveSlots(1)
	var cur iter.Cursor
	a.Table.SpawnCursor(&cur, idx, 1, nil)
	a.Table.ReleaseSlots()

	var btc BakedTablesCatalog
	btc.Add(a, []comp.Def{posMeta})
	bt := btc.Get(archID)
	return bt, archID
}

func TestBakedTable_Len(t *testing.T) {
	bt, _ := setupBakedTable(t)

	if bt.Table.Len() != 1 {
		t.Errorf("expected Len 1, got %d", bt.Table.Len())
	}
}

func TestBakedTable_CompOffsets(t *testing.T) {
	bt, _ := setupBakedTable(t)

	if len(bt.CompOffsets) != 1 {
		t.Fatalf("expected 1 CompOffset, got %d", len(bt.CompOffsets))
	}
	// offset must be > 0 — entity column occupies offset 0
	if bt.CompOffsets[0] == 0 {
		t.Error("expected CompOffset[0] > 0 (entity column is at 0)")
	}
}
