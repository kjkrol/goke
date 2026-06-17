package query

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/mem"
)

func setupBakedTable(t *testing.T) (*BakedTable, arch.ID, unsafe.Pointer) {
	t.Helper()
	var cc comp.MetaIndex
	cc.Init()
	type Pos struct{ X, Y float32 }
	posMeta := cc.Intern(reflect.TypeFor[Pos]())

	var ac arch.Catalog
	ac.Init(func(*arch.Archetype) {})
	archID := ac.Upsert(comp.Composition{}.With(posMeta))
	a := &ac.Archetypes[archID]
	a.Table.AddEntity(uid.UID64(1))

	var btc BakedTablesCatalog
	btc.Add(a, []comp.Meta{posMeta})
	bt := btc.Get(archID)
	return bt, archID, a.Table.ChunkPtr(0)
}

func TestBakedTable_ChunkPtr(t *testing.T) {
	bt, _, expectedPtr := setupBakedTable(t)

	if bt.Table.NumChunks() != 1 {
		t.Fatalf("expected 1 chunk, got %d", bt.Table.NumChunks())
	}
	if bt.Table.ChunkPtr(0) != expectedPtr {
		t.Error("ChunkPtr(0) does not match table's base pointer")
	}
}

func TestBakedTable_ChunkLen(t *testing.T) {
	bt, _, _ := setupBakedTable(t)

	if l := bt.Table.ChunkLen(mem.ChunkIdx(0)); l != 1 {
		t.Errorf("expected ChunkLen 1, got %d", l)
	}
}

func TestBakedTable_CompOffsets(t *testing.T) {
	bt, _, _ := setupBakedTable(t)

	if len(bt.CompOffsets) != 1 {
		t.Fatalf("expected 1 CompOffset, got %d", len(bt.CompOffsets))
	}
	// offset must be > 0 — entity column occupies offset 0
	if bt.CompOffsets[0] == 0 {
		t.Error("expected CompOffset[0] > 0 (entity column is at 0)")
	}
}

func TestBakedTable_PointerArithmetic(t *testing.T) {
	bt, _, _ := setupBakedTable(t)

	chunkPtr := bt.Table.ChunkPtr(0)
	offset := bt.CompOffsets[0]

	// pointer at slot 0: chunkPtr + offset + 0*sizeof(Pos)
	type Pos struct{ X, Y float32 }
	size := unsafe.Sizeof(Pos{})
	expected := unsafe.Add(chunkPtr, offset+uintptr(0)*size)
	got := unsafe.Add(chunkPtr, offset)

	if got != expected {
		t.Errorf("pointer arithmetic mismatch: got %p, expected %p", got, expected)
	}
}
