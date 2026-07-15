package addr

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/colstore"
)

// fakeChunks backs fakePtr with real memory — fabricating pointers from
// integers is invalid for the GC and flagged by vet's unsafeptr check.
var fakeChunks [4]byte

// fakePtr returns a distinct non-nil unsafe.Pointer for use as a fake chunk address in tests.
func fakePtr(n uintptr) unsafe.Pointer { return unsafe.Pointer(&fakeChunks[n]) }

func TestBook_SeedAndGet(t *testing.T) {
	var b Book
	b.Init(8, 8)

	dst := make([]uid.UID64, 3)
	ptr := fakePtr(0)
	b.Seed(dst, arch.ID(5), ptr, colstore.Slot(0))

	for i, id := range dst {
		entry, ok := b.Get(id)
		if !ok {
			t.Fatalf("expected entry for seeded id %d", i)
		}
		if entry.ArchID != arch.ID(5) {
			t.Errorf("id %d: expected ArchID 5, got %d", i, entry.ArchID)
		}
		if entry.Slot != colstore.Slot(i) {
			t.Errorf("id %d: expected Slot %d, got %d", i, i, entry.Slot)
		}
		if entry.ChunkPtr != ptr {
			t.Errorf("id %d: expected Ptr %v, got %v", i, ptr, entry.ChunkPtr)
		}
	}
}

// Seeding more entities than the index's initial capacity must grow it
// transparently (exercises Book.Seed -> Index.EnsureCap's grow branch).
func TestBook_SeedGrowsIndexCapacity(t *testing.T) {
	var b Book
	b.Init(2, 2)

	dst := make([]uid.UID64, 10)
	b.Seed(dst, arch.ID(1), fakePtr(0), colstore.Slot(0))

	for i, id := range dst {
		if _, ok := b.Get(id); !ok {
			t.Errorf("expected entry for seeded id %d after growth", i)
		}
	}
}

func TestBook_Move(t *testing.T) {
	var b Book
	b.Init(8, 8)
	dst := make([]uid.UID64, 1)
	b.Seed(dst, arch.ID(1), fakePtr(0), colstore.Slot(0))

	movePtr := fakePtr(3)
	b.Move(dst[0], arch.ID(2), movePtr, colstore.Slot(4))

	entry, ok := b.Get(dst[0])
	if !ok {
		t.Fatal("expected entry after Move")
	}
	if entry.ArchID != arch.ID(2) || entry.ChunkPtr != movePtr || entry.Slot != colstore.Slot(4) {
		t.Errorf("expected updated address, got %+v", entry)
	}
}

func TestBook_Delete(t *testing.T) {
	var b Book
	b.Init(8, 8)
	dst := make([]uid.UID64, 1)
	b.Seed(dst, arch.ID(1), fakePtr(0), colstore.Slot(0))
	deletedID := dst[0]

	b.Delete(deletedID)

	if _, ok := b.Get(deletedID); ok {
		t.Error("expected entry to be gone after Delete")
	}

	// The recycled index must be reusable, with an incremented generation.
	dst2 := make([]uid.UID64, 1)
	b.Seed(dst2, arch.ID(2), fakePtr(0), colstore.Slot(0))
	if dst2[0].Index() != deletedID.Index() {
		t.Errorf("expected recycled index %d to be reused, got %d", deletedID.Index(), dst2[0].Index())
	}
	if dst2[0].Generation() == deletedID.Generation() {
		t.Error("expected generation to increment after recycling")
	}
}

func TestBook_Reset(t *testing.T) {
	var b Book
	b.Init(8, 8)
	dst := make([]uid.UID64, 1)
	b.Seed(dst, arch.ID(1), fakePtr(0), colstore.Slot(0))

	b.Reset()

	if _, ok := b.Get(dst[0]); ok {
		t.Error("expected no entry after Reset")
	}
}

func TestBook_MoveUnchecked(t *testing.T) {
	var b Book
	b.Init(8, 8)

	dst := make([]uid.UID64, 3)
	b.Seed(dst, arch.ID(1), fakePtr(0), colstore.Slot(0))

	want := []struct {
		archID arch.ID
		ptr    unsafe.Pointer
		slot   colstore.Slot
	}{
		{arch.ID(2), fakePtr(1), 10},
		{arch.ID(3), fakePtr(2), 20},
		{arch.ID(4), fakePtr(3), 30},
	}
	for i, w := range want {
		b.MoveUnchecked(dst[i], w.archID, w.ptr, w.slot)
	}

	for i, w := range want {
		entry, ok := b.Get(dst[i])
		if !ok {
			t.Fatalf("entity %d missing after MoveUnchecked", i)
		}
		if entry.ArchID != w.archID {
			t.Errorf("entity %d: ArchID = %d; want %d", i, entry.ArchID, w.archID)
		}
		if entry.ChunkPtr != w.ptr || entry.Slot != w.slot {
			t.Errorf("entity %d: Ptr=%v Slot=%d; want Ptr=%v Slot=%d", i, entry.ChunkPtr, entry.Slot, w.ptr, w.slot)
		}
	}
}
