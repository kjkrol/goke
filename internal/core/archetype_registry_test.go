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

	reg.AddEntity(e1, RootArchetypeId)
	reg.AddEntity(e2, RootArchetypeId)

	// First assignment: Should build the graph edge (Slow Path)
	if ptr, err := reg.AllocateComponentMemory(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}

	arch1, _ := reg.EntityLinkStore.Get(e1)
	if arch1.ArchId == RootArchetypeId {
		t.Fatal("entity E1 should have moved from rootArch")
	}

	// Case: Verify edge was cached in rootArch
	rootArch := &reg.Archetypes[RootArchetypeId]
	if nextEdge := rootArch.edgesNext[posTypeInfo.ID]; nextEdge == RootArchetypeId {
		t.Fatal("fast path edge was not cached in rootArch")
	}

	// Second assignment: Should follow the edge (Fast Path)
	if ptr, err := reg.AllocateComponentMemory(e2, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}
	arch2, _ := reg.EntityLinkStore.Get(e2)

	if arch1.ArchId != arch2.ArchId {
		t.Error("E1 and E2 should share the same archetype instance via graph edges")
	}
}

// 2. Bidirectional Cycle
func TestArchetypeRegistry_CycleConsistency(t *testing.T) {
	reg := setupTestRegistry()
	e := Entity(10)
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())

	reg.AddEntity(e, RootArchetypeId)

	// Root -> +Pos -> ArchA
	// Passing a valid pointer to avoid panic in column.setData
	posData := position{x: 10, y: 20}
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}
	linkA, _ := reg.EntityLinkStore.Get(e)

	if linkA.ArchId == NullArchetypeId || linkA.ArchId == RootArchetypeId {
		t.Fatal("entity failed to move to a new archetype")
	}

	// Case: Verify back-link exists in the graph
	linkAArch := &reg.Archetypes[linkA.ArchId]
	if linkAArch.edgesPrev[posTypeInfo.ID] != RootArchetypeId {
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

	reg.AddEntity(e1, RootArchetypeId)
	reg.AddEntity(e2, RootArchetypeId)

	// Branching from Root to two different archetypes
	// Use actual pointers instead of nil
	if ptr, err := reg.AllocateComponentMemory(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	}
	if ptr, err := reg.AllocateComponentMemory(e2, velTypeInfo); err == nil {
		*(*velocity)(ptr) = vData
	}

	// Case: Root should have 2 independent outgoing edges
	rootArch := &reg.Archetypes[RootArchetypeId]
	if count := rootArch.CountNextEdges(); count != 2 {
		t.Errorf("expected 2 outgoing edges from Root, got %d", count)
	}

	archPos := rootArch.edgesNext[posTypeInfo.ID]
	archVel := rootArch.edgesNext[velTypeInfo.ID]

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

	reg.AddEntity(e, RootArchetypeId)

	// Validating if Assign works with the new error-returning signature
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	} else {
		t.Fatalf("failed to assign component: %v", err)
	}

	// Case: UnAssigning the only component should trigger EntityRemove
	reg.UnAssign(e, posTypeInfo)

	link, _ := reg.EntityLinkStore.Get(e)

	if link.ArchId != NullArchetypeId {
		linkArch := &reg.Archetypes[link.ArchId]
		t.Errorf("entity should be removed (arch == nil), but still linked to archetype with mask: %v", linkArch.Mask)
	}
}

// 5. Data Idempotency
func TestArchetypeRegistry_OverwriteIdempotency(t *testing.T) {
	reg := setupTestRegistry()
	e := Entity(7)
	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	p1 := position{1, 1}
	p2 := position{2, 2}

	reg.AddEntity(e, RootArchetypeId)
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = p1
	}

	linkBefore, _ := reg.EntityLinkStore.Get(e)

	// Case: Assign same component again (update data)
	if ptr, err := reg.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = p2
	}

	linkAfter, _ := reg.EntityLinkStore.Get(e)

	if linkBefore.ArchId != linkAfter.ArchId || linkBefore.Row != linkAfter.Row {
		t.Error("re-assigning same component should not move entity in graph")
	}

	linkAfterArch := &reg.Archetypes[linkAfter.ArchId]
	gotData := *(*position)(linkAfterArch.GetColumn(posTypeInfo.ID).GetElement(linkAfter.Row))
	if gotData != p2 {
		t.Errorf("data update failed: got %+v, want %+v", gotData, p2)
	}
}

// 6. Structural Integrity (Swap-and-Pop link update)
func TestArchetypeRegistry_SwapPopIntegrity(t *testing.T) {
	reg := setupTestRegistry()

	posTypeInfo := reg.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())

	mask := NewArchetypeMask(posTypeInfo.ID)
	archId := reg.InitArchetype(mask, 4)

	// Dane testowe
	e0, e1, e2 := Entity(10), Entity(11), Entity(12)
	pData := position{x: 1, y: 1}
	pData2 := position{x: 2, y: 2}

	setPos := func(e Entity, p position) {
		reg.AddEntity(e, archId)
		ptr, err := reg.AllocateComponentMemory(e, posTypeInfo)
		if err != nil {
			t.Fatalf("Failed to allocate memory for entity %d: %v", e, err)
		}
		*(*position)(ptr) = p
	}

	setPos(e0, pData)
	setPos(e1, pData)
	setPos(e2, pData2)

	link2_pre, _ := reg.EntityLinkStore.Get(e2)
	if link2_pre.Row != 2 {
		t.Fatalf("Setup error: e2 should be at row 2, got %d", link2_pre.Row)
	}

	reg.UnlinkEntity(e1)

	link2_post, ok := reg.EntityLinkStore.Get(e2)
	if !ok {
		t.Fatal("Entity e2 lost from LinkStore")
	}
	if link2_post.Row != 1 {
		t.Errorf("SwapPop failed: E2 should move to row 1, got %d", link2_post.Row)
	}

	ptr, err := reg.AllocateComponentMemory(e2, posTypeInfo)
	if err != nil {
		t.Fatalf("Failed to access memory for e2: %v", err)
	}

	gotVal := *(*position)(ptr)
	if gotVal != pData2 {
		t.Errorf("Data Integrity Lost: Expected %+v, got %+v. (Did e2 data overwrite e1 data correctly?)", pData2, gotVal)
	}
}

func setupTestRegistry() *ArchetypeRegistry {
	compReg := NewComponentsRegistry()
	compReg.GetOrRegister(reflect.TypeFor[position]())
	compReg.GetOrRegister(reflect.TypeFor[velocity]())

	// Mock ViewRegistry to avoid panics
	viewReg := &ViewRegistry{}

	return NewArchetypeRegistry(&compReg, viewReg, DefaultRegistryConfig())
}

func TestArchetypeRegistry_AssignValidation(t *testing.T) {
	reg := setupTestRegistry()
	e := Entity(1)
	reg.AddEntity(e, RootArchetypeId)
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
