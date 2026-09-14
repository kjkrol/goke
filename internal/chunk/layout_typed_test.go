package chunk

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kjkrol/goke/v3/internal/comp"
)

// opaqueType must reproduce a component's footprint exactly, since it stands in
// for one whose reflect.Type is unknown.
func TestOpaqueType_MatchesSizeAndAlignment(t *testing.T) {
	cases := []struct{ size, align uintptr }{
		{24, 8}, {12, 4}, {6, 2}, {3, 1}, {8, 8}, {4, 4}, {7, 1},
	}
	for _, tc := range cases {
		got := opaqueType(tc.size, tc.align)
		if got.Size() != tc.size {
			t.Errorf("opaqueType(%d, %d).Size() = %d, want %d", tc.size, tc.align, got.Size(), tc.size)
		}
		if uintptr(got.Align()) > tc.align {
			t.Errorf("opaqueType(%d, %d).Align() = %d, want at most %d", tc.size, tc.align, got.Align(), tc.align)
		}
	}
}

// A Def carrying its reflect.Type yields that type; one without falls back to an
// opaque stand-in, and may not do so when the collector has to look inside it.
func TestColumnElem_FallsBackOnlyForPointerFreeDefs(t *testing.T) {
	typed := comp.Def{ID: 1, Size: 8, Align: 8, Type: reflect.TypeFor[uint64]()}
	if got := columnElem(typed); got != reflect.TypeFor[uint64]() {
		t.Errorf("columnElem(typed) = %v, want uint64", got)
	}

	untyped := comp.Def{ID: 2, Size: 16, Align: 8}
	if got := columnElem(untyped); got.Size() != 16 {
		t.Errorf("columnElem(untyped).Size() = %d, want 16", got.Size())
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected columnElem to reject a NeedsScan Def with no reflect.Type")
		}
		if msg, _ := r.(string); !strings.Contains(msg, "NeedsScan") {
			t.Errorf("panic message %q does not mention NeedsScan", msg)
		}
	}()
	columnElem(comp.Def{ID: 3, Size: 16, Align: 8, NeedsScan: true})
}

// buildChunkType asserts that the generated type agrees with the layout Init
// computed; a disagreement must surface there, not as misaddressed columns.
func TestBuildChunkType_RejectsLayoutDisagreement(t *testing.T) {
	defs := []comp.Def{{ID: 1, Size: 8, Align: 8, Type: reflect.TypeFor[uint64]()}}

	var good Layout
	good.Init(defs)

	cases := map[string]func(l *Layout){
		"wrong column offset": func(l *Layout) { l.Offsets[1]++ },
		"wrong chunk size":    func(l *Layout) { l.ChunkBytes++ },
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			l := Layout{
				ChunkCap:   good.ChunkCap,
				ChunkBytes: good.ChunkBytes,
				Offsets:    append([]uintptr(nil), good.Offsets...),
			}
			corrupt(&l)

			defer func() {
				if recover() == nil {
					t.Error("expected buildChunkType to panic")
				}
			}()
			buildChunkType(defs, uintptr(good.ChunkCap), &l)
		})
	}
}
