package comp

import (
	"reflect"
	"testing"
)

func TestMask_BasicOperations(t *testing.T) {
	t.Run("Bit Activation", func(t *testing.T) {
		mask := Mask{}
		id := ID(10)
		mask = mask.Set(id)
		if !mask.IsSet(id) {
			t.Errorf("expected bit %d to be set", id)
		}
	})

	t.Run("Bit Deactivation", func(t *testing.T) {
		id := ID(20)
		mask := Mask{}.Set(id)
		mask = mask.Clear(id)
		if mask.IsSet(id) {
			t.Errorf("expected bit %d to be cleared", id)
		}
	})

	t.Run("Bit Independence", func(t *testing.T) {
		mask := Mask{}.Set(10)
		if mask.IsSet(11) {
			t.Error("setting bit 10 should not set bit 11")
		}
	})

	t.Run("Word Boundaries", func(t *testing.T) {
		boundaries := []ID{0, 63, 64, 127}
		mask := Mask{}
		for _, id := range boundaries {
			mask = mask.Set(id)
			if !mask.IsSet(id) {
				t.Errorf("failed to set bit at word boundary: %d", id)
			}
		}
	})

	t.Run("Out of Bounds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("code panicked on out of bounds bit: %v", r)
			}
		}()
		mask := Mask{}
		oobID := ID(150)
		mask = mask.Set(oobID)
		if mask.IsSet(oobID) {
			t.Error("IsSet should return false for out of bounds IDs")
		}
	})
}

func TestMask_GroupLogic(t *testing.T) {
	t.Run("IsEmpty", func(t *testing.T) {
		mask := Mask{}
		if !mask.IsEmpty() {
			t.Error("newly created mask should be empty")
		}
		mask = mask.Set(1).Clear(1)
		if !mask.IsEmpty() {
			t.Error("mask should be empty after setting and clearing the same bit")
		}
	})

	t.Run("Equals", func(t *testing.T) {
		m1 := Mask{}.Set(10).Set(20)
		m2 := Mask{}.Set(20).Set(10)
		m3 := Mask{}.Set(10).Set(30)

		if !m1.Equals(m2) {
			t.Error("masks with same bits should be equal regardless of insertion order")
		}
		if m1.Equals(m3) {
			t.Error("different masks should not be equal")
		}
	})

	t.Run("Contains", func(t *testing.T) {
		main := Mask{}.Set(1).Set(10).Set(100)
		sub := Mask{}.Set(1).Set(10)
		empty := Mask{}

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

func TestMask_Iteration(t *testing.T) {
	t.Run("Empty Mask", func(t *testing.T) {
		mask := Mask{}
		calls := 0
		for range mask.AllSet() {
			calls++
		}
		if calls != 0 {
			t.Errorf("expected 0 iteration calls, got %d", calls)
		}
	})

	t.Run("Scattered Bits", func(t *testing.T) {
		input := []ID{5, 70, 100}
		mask := Mask{}
		for _, id := range input {
			mask = mask.Set(id)
		}
		var output []ID

		for id := range mask.AllSet() {
			output = append(output, id)
		}

		if !reflect.DeepEqual(input, output) {
			t.Errorf("iteration mismatch: got %v, want %v", output, input)
		}
	})

	t.Run("Full Word Iteration", func(t *testing.T) {
		mask := Mask{}
		for i := 0; i < 64; i++ {
			mask = mask.Set(ID(i))
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

func TestMask_Immutability(t *testing.T) {
	t.Run("Immutability Check", func(t *testing.T) {
		original := Mask{}
		_ = original.Set(10)

		if original.IsSet(10) {
			t.Error("Mask should be immutable")
		}
	})
}
