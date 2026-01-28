package core

import (
	"reflect"
	"testing"
	"unsafe"
)

// Internal helper to inject columns into archetype for testing
func addTestColumn[T any](a *Archetype, id ComponentID) {
	t := reflect.TypeFor[T]()
	a.Columns[id] = &Column{
		dataType: t,
		ItemSize: t.Size(),
	}
	a.Columns[id].growTo(a.cap)
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
	initCapacity := 128
	mask := NewArchetypeMask(1)
	arch := NewArchetype(mask, initCapacity)
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
	if arch.Columns[1].cap != expectedCap {
		t.Errorf("column capacity should match archetype capacity: expected %d, got %d", expectedCap, arch.Columns[1].cap)
	}
}

// Scenario 2: AddEntity & Registration
func TestArchetype_AddEntityAndRegistration(t *testing.T) {
	mask := NewArchetypeMask(1, 2)
	initCapacity := 128
	arch := NewArchetype(mask, initCapacity)
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

	if link0.Row != 0 || link1.Row != 1 {
		t.Errorf("expected sequential rows 0 and 1, got %d and %d", link0.Row, link1.Row)
	}

	// Case: Column data routing
	posData := position{x: 100, y: 200}
	e2 := Entity(12)
	link2 := arch.AddEntity(e2, compID1, unsafe.Pointer(&posData))

	gotP := *(*position)(arch.Columns[compID1].GetElement(link2.Row))
	if gotP != posData {
		t.Errorf("routing failed: data not found in target column")
	}

	gotV := *(*velocity)(arch.Columns[compID2].GetElement(link2.Row))
	if gotV.vx != 0 || gotV.vy != 0 {
		t.Error("routing failed: non-target column was not zeroed")
	}

	// Case: Link integrity
	if link2.Arch != arch {
		t.Error("EntityArchLink should point to the correct archetype")
	}
}

// Scenario 3: Swap-and-Pop Logic
func TestArchetype_SwapRemoveWithLenSync(t *testing.T) {
	mask := NewArchetypeMask(1)
	arch := NewArchetype(mask, 10)
	addTestColumn[int32](arch, 1)

	// Przygotowujemy dane, żeby nie słać nil
	val1, val2 := int32(10), int32(20)

	// 1. Dodajemy encje
	arch.AddEntity(Entity(1), 1, unsafe.Pointer(&val1))
	arch.AddEntity(Entity(2), 1, unsafe.Pointer(&val2))

	// Sprawdzamy stan początkowy przed usunięciem
	if arch.len != 2 {
		t.Fatalf("Setup failed: Arch len should be 2, got %d", arch.len)
	}
	if arch.Columns[1].len != 2 {
		t.Fatalf("Setup failed: Column len should be 2, got %d", arch.Columns[1].len)
	}

	// 2. Wykonujemy Swap-Remove
	arch.SwapRemoveEntity(0) // Usuwamy pierwszą encję

	// 3. Weryfikacja zmiany (to czego brakowało)
	if arch.len != 1 {
		t.Errorf("Archetype len should decrease to 1, got %d", arch.len)
	}

	if arch.Columns[1].len != 1 {
		t.Errorf("Column len should decrease to 1, got %d", arch.Columns[1].len)
	}

	// Dodatkowe sprawdzenie spójności między strukturami
	if arch.Columns[1].len != arch.len {
		t.Errorf("Critical Sync Error: Archetype len (%d) != Column len (%d)",
			arch.len, arch.Columns[1].len)
	}
}

// Scenario 4: Multi-Column Integrity
func TestArchetype_MultiColumnIntegrity(t *testing.T) {
	mask := NewArchetypeMask(1, 2)
	initCapacity := 128
	arch := NewArchetype(mask, initCapacity)
	posID, velID := ComponentID(1), ComponentID(2)
	addTestColumn[position](arch, posID)
	addTestColumn[velocity](arch, velID)

	// Case: Row integrity after swap
	// Fill E0 and E1
	p0, v0 := position{1, 1}, velocity{10, 10}
	p1, v1 := position{2, 2}, velocity{20, 20}
	arch.AddEntity(Entity(0), posID, unsafe.Pointer(&p0))
	arch.Columns[velID].setData(0, unsafe.Pointer(&v0))
	arch.AddEntity(Entity(1), posID, unsafe.Pointer(&p1))
	arch.Columns[velID].setData(1, unsafe.Pointer(&v1))

	arch.SwapRemoveEntity(0) // E1 moves to row 0

	// Case: Data isolation (E1 components must remain paired correctly)
	resP := *(*position)(arch.Columns[posID].GetElement(0))
	resV := *(*velocity)(arch.Columns[velID].GetElement(0))

	if resP != p1 || resV != v1 {
		t.Errorf("multi-column integrity lost: row 0 has mismatched components %+v, %+v", resP, resV)
	}
}

// Scenario 5: Edge Cases
func TestArchetype_EdgeCases(t *testing.T) {
	// Case: Columnless Archetype Operations
	t.Run("Columnless operations", func(t *testing.T) {
		initCapacity := 128
		arch := NewArchetype(ArchetypeMask{}, initCapacity)
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
		initCapacity := 128
		arch := NewArchetype(NewArchetypeMask(1), initCapacity)
		addTestColumn[int32](arch, 1)
		arch.registerEntity(Entity(5))

		moved, swapped := arch.SwapRemoveEntity(0)
		if swapped || moved != 0 || arch.len != 0 {
			t.Error("invalid state after removing the only entity")
		}
	})
}

// Scenario: Memory Alignment
func TestArchetype_Alignment(t *testing.T) {
	// Case: Testing with a struct that has odd size
	type OddStruct struct {
		A int8 // 1 byte
	}

	mask := NewArchetypeMask(1)
	initCapacity := 128
	arch := NewArchetype(mask, initCapacity)
	id := ComponentID(1)

	info := reflect.TypeFor[OddStruct]()
	arch.Columns[id] = &Column{
		dataType: info,
		ItemSize: info.Size(),
	}
	arch.Columns[id].growTo(arch.cap)

	val := OddStruct{A: 42}
	arch.AddEntity(Entity(1), id, unsafe.Pointer(&val))

	// Check if we can retrieve it without corruption
	res := *(*OddStruct)(arch.Columns[id].GetElement(0))
	if res.A != 42 {
		t.Errorf("alignment or size issue: expected 42, got %d", res.A)
	}
}

func TestArchetype_DataIntegrityAfterGrowth(t *testing.T) {
	// 1. Setup z małym cap, żeby szybko wymusić grow
	initCap := 2
	mask := NewArchetypeMask(1)
	arch := NewArchetype(mask, initCap)
	addTestColumn[int32](arch, 1)

	// 2. Dodajemy encje do pełna (0 i 1)
	val0, val1 := int32(100), int32(200)
	arch.AddEntity(Entity(0), 1, unsafe.Pointer(&val0))
	arch.AddEntity(Entity(1), 1, unsafe.Pointer(&val1))

	// 3. To wywoła growTo(4).
	// Jeśli col.len nie było inkrementowane, growTo NIE SKOPIUJE val0 i val1!
	val2 := int32(300)
	arch.AddEntity(Entity(2), 1, unsafe.Pointer(&val2))

	// 4. WERYFIKACJA: Czy dane z pierwszego bloku pamięci nadal tam są?
	got0 := *(*int32)(arch.Columns[1].GetElement(0))
	got1 := *(*int32)(arch.Columns[1].GetElement(1))
	got2 := *(*int32)(arch.Columns[1].GetElement(2))

	if got0 != 100 {
		t.Errorf("DATA LOSS: Entity 0 data corrupted after growth. Expected 100, got %d", got0)
	}
	if got1 != 200 {
		t.Errorf("DATA LOSS: Entity 1 data corrupted after growth. Expected 200, got %d", got1)
	}
	if got2 != 300 {
		t.Errorf("Entity 2 (new) data mismatch. Got %d", got2)
	}
}

func TestArchetype_RegisterAndSetData(t *testing.T) {
	arch := NewArchetype(NewArchetypeMask(1), 2) // Cap = 2
	addTestColumn[int32](arch, 1)

	// Wypełniamy
	arch.registerEntity(Entity(0))
	arch.registerEntity(Entity(1))

	// Ta rejestracja wywoła growTo
	newRow := arch.registerEntity(Entity(2))

	// Symulujemy setData z moveEntity
	val := int32(777)
	arch.Columns[1].setData(newRow, unsafe.Pointer(&val))

	got := *(*int32)(arch.Columns[1].GetElement(newRow))
	if got != 777 {
		t.Errorf("setData failed after grow: expected 777, got %d", got)
	}
}

func TestArchetype_RemoveLastEntityLenSync(t *testing.T) {
	arch := NewArchetype(NewArchetypeMask(1), 5)
	addTestColumn[int32](arch, 1)

	arch.registerEntity(Entity(1)) // len = 1

	// Usuwamy jedyną (ostatnią) encję
	arch.SwapRemoveEntity(0)

	if arch.len != 0 || arch.Columns[1].len != 0 {
		t.Errorf("Failed to sync len when removing last entity: arch=%d, col=%d",
			arch.len, arch.Columns[1].len)
	}
}
