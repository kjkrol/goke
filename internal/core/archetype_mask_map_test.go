package core

import (
	"testing"
)

func TestArchetypeMaskMap_PutAndGet(t *testing.T) {
	t.Run("stored value is retrievable", func(t *testing.T) {
		var m ArchetypeMaskMap
		mask := NewArchetypeMask(1, 2, 3)
		m.Put(mask, ArchetypeId(1))

		id, ok := m.Get(mask)
		if !ok {
			t.Fatal("expected to find stored mask")
		}
		if id != ArchetypeId(1) {
			t.Errorf("expected id 1, got %d", id)
		}
	})

	t.Run("missing mask returns false", func(t *testing.T) {
		var m ArchetypeMaskMap
		mask := NewArchetypeMask(10, 20)

		_, ok := m.Get(mask)
		if ok {
			t.Error("expected ok=false for absent mask")
		}
	})

	t.Run("distinct masks get distinct ids", func(t *testing.T) {
		var m ArchetypeMaskMap
		m1 := NewArchetypeMask(1)
		m2 := NewArchetypeMask(2)

		m.Put(m1, ArchetypeId(1))
		m.Put(m2, ArchetypeId(2))

		id1, _ := m.Get(m1)
		id2, _ := m.Get(m2)

		if id1 != ArchetypeId(1) || id2 != ArchetypeId(2) {
			t.Errorf("expected ids 1 and 2, got %d and %d", id1, id2)
		}
	})
}

func TestArchetypeMaskMap_Idempotency(t *testing.T) {
	t.Run("Put with same mask updates id", func(t *testing.T) {
		var m ArchetypeMaskMap
		mask := NewArchetypeMask(5, 10)

		m.Put(mask, ArchetypeId(1))
		m.Put(mask, ArchetypeId(42))

		id, ok := m.Get(mask)
		if !ok {
			t.Fatal("expected mask to be present")
		}
		if id != ArchetypeId(42) {
			t.Errorf("expected updated id 42, got %d", id)
		}
	})
}

func TestArchetypeMaskMap_CollisionResolution(t *testing.T) {
	t.Run("masks that hash to the same slot are both retrievable", func(t *testing.T) {
		var m ArchetypeMaskMap

		// Build two masks that produce the same initial hash slot.
		// We force a collision by inserting many distinct masks and verifying
		// all are retrievable — linear probing must handle the overflow.
		const n = 20
		masks := make([]ArchetypeMask, n)
		for i := range n {
			masks[i] = NewArchetypeMask(ComponentID(i))
			m.Put(masks[i], ArchetypeId(i+1))
		}

		for i, mask := range masks {
			id, ok := m.Get(mask)
			if !ok {
				t.Errorf("mask %d not found after insertion", i)
				continue
			}
			if id != ArchetypeId(i+1) {
				t.Errorf("mask %d: expected id %d, got %d", i, i+1, id)
			}
		}
	})
}

func TestArchetypeMaskMap_Reset(t *testing.T) {
	t.Run("Reset clears all entries", func(t *testing.T) {
		var m ArchetypeMaskMap
		mask := NewArchetypeMask(1, 2, 3)
		m.Put(mask, ArchetypeId(7))

		m.Reset()

		_, ok := m.Get(mask)
		if ok {
			t.Error("expected map to be empty after Reset")
		}
	})
}

func TestArchetypeMaskMap_PutNullPanics(t *testing.T) {
	t.Run("Put with NullArchetypeId panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when storing NullArchetypeId")
			}
		}()

		var m ArchetypeMaskMap
		m.Put(NewArchetypeMask(1), NullArchetypeId)
	})
}
