package ecs

import (
	"reflect"
	"testing"
	"unsafe"
)

// 1. Fast Path Discovery
func TestArchetypeRegistry_FastPath(t *testing.T) {
	reg := setupTestRegistry()
	e1, e2 := Entity(1), Entity(2)
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	posData := position{10, 20}

	reg.AddEntity(e1)
	reg.AddEntity(e2)

	// First assignment: Should build the graph edge (Slow Path)
	reg.Assign(e1, posTypeInfo, unsafe.Pointer(&posData))

	arch1 := reg.entityArchLinks[e1.Index()].arch
	if arch1 == reg.rootArch {
		t.Fatal("entity E1 should have moved from rootArch")
	}

	// Case: Verify edge was cached in rootArch
	if _, ok := reg.rootArch.edgesNext[posTypeInfo.ID]; !ok {
		t.Fatal("fast path edge was not cached in rootArch")
	}

	// Second assignment: Should follow the edge (Fast Path)
	reg.Assign(e2, posTypeInfo, unsafe.Pointer(&posData))
	arch2 := reg.entityArchLinks[e2.Index()].arch

	if arch1 != arch2 {
		t.Error("E1 and E2 should share the same archetype instance via graph edges")
	}
}

// 2. Bidirectional Cycle
// 2. Bidirectional Cycle
func TestArchetypeRegistry_CycleConsistency(t *testing.T) {
	reg := setupTestRegistry()
	e := Entity(10)
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())

	reg.AddEntity(e)

	// Root -> +Pos -> ArchA
	// Passing a valid pointer to avoid panic in column.setData
	posData := position{x: 10, y: 20}
	reg.Assign(e, posTypeInfo, unsafe.Pointer(&posData))
	archA := reg.entityArchLinks[e.Index()].arch

	if archA == nil || archA == reg.rootArch {
		t.Fatal("entity failed to move to a new archetype")
	}

	// Case: Verify back-link exists in the graph
	if archA.edgesPrev[posTypeInfo.ID] != reg.rootArch {
		t.Error("bidirectional link (edgesPrev) from ArchA to Root not established")
	}
}

// 3. Graph Branching
func TestArchetypeRegistry_GraphBranching(t *testing.T) {
	reg := setupTestRegistry()
	e1, e2 := Entity(100), Entity(101)
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	velTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[velocity]())

	// Valid data to avoid nil pointer dereference in unsafe operations
	pData := position{x: 1, y: 1}
	vData := velocity{vx: 10, vy: 10}

	reg.AddEntity(e1)
	reg.AddEntity(e2)

	// Branching from Root to two different archetypes
	// Use actual pointers instead of nil
	reg.Assign(e1, posTypeInfo, unsafe.Pointer(&pData))
	reg.Assign(e2, velTypeInfo, unsafe.Pointer(&vData))

	// Case: Root should have 2 independent outgoing edges
	if len(reg.rootArch.edgesNext) != 2 {
		t.Errorf("expected 2 outgoing edges from Root, got %d", len(reg.rootArch.edgesNext))
	}

	archPos := reg.rootArch.edgesNext[posTypeInfo.ID]
	archVel := reg.rootArch.edgesNext[velTypeInfo.ID]

	if archPos == archVel {
		t.Error("different components must lead to distinct archetypes")
	}
}

// 4. Removal Strategy (Empty Mask = Removal)
func TestArchetypeRegistry_RemovalStrategy(t *testing.T) {
	reg := setupTestRegistry()
	e := Entity(50)
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	pData := position{x: 1, y: 1}

	reg.AddEntity(e)

	// Validating if Assign works with the new error-returning signature
	if err := reg.Assign(e, posTypeInfo, unsafe.Pointer(&pData)); err != nil {
		t.Fatalf("failed to assign component: %v", err)
	}

	// Case: UnAssigning the only component should trigger RemoveEntity
	reg.UnAssign(e, posTypeInfo)

	index := e.Index()
	// Fix: changed 'r' to 'reg'
	if int(index) >= len(reg.entityArchLinks) {
		t.Fatal("entity link index out of bounds")
	}

	link := reg.entityArchLinks[index]
	if link.arch != nil {
		t.Errorf("entity should be removed (arch == nil), but still linked to archetype with mask: %v", link.arch.mask)
	}
}

// 5. Data Idempotency
func TestArchetypeRegistry_OverwriteIdempotency(t *testing.T) {
	reg := setupTestRegistry()
	e := Entity(7)
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	p1 := position{1, 1}
	p2 := position{2, 2}

	reg.AddEntity(e)
	reg.Assign(e, posTypeInfo, unsafe.Pointer(&p1))

	linkBefore := reg.entityArchLinks[e.Index()]

	// Case: Assign same component again (update data)
	reg.Assign(e, posTypeInfo, unsafe.Pointer(&p2))

	linkAfter := reg.entityArchLinks[e.Index()]

	if linkBefore.arch != linkAfter.arch || linkBefore.row != linkAfter.row {
		t.Error("re-assigning same component should not move entity in graph")
	}

	gotData := *(*position)(linkAfter.arch.columns[posTypeInfo.ID].GetElement(linkAfter.row))
	if gotData != p2 {
		t.Errorf("data update failed: got %+v, want %+v", gotData, p2)
	}
}

// 6. Structural Integrity (Swap-and-Pop link update)
func TestArchetypeRegistry_SwapPopIntegrity(t *testing.T) {
	reg := setupTestRegistry()
	e0, e1, e2 := Entity(0), Entity(1), Entity(2)
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	pData := position{x: 1, y: 1}

	// Fix: replace nil with pointer in all Assign calls
	reg.AddEntity(e0)
	reg.Assign(e0, posTypeInfo, unsafe.Pointer(&pData))
	reg.AddEntity(e1)
	reg.Assign(e1, posTypeInfo, unsafe.Pointer(&pData))
	reg.AddEntity(e2)
	reg.Assign(e2, posTypeInfo, unsafe.Pointer(&pData))

	reg.RemoveEntity(e1)

	link2 := reg.entityArchLinks[e2.Index()]
	if link2.row != 1 {
		t.Errorf("E2 should be at row 1, got %d", link2.row)
	}
}

func setupTestRegistry() *ArchetypeRegistry {
	compReg := NewComponentsRegistry()
	compReg.GetOrRegister(reflect.TypeFor[position]())
	compReg.GetOrRegister(reflect.TypeFor[velocity]())

	// Mock ViewRegistry to avoid panics
	viewReg := &ViewRegistry{}

	return NewArchetypeRegistry(compReg, viewReg)
}

func TestArchetypeRegistry_AssignValidation(t *testing.T) {
	reg := setupTestRegistry()
	e := Entity(1)
	reg.AddEntity(e)
	type void struct{}
	voidTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[void]())
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())

	// Case: Passing tag
	err := reg.Assign(e, voidTypeInfo, nil)
	if err != nil {
		t.Errorf("unexpected error when assigning tag component: tags should allow nil data, but got: %v", err)
	}

	// Case: Passing nil data
	err = reg.Assign(e, posTypeInfo, nil)
	if err != ErrNilComponentData {
		t.Errorf("expected ErrNilComponentData, got %v", err)
	}

	// Case: Passing valid data
	pos := position{1, 2}
	err = reg.Assign(e, posTypeInfo, unsafe.Pointer(&pos))
	if err != nil {
		t.Errorf("unexpected error for valid assign: %v", err)
	}
}
