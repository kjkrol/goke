package core

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

type position struct {
	x, y float64
}

type velocity struct {
	vx, vy float64
}

// 1. Basic Allocation and Growth (MemoryBlock)
func TestMemoryBlock_AllocationAndGrowth(t *testing.T) {
	// Setup: Define component info
	posType := reflect.TypeFor[position]()
	posInfo := ComponentInfo{
		ID:   1, // Arbitrary ID
		Type: posType,
		Size: posType.Size(),
	}

	// Initialize MemoryBlock with 0 capacity initially to test growth
	var block MemoryBlock
	block.Init(0, []ComponentInfo{posInfo})

	// Helper to get the position column (Index 1, because 0 is Entity)
	getPosColumn := func() *Column {
		return &block.Columns[1]
	}

	// Case: Initial Growth
	// Checking if data is available and pointer is not nil after first Resize.
	t.Run("Initial Growth", func(t *testing.T) {
		block.Resize(10)

		col := getPosColumn()
		if col.Data == nil {
			t.Fatal("expected column data pointer to be allocated, got nil")
		}
		if block.Cap != 10 {
			t.Errorf("expected capacity 10, got %d", block.Cap)
		}
	})

	// Case: Data Preservation
	// Verifying that data is correctly copied to the new memory location after growth.
	t.Run("Data Preservation", func(t *testing.T) {
		col := getPosColumn()

		val := position{x: 1.23, y: 4.56}
		// Write data to row 0
		col.SetData(0, unsafe.Pointer(&val))

		// Manually update Len (usually done by Archetype)
		block.Len = 1

		// Trigger reallocation
		block.Resize(20)

		// Check if data is still there (col.Data should have changed, but point to copied data)
		// Note: We don't need to re-fetch 'col' pointer because MemoryBlock.Columns slice header stays stable,
		// only the internal .Data pointer of the column changes.
		gotPtr := col.GetElement(0)
		got := *(*position)(gotPtr)

		if got != val {
			t.Errorf("data corrupted after growth: got %+v, want %+v", got, val)
		}
	})

	// Case: Capacity vs Length
	// Checking if Resize only affects capacity and doesn't implicitly change length.
	t.Run("Capacity vs Length", func(t *testing.T) {
		initialLen := block.Len

		block.Resize(30)

		if block.Len != initialLen {
			t.Errorf("Resize should not change length: got %d, want %d", block.Len, initialLen)
		}
		if block.Cap != 30 {
			t.Errorf("expected capacity 30, got %d", block.Cap)
		}
	})
}

// 2. Data Manipulation (GetElement / SetData)
func TestColumn_DataManipulation(t *testing.T) {
	// Local struct for support testing
	type largeStruct struct {
		A      float64
		B      float64
		Active bool
	}

	// 1. Setup MemoryBlock containing our test component
	dataType := reflect.TypeFor[largeStruct]()

	// We need ComponentInfo to initialize the block
	compInfo := ComponentInfo{
		ID:   1, // Arbitrary ID
		Type: dataType,
		Size: dataType.Size(),
	}

	var block MemoryBlock
	// Initialize block with capacity 10 immediately
	block.Init(10, []ComponentInfo{compInfo})

	// 2. Retrieve the column to test
	// Index 0 is always EntityID, so our component is at Index 1
	col := &block.Columns[1]

	// Case: Pointer Arithmetic
	// Checking if GetElement returns addresses shifted exactly by ItemSize * index.
	t.Run("Pointer Arithmetic", func(t *testing.T) {
		// We cast to uintptr to perform arithmetic checks
		ptr0 := uintptr(col.GetElement(0))
		ptr1 := uintptr(col.GetElement(1))
		ptr5 := uintptr(col.GetElement(5))

		// Check offset for 1 element
		if ptr1-ptr0 != col.ItemSize {
			t.Errorf("pointer shift mismatch: got %d, want %d", ptr1-ptr0, col.ItemSize)
		}

		// Check offset for 5 elements
		if ptr5-ptr0 != col.ItemSize*5 {
			t.Errorf("multi-row pointer shift mismatch: got %d, want %d", ptr5-ptr0, col.ItemSize*5)
		}
	})

	// Case: Value Integrity
	// Writing a value via SetData and reading it back via GetElement.
	t.Run("Value Integrity", func(t *testing.T) {
		val := largeStruct{A: 10.5, B: 20.5, Active: true}

		// Note: SetData is now capitalized (public method)
		col.SetData(2, unsafe.Pointer(&val))

		got := *(*largeStruct)(col.GetElement(2))
		if got != val {
			t.Errorf("value mismatch: got %+v, want %+v", got, val)
		}
	})

	// Case: Struct Support
	// Ensuring copyMemory copies the entire block, including fields with different alignments.
	t.Run("Struct Support", func(t *testing.T) {
		val := largeStruct{A: 99.9, B: -99.9, Active: false}
		col.SetData(9, unsafe.Pointer(&val))

		got := *(*largeStruct)(col.GetElement(9))

		// Check fields individually just to be sure about alignment handling
		if got.A != val.A || got.B != val.B || got.Active != val.Active {
			t.Errorf("struct data corrupted: got %+v, want %+v", got, val)
		}
	})
}

// 3. Lifecycle Management (ZeroData / CopyData)
func TestColumn_LifecycleManagement(t *testing.T) {
	// 1. Setup MemoryBlock
	dataType := reflect.TypeFor[int64]()
	compInfo := ComponentInfo{
		ID:   1,
		Type: dataType,
		Size: dataType.Size(),
	}

	var block MemoryBlock
	block.Init(10, []ComponentInfo{compInfo}) // Capacity 10

	// Get the component column (Index 1)
	col := &block.Columns[1]

	// Case: Zeroing Memory
	// Verifying that ZeroData fills the memory with zeros.
	t.Run("Zeroing Memory", func(t *testing.T) {
		val := int64(0xDEADBEEF)

		// Write dirty value
		col.SetData(0, unsafe.Pointer(&val))

		// Clear it
		col.ZeroData(0)

		got := *(*int64)(col.GetElement(0))
		if got != 0 {
			t.Errorf("ZeroData failed: expected 0, got %X", got)
		}
	})

	// Case: Access at Offset (formerly Implicit Length Growth)
	// NOTE: In the new architecture, Column.ZeroData strictly clears memory.
	// It does NOT update 'Len'. Length management is now the responsibility of Archetype.
	// This test now simply verifies we can manipulate memory at higher indices within capacity.
	t.Run("Access at Offset", func(t *testing.T) {
		idx := ArchRow(5)
		val := int64(999)

		// Write value at index 5
		col.SetData(idx, unsafe.Pointer(&val))

		// Clear value at index 5
		col.ZeroData(idx)

		got := *(*int64)(col.GetElement(idx))
		if got != 0 {
			t.Errorf("ZeroData at offset failed: expected 0, got %d", got)
		}

		// We do NOT check block.Len here, because ZeroData works on raw pointers
		// and has no side effects on the block's state.
	})

	// Case: Internal Copy
	// Testing CopyData (swap-and-pop simulation) within the same column.
	t.Run("Internal Copy", func(t *testing.T) {
		vSource := int64(100)
		vDest := int64(200)

		col.SetData(1, unsafe.Pointer(&vSource)) // Source at index 1
		col.SetData(0, unsafe.Pointer(&vDest))   // Dest at index 0

		// Overwrite index 0 with data from index 1
		col.CopyData(0, 1)

		got := *(*int64)(col.GetElement(0))
		if got != vSource {
			t.Errorf("CopyData failed: expected %d, got %d", vSource, got)
		}
	})
}

// 4. Garbage Collector Persistence (GC Persistence)
func TestColumn_GCPersistence(t *testing.T) {
	// Case: GC Stress Test
	// Ensuring MemoryBlock.Meta.rawSlice prevents GC from reclaiming col.Data.
	t.Run("GC Stress Test", func(t *testing.T) {
		// 1. Setup MemoryBlock
		dataType := reflect.TypeFor[position]()
		info := ComponentInfo{ID: 1, Type: dataType, Size: dataType.Size()}

		var block MemoryBlock
		block.Init(100, []ComponentInfo{info}) // Pre-allocate 100 slots

		col := &block.Columns[1] // Get pointer to the component column

		// 2. Fill Data
		for i := 0; i < 100; i++ {
			val := position{x: float64(i), y: float64(i)}
			col.SetData(ArchRow(i), unsafe.Pointer(&val))
		}

		// 3. Force aggressive GC
		// The block variable is still in scope, so Go should keep the structure alive.
		// We are testing if the underlying array (pointed to by col.Data) is kept alive
		// by the reflect.Value in block.Meta, even though we only use unsafe pointers here.
		runtime.GC()
		runtime.GC()

		// 4. Verify Data Integrity
		for i := 0; i < 100; i++ {
			gotPtr := col.GetElement(ArchRow(i))
			got := (*position)(gotPtr)

			if got.x != float64(i) {
				t.Fatalf("GC corruption at index %d: expected %f, got %f", i, float64(i), got.x)
			}
		}

		// Keep block alive explicitly (though mostly handled by compiler in this scope)
		runtime.KeepAlive(&block)
	})
}

// 5. Data Types and Alignment
func TestColumn_AlignmentAndTypes(t *testing.T) {
	// Case: Various Sizes (non-power of two)
	// Testing types that don't fit power-of-two boundaries.
	t.Run("Various Sizes", func(t *testing.T) {
		type oddStruct struct {
			Data [13]byte
		}
		dataType := reflect.TypeFor[oddStruct]()
		info := ComponentInfo{ID: 1, Type: dataType, Size: dataType.Size()}

		var block MemoryBlock
		block.Init(2, []ComponentInfo{info})

		col := &block.Columns[1]

		val := oddStruct{}
		val.Data[0] = 0xFF
		val.Data[12] = 0xAA

		col.SetData(1, unsafe.Pointer(&val))
		got := (*oddStruct)(col.GetElement(1))

		if got.Data[0] != 0xFF || got.Data[12] != 0xAA {
			t.Error("data corruption in non-power-of-two struct")
		}
	})

	// Case: Zero-size types
	// Checking behavior with struct{}.
	t.Run("Zero-size types", func(t *testing.T) {
		type empty struct{}
		dataType := reflect.TypeFor[empty]()
		info := ComponentInfo{ID: 1, Type: dataType, Size: dataType.Size()}

		var block MemoryBlock
		block.Init(10, []ComponentInfo{info})

		col := &block.Columns[1]

		// Arithmetic should not panic, though pointers will be identical
		ptr0 := col.GetElement(0)
		ptr1 := col.GetElement(1)

		// In Go, zero-sized objects may have the same address.
		// Also, since itemSize is 0, ptr + (1 * 0) == ptr.
		if ptr0 != ptr1 {
			// This check depends on implementation, but mathematically they should be equal
			// if they come from the same base pointer.
			// If Go allocates distinct 0-size objects, that's fine too, but in this slab
			// allocator logic, they share the base pointer logic.
			if col.ItemSize == 0 && ptr0 != ptr1 {
				t.Log("Warning: Zero-sized elements have different addresses (unexpected for slab logic but safe)")
			}
		}

		// Ensure writing doesn't crash
		val := empty{}
		col.SetData(5, unsafe.Pointer(&val))
	})
}
