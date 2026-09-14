package chunk

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/comp"
)

// A chunk allocated through the archetype's generated type must behave like an
// ordinary byte chunk under the normal offset arithmetic.
func TestPack_AddChunks_TypedLayout(t *testing.T) {
	var layout Layout
	layout.Init([]comp.Def{
		{ID: 1, Size: unsafe.Sizeof(layoutTestOffChunk{}), Align: 8, Type: reflect.TypeFor[layoutTestOffChunk](), NeedsScan: true},
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
