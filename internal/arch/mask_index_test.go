package arch

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func TestMaskIndex_UpsertAndGet(t *testing.T) {
	t.Run("stored value is retrievable", func(t *testing.T) {
		var m MaskIndex
		mask := comp.NewMask(1, 2, 3)
		m.Upsert(mask, ID(1))

		id, ok := m.Get(mask)
		if !ok {
			t.Fatal("expected to find stored mask")
		}
		if id != ID(1) {
			t.Errorf("expected id 1, got %d", id)
		}
	})

	t.Run("missing mask returns false", func(t *testing.T) {
		var m MaskIndex
		mask := comp.NewMask(10, 20)

		_, ok := m.Get(mask)
		if ok {
			t.Error("expected ok=false for absent mask")
		}
	})

	t.Run("distinct masks get distinct ids", func(t *testing.T) {
		var m MaskIndex
		m1 := comp.NewMask(1)
		m2 := comp.NewMask(2)

		m.Upsert(m1, ID(1))
		m.Upsert(m2, ID(2))

		id1, _ := m.Get(m1)
		id2, _ := m.Get(m2)

		if id1 != ID(1) || id2 != ID(2) {
			t.Errorf("expected ids 1 and 2, got %d and %d", id1, id2)
		}
	})
}

func TestMaskIndex_Idempotency(t *testing.T) {
	t.Run("Upsert with same mask updates id", func(t *testing.T) {
		var m MaskIndex
		mask := comp.NewMask(5, 10)

		m.Upsert(mask, ID(1))
		m.Upsert(mask, ID(42))

		id, ok := m.Get(mask)
		if !ok {
			t.Fatal("expected mask to be present")
		}
		if id != ID(42) {
			t.Errorf("expected updated id 42, got %d", id)
		}
	})
}

func TestMaskIndex_CollisionResolution(t *testing.T) {
	t.Run("masks that hash to the same slot are both retrievable", func(t *testing.T) {
		var m MaskIndex

		const n = 20
		masks := make([]comp.Mask, n)
		for i := range n {
			masks[i] = comp.NewMask(comp.ID(i))
			m.Upsert(masks[i], ID(i+1))
		}

		for i, mask := range masks {
			id, ok := m.Get(mask)
			if !ok {
				t.Errorf("mask %d not found after insertion", i)
				continue
			}
			if id != ID(i+1) {
				t.Errorf("mask %d: expected id %d, got %d", i, i+1, id)
			}
		}
	})
}

func TestMaskIndex_Reset(t *testing.T) {
	t.Run("Reset clears all entries", func(t *testing.T) {
		var m MaskIndex
		mask := comp.NewMask(1, 2, 3)
		m.Upsert(mask, ID(7))

		m.Reset()

		_, ok := m.Get(mask)
		if ok {
			t.Error("expected map to be empty after Reset")
		}
	})
}

func TestMaskIndex_UpsertNullPanics(t *testing.T) {
	t.Run("Upsert with NullID panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when storing NullID")
			}
		}()

		var m MaskIndex
		m.Upsert(comp.NewMask(1), NullID)
	})
}
