package ecs

import (
	"testing"
)

func TestEntity_Unpack(t *testing.T) {
	var wantIndex uint32 = 0x12345678
	var wantGen uint32 = 0xABCDEF01

	e := NewEntity(wantGen, wantIndex)
	index, gen := e.Unpack()

	if index != wantIndex {
		t.Errorf("unpack index: got %d, want %d", index, wantIndex)
	}
	if gen != wantGen {
		t.Errorf("unpack generation: got %d, want %d", gen, wantGen)
	}
}

// Scenario: New generation with the same index (slot recycling)
func TestEntity_GenerationIncrement(t *testing.T) {
	index := uint32(42)
	oldGen := uint32(1)
	newGen := uint32(2)

	oldEntity := NewEntity(oldGen, index)
	newEntity := NewEntity(newGen, index)

	if oldEntity.Index() != newEntity.Index() {
		t.Errorf("entities should share the same index: %d == %d", oldEntity.Index(), newEntity.Index())
	}

	if oldEntity == newEntity {
		t.Error("entities with different generations must not be equal")
	}

	if newEntity.Generation() <= oldEntity.Generation() {
		t.Errorf("new entity should have higher generation: %d > %d", newEntity.Generation(), oldEntity.Generation())
	}
}

func TestEntity_VirtualStatus(t *testing.T) {
	t.Run("Standard entity should not be virtual", func(t *testing.T) {
		e := NewEntity(10, 500)
		if e.IsVirtual() {
			t.Error("expected IsVirtual() to be false for standard entity")
		}
	})

	t.Run("Virtual entity should be correctly identified", func(t *testing.T) {
		id := uint32(99)
		ve := NewVirtualEntity(id)

		if !ve.IsVirtual() {
			t.Error("expected IsVirtual() to be true for virtual entity")
		}

		// Verify that the virtual bit is actually set in the index
		if (ve.Index() & VirtualEntityMask) == 0 {
			t.Error("virtual bit missing in entity index")
		}
	})
}

func TestEntity_BitMasks(t *testing.T) {
	// Check if VirtualEntityMask is within uint32 range and uses the correct bit
	if VirtualEntityMask != 0x80000000 {
		t.Errorf("unexpected VirtualEntityMask value: %X", VirtualEntityMask)
	}

	// Ensure NewVirtualEntity doesn't corrupt the generation part (upper 32 bits)
	ve := NewVirtualEntity(1)
	if ve.Generation() != 0 {
		t.Errorf("virtual entity should have generation 0 by default, got %d", ve.Generation())
	}
}

func TestEntity_MaxValues(t *testing.T) {
	// Casting IndexMask to uint32 to match NewEntity signature
	maxIndex := uint32(IndexMask)
	maxGen := uint32(0xFFFFFFFF)

	e := NewEntity(maxGen, maxIndex)
	index, gen := e.Unpack()

	if index != maxIndex {
		t.Errorf("max index mismatch: got %d, want %d", index, maxIndex)
	}
	if gen != maxGen {
		t.Errorf("max generation mismatch: got %d, want %d", gen, maxGen)
	}
}

func TestEntity_IndexOverflow(t *testing.T) {
	// Verifying that NewEntity handles the index correctly
	// even if bits beyond IndexMask are set in the input uint32
	gen := uint32(5)
	indexWithExtraBits := uint32(0x12345678)

	e := NewEntity(gen, indexWithExtraBits)

	if e.Index() != indexWithExtraBits {
		t.Errorf("index mismatch: got %x, want %x", e.Index(), indexWithExtraBits)
	}

	if e.Generation() != gen {
		t.Errorf("generation corrupted by index: got %d, want %d", e.Generation(), gen)
	}
}
