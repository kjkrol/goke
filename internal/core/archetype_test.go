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

// Scenario 1: Capacity & Growth
func TestArchetype_CapacityAndGrowth(t *testing.T) {
	initCapacity := 128
	mask := NewArchetypeMask(1)
	archReg := setupTestRegistry()
	archId := archReg.InitArchetype(mask, initCapacity)
	arch := &archReg.Archetypes[archId]
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

// Scenario 2: Manual Registration & Data Setting
func TestArchetype_ManualRegistration(t *testing.T) {
	mask := NewArchetypeMask(1, 2)
	archReg := setupTestRegistry()
	archId := archReg.InitArchetype(mask, 128)
	arch := &archReg.Archetypes[archId]
	compID1, compID2 := ComponentID(1), ComponentID(2)
	addTestColumn[position](arch, compID1)
	addTestColumn[velocity](arch, compID2)

	e0 := Entity(10)
	p0 := position{x: 1, y: 1}

	// 1. Register
	row := arch.registerEntity(e0)

	// 2. Set Data manually
	arch.Columns[compID1].setData(row, unsafe.Pointer(&p0))
	arch.Columns[compID2].zeroData(row) // Manually zero if needed

	// 3. Verify
	gotP := *(*position)(arch.Columns[compID1].GetElement(row))
	if gotP != p0 {
		t.Errorf("expected %+v, got %+v", p0, gotP)
	}
}

// Scenario 3: Swap-and-Pop Logic
func TestArchetype_SwapRemoveWithLenSync(t *testing.T) {
	mask := NewArchetypeMask(1)
	archReg := setupTestRegistry()
	archId := archReg.InitArchetype(mask, 10)
	arch := &archReg.Archetypes[archId]
	addTestColumn[int32](arch, 1)

	val1, val2 := int32(10), int32(20)

	// 1. Add entities using the actual API: register + setData
	row0 := arch.registerEntity(Entity(1))
	arch.Columns[1].setData(row0, unsafe.Pointer(&val1))

	row1 := arch.registerEntity(Entity(2))
	arch.Columns[1].setData(row1, unsafe.Pointer(&val2))

	// Verify initial state
	if arch.len != 2 {
		t.Fatalf("Setup failed: Arch len should be 2, got %d", arch.len)
	}
	if arch.Columns[1].len != 2 {
		t.Fatalf("Setup failed: Column len should be 2, got %d", arch.Columns[1].len)
	}

	// 2. Perform Swap-Remove
	// This will move Entity(2) from row 1 to row 0
	arch.SwapRemoveEntity(0)

	// 3. Verify sync and length reduction
	if arch.len != 1 {
		t.Errorf("Archetype len should decrease to 1, got %d", arch.len)
	}

	if arch.Columns[1].len != 1 {
		t.Errorf("Column len should decrease to 1, got %d", arch.Columns[1].len)
	}

	// Critical check: Did the data actually swap?
	gotVal := *(*int32)(arch.Columns[1].GetElement(0))
	if gotVal != val2 {
		t.Errorf("Swap failed: expected val2 (20) at row 0, got %d", gotVal)
	}

	if arch.Columns[1].len != arch.len {
		t.Errorf("Critical Sync Error: Archetype len (%d) != Column len (%d)",
			arch.len, arch.Columns[1].len)
	}
}

// Scenario 4: Multi-Column Integrity
func TestArchetype_MultiColumnIntegrity(t *testing.T) {
	archReg := setupTestRegistry()
	positionInfo := EnsureComponentRegistered[position](archReg.componentsRegistry)
	velocityInfo := EnsureComponentRegistered[velocity](archReg.componentsRegistry)
	mask := NewArchetypeMask(positionInfo.ID, velocityInfo.ID)
	initCapacity := 128

	archId := archReg.InitArchetype(mask, initCapacity)
	arch := &archReg.Archetypes[archId]

	addTestColumn[position](arch, positionInfo.ID)
	addTestColumn[velocity](arch, velocityInfo.ID)

	// Data for two separate entities
	p0, v0 := position{1, 1}, velocity{10, 10}
	p1, v1 := position{2, 2}, velocity{20, 20}

	// 1. Setup Entity 0 at Row 0
	row0 := arch.registerEntity(Entity(0))
	arch.Columns[positionInfo.ID].setData(row0, unsafe.Pointer(&p0))
	arch.Columns[velocityInfo.ID].setData(row0, unsafe.Pointer(&v0))

	// 2. Setup Entity 1 at Row 1
	row1 := arch.registerEntity(Entity(1))
	arch.Columns[positionInfo.ID].setData(row1, unsafe.Pointer(&p1))
	arch.Columns[velocityInfo.ID].setData(row1, unsafe.Pointer(&v1))

	// 3. Swap-and-Pop: Remove Entity 0
	// This moves Entity 1 from Row 1 to Row 0 across ALL columns
	arch.SwapRemoveEntity(row0)

	// 4. Case: Data isolation (E1 components must remain paired correctly at Row 0)
	resP := *(*position)(arch.Columns[positionInfo.ID].GetElement(0))
	resV := *(*velocity)(arch.Columns[velocityInfo.ID].GetElement(0))

	if resP != p1 {
		t.Errorf("Position integrity lost: expected %+v, got %+v", p1, resP)
	}
	if resV != v1 {
		t.Errorf("Velocity integrity lost: expected %+v, got %+v", v1, resV)
	}
}

// Scenario 5: Edge Cases
func TestArchetype_EdgeCases(t *testing.T) {
	// Case: Columnless Archetype Operations
	t.Run("Columnless operations", func(t *testing.T) {
		initCapacity := 128
		archReg := setupTestRegistry()
		archId := archReg.InitArchetype(ArchetypeMask{}, initCapacity)
		arch := &archReg.Archetypes[archId]
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

		archReg := setupTestRegistry()
		archId := archReg.InitArchetype(NewArchetypeMask(1), initCapacity)
		arch := &archReg.Archetypes[archId]

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

	archReg := setupTestRegistry()
	archId := archReg.InitArchetype(mask, initCapacity)
	arch := &archReg.Archetypes[archId]

	id := ComponentID(1)

	info := reflect.TypeFor[OddStruct]()
	arch.Columns[id] = &Column{
		dataType: info,
		ItemSize: info.Size(),
	}
	arch.Columns[id].growTo(arch.cap)

	val := OddStruct{A: 42}

	// 1. Manually register and set data
	row := arch.registerEntity(Entity(1))
	arch.Columns[id].setData(row, unsafe.Pointer(&val))

	// 2. Check if we can retrieve it without corruption
	// GetElement must return the exact pointer for row 0
	res := *(*OddStruct)(arch.Columns[id].GetElement(row))
	if res.A != 42 {
		t.Errorf("alignment or size issue: expected 42, got %d", res.A)
	}
}

func TestArchetype_DataIntegrityAfterGrowth(t *testing.T) {
	// 1. Setup with small capacity to force growth quickly
	initCap := 2
	mask := NewArchetypeMask(1)

	archReg := setupTestRegistry()

	archId := archReg.InitArchetype(mask, initCap)
	arch := &archReg.Archetypes[archId]

	addTestColumn[int32](arch, 1)

	// 2. Add entities until full (0 and 1)
	val0, val1 := int32(100), int32(200)

	row0 := arch.registerEntity(Entity(0))
	arch.Columns[1].setData(row0, unsafe.Pointer(&val0))

	row1 := arch.registerEntity(Entity(1))
	arch.Columns[1].setData(row1, unsafe.Pointer(&val1))

	// 3. This call triggers ensureCapacity() -> growTo(4).
	// If col.len wasn't incremented in registerEntity,
	// growTo WON'T COPY val0 and val1 to the new memory block!
	val2 := int32(300)
	row2 := arch.registerEntity(Entity(2))
	arch.Columns[1].setData(row2, unsafe.Pointer(&val2))

	// 4. VERIFICATION: Is the data from the first memory block still there?
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
	archReg := setupTestRegistry()
	archId := archReg.InitArchetype(NewArchetypeMask(1), 2) // Cap = 2
	arch := &archReg.Archetypes[archId]
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
	archReg := setupTestRegistry()
	archId := archReg.InitArchetype(NewArchetypeMask(1), 5) // Cap = 2
	arch := &archReg.Archetypes[archId]

	addTestColumn[int32](arch, 1)

	arch.registerEntity(Entity(1)) // len = 1

	// Usuwamy jedyną (ostatnią) encję
	arch.SwapRemoveEntity(0)

	if arch.len != 0 || arch.Columns[1].len != 0 {
		t.Errorf("Failed to sync len when removing last entity: arch=%d, col=%d",
			arch.len, arch.Columns[1].len)
	}
}
