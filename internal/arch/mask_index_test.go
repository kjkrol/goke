package arch

import (
	"testing"

	"github.com/kjkrol/goke/internal/comp"
)

func TestMaskIndex_UpsertAndGet(t *testing.T) {
	t.Run("stored value is retrievable", func(t *testing.T) {
		var m MaskIndex
		mask := comp.Mask{}.Set(1).Set(2).Set(3)
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
		mask := comp.Mask{}.Set(10).Set(20)

		_, ok := m.Get(mask)
		if ok {
			t.Error("expected ok=false for absent mask")
		}
	})

	t.Run("distinct masks get distinct ids", func(t *testing.T) {
		var m MaskIndex
		m1 := comp.Mask{}.Set(1)
		m2 := comp.Mask{}.Set(2)

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
		mask := comp.Mask{}.Set(5).Set(10)

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

// findCollidingMasks brute-forces n distinct masks that all hash to the same
// MaskIndex bucket, using the real hashMask function — deterministic, unlike
// hoping a handful of arbitrary masks happen to collide (with 8192 buckets,
// a handful of random-ish masks usually won't).
func findCollidingMasks(t *testing.T, n int) []comp.Mask {
	t.Helper()
	buckets := make(map[uint64][]comp.Mask)
	for i := uint64(1); i < 1_000_000; i++ {
		mask := comp.Mask{i, 0}
		bucket := hashMask(mask) & HashMask
		buckets[bucket] = append(buckets[bucket], mask)
		if len(buckets[bucket]) >= n {
			return buckets[bucket][:n]
		}
	}
	t.Fatalf("failed to find %d colliding masks within search bound", n)
	return nil
}

func TestMaskIndex_CollisionResolution(t *testing.T) {
	t.Run("masks that hash to the same slot are all retrievable via linear probing", func(t *testing.T) {
		var m MaskIndex
		masks := findCollidingMasks(t, 3)

		for i, mask := range masks {
			m.Upsert(mask, ID(i+1))
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
		mask := comp.Mask{}.Set(1).Set(2).Set(3)
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
		m.Upsert(comp.Mask{}.Set(1), NullID)
	})
}
