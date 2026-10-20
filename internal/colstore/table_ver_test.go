package colstore

import (
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
)

// a cursor filled by a walk or a spawn points at the counter of its own chunk
func TestTable_CursorsPointAtTheirChunksCounter(t *testing.T) {
	tbl := newTestTable(t, []comp.Def{{ID: 1, Size: 8, Align: 8}})
	cur := newCursor(1)
	chunkCap := int(tbl.chunkPack.Layout.ChunkCap)
	tbl.ReserveSlots(chunkCap + 1)
	tbl.SpawnCursor(cur, 0, chunkCap, []ColBake{{Offset: 0, CompSize: 8}})
	if cur.Ver != tbl.VerAt(tbl.ChunkPtrAt(0)) {
		t.Error("SpawnCursor's Ver is not chunk 0's counter")
	}
	tbl.SpawnCursor(cur, 1, 1, []ColBake{{Offset: 0, CompSize: 8}})

	idx, ok := tbl.FillCursorNext(cur, 1, []uintptr{0})
	if !ok || idx != 1 {
		t.Fatalf("FillCursorNext(1) = %d, %v", idx, ok)
	}
	if cur.Ver != tbl.VerAt(tbl.ChunkPtrAt(1)) {
		t.Error("FillCursorNext's Ver is not chunk 1's counter")
	}
	if *cur.Ver != 1 {
		t.Errorf("chunk 1 counts %d after one spawn, want 1", *cur.Ver)
	}
}

// the bytes a check hashes are every column of the chunk, the counter left out
func TestTable_ChunkBytesStopShortOfTheCounter(t *testing.T) {
	tbl := newTestTable(t, []comp.Def{{ID: 1, Size: 8, Align: 8}})
	tbl.AllocSlots(0, 1)
	tbl.ReserveSlots(int(tbl.chunkPack.Layout.ChunkCap) + 1)

	if got := tbl.NumChunks(); got != 2 {
		t.Fatalf("NumChunks = %d, want 2", got)
	}
	b := tbl.ChunkBytes(1)
	if uintptr(len(b)) != tbl.VerOffset() {
		t.Errorf("ChunkBytes holds %d bytes, want %d before the counter", len(b), tbl.VerOffset())
	}
	if &b[0] != (*byte)(tbl.ChunkPtrAt(1)) {
		t.Error("ChunkBytes(1) does not start at chunk 1")
	}
	before := *tbl.VerAt(tbl.ChunkPtrAt(0))
	tbl.AllocSlots(0, 1)
	if got := *tbl.VerAt(tbl.ChunkPtrAt(0)); got != before+1 {
		t.Errorf("an append counts %d on its chunk, want %d", got, before+1)
	}
	if tbl.Version() != 0 {
		t.Errorf("appends moved the table version to %d", tbl.Version())
	}
}
