package arch

import (
	"fmt"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/colstore"
	"github.com/kjkrol/goke/internal/comp"
)

type Catalog struct {
	maskIndex          MaskIndex
	Archetypes         [MaxID]Archetype
	lastArchetypeId    ID
	onArchetypeCreated func(*Archetype)
}

func (r *Catalog) Init(onArchetypeCreated func(*Archetype)) {
	r.onArchetypeCreated = onArchetypeCreated
	r.lastArchetypeId = RootID
	r.addArchetype(comp.Composition{})
}

func (r *Catalog) Len() ID {
	return r.lastArchetypeId
}

func (r *Catalog) Upsert(set comp.Composition) ID {
	if found, ok := r.maskIndex.Get(set.Mask); ok {
		return found
	}
	archID := r.addArchetype(set)
	r.onArchetypeCreated(&r.Archetypes[archID])
	return archID
}

// RemoveEntity swap-removes the entity at (archID, pos) from its archetype's table.
// Returns the entity displaced by the swap and whether a swap occurred.
// The displaced entity moves to pos — the caller must update the entity index accordingly.
func (r *Catalog) RemoveEntity(archID ID, pos colstore.Pos) (uid.UID64, bool) {
	return r.Archetypes[archID].Table.RemoveAt(pos)
}

// MigrateEntity moves entityID from (srcArchID, srcPos) to dstArchID.
// Returns the new storage position, the entity displaced by swap-remove in the source table,
// and whether a swap occurred. The displaced entity moves to srcPos.
func (r *Catalog) MigrateEntity(entityID uid.UID64, srcArchID ID, srcPos colstore.Pos, dstArchID ID) (colstore.Pos, uid.UID64, bool) {
	return r.Archetypes[dstArchID].Table.MoveEntityFrom(&r.Archetypes[srcArchID].Table, entityID, srcPos)
}

// EnsureEdgeNext returns the archetype ID reached by adding compDef to archID.
// Creates the archetype and caches the graph edge if not already present.
func (r *Catalog) EnsureEdgeNext(compDef comp.Def, archID ID) ID {
	archetype := &r.Archetypes[archID]
	if nextEdgeID := archetype.graph.edgesNext[compDef.ID]; nextEdgeID != NullID {
		return nextEdgeID
	}
	nextArchID := r.Upsert(archetype.set.With(compDef))
	nextArch := &r.Archetypes[nextArchID]
	archetype.graph.linkNext(nextArch.graph, archetype.Id, nextArch.Id, compDef.ID)
	return nextArchID
}

// EnsureEdgePrev returns the archetype ID reached by removing compDef from archID.
// shouldUnlink is true when the resulting composition has an empty mask — the caller
// should remove the entity entirely rather than migrate it.
func (r *Catalog) EnsureEdgePrev(compDef comp.Def, archID ID) (targetID ID, shouldUnlink bool) {
	archetype := &r.Archetypes[archID]
	compID := compDef.ID

	if prevArchID := archetype.graph.edgesPrev[compID]; prevArchID != NullID {
		return prevArchID, r.Archetypes[prevArchID].set.Mask.IsEmpty()
	}

	newSet := archetype.set.Without(compID)
	if newSet.Mask.IsEmpty() {
		return NullID, true
	}

	prevArchID := r.Upsert(newSet)
	prevArch := &r.Archetypes[prevArchID]
	archetype.graph.linkPrev(prevArch.graph, archetype.Id, prevArch.Id, compID)
	return prevArchID, false
}

func (r *Catalog) Reset() {
	for i := int(RootID); i < int(r.lastArchetypeId); i++ {
		r.Archetypes[i].Reset()
	}
	clear(r.Archetypes[:])
	r.maskIndex.Reset()
	r.lastArchetypeId = RootID
	r.addArchetype(comp.Composition{})
}

func (r *Catalog) addArchetype(set comp.Composition) ID {
	if r.lastArchetypeId >= MaxID {
		panic(fmt.Sprintf("Max archetype number exceeded: %d", MaxID))
	}
	archID := r.lastArchetypeId
	r.Archetypes[archID].Init(archID, set)
	r.maskIndex.Upsert(set.Mask, archID)
	r.lastArchetypeId++
	return archID
}
