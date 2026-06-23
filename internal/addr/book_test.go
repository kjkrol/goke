package addr

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/colstore"
)

func TestBook_SeedAndGet(t *testing.T) {
	var b Book
	b.Init(8, 8)

	dst := make([]uid.UID64, 3)
	b.Seed(dst, arch.ID(5), colstore.Pos{Idx: 0, Slot: 0})

	for i, id := range dst {
		entry, ok := b.Get(id)
		if !ok {
			t.Fatalf("expected entry for seeded id %d", i)
		}
		if entry.ArchId != arch.ID(5) {
			t.Errorf("id %d: expected ArchId 5, got %d", i, entry.ArchId)
		}
		if entry.Pos.Slot != colstore.Slot(i) {
			t.Errorf("id %d: expected Slot %d, got %d", i, i, entry.Pos.Slot)
		}
	}
}

// Seeding more entities than the index's initial capacity must grow it
// transparently (exercises Book.Seed -> Index.EnsureCap's grow branch).
func TestBook_SeedGrowsIndexCapacity(t *testing.T) {
	var b Book
	b.Init(2, 2)

	dst := make([]uid.UID64, 10)
	b.Seed(dst, arch.ID(1), colstore.Pos{})

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
	b.Seed(dst, arch.ID(1), colstore.Pos{})

	b.Move(dst[0], arch.ID(2), colstore.Pos{Idx: 3, Slot: 4})

	entry, ok := b.Get(dst[0])
	if !ok {
		t.Fatal("expected entry after Move")
	}
	if entry.ArchId != arch.ID(2) || entry.Pos.Idx != colstore.Idx(3) || entry.Pos.Slot != colstore.Slot(4) {
		t.Errorf("expected updated address, got %+v", entry)
	}
}

func TestBook_Delete(t *testing.T) {
	var b Book
	b.Init(8, 8)
	dst := make([]uid.UID64, 1)
	b.Seed(dst, arch.ID(1), colstore.Pos{})
	deletedID := dst[0]

	b.Delete(deletedID)

	if _, ok := b.Get(deletedID); ok {
		t.Error("expected entry to be gone after Delete")
	}

	// The recycled index must be reusable, with an incremented generation.
	dst2 := make([]uid.UID64, 1)
	b.Seed(dst2, arch.ID(2), colstore.Pos{})
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
	b.Seed(dst, arch.ID(1), colstore.Pos{})

	b.Reset()

	if _, ok := b.Get(dst[0]); ok {
		t.Error("expected no entry after Reset")
	}
}
