package ent

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/colstore"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/goke/v2/iter"
)

// Editor applies a fixed set of structural changes — adding and removing
// components — to an entity in a single archetype migration, then positions
// Cursor so the added components' values can be written via ArrayRef.At.
//
// Migration cost scales with the width of the source and destination
// archetypes, not with how many components the edit changes: vacating the
// source row touches every column the source tracks, and the copy into the
// destination row touches every column already shared with the source.
// Removing a few components from a wide archetype therefore costs about as
// much as removing many; adding components onto a narrow one stays cheap
// regardless of how many are added.
type Editor struct {
	Cursor   iter.Cursor
	manager  *Manager
	addDefs  []comp.Def
	delDefs  []comp.Def
	addIDs   []comp.ID
	offsets  [arch.MaxID][]uintptr
	table    *colstore.Table
	lastArch arch.ID
}

// CreateEditor builds an Editor from spec.
func (m *Manager) CreateEditor(spec comp.EditSpec) *Editor {
	e := &Editor{
		manager:  m,
		addDefs:  spec.AddDefs,
		delDefs:  spec.DelDefs,
		addIDs:   make([]comp.ID, len(spec.AddDefs)),
		lastArch: arch.NullID,
	}
	for i, d := range spec.AddDefs {
		e.addIDs[i] = d.ID
	}
	return e
}

// Update migrates entityID to the archetype composed from its current
// components plus the added and minus the removed ones — in a single move — then
// positions Cursor so the added components' values can be written via ArrayRef.At.
// Returns false if the entity does not exist. If the edit would leave the entity
// with no components, it is removed entirely.
func (e *Editor) Update(entityID uid.UID64) bool {
	entry, ok := e.manager.AddressBook.Get(entityID)
	if !ok {
		return false
	}

	targetArchID, unlink := e.resolveTarget(entry.ArchId)
	if unlink {
		e.manager.removeFromArchetype(entityID, entry.ArchId, entry.Pos)
		return true
	}

	targetPos := entry.Pos
	if targetArchID != entry.ArchId {
		targetPos = e.manager.migrateEntity(entityID, entry.ArchId, entry.Pos, targetArchID)
	}

	// Position the cursor only when there are added columns to write into.
	// A remove-only Editor has nothing to write, so it skips this entirely.
	if len(e.addIDs) > 0 {
		if targetArchID != e.lastArch {
			e.table = &e.manager.ArchCatalog.Archetypes[targetArchID].Table
			offs := e.offsets[targetArchID]
			if offs == nil {
				offs = e.table.BakeOffsets(e.addIDs)
				e.offsets[targetArchID] = offs
			}
			e.Cursor.Offsets = offs
			e.lastArch = targetArchID
		}
		e.table.PointCursor(&e.Cursor, targetPos)
	}
	return true
}

// resolveTarget walks the archetype graph from srcArchID, applying the adds then
// the dels, and returns the final archetype. unlink is true when the result has
// no components.
func (e *Editor) resolveTarget(srcArchID arch.ID) (target arch.ID, unlink bool) {
	cat := &e.manager.ArchCatalog
	target = srcArchID
	for i := range e.addDefs {
		d := e.addDefs[i]
		if !cat.Archetypes[target].Mask().IsSet(d.ID) {
			target = cat.EnsureEdgeNext(d, target)
		}
	}
	for i := range e.delDefs {
		d := e.delDefs[i]
		if cat.Archetypes[target].Mask().IsSet(d.ID) {
			next, shouldUnlink := cat.EnsureEdgePrev(d, target)
			if shouldUnlink {
				return arch.NullID, true
			}
			target = next
		}
	}
	return target, false
}
