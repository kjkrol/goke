package ecs

import (
	"reflect"
	"testing"
	"unsafe"
)

// Internal helper to inject columns into archetype for testing
func addTestColumn[T any](a *Archetype, id ComponentID) {
	t := reflect.TypeFor[T]()
	a.columns[id] = &column{
		dataType: t,
		itemSize: t.Size(),
	}
	a.columns[id].growTo(a.cap)
}

// Helper to verify if memory at pointer is zeroed
func isMemoryZeroed(ptr unsafe.Pointer, size uintptr) bool {
	bytes := unsafe.Slice((*byte)(ptr), size)
	for _, b := range bytes {
		if b != 0 {
			return false
		}
	}
	return true
}

// Scenario 1: Capacity & Growth
func TestArchetype_CapacityAndGrowth(t *testing.T) {
	mask := NewArchetypeMask(1)
	arch := NewArchetype(mask)
	addTestColumn[position](arch, 1)

	// Case: Initial capacity
	if arch.cap != initCapacity {
		t.Errorf("expected initial capacity %d, got %d", initCapacity, arch.cap)
	}

	// Case: Auto-growth - exceed initial capacity to trigger doubling
	for i := 0; i < initCapacity+1; i++ {
		arch.registerEntity(Entity(i))
	}

	expectedCap := initCapacity * 2
	if arch.cap != expectedCap {
		t.Errorf("expected capacity to grow to %d, got %d", expectedCap, arch.cap)
	}
	if arch.columns[1].cap != expectedCap {
		t.Errorf("column capacity should match archetype capacity: expected %d, got %d", expectedCap, arch.columns[1].cap)
	}
}

// Scenario 2: AddEntity & Registration
func TestArchetype_AddEntityAndRegistration(t *testing.T) {
	mask := NewArchetypeMask(1, 2)
	arch := NewArchetype(mask)
	compID1, compID2 := ComponentID(1), ComponentID(2)
	addTestColumn[position](arch, compID1)
	addTestColumn[velocity](arch, compID2)

	// Case: Sequential adding
	e0 := Entity(10)
	e1 := Entity(11)
	p0 := position{x: 1, y: 1}
	p1 := position{x: 2, y: 2}

	link0 := arch.AddEntity(e0, compID1, unsafe.Pointer(&p0))
	link1 := arch.AddEntity(e1, compID1, unsafe.Pointer(&p1))

	if link0.row != 0 || link1.row != 1 {
		t.Errorf("expected sequential rows 0 and 1, got %d and %d", link0.row, link1.row)
	}

	// Case: Column data routing
	posData := position{x: 100, y: 200}
	e2 := Entity(12)
	link2 := arch.AddEntity(e2, compID1, unsafe.Pointer(&posData))

	gotP := *(*position)(arch.columns[compID1].GetElement(link2.row))
	if gotP != posData {
		t.Errorf("routing failed: data not found in target column")
	}

	gotV := *(*velocity)(arch.columns[compID2].GetElement(link2.row))
	if gotV.vx != 0 || gotV.vy != 0 {
		t.Error("routing failed: non-target column was not zeroed")
	}

	// Case: Link integrity
	if link2.arch != arch {
		t.Error("EntityArchLink should point to the correct archetype")
	}
}

// Scenario 3: Swap-and-Pop Logic
func TestArchetype_SwapRemoveLogic(t *testing.T) {
	mask := NewArchetypeMask(1)
	arch := NewArchetype(mask)
	compID := ComponentID(1)
	addTestColumn[int32](arch, compID)

	// Populate [E0:0, E1:10, E2:20]
	for i := 0; i < 3; i++ {
		val := int32(i * 10)
		arch.AddEntity(Entity(i), compID, unsafe.Pointer(&val))
	}

	// Case: Remove last entity
	_, swappedLast := arch.SwapRemoveEntity(2)
	if swappedLast {
		t.Error("removing the last entity should not trigger a swap")
	}

	// Case: Remove middle entity
	// State is now [E0:0, E1:10]. Let's add E2 back to have [E0, E1, E2]
	val2 := int32(20)
	arch.AddEntity(Entity(2), compID, unsafe.Pointer(&val2))

	movedEntity, swappedMid := arch.SwapRemoveEntity(0) // Remove E0
	if !swappedMid || movedEntity != Entity(2) {
		t.Errorf("expected E2 to move to row 0, got entity %d", movedEntity)
	}
	if arch.entities[0] != Entity(2) {
		t.Errorf("entities slice mismatch: row 0 should be E2, got %d", arch.entities[0])
	}

	// Case: Memory sanitization
	if arch.entities[2] != 0 {
		t.Error("old last slot in entities slice should be cleared")
	}
	if !isMemoryZeroed(arch.columns[compID].GetElement(2), arch.columns[compID].itemSize) {
		t.Error("old last slot in column should be zeroed")
	}
}

// Scenario 4: Multi-Column Integrity
func TestArchetype_MultiColumnIntegrity(t *testing.T) {
	mask := NewArchetypeMask(1, 2)
	arch := NewArchetype(mask)
	posID, velID := ComponentID(1), ComponentID(2)
	addTestColumn[position](arch, posID)
	addTestColumn[velocity](arch, velID)

	// Case: Row integrity after swap
	// Fill E0 and E1
	p0, v0 := position{1, 1}, velocity{10, 10}
	p1, v1 := position{2, 2}, velocity{20, 20}
	arch.AddEntity(Entity(0), posID, unsafe.Pointer(&p0))
	arch.columns[velID].setData(0, unsafe.Pointer(&v0))
	arch.AddEntity(Entity(1), posID, unsafe.Pointer(&p1))
	arch.columns[velID].setData(1, unsafe.Pointer(&v1))

	arch.SwapRemoveEntity(0) // E1 moves to row 0

	// Case: Data isolation (E1 components must remain paired correctly)
	resP := *(*position)(arch.columns[posID].GetElement(0))
	resV := *(*velocity)(arch.columns[velID].GetElement(0))

	if resP != p1 || resV != v1 {
		t.Errorf("multi-column integrity lost: row 0 has mismatched components %+v, %+v", resP, resV)
	}
}

// Scenario 5: Edge Cases
func TestArchetype_EdgeCases(t *testing.T) {
	// Case: Columnless Archetype Operations
	t.Run("Columnless operations", func(t *testing.T) {
		arch := NewArchetype(ArchetypeMask{})
		arch.registerEntity(Entity(100))
		if arch.len != 1 {
			t.Fatal("failed to register entity in columnless archetype")
		}
		arch.SwapRemoveEntity(0)
		if arch.len != 0 {
			t.Error("failed to remove entity from columnless archetype")
		}
	})

	// Case: Remove from single-entity archetype
	t.Run("Single-entity remove", func(t *testing.T) {
		arch := NewArchetype(NewArchetypeMask(1))
		addTestColumn[int32](arch, 1)
		arch.registerEntity(Entity(5))

		moved, swapped := arch.SwapRemoveEntity(0)
		if swapped || moved != 0 || arch.len != 0 {
			t.Error("invalid state after removing the only entity")
		}
	})
}
