package arch

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/chunk"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

type position struct {
	x, y float64
}

type velocity struct {
	vx, vy float64
}

func newCatalog() comp.DefIndex {
	var c comp.DefIndex
	c.Init()
	return c
}

func testMetas() (pos, vel comp.Def) {
	mi := newCatalog()
	pos = mi.Intern(reflect.TypeFor[position]())
	vel = mi.Intern(reflect.TypeFor[velocity]())
	return
}

// testEntityEntry is a test-local entity location record. It mirrors entity.Link
// without importing the entity package, which imports arch and would create a cycle
// in this package's test binary.
type testEntityEntry struct {
	ArchId ID
	Pos    chunk.Pos
	gen    uint32
}

// testEntityIndex is a minimal entity index for use in arch package tests.
type testEntityIndex struct {
	entries []testEntityEntry
}

func newTestEntityIndex(initialCap int) testEntityIndex {
	return testEntityIndex{entries: make([]testEntityEntry, initialCap)}
}

func (t *testEntityIndex) Get(entityID uid.UID64) (testEntityEntry, bool) {
	idx, gen := entityID.Unpack()
	if int(idx) >= len(t.entries) {
		return testEntityEntry{}, false
	}
	e := t.entries[idx]
	if e.ArchId == NullID || e.gen != gen {
		return testEntityEntry{}, false
	}
	return e, true
}

func (t *testEntityIndex) Upsert(entityID uid.UID64, archID ID, pos chunk.Pos) {
	idx, gen := entityID.Unpack()
	for int(idx) >= len(t.entries) {
		t.entries = append(t.entries, testEntityEntry{})
	}
	t.entries[idx] = testEntityEntry{ArchId: archID, Pos: pos, gen: gen}
}

func (t *testEntityIndex) Clear(entityID uid.UID64) {
	idx, gen := entityID.Unpack()
	if int(idx) >= len(t.entries) {
		return
	}
	if t.entries[idx].gen == gen {
		t.entries[idx] = testEntityEntry{ArchId: NullID}
	}
}

// testEnv wraps Catalog with a local entity tracker to mirror what reg.Registry does.
type testEnv struct {
	catalog     *Catalog
	entityIndex testEntityIndex
}

func newTestEnv() *testEnv {
	catalog := &Catalog{}
	catalog.Init(func(*Archetype) {})
	return &testEnv{
		catalog:     catalog,
		entityIndex: newTestEntityIndex(1000),
	}
}

func (env *testEnv) addEntity(entityID uid.UID64, archID ID) {
	table := &env.catalog.Archetypes[archID].Table
	table.SetIDSeeder(func(dst []uid.UID64, _ chunk.Pos) { dst[0] = entityID })
	idx, _, _ := table.ReserveSlots(1)
	var cur iter.Cursor
	_, pos := table.SpawnCursor(&cur, idx, 1, nil)
	table.ReleaseSlots()
	env.entityIndex.Upsert(entityID, archID, pos)
}

func (env *testEnv) upsertComp(entityID uid.UID64, compDef comp.Def) (unsafe.Pointer, bool) {
	link, ok := env.entityIndex.Get(entityID)
	if !ok {
		panic("upsertComp: entity not in EntityIndex")
	}

	targetArchID := link.ArchId
	targetPos := link.Pos

	if !env.catalog.Archetypes[link.ArchId].Mask().IsSet(compDef.ID) {
		targetArchID = env.catalog.EnsureEdgeNext(compDef, link.ArchId)
		newPos, swappedEntity, swapped := env.catalog.MigrateEntity(entityID, link.ArchId, link.Pos, targetArchID)
		if swapped {
			env.entityIndex.Upsert(swappedEntity, link.ArchId, link.Pos)
		}
		env.entityIndex.Upsert(entityID, targetArchID, newPos)
		targetPos = newPos
	}

	if compDef.Size == 0 {
		return nil, true
	}

	return env.catalog.Archetypes[targetArchID].Table.ComponentAt(targetPos, compDef.ID), true
}

func (env *testEnv) removeComp(entityID uid.UID64, compDef comp.Def) {
	link, ok := env.entityIndex.Get(entityID)
	if !ok {
		return
	}
	if !env.catalog.Archetypes[link.ArchId].Mask().IsSet(compDef.ID) {
		return
	}

	targetArchID, shouldUnlink := env.catalog.EnsureEdgePrev(compDef, link.ArchId)
	if shouldUnlink {
		swappedEntity, swapped := env.catalog.RemoveEntity(link.ArchId, link.Pos)
		if swapped {
			env.entityIndex.Upsert(swappedEntity, link.ArchId, link.Pos)
		}
		env.entityIndex.Clear(entityID)
		return
	}

	newPos, swappedEntity, swapped := env.catalog.MigrateEntity(entityID, link.ArchId, link.Pos, targetArchID)
	if swapped {
		env.entityIndex.Upsert(swappedEntity, link.ArchId, link.Pos)
	}
	env.entityIndex.Upsert(entityID, targetArchID, newPos)
}

func (env *testEnv) unlinkEntity(entityID uid.UID64) {
	link, ok := env.entityIndex.Get(entityID)
	if !ok {
		return
	}
	swappedEntity, swapped := env.catalog.RemoveEntity(link.ArchId, link.Pos)
	if swapped {
		env.entityIndex.Upsert(swappedEntity, link.ArchId, link.Pos)
	}
	env.entityIndex.Clear(entityID)
}

func TestRegistry_FastPath(t *testing.T) {
	env := newTestEnv()
	e1, e2 := uid.UID64(1), uid.UID64(2)
	posTypeInfo, _ := testMetas()
	posData := position{10, 20}

	env.addEntity(e1, RootID)
	env.addEntity(e2, RootID)

	if ptr, ok := env.upsertComp(e1, posTypeInfo); ok {
		*(*position)(ptr) = posData
	}

	arch1, _ := env.entityIndex.Get(e1)
	if arch1.ArchId == RootID {
		t.Fatal("entity E1 should have moved from rootArch")
	}

	rootArch := &env.catalog.Archetypes[RootID]
	if nextEdge := rootArch.graph.edgesNext[posTypeInfo.ID]; nextEdge == RootID {
		t.Fatal("fast path edge was not cached in rootArch")
	}

	if ptr, ok := env.upsertComp(e2, posTypeInfo); ok {
		*(*position)(ptr) = posData
	}
	arch2, _ := env.entityIndex.Get(e2)

	if arch1.ArchId != arch2.ArchId {
		t.Error("E1 and E2 should share the same archetype instance via graph edges")
	}
}

func TestRegistry_CycleConsistency(t *testing.T) {
	env := newTestEnv()
	e := uid.UID64(10)
	posTypeInfo, _ := testMetas()

	env.addEntity(e, RootID)

	posData := position{x: 10, y: 20}
	if ptr, ok := env.upsertComp(e, posTypeInfo); ok {
		*(*position)(ptr) = posData
	}
	linkA, _ := env.entityIndex.Get(e)

	if linkA.ArchId == NullID || linkA.ArchId == RootID {
		t.Fatal("entity failed to move to a new archetype")
	}

	linkAArch := &env.catalog.Archetypes[linkA.ArchId]
	if linkAArch.graph.edgesPrev[posTypeInfo.ID] != RootID {
		t.Error("bidirectional link (edgesPrev) from ArchA to Root not established")
	}
}

func TestRegistry_GraphBranching(t *testing.T) {
	env := newTestEnv()
	e1, e2 := uid.UID64(100), uid.UID64(101)
	posTypeInfo, velTypeInfo := testMetas()

	env.addEntity(e1, RootID)
	env.addEntity(e2, RootID)

	if ptr, ok := env.upsertComp(e1, posTypeInfo); ok {
		*(*position)(ptr) = position{x: 1, y: 1}
	}
	if ptr, ok := env.upsertComp(e2, velTypeInfo); ok {
		*(*velocity)(ptr) = velocity{vx: 10, vy: 10}
	}

	rootArch := &env.catalog.Archetypes[RootID]
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
	env := newTestEnv()
	e := uid.UID64(50)
	posTypeInfo, _ := testMetas()

	env.addEntity(e, RootID)

	if ptr, ok := env.upsertComp(e, posTypeInfo); ok {
		*(*position)(ptr) = position{x: 1, y: 1}
	} else {
		t.Fatal("failed to assign component")
	}

	env.removeComp(e, posTypeInfo)

	link, _ := env.entityIndex.Get(e)
	if link.ArchId != NullID {
		linkArch := &env.catalog.Archetypes[link.ArchId]
		t.Errorf("entity should be removed (arch == nil), but still linked to archetype with mask: %v", linkArch.Mask())
	}
}

func TestRegistry_OverwriteIdempotency(t *testing.T) {
	env := newTestEnv()
	e := uid.UID64(7)
	posTypeInfo, _ := testMetas()

	env.addEntity(e, RootID)
	if ptr, ok := env.upsertComp(e, posTypeInfo); ok {
		*(*position)(ptr) = position{1, 1}
	}

	linkBefore, _ := env.entityIndex.Get(e)

	if ptr, ok := env.upsertComp(e, posTypeInfo); ok {
		*(*position)(ptr) = position{2, 2}
	}

	linkAfter, _ := env.entityIndex.Get(e)

	if linkBefore.ArchId != linkAfter.ArchId || linkBefore.Pos.Slot != linkAfter.Pos.Slot {
		t.Error("re-assigning same component should not move entity in graph")
	}

	gotData := *(*position)(env.catalog.Archetypes[linkAfter.ArchId].Table.ComponentAt(linkAfter.Pos, posTypeInfo.ID))
	if gotData != (position{2, 2}) {
		t.Errorf("data update failed: got %+v, want {2 2}", gotData)
	}
}

func TestRegistry_SwapPopIntegrity(t *testing.T) {
	env := newTestEnv()
	posTypeInfo, _ := testMetas()

	spec := comp.Composition{}.With(posTypeInfo)
	archID := env.catalog.Upsert(spec)

	e0, e1, e2 := uid.UID64(10), uid.UID64(11), uid.UID64(12)

	setPos := func(entityID uid.UID64, p position) {
		env.addEntity(entityID, archID)
		ptr, ok := env.upsertComp(entityID, posTypeInfo)
		if !ok {
			t.Fatalf("failed to allocate memory for entity %d", entityID)
		}
		*(*position)(ptr) = p
	}

	setPos(e0, position{x: 1, y: 1})
	setPos(e1, position{x: 1, y: 1})
	setPos(e2, position{x: 2, y: 2})

	link2Pre, _ := env.entityIndex.Get(e2)
	if link2Pre.Pos.Slot != 2 {
		t.Fatalf("setup error: e2 should be at slot 2, got %d", link2Pre.Pos.Slot)
	}

	env.unlinkEntity(e1)

	link2Post, ok := env.entityIndex.Get(e2)
	if !ok {
		t.Fatal("entity e2 lost from EntityIndex")
	}
	if link2Post.Pos.Slot != 1 {
		t.Errorf("swap-pop failed: e2 should move to slot 1, got %d", link2Post.Pos.Slot)
	}

	ptr, ok := env.upsertComp(e2, posTypeInfo)
	if !ok {
		t.Fatal("failed to access memory for e2")
	}
	if gotVal := *(*position)(ptr); gotVal != (position{x: 2, y: 2}) {
		t.Errorf("data integrity lost: got %+v, want {2 2}", gotVal)
	}
}

func TestRegistry_AssignValidation(t *testing.T) {
	env := newTestEnv()
	e := uid.UID64(1)
	env.addEntity(e, RootID)

	mi := newCatalog()
	type void struct{}
	voidTypeInfo := mi.Intern(reflect.TypeFor[void]())
	posTypeInfo := mi.Intern(reflect.TypeFor[position]())

	if _, ok := env.upsertComp(e, voidTypeInfo); !ok {
		t.Error("unexpected failure when assigning tag component")
	}

	eX := uid.UID64(3123)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for unknown entityID, got none")
			}
		}()
		env.upsertComp(eX, posTypeInfo)
	}()

	if ptr, ok := env.upsertComp(e, posTypeInfo); ok {
		*(*position)(ptr) = position{1, 2}
	} else {
		t.Error("unexpected failure for valid assign")
	}
}

func setupTestArchCatalog() *Catalog {
	catalog := &Catalog{}
	catalog.Init(func(*Archetype) {})
	return catalog
}
