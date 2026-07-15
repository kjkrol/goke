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
// Migration cost scales with the width of the source and destination
// archetypes, not with how many components the edit changes.
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

// Update applies the edit to entityID in a single migration and positions
// Cursor over the added columns. Returns false if the entity does not exist;
// an edit that leaves no components removes the entity entirely.
func (e *Editor) Update(entityID uid.UID64) bool {
	entry, ok := e.manager.AddressBook.Get(entityID)
	if !ok {
		return false
	}

	targetArchID, unlink := e.resolveTarget(entry.ArchID)
	if unlink {
		e.manager.removeFromArchetype(entityID, entry.ArchID, entry.ChunkPtr, entry.Slot)
		return true
	}

	targetPtr := entry.ChunkPtr
	targetSlot := entry.Slot
	if targetArchID != entry.ArchID {
		targetPtr, targetSlot = e.manager.migrateEntity(entityID, entry.ArchID, entry.ChunkPtr, entry.Slot, targetArchID)
	}

	// A remove-only Editor has nothing to write into — skip cursor setup.
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
		e.Cursor.Set(targetPtr, uintptr(targetSlot))
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
