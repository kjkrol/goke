package goke_test

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// Padded has trailing bools, leaving six bytes of alignment padding.
type Padded struct {
	A, B, C, D uint64
	X, Y       bool
}

// Named forces the archetype's chunk to be GC-scanned.
type Named struct {
	N     uint64
	Label string
}

// Padded and Named share one archetype, so its chunk is GC-scanned. A stale
// address parked in Padded's padding must not be followed. A regression here
// aborts the test binary rather than failing.
func TestTypedChunks_PaddingIsNotScannedAsPointers(t *testing.T) {
	ecs := goke.New()
	_ = ecs.RegComp[Padded]()
	_ = ecs.RegComp[Named]()

	// Inside the heap, but no longer pointing at a live object.
	junk := make([]byte, 1<<20)
	stale := uintptr(unsafe.Pointer(&junk[1<<19]))
	junk = nil
	runtime.GC()
	runtime.GC()

	const n = 64
	var padded goke.Comp[Padded]
	var named goke.Comp[Named]
	var ids []uid.UID64
	var q *goke.Query

	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&padded, &named)
		f.Create(n)
		for f.Next() {
			cur := &f.Cursor
			ps, ns := padded.Slice(cur), named.Slice(cur)
			for i := range f.IDs {
				ps[i] = Padded{A: uint64(i), X: true}
				ns[i] = Named{N: uint64(i), Label: "keep me alive"}

				// The bools-plus-padding word.
				*(*uintptr)(unsafe.Add(unsafe.Pointer(&ps[i]), 32)) = stale
			}
			ids = append(ids, f.IDs...)
		}
		q = si.NewQueryBuilder(&named).Build()
	}})

	runtime.GC()
	runtime.GC()

	if len(ids) != n {
		t.Fatalf("spawned %d entities, want %d", len(ids), n)
	}

	// The real pointers must still be intact.
	seen := 0
	q.All()
	for q.Next() {
		for _, v := range named.Slice(q.Cursor()) {
			if v.Label != "keep me alive" {
				t.Fatalf("string component read back as %q — its backing array was collected", v.Label)
			}
			seen++
		}
	}
	if seen != n {
		t.Errorf("read back %d string components, want %d", seen, n)
	}
}
