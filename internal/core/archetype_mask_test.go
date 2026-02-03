package core

import (
	"reflect"
	"testing"
)

// 1. Basic Operations (Set/Clear/IsSet)
func TestArchetypeMask_BasicOperations(t *testing.T) {
	// Case: Bit activation
	t.Run("Bit Activation", func(t *testing.T) {
		mask := ArchetypeMask{}
		id := ComponentID(10)
		mask = mask.Set(id)
		if !mask.IsSet(id) {
			t.Errorf("expected bit %d to be set", id)
		}
	})

	// Case: Bit deactivation
	t.Run("Bit Deactivation", func(t *testing.T) {
		id := ComponentID(20)
		mask := NewArchetypeMask(id)
		mask = mask.Clear(id)
		if mask.IsSet(id) {
			t.Errorf("expected bit %d to be cleared", id)
		}
	})

	// Case: Bit independence
	t.Run("Bit Independence", func(t *testing.T) {
		mask := ArchetypeMask{}.Set(10)
		if mask.IsSet(11) {
			t.Error("setting bit 10 should not set bit 11")
		}
	})

	// Case: Word boundaries
	// Updated for MaxComponents = 128 (bits 0-127)
	t.Run("Word Boundaries", func(t *testing.T) {
		boundaries := []ComponentID{0, 63, 64, 127}
		mask := ArchetypeMask{}
		for _, id := range boundaries {
			mask = mask.Set(id)
			if !mask.IsSet(id) {
				t.Errorf("failed to set bit at word boundary: %d", id)
			}
		}
	})

	// Case: Out of bounds
	// MaxComponents is 128, so ID 128 and 200 are out of bounds.
	t.Run("Out of Bounds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("code panicked on out of bounds bit: %v", r)
			}
		}()
		mask := ArchetypeMask{}
		oobID := ComponentID(150)
		mask = mask.Set(oobID)
		if mask.IsSet(oobID) {
			t.Error("IsSet should return false for out of bounds IDs")
		}
	})
}

// 2. Group Logic (Contains/Equals/IsEmpty)
func TestArchetypeMask_GroupLogic(t *testing.T) {
	t.Run("IsEmpty", func(t *testing.T) {
		mask := ArchetypeMask{}
		if !mask.IsEmpty() {
			t.Error("newly created mask should be empty")
		}
		mask = mask.Set(1).Clear(1)
		if !mask.IsEmpty() {
			t.Error("mask should be empty after setting and clearing the same bit")
		}
	})

	t.Run("Equals", func(t *testing.T) {
		m1 := NewArchetypeMask(10, 20)
		m2 := NewArchetypeMask(20, 10)
		m3 := NewArchetypeMask(10, 30)

		if !m1.Equals(m2) {
			t.Error("masks with same bits should be equal regardless of insertion order")
		}
		if m1.Equals(m3) {
			t.Error("different masks should not be equal")
		}
	})

	t.Run("Contains", func(t *testing.T) {
		main := NewArchetypeMask(1, 10, 100)
		sub := NewArchetypeMask(1, 10)
		empty := ArchetypeMask{}

		if !main.Contains(main) {
			t.Error("mask should contain itself")
		}
		if !main.Contains(empty) {
			t.Error("any mask should contain an empty mask")
		}
		if !main.Contains(sub) {
			t.Error("mask {1, 10, 100} should contain {1, 10}")
		}
		if sub.Contains(main) {
			t.Error("mask {1, 10} should not contain {1, 10, 100}")
		}
	})
}

// 4. Iteration (AllSet)
func TestArchetypeMask_Iteration(t *testing.T) {
	t.Run("Empty Mask", func(t *testing.T) {
		mask := ArchetypeMask{}
		calls := 0
		for range mask.AllSet() {
			calls++
		}
		if calls != 0 {
			t.Errorf("expected 0 iteration calls, got %d", calls)
		}
	})

	t.Run("Scattered Bits", func(t *testing.T) {
		input := []ComponentID{5, 70, 100} // Updated ID from 200 to 100
		mask := NewArchetypeMask(input...)
		var output []ComponentID

		for id := range mask.AllSet() {
			output = append(output, id)
		}

		if !reflect.DeepEqual(input, output) {
			t.Errorf("iteration mismatch: got %v, want %v", output, input)
		}
	})

	t.Run("Full Word Iteration", func(t *testing.T) {
		mask := ArchetypeMask{}
		for i := 0; i < 64; i++ {
			mask = mask.Set(ComponentID(i))
		}
		count := 0
		for range mask.AllSet() {
			count++
		}
		if count != 64 {
			t.Errorf("expected to iterate over 64 bits, got %d", count)
		}
	})
}

// 5. Immutability
func TestArchetypeMask_Immutability(t *testing.T) {
	t.Run("Immutability Check", func(t *testing.T) {
		original := ArchetypeMask{}
		_ = original.Set(10)

		if original.IsSet(10) {
			t.Error("ArchetypeMask should be immutable")
		}
	})
}
