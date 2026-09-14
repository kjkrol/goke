package orch

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/uid"
)

// padded has trailing bools, leaving six bytes of alignment padding.
type padded struct {
	A, B uint64
	X, Y bool
}

// withString's backing array is reachable only through the staged payload
// until Sync runs.
type withString struct {
	A    uint64
	Name string
}

// scanTestBuf returns a CmdBuf wired to a registry holding padded and
// withString, plus their ids.
func scanTestBuf() (*CmdBuf, comp.ID, comp.ID) {
	var defs comp.DefIndex
	defs.Init()
	paddedID := defs.Intern(reflect.TypeFor[padded]()).ID
	stringID := defs.Intern(reflect.TypeFor[withString]()).ID
	cb := NewCmdBuf()
	cb.SetDefs(&defs)
	return cb, paddedID, stringID
}

func TestCmdBuf_PointerFreePayloadsUseTheArena(t *testing.T) {
	cb, paddedID, stringID := scanTestBuf()

	AddOne(cb, uid.UID64(1), paddedID, padded{A: 1, X: true})
	if cb.plain.offset == 0 {
		t.Error("pointer-free payload did not land in the arena")
	}

	before := cb.plain.offset
	AddOne(cb, uid.UID64(2), stringID, withString{Name: "keep me alive"})
	if cb.plain.offset != before {
		t.Error("payload with a string field consumed arena space; it must be staged in its own typed allocation")
	}
}

// A staged component whose padding holds the bytes of a dead heap address must
// survive a GC cycle. A regression here kills the test binary.
func TestCmdBuf_StalePointerInPaddingSurvivesGC(t *testing.T) {
	junk := make([]byte, 1<<20)
	stale := uintptr(unsafe.Pointer(&junk[1<<19]))
	junk = nil
	runtime.GC()
	runtime.GC()

	cb, paddedID, _ := scanTestBuf()
	AddOne(cb, uid.UID64(1), paddedID, padded{A: 1, B: 2, X: true})

	// Offset 16 is the bools-plus-padding word.
	*(*uintptr)(unsafe.Add(cb.cmds[0].dataPtr, 16)) = stale

	runtime.GC()

	if got := *(*uintptr)(unsafe.Add(cb.cmds[0].dataPtr, 16)); got != stale {
		t.Errorf("padding word = %#x, want %#x — the payload must survive verbatim", got, stale)
	}
}

// A staged string must stay alive until Sync even when the caller drops every
// other reference to it.
func TestCmdBuf_StagedStringSurvivesGCUntilSync(t *testing.T) {
	cb, _, stringID := scanTestBuf()

	label := string([]byte("keep me alive")) // heap-allocated, not a literal
	AddOne(cb, uid.UID64(1), stringID, withString{A: 7, Name: label})
	label = ""

	runtime.GC()
	runtime.GC()

	got := (*withString)(cb.cmds[0].dataPtr)
	if got.A != 7 || got.Name != "keep me alive" {
		t.Errorf("staged payload = %+v, want A=7 and its string intact — the backing array was collected", *got)
	}
}

func TestCmdBuf_AddOne_PanicsOnCompIDTypeMismatch(t *testing.T) {
	cb, paddedID, stringID := scanTestBuf()

	// A mismatched type would memcpy the wrong size into padded's column.
	assertPanics(t, "mismatched type", func() {
		AddOne(cb, uid.UID64(1), paddedID, withString{Name: "wrong column"})
	})

	assertPanics(t, "unregistered id", func() {
		AddOne(cb, uid.UID64(1), comp.ID(120), padded{A: 1})
	})

	AddOne(cb, uid.UID64(1), stringID, withString{Name: "right column"})
}

func assertPanics(t *testing.T, what string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected AddOne to panic", what)
		}
	}()
	fn()
}

// A payload the collector must follow into is staged in its own allocation of
// the element type, not carved out of the pointer-free arena.
func TestCmdBuf_AddCompValue_StagesScannedPayloadSeparately(t *testing.T) {
	cb := NewCmdBuf()
	ids := []uid.UID64{1, 2, 3}
	var zero withString

	before := cb.plain.offset
	ptr := cb.AddCompValue(&stubValueMigrator{}, bulk.ChunkSnapshot{}, ids,
		unsafe.Sizeof(zero), unsafe.Alignof(zero), reflect.TypeFor[withString](), true)
	if ptr == nil {
		t.Fatal("expected a non-nil payload pointer")
	}

	// Only the id copies may come from the arena; the payload must not.
	idBytes := len(ids) * int(unsafe.Sizeof(uid.UID64(0)))
	if grew := cb.plain.offset - before; grew > idBytes {
		t.Errorf("arena grew by %d bytes, want at most %d — the payload took arena space", grew, idBytes)
	}

	vals := unsafe.Slice((*withString)(ptr), len(ids))
	for i := range vals {
		vals[i] = withString{A: uint64(i), Name: "keep me alive"}
	}
	runtime.GC()
	if vals[2].Name != "keep me alive" {
		t.Errorf("payload string read back as %q", vals[2].Name)
	}
}
