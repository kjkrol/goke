package arch

import (
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
)

type position struct {
	x, y float64
}

type velocity struct {
	vx, vy float64
}

func testMetas() (pos, vel comp.Meta) {
	mi := comp.NewCatalog()
	pos = mi.Register(reflect.TypeFor[position]())
	vel = mi.Register(reflect.TypeFor[velocity]())
	return
}

func TestRegistry_FastPath(t *testing.T) {
	ar := setupTestArchCatalog()
	e1, e2 := uid.UID64(1), uid.UID64(2)
	posTypeInfo, _ := testMetas()
	posData := position{10, 20}

	ar.AddEntity(e1, RootID)
	ar.AddEntity(e2, RootID)

	if ptr, err := ar.UpsertComp(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}

	arch1, _ := ar.EntityIndex.Get(e1)
	if arch1.ArchId == RootID {
		t.Fatal("entity E1 should have moved from rootArch")
	}

	rootArch := &ar.Archetypes[RootID]
	if nextEdge := rootArch.graph.edgesNext[posTypeInfo.ID]; nextEdge == RootID {
		t.Fatal("fast path edge was not cached in rootArch")
	}

	if ptr, err := ar.UpsertComp(e2, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}
	arch2, _ := ar.EntityIndex.Get(e2)

	if arch1.ArchId != arch2.ArchId {
		t.Error("E1 and E2 should share the same archetype instance via graph edges")
	}
}

func TestRegistry_CycleConsistency(t *testing.T) {
	ar := setupTestArchCatalog()
	e := uid.UID64(10)
	posTypeInfo, _ := testMetas()

	ar.AddEntity(e, RootID)

	posData := position{x: 10, y: 20}
	if ptr, err := ar.UpsertComp(e, posTypeInfo); err == nil {
		*(*position)(ptr) = posData
	}
	linkA, _ := ar.EntityIndex.Get(e)

	if linkA.ArchId == NullID || linkA.ArchId == RootID {
		t.Fatal("entity failed to move to a new archetype")
	}

	linkAArch := &ar.Archetypes[linkA.ArchId]
	if linkAArch.graph.edgesPrev[posTypeInfo.ID] != RootID {
		t.Error("bidirectional link (edgesPrev) from ArchA to Root not established")
	}
}

func TestRegistry_GraphBranching(t *testing.T) {
	ar := setupTestArchCatalog()
	e1, e2 := uid.UID64(100), uid.UID64(101)
	posTypeInfo, velTypeInfo := testMetas()

	pData := position{x: 1, y: 1}
	vData := velocity{vx: 10, vy: 10}

	ar.AddEntity(e1, RootID)
	ar.AddEntity(e2, RootID)

	if ptr, err := ar.UpsertComp(e1, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	}
	if ptr, err := ar.UpsertComp(e2, velTypeInfo); err == nil {
		*(*velocity)(ptr) = vData
	}

	rootArch := &ar.Archetypes[RootID]
	if count := rootArch.graph.CountNextEdges(); count != 2 {
		t.Errorf("expected 2 outgoing edges from Root, got %d", count)
	}

	archPos := rootArch.graph.edgesNext[posTypeInfo.ID]
	archVel := rootArch.graph.edgesNext[velTypeInfo.ID]

	if archPos == archVel {
		t.Error("different components must lead to distinct archetypes")
	}
}

func TestRegistry_RemovalStrategy(t *testing.T) {
	ar := setupTestArchCatalog()
	e := uid.UID64(50)
	posTypeInfo, _ := testMetas()
	pData := position{x: 1, y: 1}

	ar.AddEntity(e, RootID)

	if ptr, err := ar.UpsertComp(e, posTypeInfo); err == nil {
		*(*position)(ptr) = pData
	} else {
		t.Fatalf("failed to assign component: %v", err)
	}

	ar.RemoveComp(e, posTypeInfo)

	link, _ := ar.EntityIndex.Get(e)

	if link.ArchId != NullID {
		linkArch := &ar.Archetypes[link.ArchId]
		t.Errorf("entity should be removed (arch == nil), but still linked to archetype with mask: %v", linkArch.Mask())
	}
}

func TestRegistry_OverwriteIdempotency(t *testing.T) {
	ar := setupTestArchCatalog()
	e := uid.UID64(7)
	posTypeInfo, _ := testMetas()
	p1 := position{1, 1}
	p2 := position{2, 2}

	ar.AddEntity(e, RootID)
	if ptr, err := ar.UpsertComp(e, posTypeInfo); err == nil {
		*(*position)(ptr) = p1
	}

	linkBefore, _ := ar.EntityIndex.Get(e)

	if ptr, err := ar.UpsertComp(e, posTypeInfo); err == nil {
		*(*position)(ptr) = p2
	}

	linkAfter, _ := ar.EntityIndex.Get(e)

	if linkBefore.ArchId != linkAfter.ArchId || linkBefore.Pos.ChunkSlot != linkAfter.Pos.ChunkSlot {
		t.Error("re-assigning same component should not move entity in graph")
	}

	linkAfterArch := &ar.Archetypes[linkAfter.ArchId]
	targetPage := linkAfterArch.Table.GetChunk(linkAfter.Pos.ChunkIdx)

	col := linkAfterArch.Table.GetColumn(posTypeInfo.ID)
	ptr := col.At(targetPage, linkAfter.Pos.ChunkSlot)

	gotData := *(*position)(ptr)
	if gotData != p2 {
		t.Errorf("data update failed: got %+v, want %+v", gotData, p2)
	}
}

func TestRegistry_SwapPopIntegrity(t *testing.T) {
	ar := setupTestArchCatalog()
	posTypeInfo, _ := testMetas()

	spec := comp.Composition{}.With(posTypeInfo)
	archId := ar.Upsert(spec)

	e0, e1, e2 := uid.UID64(10), uid.UID64(11), uid.UID64(12)
	pData := position{x: 1, y: 1}
	pData2 := position{x: 2, y: 2}

	setPos := func(e uid.UID64, p position) {
		ar.AddEntity(e, archId)
		ptr, err := ar.UpsertComp(e, posTypeInfo)
		if err != nil {
			t.Fatalf("Failed to allocate memory for entity %d: %v", e, err)
		}
		*(*position)(ptr) = p
	}

	setPos(e0, pData)
	setPos(e1, pData)
	setPos(e2, pData2)

	link2_pre, _ := ar.EntityIndex.Get(e2)
	if link2_pre.Pos.ChunkSlot != 2 {
		t.Fatalf("Setup error: e2 should be at slot 2, got %d", link2_pre.Pos.ChunkSlot)
	}

	ar.UnlinkEntity(e1)

	link2_post, ok := ar.EntityIndex.Get(e2)
	if !ok {
		t.Fatal("Entity e2 lost from LinkStore")
	}
	if link2_post.Pos.ChunkSlot != 1 {
		t.Errorf("SwapPop failed: E2 should move to slot 1, got %d", link2_post.Pos.ChunkSlot)
	}

	ptr, err := ar.UpsertComp(e2, posTypeInfo)
	if err != nil {
		t.Fatalf("Failed to access memory for e2: %v", err)
	}

	gotVal := *(*position)(ptr)
	if gotVal != pData2 {
		t.Errorf("Data Integrity Lost: Expected %+v, got %+v.", pData2, gotVal)
	}
}

func TestRegistry_AssignValidation(t *testing.T) {
	ar := setupTestArchCatalog()
	e := uid.UID64(1)
	ar.AddEntity(e, RootID)

	mi := comp.NewCatalog()
	type void struct{}
	voidTypeInfo := mi.Register(reflect.TypeFor[void]())
	posTypeInfo := mi.Register(reflect.TypeFor[position]())

	if _, err := ar.UpsertComp(e, voidTypeInfo); err != nil {
		t.Errorf("unexpected error when assigning tag component: %v", err)
	}

	eX := uid.UID64(3123)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for unknown entityID, got none")
			}
		}()
		ar.UpsertComp(eX, posTypeInfo) //nolint
	}()

	pos := position{1, 2}
	if ptr, err := ar.UpsertComp(e, posTypeInfo); err == nil {
		*(*position)(ptr) = pos
	} else {
		t.Errorf("unexpected error for valid assign: %v", err)
	}
}

type noopObserver struct{}

func (noopObserver) OnArchetypeCreated(*Archetype) {}

func setupTestArchCatalog() *Catalog {
	r := &Catalog{}
	r.Init(noopObserver{}, 1000)
	return r
}
