package ecs

import (
	"reflect"
	"testing"
)

// 1. Basic Operations (Set/Clear/IsSet)
func TestArchetypeMask_BasicOperations(t *testing.T) {
	// Case: Bit activation
	// Checking if Set(id) makes IsSet(id) return true.
	t.Run("Bit Activation", func(t *testing.T) {
		mask := ArchetypeMask{}
		id := ComponentID(10)
		mask = mask.Set(id)
		if !mask.IsSet(id) {
			t.Errorf("expected bit %d to be set", id)
		}
	})

	// Case: Bit deactivation
	// Checking if Clear(id) on a set bit makes IsSet(id) return false.
	t.Run("Bit Deactivation", func(t *testing.T) {
		id := ComponentID(20)
		mask := NewArchetypeMask(id)
		mask = mask.Clear(id)
		if mask.IsSet(id) {
			t.Errorf("expected bit %d to be cleared", id)
		}
	})

	// Case: Bit independence
	// Ensuring that setting bit 10 does not affect the state of bit 11.
	t.Run("Bit Independence", func(t *testing.T) {
		mask := ArchetypeMask{}.Set(10)
		if mask.IsSet(11) {
			t.Error("setting bit 10 should not set bit 11")
		}
	})

	// Case: Word boundaries
	// Checking bits at the edges of uint64 (63, 64, 127, 128) to verify indexing.
	t.Run("Word Boundaries", func(t *testing.T) {
		boundaries := []ComponentID{63, 64, 127, 128}
		mask := ArchetypeMask{}
		for _, id := range boundaries {
			mask = mask.Set(id)
			if !mask.IsSet(id) {
				t.Errorf("failed to set bit at word boundary: %d", id)
			}
		}
	})

	// Case: Out of bounds
	// Attempting to set a bit larger than 256 ($4 \times 64$) to ensure no panic.
	t.Run("Out of Bounds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("code panicked on out of bounds bit: %v", r)
			}
		}()
		mask := ArchetypeMask{}
		mask = mask.Set(300)
		if mask.IsSet(300) {
			t.Error("IsSet should return false for out of bounds IDs")
		}
	})
}

// 2. Group Logic (Contains/Equals/IsEmpty)
func TestArchetypeMask_GroupLogic(t *testing.T) {
	// Case: IsEmpty
	// Checking the mask immediately after creation and after clearing bits.
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

	// Case: Equals
	// Comparing masks set with the same IDs in different order and different masks.
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

	// Case: Contains (Subsets)
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

// 3. Constructor (NewArchetypeMask)
func TestArchetypeMask_Constructor(t *testing.T) {
	// Case: Variadic arguments
	// Checking if passing multiple IDs to the constructor matches manual sequential Sets.
	t.Run("Variadic Arguments", func(t *testing.T) {
		ids := []ComponentID{5, 15, 25}
		m1 := NewArchetypeMask(ids...)

		m2 := ArchetypeMask{}
		for _, id := range ids {
			m2 = m2.Set(id)
		}

		if !m1.Equals(m2) {
			t.Error("constructor output does not match manual sequential Set calls")
		}
	})
}

// 4. Iteration (ForEachSet)
func TestArchetypeMask_Iteration(t *testing.T) {
	// Case: Empty mask
	// Ensuring the iteration function is never called for an empty mask.
	t.Run("Empty Mask", func(t *testing.T) {
		mask := ArchetypeMask{}
		calls := 0
		mask.ForEachSet(func(id ComponentID) {
			calls++
		})
		if calls != 0 {
			t.Errorf("expected 0 iteration calls, got %d", calls)
		}
	})

	// Case: Scattered bits
	// Verifying that ForEachSet visits exactly the specified IDs across different words.
	t.Run("Scattered Bits", func(t *testing.T) {
		input := []ComponentID{5, 70, 200}
		mask := NewArchetypeMask(input...)
		var output []ComponentID
		mask.ForEachSet(func(id ComponentID) {
			output = append(output, id)
		})

		if !reflect.DeepEqual(input, output) {
			t.Errorf("iteration mismatch: got %v, want %v", output, input)
		}
	})

	// Case: All bits in a word
	// Setting all 64 bits in a word to verify trailing zero logic and full iteration.
	t.Run("Full Word Iteration", func(t *testing.T) {
		mask := ArchetypeMask{}
		for i := 0; i < 64; i++ {
			mask = mask.Set(ComponentID(i))
		}
		count := 0
		mask.ForEachSet(func(id ComponentID) {
			count++
		})
		if count != 64 {
			t.Errorf("expected to iterate over 64 bits, got %d", count)
		}
	})
}

// 5. Immutability
func TestArchetypeMask_Immutability(t *testing.T) {
	// Case: Value Receiver Modification
	// Verifying that calling Set without assignment does not modify the original mask.
	t.Run("Immutability Check", func(t *testing.T) {
		original := ArchetypeMask{}
		// Call Set but ignore result
		_ = original.Set(10)

		if original.IsSet(10) {
			t.Error("ArchetypeMask should be immutable; calling Set should not modify the original variable")
		}
	})
}
