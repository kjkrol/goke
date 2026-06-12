package arch

import (
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/core"
)

type position struct {
	x, y float64
}

type velocity struct {
	vx, vy float64
}

func TestArchetypeRegistry_FastPath(t *testing.T) {
	ar := setupTestArchRegistry()
	e1, e2 := uid.UID64(1), uid.UID64(2)
	posTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	posData := position{10, 20}

	ar.AddEntity(e1, core.RootArchetypeId)
	ar.AddEntity(e2, core.RootArchetypeId)

	if ptr, err := ar.AllocateComponentMemory(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}

	arch1, _ := ar.EntityLinkStore.Get(e1)
	if arch1.ArchId == core.RootArchetypeId {
		t.Fatal("entity E1 should have moved from rootArch")
	}

	rootArch := &ar.Archetypes[core.RootArchetypeId]
	if nextEdge := rootArch.graph.edgesNext[posTypeInfo.ID]; nextEdge == core.RootArchetypeId {
		t.Fatal("fast path edge was not cached in rootArch")
	}

	if ptr, err := ar.AllocateComponentMemory(e2, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}
	arch2, _ := ar.EntityLinkStore.Get(e2)

	if arch1.ArchId != arch2.ArchId {
		t.Error("E1 and E2 should share the same archetype instance via graph edges")
	}
}

func TestArchetypeRegistry_CycleConsistency(t *testing.T) {
	ar := setupTestArchRegistry()
	e := uid.UID64(10)
	posTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())

	ar.AddEntity(e, core.RootArchetypeId)

	posData := position{x: 10, y: 20}
	if ptr, err := ar.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}
	linkA, _ := ar.EntityLinkStore.Get(e)

	if linkA.ArchId == core.NullArchetypeId || linkA.ArchId == core.RootArchetypeId {
		t.Fatal("entity failed to move to a new archetype")
	}

	linkAArch := &ar.Archetypes[linkA.ArchId]
	if linkAArch.graph.edgesPrev[posTypeInfo.ID] != core.RootArchetypeId {
		t.Error("bidirectional link (edgesPrev) from ArchA to Root not established")
	}
}

func TestArchetypeRegistry_GraphBranching(t *testing.T) {
	ar := setupTestArchRegistry()
	e1, e2 := uid.UID64(100), uid.UID64(101)
	posTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	velTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[velocity]())

	pData := position{x: 1, y: 1}
	vData := velocity{vx: 10, vy: 10}

	ar.AddEntity(e1, core.RootArchetypeId)
	ar.AddEntity(e2, core.RootArchetypeId)

	if ptr, err := ar.AllocateComponentMemory(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	}
	if ptr, err := ar.AllocateComponentMemory(e2, velTypeInfo); err == nil {
		*(*velocity)(ptr) = vData
	}

	rootArch := &ar.Archetypes[core.RootArchetypeId]
	if count := rootArch.graph.CountNextEdges(); count != 2 {
		t.Errorf("expected 2 outgoing edges from Root, got %d", count)
	}

	archPos := rootArch.graph.edgesNext[posTypeInfo.ID]
	archVel := rootArch.graph.edgesNext[velTypeInfo.ID]

	if archPos == archVel {
		t.Error("different components must lead to distinct archetypes")
	}
}

func TestArchetypeRegistry_RemovalStrategy(t *testing.T) {
	ar := setupTestArchRegistry()
	e := uid.UID64(50)
	posTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	pData := position{x: 1, y: 1}

	ar.AddEntity(e, core.RootArchetypeId)

	if ptr, err := ar.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	} else {
		t.Fatalf("failed to assign component: %v", err)
	}

	ar.UnAssign(e, posTypeInfo)

	link, _ := ar.EntityLinkStore.Get(e)

	if link.ArchId != core.NullArchetypeId {
		linkArch := &ar.Archetypes[link.ArchId]
		t.Errorf("entity should be removed (arch == nil), but still linked to archetype with mask: %v", linkArch.Mask)
	}
}

func TestArchetypeRegistry_OverwriteIdempotency(t *testing.T) {
	ar := setupTestArchRegistry()
	e := uid.UID64(7)
	posTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())
	p1 := position{1, 1}
	p2 := position{2, 2}

	ar.AddEntity(e, core.RootArchetypeId)
	if ptr, err := ar.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = p1
	}

	linkBefore, _ := ar.EntityLinkStore.Get(e)

	if ptr, err := ar.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = p2
	}

	linkAfter, _ := ar.EntityLinkStore.Get(e)

	if linkBefore.ArchId != linkAfter.ArchId || linkBefore.PageSlot != linkAfter.PageSlot {
		t.Error("re-assigning same component should not move entity in graph")
	}

	linkAfterArch := &ar.Archetypes[linkAfter.ArchId]
	targetPage := &linkAfterArch.Memory.Pages[linkAfter.PageIdx]

	col := linkAfterArch.GetColumn(posTypeInfo.ID)
	ptr := col.GetPointer(targetPage, linkAfter.PageSlot)

	gotData := *(*position)(ptr)
	if gotData != p2 {
		t.Errorf("data update failed: got %+v, want %+v", gotData, p2)
	}
}

func TestArchetypeRegistry_SwapPopIntegrity(t *testing.T) {
	ar := setupTestArchRegistry()

	posTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())

	mask := core.NewArchetypeMask(posTypeInfo.ID)
	archId := ar.InitArchetype(mask)

	e0, e1, e2 := uid.UID64(10), uid.UID64(11), uid.UID64(12)
	pData := position{x: 1, y: 1}
	pData2 := position{x: 2, y: 2}

	setPos := func(e uid.UID64, p position) {
		ar.AddEntity(e, archId)
		ptr, err := ar.AllocateComponentMemory(e, posTypeInfo)
		if err != nil {
			t.Fatalf("Failed to allocate memory for entity %d: %v", e, err)
		}
		*(*position)(ptr) = p
	}

	setPos(e0, pData)
	setPos(e1, pData)
	setPos(e2, pData2)

	link2_pre, _ := ar.EntityLinkStore.Get(e2)
	if link2_pre.PageSlot != 2 {
		t.Fatalf("Setup error: e2 should be at slot 2, got %d", link2_pre.PageSlot)
	}

	ar.UnlinkEntity(e1)

	link2_post, ok := ar.EntityLinkStore.Get(e2)
	if !ok {
		t.Fatal("Entity e2 lost from LinkStore")
	}
	if link2_post.PageSlot != 1 {
		t.Errorf("SwapPop failed: E2 should move to slot 1, got %d", link2_post.PageSlot)
	}

	ptr, err := ar.AllocateComponentMemory(e2, posTypeInfo)
	if err != nil {
		t.Fatalf("Failed to access memory for e2: %v", err)
	}

	gotVal := *(*position)(ptr)
	if gotVal != pData2 {
		t.Errorf("Data Integrity Lost: Expected %+v, got %+v.", pData2, gotVal)
	}
}

func TestArchetypeRegistry_AssignValidation(t *testing.T) {
	ar := setupTestArchRegistry()
	e := uid.UID64(1)
	ar.AddEntity(e, core.RootArchetypeId)
	type void struct{}
	voidTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[void]())
	posTypeInfo := ar.componentsRegistry.GetOrRegister(reflect.TypeFor[position]())

	if _, err := ar.AllocateComponentMemory(e, voidTypeInfo); err != nil {
		t.Errorf("unexpected error when assigning tag component: %v", err)
	}

	eX := uid.UID64(3123)
	if _, err := ar.AllocateComponentMemory(eX, posTypeInfo); err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}

	pos := position{1, 2}
	if ptr, err := ar.AllocateComponentMemory(e, posTypeInfo); err == nil {
		*(*position)(ptr) = pos
	} else {
		t.Errorf("unexpected error for valid assign: %v", err)
	}
}


type noopObserver struct{}

func (noopObserver) OnArchetypeCreated(*Archetype) {}

func setupTestArchRegistry() *ArchetypeRegistry {
	compReg := core.NewComponentsRegistry()
	compReg.GetOrRegister(reflect.TypeFor[position]())
	compReg.GetOrRegister(reflect.TypeFor[velocity]())

	return NewArchetypeRegistry(&compReg, noopObserver{}, 1000)
}
