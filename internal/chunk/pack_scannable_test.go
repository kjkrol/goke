package chunk

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/comp"
)

// TestPack_AddChunks_ScannableLayout exercises AddChunks' NeedsScan branch
// (and, through it, ScannableBytes) end-to-end: a chunk allocated as
// GC-scanned memory must still behave exactly like an ordinary byte chunk
// for reads/writes through the normal offset arithmetic.
func TestPack_AddChunks_ScannableLayout(t *testing.T) {
	var layout Layout
	layout.Init([]comp.Def{
		{ID: 1, Size: unsafe.Sizeof(layoutTestOffChunk{}), Align: 8, Type: reflect.TypeFor[layoutTestOffChunk]()},
	})
	if !layout.NeedsScan {
		t.Fatal("expected NeedsScan true for this layout")
	}

	var g Pack
	g.Init(layout)

	ptr := g.ChunkPtr(0)
	if ptr == nil {
		t.Fatal("expected a non-nil chunk pointer")
	}

	fieldPtr := (*layoutTestOffChunk)(unsafe.Add(ptr, layout.Offsets[1]))
	*fieldPtr = layoutTestOffChunk{Name: "hello"}
	if fieldPtr.Name != "hello" {
		t.Errorf("expected to read back %q, got %q", "hello", fieldPtr.Name)
	}

	// Growing by more than one chunk also goes through the NeedsScan branch
	// of AddChunks, not just the single-chunk spare-reuse path.
	g.AddChunks(2)
	if g.NumChunks() != 3 {
		t.Fatalf("expected 3 chunks, got %d", g.NumChunks())
	}
	for i := Idx(0); i < 3; i++ {
		if g.ChunkPtr(i) == nil {
			t.Errorf("expected non-nil ChunkPtr(%d)", i)
		}
	}
}
