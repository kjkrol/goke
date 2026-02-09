package core

import (
	"reflect"
	"testing"
	"unsafe"
)

// Scenario 1: Capacity & Growth
func TestArchetype_CapacityAndGrowth(t *testing.T) {
	archReg := setupTestRegistry()

	posType := reflect.TypeFor[position]()
	posInfo := archReg.componentsRegistry.GetOrRegister(posType)
	posID := posInfo.ID

	initCapacity := 128
	mask := NewArchetypeMask(posID)

	archId := archReg.InitArchetype(mask, initCapacity)
	arch := &archReg.Archetypes[archId]

	// --- Case 1: Initial Capacity ---
	if arch.cap != initCapacity {
		t.Errorf("expected initial capacity %d, got %d", initCapacity, arch.cap)
	}

	col := arch.GetColumn(posID)
	if col == nil {
		t.Fatalf("InitArchetype failed to create column for component ID %d", posID)
	}

	if col.cap != initCapacity {
		t.Errorf("expected initial column capacity %d, got %d", initCapacity, col.cap)
	}

	// --- Case 2: Auto-growth ---
	for i := 0; i < initCapacity+1; i++ {
		arch.registerEntity(NewEntity(0, uint32(i+1)))
	}

	expectedCap := initCapacity * 2
	if arch.cap != expectedCap {
		t.Errorf("expected archetype capacity to grow to %d, got %d", expectedCap, arch.cap)
	}

	col = arch.GetColumn(posID)

	if col.cap != expectedCap {
		t.Errorf("column capacity should match archetype capacity: expected %d, got %d", expectedCap, col.cap)
	}
}

// Scenario 2: Manual Registration & Data Setting
// Scenario 2: Manual Registration & Data Setting
func TestArchetype_ManualRegistration(t *testing.T) {
	// 1. Setup Registry & Resolve Component IDs
	archReg := setupTestRegistry()

	posType := reflect.TypeFor[position]()
	posInfo := archReg.componentsRegistry.GetOrRegister(posType)
	posID := posInfo.ID

	velType := reflect.TypeFor[velocity]()
	velInfo := archReg.componentsRegistry.GetOrRegister(velType)
	velID := velInfo.ID

	// 2. Initialize Archetype
	// InitArchetype automatically allocates dense columns for posID and velID
	mask := NewArchetypeMask(posID, velID)
	archId := archReg.InitArchetype(mask, 128)
	arch := &archReg.Archetypes[archId]

	// 3. Prepare Data
	e0 := NewEntity(0, 10)
	p0 := position{x: 1, y: 1}

	// 4. Register Entity (Allocates row, data is zero-value initially)
	row := arch.registerEntity(e0)

	// 5. Set Data manually
	// We must use GetColumn because columns are now in a dense slice
	// and mapped via columnMap[ID]
	colPos := arch.GetColumn(posID)
	if colPos == nil {
		t.Fatalf("Column for Position (ID %d) missing", posID)
	}

	// Unsafe operation: Copy data from p0 to the archetype memory
	colPos.setData(row, unsafe.Pointer(&p0))

	colVel := arch.GetColumn(velID)
	if colVel == nil {
		t.Fatalf("Column for Velocity (ID %d) missing", velID)
	}
	colVel.zeroData(row) // Explicitly zero (optional, usually memory is already zeroed)

	// 6. Verify Position Data
	// Pointer arithmetic: Get address -> Cast to *position -> Dereference
	gotPtr := colPos.GetElement(row)
	gotP := *(*position)(gotPtr)

	if gotP != p0 {
		t.Errorf("expected position %+v, got %+v", p0, gotP)
	}

	// 7. Verify Velocity Data (Should be zero)
	gotVelPtr := colVel.GetElement(row)
	gotVel := *(*velocity)(gotVelPtr)
	emptyVel := velocity{}

	if gotVel != emptyVel {
		t.Errorf("expected empty velocity %+v, got %+v", emptyVel, gotVel)
	}
}

// Scenario 3: Swap-and-Pop Logic
func TestArchetype_SwapRemoveWithLenSync(t *testing.T) {
	// 1. Setup Registry
	archReg := setupTestRegistry()

	// 2. Register 'int32' type dynamically for this test
	// We need to do this so InitArchetype knows how to allocate columns for it.
	int32Type := reflect.TypeFor[int32]()
	int32Info := archReg.componentsRegistry.GetOrRegister(int32Type)
	int32ID := int32Info.ID

	// 3. Init Archetype
	mask := NewArchetypeMask(int32ID)
	archId := archReg.InitArchetype(mask, 10)
	arch := &archReg.Archetypes[archId]

	// Prepare data
	val1, val2 := int32(10), int32(20)

	// Helper to get the column safely
	col := arch.GetColumn(int32ID)
	if col == nil {
		t.Fatalf("Column for int32 (ID %d) not created", int32ID)
	}

	// 4. Add entities using the actual API
	// Row 0 -> Val 10
	row0 := arch.registerEntity(NewEntity(0, 1))
	col.setData(row0, unsafe.Pointer(&val1))

	// Row 1 -> Val 20
	row1 := arch.registerEntity(NewEntity(0, 2))
	col.setData(row1, unsafe.Pointer(&val2))

	// 5. Verify initial state
	if arch.len != 2 {
		t.Fatalf("Setup failed: Arch len should be 2, got %d", arch.len)
	}
	if col.len != 2 {
		t.Fatalf("Setup failed: Column len should be 2, got %d", col.len)
	}

	// 6. Perform Swap-Remove on Row 0
	// Logic: Entity at last row (row 1, val 20) moves to row 0.
	// Entity at row 0 is removed.
	swappedEntity, swapped := arch.SwapRemoveEntity(0)

	// 7. Verify sync and length reduction
	if arch.len != 1 {
		t.Errorf("Archetype len should decrease to 1, got %d", arch.len)
	}

	if col.len != 1 {
		t.Errorf("Column len should decrease to 1, got %d", col.len)
	}

	// 8. Critical check: Did the data actually swap?
	// We expect val2 (20) to be now at row 0
	gotPtr := col.GetElement(0)
	gotVal := *(*int32)(gotPtr)

	if gotVal != val2 {
		t.Errorf("Swap failed: expected val2 (20) at row 0, got %d", gotVal)
	}

	// 9. Verify Entity Swap Return
	if !swapped {
		t.Errorf("Expected swapped=true")
	}
	if swappedEntity.Index() != 2 {
		t.Errorf("Expected Entity(2) to be swapped into the gap, got Entity(%d)", swappedEntity.Index())
	}

	// 10. Safety Check: Column vs Archetype sync
	if col.len != arch.len {
		t.Errorf("Critical Sync Error: Archetype len (%d) != Column len (%d)",
			arch.len, col.len)
	}
}

// Scenario 4: Multi-Column Integrity
func TestArchetype_MultiColumnIntegrity(t *testing.T) {
	// 1. Setup Registry
	archReg := setupTestRegistry()

	// 2. Resolve Component IDs
	// Używamy GetOrRegister, aby mieć pewność, że mamy poprawne ID i Typy
	posType := reflect.TypeFor[position]()
	posInfo := archReg.componentsRegistry.GetOrRegister(posType)

	velType := reflect.TypeFor[velocity]()
	velInfo := archReg.componentsRegistry.GetOrRegister(velType)

	// 3. Init Archetype
	// InitArchetype automatycznie tworzy gęste kolumny dla obu komponentów
	mask := NewArchetypeMask(posInfo.ID, velInfo.ID)
	initCapacity := 128

	archId := archReg.InitArchetype(mask, initCapacity)
	arch := &archReg.Archetypes[archId]

	// 4. Prepare Data
	// Entity 0 -> Pos{1,1}, Vel{10,10}
	// Entity 1 -> Pos{2,2}, Vel{20,20}
	p0, v0 := position{1, 1}, velocity{10, 10}
	p1, v1 := position{2, 2}, velocity{20, 20}

	// Helpery do kolumn (Pointer caching)
	colPos := arch.GetColumn(posInfo.ID)
	colVel := arch.GetColumn(velInfo.ID)

	if colPos == nil || colVel == nil {
		t.Fatal("Failed to initialize columns")
	}

	// 5. Setup Entity 0 at Row 0
	e0 := NewEntity(0, 100)
	row0 := arch.registerEntity(e0) // row0 == 0
	colPos.setData(row0, unsafe.Pointer(&p0))
	colVel.setData(row0, unsafe.Pointer(&v0))

	// 6. Setup Entity 1 at Row 1
	e1 := NewEntity(0, 200)
	row1 := arch.registerEntity(e1) // row1 == 1
	colPos.setData(row1, unsafe.Pointer(&p1))
	colVel.setData(row1, unsafe.Pointer(&v1))

	// Verify setup sanity
	if arch.len != 2 {
		t.Fatalf("Setup failed, len %d", arch.len)
	}

	// 7. Swap-and-Pop: Remove Entity 0 (Row 0)
	// Logic:
	// - Entity 0 is removed.
	// - Entity 1 (last element) is moved from Row 1 to Row 0.
	// - This move MUST happen in BOTH columns (Position and Velocity).
	swappedEntity, swapped := arch.SwapRemoveEntity(row0)

	if !swapped {
		t.Error("Expected swapped=true")
	}
	if swappedEntity.Index() != e1.Index() {
		t.Errorf("Expected Entity 1 (ID %d) to be swapped, got ID %d", e1.Index(), swappedEntity.Index())
	}

	// 8. Case: Data integrity (E1 components must remain paired correctly at Row 0)
	// Sprawdzamy, czy pod indeksem 0 leżą teraz dane encji nr 1.

	// Check Position at Row 0
	gotPtrP := colPos.GetElement(0)
	gotP := *(*position)(gotPtrP)

	// Check Velocity at Row 0
	gotPtrV := colVel.GetElement(0)
	gotV := *(*velocity)(gotPtrV)

	if gotP != p1 {
		t.Errorf("Position integrity lost: expected %+v (Entity 1), got %+v", p1, gotP)
	}
	if gotV != v1 {
		t.Errorf("Velocity integrity lost: expected %+v (Entity 1), got %+v", v1, gotV)
	}

	// 9. Verify Lengths
	if arch.len != 1 {
		t.Errorf("Archetype len expected 1, got %d", arch.len)
	}
	if colPos.len != 1 || colVel.len != 1 {
		t.Errorf("Column len desync")
	}
}

// Scenario 5: Edge Cases
func TestArchetype_EdgeCases(t *testing.T) {
	// Case: Columnless Archetype Operations (e.g., pure tagging or root arch)
	t.Run("Columnless operations (Entity Column Only)", func(t *testing.T) {
		initCapacity := 128
		archReg := setupTestRegistry()

		// Empty Mask -> Creates 1 Column (for Entities only)
		archId := archReg.InitArchetype(ArchetypeMask{}, initCapacity)
		arch := &archReg.Archetypes[archId]

		// 1. Check Initial State
		// NOWA LOGIKA: Pusty archetyp ma 1 kolumnę (indeks 0 to encje)
		if len(arch.columns) != 1 {
			t.Fatalf("expected 1 column (Entity Column) in empty-mask archetype, got %d", len(arch.columns))
		}

		// 2. Register
		entity := NewEntity(0, 100)
		arch.registerEntity(entity)

		if arch.len != 1 {
			t.Fatal("failed to register entity in columnless archetype")
		}

		// Verify Entity is stored in column 0
		storedEntity := *(*Entity)(arch.columns[0].Data)
		if storedEntity != entity {
			t.Fatalf("Entity storage corrupted: expected %v, got %v", entity, storedEntity)
		}

		// 3. Remove
		arch.SwapRemoveEntity(0)

		if arch.len != 0 {
			t.Error("failed to remove entity from columnless archetype")
		}
	})

	// Case: Remove from single-entity archetype
	// Tests if SwapRemove handles the case where row == lastRow correctly (no swap needed)
	t.Run("Single-entity remove", func(t *testing.T) {
		initCapacity := 128
		archReg := setupTestRegistry()

		// Resolve component ID (using position which is already registered by setupTestRegistry)
		posType := reflect.TypeFor[position]()
		posInfo := archReg.componentsRegistry.GetOrRegister(posType)

		// Init Archetype with 1 component
		mask := NewArchetypeMask(posInfo.ID)
		archId := archReg.InitArchetype(mask, initCapacity)
		arch := &archReg.Archetypes[archId]

		// Add 1 Entity
		arch.registerEntity(NewEntity(0, 5))

		// Verify setup
		if arch.len != 1 {
			t.Fatal("Setup failed: len should be 1")
		}

		// Perform SwapRemove on the ONLY element (index 0)
		moved, swapped := arch.SwapRemoveEntity(0)

		// Expectations:
		// 1. swapped == false (because we removed the last element, nothing swapped into its place)
		// 2. moved == Entity{} (Zero value, because nothing moved)
		// 3. len == 0
		if swapped {
			t.Error("swapped should be false when removing the last element")
		}
		if moved != 0 {
			t.Errorf("moved entity should be empty/zero, got %+v", moved)
		}
		if arch.len != 0 {
			t.Error("invalid len after removing the only entity")
		}
	})
}

// Scenario: Memory Alignment
func TestArchetype_Alignment(t *testing.T) {
	// Case: Testing with a struct that has odd size (1 byte)
	// This ensures that the engine handles non-word-aligned sizes correctly.
	type OddStruct struct {
		A int8 // Size: 1 byte, Alignment: 1
	}

	// 1. Setup Registry
	archReg := setupTestRegistry()

	// 2. Register the custom type dynamically
	// We MUST register it so InitArchetype knows the size (1 byte)
	// and can allocate the column correctly.
	oddType := reflect.TypeFor[OddStruct]()
	oddInfo := archReg.componentsRegistry.GetOrRegister(oddType)
	oddID := oddInfo.ID

	// 3. Init Archetype
	mask := NewArchetypeMask(oddID)
	initCapacity := 128

	// InitArchetype will create a dense column for oddID with ItemSize=1
	archId := archReg.InitArchetype(mask, initCapacity)
	arch := &archReg.Archetypes[archId]

	// 4. Get the column safely
	col := arch.GetColumn(oddID)
	if col == nil {
		t.Fatalf("Column for OddStruct (ID %d) was not created", oddID)
	}

	// Verify ItemSize matches the struct size
	if col.ItemSize != 1 {
		t.Errorf("Expected ItemSize 1, got %d", col.ItemSize)
	}

	// 5. Create Data
	val := OddStruct{A: 42}

	// 6. Register Entity & Set Data
	row := arch.registerEntity(NewEntity(0, 1))

	// Use unsafe pointer to write 1 byte into the column memory
	col.setData(row, unsafe.Pointer(&val))

	// 7. Check if we can retrieve it without corruption
	// GetElement must return the exact pointer for the row.
	// If striding was wrong (e.g. assuming 8 bytes), this would read garbage.
	ptr := col.GetElement(row)
	res := *(*OddStruct)(ptr)

	if res.A != 42 {
		t.Errorf("alignment or size issue: expected 42, got %d", res.A)
	}
}

func TestArchetype_DataIntegrityAfterGrowth(t *testing.T) {
	// 1. Setup Registry
	archReg := setupTestRegistry()

	// Register 'int32' to get a valid ID and ComponentInfo
	int32Type := reflect.TypeFor[int32]()
	int32Info := archReg.componentsRegistry.GetOrRegister(int32Type)
	int32ID := int32Info.ID

	// 2. Setup with SMALL capacity to force growth quickly
	initCap := 2
	mask := NewArchetypeMask(int32ID)

	// InitArchetype automatically creates the column for int32ID
	archId := archReg.InitArchetype(mask, initCap)
	arch := &archReg.Archetypes[archId]

	// 3. Add entities until full (Capacity 2 -> 2 entities)
	val0, val1 := int32(100), int32(200)

	// Add Entity 0
	row0 := arch.registerEntity(NewEntity(0, 10))
	// We use GetColumn to ensure we get the valid pointer to the column struct
	col := arch.GetColumn(int32ID)
	col.setData(row0, unsafe.Pointer(&val0))

	// Add Entity 1
	row1 := arch.registerEntity(NewEntity(0, 11))
	col.setData(row1, unsafe.Pointer(&val1))

	// Verify we are exactly at capacity
	if arch.len != 2 || arch.cap != 2 {
		t.Fatalf("Setup failed: expected len=2, cap=2. Got len=%d, cap=%d", arch.len, arch.cap)
	}

	// 4. TRIGGER GROWTH (Add 3rd Entity)
	// This call triggers ensureCapacity() -> which calls col.growTo(4).
	// The internal 'Data' pointer in the column changes here!
	// The implementation MUST copy val0 and val1 to the new memory block.
	val2 := int32(300)
	row2 := arch.registerEntity(NewEntity(0, 12)) // New capacity should be 4

	// Re-fetch column pointer just to be safe (though the struct in slice is stable-ish, internal Data changed)
	col = arch.GetColumn(int32ID)
	col.setData(row2, unsafe.Pointer(&val2))

	// 5. VERIFICATION: Is the data from the first memory block still there?
	// If the copy logic was missing or broken, these would read garbage or zeros.

	// Check Entity 0 (Old Data)
	got0 := *(*int32)(col.GetElement(0))
	if got0 != 100 {
		t.Errorf("DATA LOSS: Entity 0 data corrupted after growth. Expected 100, got %d", got0)
	}

	// Check Entity 1 (Old Data)
	got1 := *(*int32)(col.GetElement(1))
	if got1 != 200 {
		t.Errorf("DATA LOSS: Entity 1 data corrupted after growth. Expected 200, got %d", got1)
	}

	// Check Entity 2 (New Data)
	got2 := *(*int32)(col.GetElement(2))
	if got2 != 300 {
		t.Errorf("Entity 2 (new) data mismatch. Expected 300, got %d", got2)
	}

	// Verify new capacity
	if arch.cap != 4 {
		t.Errorf("Expected capacity to double to 4, got %d", arch.cap)
	}
}

func TestArchetype_RegisterAndSetData(t *testing.T) {
	// 1. Setup Registry
	archReg := setupTestRegistry()

	// 2. Register 'int32' type dynamically
	int32Type := reflect.TypeFor[int32]()
	int32Info := archReg.componentsRegistry.GetOrRegister(int32Type)
	int32ID := int32Info.ID

	// 3. Init Archetype with small capacity (Cap = 2)
	// InitArchetype automatically creates the column for int32ID
	archId := archReg.InitArchetype(NewArchetypeMask(int32ID), 2)
	arch := &archReg.Archetypes[archId]

	// 4. Fill to capacity (Len becomes 2, Cap is 2)
	arch.registerEntity(NewEntity(0, 0))
	arch.registerEntity(NewEntity(0, 1))

	// 5. Trigger Growth
	// This registration calls ensureCapacity -> growTo(4).
	// The internal column 'Data' pointer changes to a new, larger memory block.
	newRow := arch.registerEntity(NewEntity(0, 2))

	// 6. Simulate setData (like in moveEntity)
	// We get the column pointer (which points to the struct in the slice)
	col := arch.GetColumn(int32ID)
	if col == nil {
		t.Fatal("Column missing")
	}

	val := int32(777)

	// CRITICAL: This setData must write to the NEW memory block.
	// If the column struct wasn't updated during growth, this would write
	// to old freed memory or crash.
	col.setData(newRow, unsafe.Pointer(&val))

	// 7. Verify read back
	gotPtr := col.GetElement(newRow)
	got := *(*int32)(gotPtr)

	if got != 777 {
		t.Errorf("setData failed after grow: expected 777, got %d", got)
	}
}

func TestArchetype_RemoveLastEntityLenSync(t *testing.T) {
	// 1. Setup Registry
	archReg := setupTestRegistry()

	// 2. Register 'int32' dynamically
	// Musimy zarejestrować typ, aby InitArchetype wiedział, że ma utworzyć kolumnę
	int32Type := reflect.TypeFor[int32]()
	int32Info := archReg.componentsRegistry.GetOrRegister(int32Type)
	int32ID := int32Info.ID

	// 3. Init Archetype
	// Tworzymy archetyp z pojemnością 5 i maską zawierającą int32
	mask := NewArchetypeMask(int32ID)
	archId := archReg.InitArchetype(mask, 5)
	arch := &archReg.Archetypes[archId]

	// (Usunięto addTestColumn - InitArchetype tworzy kolumny automatycznie)

	// 4. Add the ONLY entity
	// len: 0 -> 1
	arch.registerEntity(NewEntity(0, 1))

	// Safety check przed operacją
	col := arch.GetColumn(int32ID)
	if arch.len != 1 || col.len != 1 {
		t.Fatalf("Setup failed: expected len=1, got arch=%d, col=%d", arch.len, col.len)
	}

	// 5. Remove the only entity (Index 0)
	// SwapRemoveEntity(0) na tablicy o długości 1 po prostu zmniejsza len do 0.
	// Nie ma nic do "swapowania" (bo ostatni element jest tym usuwanym).
	arch.SwapRemoveEntity(0)

	// 6. Verify Sync
	if arch.len != 0 {
		t.Errorf("Archetype len should be 0, got %d", arch.len)
	}

	if col.len != 0 {
		t.Errorf("Column len should be 0, got %d", col.len)
	}
}
