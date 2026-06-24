package arch

import (
	"reflect"
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/chunk"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/iter"
)

type position struct {
	x, y float64
}

type velocity struct {
	vx, vy float64
}

func newDefIndex() comp.DefIndex {
	var c comp.DefIndex
	c.Init()
	return c
}

func testMetas() (pos, vel comp.Def) {
	mi := newDefIndex()
	pos = mi.Intern(reflect.TypeFor[position]())
	vel = mi.Intern(reflect.TypeFor[velocity]())
	return
}

// spawnEntity registers id as the next entity seeded into archetype's table
// and returns its storage position. Tests track positions themselves —
// Catalog has no entity index of its own (that's addr.Index's job, one
// layer up).
func spawnEntity(t *testing.T, archetype *Archetype, id uid.UID64) chunk.Pos {
	t.Helper()
	archetype.Table.SetIDSeeder(func(dst []uid.UID64, _ chunk.Pos) { dst[0] = id })
	idx, _, _ := archetype.Table.ReserveSlots(1)
	var cur iter.Cursor
	_, pos := archetype.Table.SpawnCursor(&cur, idx, 1, nil)
	archetype.Table.ReleaseSlots()
	return pos
}

func newTestCatalog() *Catalog {
	cat := &Catalog{}
	cat.Init(func(*Archetype) {})
	return cat
}

func TestCatalog_Init(t *testing.T) {
	cat := newTestCatalog()

	// Len() reports the next free archetype ID, i.e. one past the root
	// archetype that Init creates.
	if cat.Len() != RootID+1 {
		t.Errorf("expected Len() == RootID+1 after Init, got %d", cat.Len())
	}
	if !cat.Archetypes[RootID].Mask().IsEmpty() {
		t.Error("expected the root archetype to have an empty mask")
	}
}

func TestCatalog_Upsert_CachesByMask(t *testing.T) {
	cat := newTestCatalog()
	posDef, _ := testMetas()
	composition := comp.Composition{}.With(posDef)

	id1 := cat.Upsert(composition)
	id2 := cat.Upsert(composition)

	if id1 != id2 {
		t.Errorf("expected Upsert to return the same archetype for an identical mask, got %d and %d", id1, id2)
	}
	if id1 == RootID {
		t.Error("expected a new archetype distinct from root")
	}
}

func TestCatalog_Upsert_InvokesCallbackOnlyForNewArchetypes(t *testing.T) {
	created := 0
	cat := &Catalog{}
	cat.Init(func(*Archetype) { created++ })
	posDef, _ := testMetas()
	composition := comp.Composition{}.With(posDef)

	cat.Upsert(composition)
	cat.Upsert(composition) // same mask — must not re-trigger the callback

	if created != 1 {
		t.Errorf("expected onArchetypeCreated to fire exactly once, got %d", created)
	}
}

func TestCatalog_EnsureEdgeNext_CachesFastPath(t *testing.T) {
	cat := newTestCatalog()
	posDef, _ := testMetas()

	target1 := cat.EnsureEdgeNext(posDef, RootID)
	target2 := cat.EnsureEdgeNext(posDef, RootID)

	if target1 != target2 {
		t.Errorf("expected EnsureEdgeNext to return the cached edge, got %d then %d", target1, target2)
	}
	if cat.Archetypes[RootID].graph.edgesNext[posDef.ID] != target1 {
		t.Error("expected the edge to be cached on the source archetype's graph")
	}
}

func TestCatalog_EnsureEdgeNext_EstablishesBidirectionalLink(t *testing.T) {
	cat := newTestCatalog()
	posDef, _ := testMetas()

	target := cat.EnsureEdgeNext(posDef, RootID)

	if cat.Archetypes[target].graph.edgesPrev[posDef.ID] != RootID {
		t.Error("expected a bidirectional edgesPrev link back to Root")
	}
}

func TestCatalog_EnsureEdgeNext_DifferentComponentsBranch(t *testing.T) {
	cat := newTestCatalog()
	posDef, velDef := testMetas()

	posTarget := cat.EnsureEdgeNext(posDef, RootID)
	velTarget := cat.EnsureEdgeNext(velDef, RootID)

	if posTarget == velTarget {
		t.Error("expected different components to lead to distinct archetypes")
	}
	if count := cat.Archetypes[RootID].graph.CountNextEdges(); count != 2 {
		t.Errorf("expected 2 outgoing edges from Root, got %d", count)
	}
}

// EnsureEdgeNext has no concept of "the caller should check the mask
// first" — callers (ent.Manager, ent.Editor) do guard against this, but
// Catalog itself stays correct even without that guard: asking to add a
// component the archetype already has resolves back to the same archetype.
func TestCatalog_EnsureEdgeNext_RedundantAddIsSelfLoop(t *testing.T) {
	cat := newTestCatalog()
	posDef, _ := testMetas()

	withPos := cat.EnsureEdgeNext(posDef, RootID)
	again := cat.EnsureEdgeNext(posDef, withPos)

	if again != withPos {
		t.Errorf("expected redundant EnsureEdgeNext to self-loop to %d, got %d", withPos, again)
	}
}

func TestCatalog_EnsureEdgePrev_CachesFastPath(t *testing.T) {
	cat := newTestCatalog()
	posDef, velDef := testMetas()
	// Built via Upsert directly, so no graph edges exist yet — the first
	// EnsureEdgePrev call below is a genuine cache miss exercising linkPrev,
	// not a fast path inherited from EnsureEdgeNext's own linkNext side effect.
	withPosVel := cat.Upsert(comp.Composition{}.With(posDef).With(velDef))
	withPos := cat.Upsert(comp.Composition{}.With(posDef))

	target1, unlink1 := cat.EnsureEdgePrev(velDef, withPosVel)
	target2, unlink2 := cat.EnsureEdgePrev(velDef, withPosVel)

	if target1 != target2 || unlink1 != unlink2 {
		t.Errorf("expected cached EnsureEdgePrev result, got (%d,%v) then (%d,%v)", target1, unlink1, target2, unlink2)
	}
	if target1 != withPos || unlink1 {
		t.Errorf("expected removing Velocity to land back at the Position-only archetype, got archID=%d unlink=%v", target1, unlink1)
	}
}

func TestCatalog_EnsureEdgePrev_UnlinksWhenMaskBecomesEmpty(t *testing.T) {
	cat := newTestCatalog()
	posDef, _ := testMetas()
	// Built via Upsert directly (not EnsureEdgeNext), so no edgesPrev cache
	// entry exists yet — this exercises the genuine cache-miss path.
	archID := cat.Upsert(comp.Composition{}.With(posDef))

	target, shouldUnlink := cat.EnsureEdgePrev(posDef, archID)

	if !shouldUnlink {
		t.Error("expected removing the only component to signal unlink")
	}
	if target != NullID {
		t.Errorf("expected NullID target on unlink, got %d", target)
	}
}

func TestCatalog_RemoveEntity_SwapPop(t *testing.T) {
	cat := newTestCatalog()
	posDef, _ := testMetas()
	archID := cat.Upsert(comp.Composition{}.With(posDef))
	archetype := &cat.Archetypes[archID]

	e0, e1, e2 := uid.UID64(10), uid.UID64(11), uid.UID64(12)
	spawnEntity(t, archetype, e0)
	pos1 := spawnEntity(t, archetype, e1)
	pos2 := spawnEntity(t, archetype, e2)

	if pos1.Slot != 1 || pos2.Slot != 2 {
		t.Fatalf("setup error: expected e1@1 e2@2, got e1@%d e2@%d", pos1.Slot, pos2.Slot)
	}

	// Remove e1 (not the last slot) — e2 must swap into its slot.
	swappedEntity, swapped := cat.RemoveEntity(archID, pos1)

	if !swapped {
		t.Fatal("expected a swap since the removed slot wasn't the last one")
	}
	if swappedEntity != e2 {
		t.Errorf("expected the swapped entity to be e2, got %v", swappedEntity)
	}
}

func TestCatalog_RemoveEntity_LastSlotNoSwap(t *testing.T) {
	cat := newTestCatalog()
	posDef, _ := testMetas()
	archID := cat.Upsert(comp.Composition{}.With(posDef))
	archetype := &cat.Archetypes[archID]

	e0 := uid.UID64(1)
	pos0 := spawnEntity(t, archetype, e0)

	_, swapped := cat.RemoveEntity(archID, pos0)
	if swapped {
		t.Error("expected no swap when removing the only/last entity")
	}
}

func TestCatalog_MigrateEntity_MovesDataAndSwapsSource(t *testing.T) {
	cat := newTestCatalog()
	posDef, velDef := testMetas()
	srcArchID := cat.Upsert(comp.Composition{}.With(posDef))
	dstArchID := cat.Upsert(comp.Composition{}.With(posDef).With(velDef))

	srcArch := &cat.Archetypes[srcArchID]
	e0, e1 := uid.UID64(1), uid.UID64(2)
	pos0 := spawnEntity(t, srcArch, e0)
	spawnEntity(t, srcArch, e1)

	*(*position)(srcArch.Table.ComponentAt(pos0, posDef.ID)) = position{x: 9, y: 9}

	newPos, swappedEntity, swapped := cat.MigrateEntity(e0, srcArchID, pos0, dstArchID)

	if !swapped {
		t.Fatal("expected e1 to swap into e0's old slot in the source table")
	}
	if swappedEntity != e1 {
		t.Errorf("expected swapped entity e1, got %v", swappedEntity)
	}

	dstArch := &cat.Archetypes[dstArchID]
	got := *(*position)(dstArch.Table.ComponentAt(newPos, posDef.ID))
	if got != (position{x: 9, y: 9}) {
		t.Errorf("expected migrated Position data to survive, got %+v", got)
	}
}

func TestCatalog_Reset(t *testing.T) {
	cat := newTestCatalog()
	posDef, _ := testMetas()
	cat.Upsert(comp.Composition{}.With(posDef))

	cat.Reset()

	if cat.Len() != RootID+1 {
		t.Errorf("expected Len() == RootID+1 after Reset, got %d", cat.Len())
	}
	if !cat.Archetypes[RootID].Mask().IsEmpty() {
		t.Error("expected a fresh root archetype with an empty mask after Reset")
	}
}

// addArchetype must panic rather than silently wrap around once the
// archetype table is exhausted — this is a real, reachable limit (MaxID
// distinct component combinations), not a defensive check for an
// impossible state.
func TestCatalog_AddArchetype_PanicsWhenMaxIDExceeded(t *testing.T) {
	cat := newTestCatalog()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected a panic when exceeding MaxID archetypes")
		}
	}()

	// Synthetic, directly-constructed masks — no real component
	// registration needed, since MaskIndex only cares about the Mask value.
	for i := uint64(1); i <= uint64(MaxID)+1; i++ {
		cat.Upsert(comp.Composition{Mask: comp.Mask{i, 0}})
	}
}
