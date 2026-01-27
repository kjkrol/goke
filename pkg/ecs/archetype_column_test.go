package ecs

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

// 1. Basic Allocation and Growth (growTo)
func TestColumn_AllocationAndGrowth(t *testing.T) {
	dataType := reflect.TypeFor[position]()
	col := &column{
		dataType: dataType,
		itemSize: dataType.Size(),
	}

	// Case: Initial Growth
	// Checking if data is available and pointer is not nil after first growTo.
	t.Run("Initial Growth", func(t *testing.T) {
		col.growTo(10)
		if col.data == nil {
			t.Fatal("expected column data pointer to be allocated, got nil")
		}
		if col.cap != 10 {
			t.Errorf("expected capacity 10, got %d", col.cap)
		}
	})

	// Case: Data Preservation
	// Verifying that data is correctly copied to the new memory location after growth.
	t.Run("Data Preservation", func(t *testing.T) {
		val := position{x: 1.23, y: 4.56}
		col.setData(0, unsafe.Pointer(&val))
		col.len = 1

		// Trigger reallocation
		col.growTo(20)

		got := *(*position)(col.GetElement(0))
		if got != val {
			t.Errorf("data corrupted after growth: got %+v, want %+v", got, val)
		}
	})

	// Case: Capacity vs Length
	// Checking if growTo only affects capacity and doesn't implicitly change length.
	t.Run("Capacity vs Length", func(t *testing.T) {
		initialLen := col.len
		col.growTo(30)
		if col.len != initialLen {
			t.Errorf("growTo should not change length: got %d, want %d", col.len, initialLen)
		}
		if col.cap != 30 {
			t.Errorf("expected capacity 30, got %d", col.cap)
		}
	})
}

// 2. Data Manipulation (GetElement / setData)
func TestColumn_DataManipulation(t *testing.T) {
	// Local struct for support testing
	type largeStruct struct {
		A      float64
		B      float64
		Active bool
	}

	dataType := reflect.TypeFor[largeStruct]()
	col := &column{
		dataType: dataType,
		itemSize: dataType.Size(),
	}
	col.growTo(10)

	// Case: Pointer Arithmetic
	// Checking if GetElement returns addresses shifted exactly by itemSize * index.
	t.Run("Pointer Arithmetic", func(t *testing.T) {
		ptr0 := uintptr(col.GetElement(0))
		ptr1 := uintptr(col.GetElement(1))
		ptr5 := uintptr(col.GetElement(5))

		if ptr1-ptr0 != col.itemSize {
			t.Errorf("pointer shift mismatch: got %d, want %d", ptr1-ptr0, col.itemSize)
		}
		if ptr5-ptr0 != col.itemSize*5 {
			t.Errorf("multi-row pointer shift mismatch: got %d, want %d", ptr5-ptr0, col.itemSize*5)
		}
	})

	// Case: Value Integrity
	// Writing a value via setData and reading it back via GetElement.
	t.Run("Value Integrity", func(t *testing.T) {
		val := largeStruct{A: 10.5, B: 20.5, Active: true}
		col.setData(2, unsafe.Pointer(&val))

		got := *(*largeStruct)(col.GetElement(2))
		if got != val {
			t.Errorf("value mismatch: got %+v, want %+v", got, val)
		}
	})

	// Case: Struct Support
	// Ensuring memmove copies the entire block, including fields with different alignments.
	t.Run("Struct Support", func(t *testing.T) {
		val := largeStruct{A: 99.9, B: -99.9, Active: false}
		col.setData(9, unsafe.Pointer(&val))

		got := *(*largeStruct)(col.GetElement(9))
		if got.A != val.A || got.B != val.B || got.Active != val.Active {
			t.Errorf("struct data corrupted: got %+v, want %+v", got, val)
		}
	})
}

// 3. Lifecycle Management (zeroData / copyData)
func TestColumn_LifecycleManagement(t *testing.T) {
	dataType := reflect.TypeFor[int64]()
	col := &column{
		dataType: dataType,
		itemSize: dataType.Size(),
	}
	col.growTo(10)

	// Case: Zeroing Memory
	// Verifying that zeroData fills the memory with zeros.
	t.Run("Zeroing Memory", func(t *testing.T) {
		val := int64(0xDEADBEEF)
		col.setData(0, unsafe.Pointer(&val))
		col.zeroData(0)

		got := *(*int64)(col.GetElement(0))
		if got != 0 {
			t.Errorf("zeroData failed: expected 0, got %X", got)
		}
	})

	// Case: Implicit Length Growth
	// Verifying that zeroData updates c.len if index is higher than current length.
	t.Run("Implicit Length Growth", func(t *testing.T) {
		col.len = 1
		col.zeroData(5)
		if col.len != 6 {
			t.Errorf("expected length to grow to 6, got %d", col.len)
		}
	})

	// Case: Internal Copy
	// Testing copyData (swap-and-pop simulation) within the same column.
	t.Run("Internal Copy", func(t *testing.T) {
		vSource := int64(100)
		vDest := int64(200)
		col.setData(1, unsafe.Pointer(&vSource)) // Source at index 1
		col.setData(0, unsafe.Pointer(&vDest))   // Dest at index 0

		col.copyData(0, 1)

		got := *(*int64)(col.GetElement(0))
		if got != vSource {
			t.Errorf("copyData failed: expected %d, got %d", vSource, got)
		}
	})
}

// 4. Garbage Collector Persistence (GC Persistence)
func TestColumn_GCPersistence(t *testing.T) {
	// Case: GC Stress Test
	// Ensuring c.rawSlice prevents GC from reclaiming col.data.
	t.Run("GC Stress Test", func(t *testing.T) {
		dataType := reflect.TypeFor[position]()
		col := &column{
			dataType: dataType,
			itemSize: dataType.Size(),
		}
		col.growTo(100)

		for i := 0; i < 100; i++ {
			val := position{x: float64(i), y: float64(i)}
			col.setData(ArchRow(i), unsafe.Pointer(&val))
		}

		// Force aggressive GC
		runtime.GC()
		runtime.GC()

		for i := 0; i < 100; i++ {
			got := (*position)(col.GetElement(ArchRow(i)))
			if got.x != float64(i) {
				t.Fatalf("GC corruption at index %d: expected %f, got %f", i, float64(i), got.x)
			}
		}
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
		col := &column{
			dataType: dataType,
			itemSize: dataType.Size(),
		}
		col.growTo(2)

		val := oddStruct{}
		val.Data[0] = 0xFF
		val.Data[12] = 0xAA

		col.setData(1, unsafe.Pointer(&val))
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
		col := &column{
			dataType: dataType,
			itemSize: dataType.Size(),
		}
		col.growTo(10)

		// Arithmetic should not panic, though pointers will be identical
		ptr0 := col.GetElement(0)
		ptr1 := col.GetElement(1)

		if ptr0 != ptr1 && col.itemSize == 0 {
			// In Go, zero-sized objects may have the same address
		}
	})
}
