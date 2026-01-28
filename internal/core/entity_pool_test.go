package core

import (
	"testing"
)

// 1. Sequential Allocation
func TestEntityGenerationalPool_SequentialAllocation(t *testing.T) {
	pool := NewEntityGenerator(10)

	// Case: New entities receive incremental indices and generation 0
	for i := 0; i < 5; i++ {
		e := pool.Next()
		index, gen := e.Unpack()

		if index != uint32(i) {
			t.Errorf("expected index %d, got %d", i, index)
		}
		if gen != 0 {
			t.Errorf("expected initial generation 0, got %d", gen)
		}
	}

	if pool.lastIndex != 5 {
		t.Errorf("expected lastIndex to be 5, got %d", pool.lastIndex)
	}
}

// 2. Index Recycling
func TestEntityGenerationalPool_Recycling(t *testing.T) {
	pool := NewEntityGenerator(10)

	// Case: Releasing and reusing an index
	e0 := pool.Next() // idx: 0, gen: 0
	pool.Release(e0)

	e0_new := pool.Next() // should be idx: 0, gen: 1
	index, gen := e0_new.Unpack()

	if index != 0 {
		t.Errorf("expected recycled index 0, got %d", index)
	}
	if gen != 1 {
		t.Errorf("expected incremented generation 1, got %d", gen)
	}

	// Case: LIFO behavior for free indices
	e1 := pool.Next() // idx: 1, gen: 0
	e2 := pool.Next() // idx: 2, gen: 0

	pool.Release(e1)
	pool.Release(e2)

	firstRecycled := pool.Next() // Should be e2 (last released)
	if firstRecycled.Index() != 2 {
		t.Errorf("expected LIFO recycling, got index %d, want 2", firstRecycled.Index())
	}
}

// 3. Validation Logic
func TestEntityGenerationalPool_Validation(t *testing.T) {
	pool := NewEntityGenerator(10)

	// Case: Active entity is valid
	e := pool.Next()
	if !pool.IsValid(e) {
		t.Error("expected active entity to be valid")
	}

	// Case: Released entity is invalid
	pool.Release(e)
	if pool.IsValid(e) {
		t.Error("expected released entity to be invalid (old generation)")
	}

	// Case: Out of bounds index
	futureEntity := NewEntity(0, 999)
	if pool.IsValid(futureEntity) {
		t.Error("expected out-of-bounds entity to be invalid")
	}

	// Case: Validating after recycling
	e_new := pool.Next() // Reuse index of 'e'
	if !pool.IsValid(e_new) {
		t.Error("expected newly recycled entity to be valid")
	}
	if pool.IsValid(e) {
		t.Error("old handle should remain invalid after recycling")
	}
}

// 4. Stress & Multiple Reuse
func TestEntityGenerationalPool_MultipleReuse(t *testing.T) {
	pool := NewEntityGenerator(1)
	idx := uint32(0)

	// Case: Cycling the same index multiple times
	for expectedGen := uint32(0); expectedGen < 10; expectedGen++ {
		e := pool.Next()
		if e.Index() != idx || e.Generation() != expectedGen {
			t.Fatalf("iteration %d: expected idx %d gen %d, got idx %d gen %d",
				expectedGen, idx, expectedGen, e.Index(), e.Generation())
		}
		pool.Release(e)
	}
}

// 5. Initial State
func TestEntityGenerationalPool_InitialState(t *testing.T) {
	capacity := 128
	pool := NewEntityGenerator(capacity)

	// Case: Internal slices should respect initial capacity
	if cap(pool.generations) < capacity {
		t.Errorf("expected generations capacity at least %d, got %d", capacity, cap(pool.generations))
	}
	if len(pool.freeIndices) != 0 {
		t.Error("expected freeIndices to be empty initially")
	}
}
