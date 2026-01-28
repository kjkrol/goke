package core

import (
	"reflect"
	"testing"
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
	if ptr, err := reg.AllocateComponentMemory(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}

	arch1 := reg.EntityLinkStore.Get(e1.Index()).Arch
	if arch1 == reg.rootArch {
		t.Fatal("entity E1 should have moved from rootArch")
	}

	// Case: Verify edge was cached in rootArch
	if _, ok := reg.rootArch.edgesNext[posTypeInfo.ID]; !ok {
		t.Fatal("fast path edge was not cached in rootArch")
	}

	// Second assignment: Should follow the edge (Fast Path)
	if ptr, err := reg.AllocateComponentMemory(e2, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}
	arch2 := reg.EntityLinkStore.Get(e2.Index()).Arch

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
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}
	archA := reg.EntityLinkStore.Get(e.Index()).Arch

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
	if ptr, err := reg.AllocateComponentMemory(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	}
	if ptr, err := reg.AllocateComponentMemory(e2, velTypeInfo); err == nil {
		*(*velocity)(ptr) = vData
	}

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
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	} else {
		t.Fatalf("failed to assign component: %v", err)
	}

	// Case: UnAssigning the only component should trigger RemoveEntity
	reg.UnAssign(e, posTypeInfo)

	index := e.Index()
	link := reg.EntityLinkStore.Get(index)

	if link.Arch != nil {
		t.Errorf("entity should be removed (arch == nil), but still linked to archetype with mask: %v", link.Arch.Mask)
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
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = p1
	}

	linkBefore := reg.EntityLinkStore.Get(e.Index())

	// Case: Assign same component again (update data)
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = p2
	}

	linkAfter := reg.EntityLinkStore.Get(e.Index())

	if linkBefore.Arch != linkAfter.Arch || linkBefore.Row != linkAfter.Row {
		t.Error("re-assigning same component should not move entity in graph")
	}

	gotData := *(*position)(linkAfter.Arch.Columns[posTypeInfo.ID].GetElement(linkAfter.Row))
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
	if ptr, err := reg.AllocateComponentMemory(e0, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	}

	reg.AddEntity(e1)
	if ptr, err := reg.AllocateComponentMemory(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	}
	reg.AddEntity(e2)
	if ptr, err := reg.AllocateComponentMemory(e2, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	}

	reg.UnlinkEntity(e1)

	link2 := reg.EntityLinkStore.Get(e2.Index())
	if link2.Row != 1 {
		t.Errorf("E2 should be at row 1, got %d", link2.Row)
	}
}

func setupTestRegistry() *ArchetypeRegistry {
	compReg := NewComponentsRegistry()
	compReg.GetOrRegister(reflect.TypeFor[position]())
	compReg.GetOrRegister(reflect.TypeFor[velocity]())

	// Mock ViewRegistry to avoid panics
	viewReg := &ViewRegistry{}

	return NewArchetypeRegistry(compReg, viewReg, DefaultRegistryConfig())
}

func TestArchetypeRegistry_AssignValidation(t *testing.T) {
	reg := setupTestRegistry()
	e := Entity(1)
	reg.AddEntity(e)
	type void struct{}
	voidTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[void]())
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())

	// Case: Passing tag
	if _, err := reg.AllocateComponentMemory(e, voidTypeInfo); err != nil {
		t.Errorf("unexpected error when assigning tag component: tags should allow nil data, but got: %v", err)
	}

	eX := Entity(3123)
	// Case: Passing nil data
	if _, err := reg.AllocateComponentMemory(eX, posTypeInfo); err != ErrEntityNotFound {
		t.Errorf("expected ErrNilComponentData, got %v", err)
	}

	// Case: Passing valid data
	pos := position{1, 2}
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = pos
	} else {
		t.Errorf("unexpected error for valid assign: %v", err)
	}
}
