package ent

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/addr"
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/colstore"
	"github.com/kjkrol/goke/internal/comp"
)

// Manager owns entity lifecycle and component composition.
// It combines the address book and archetype management into a single owner:
// entity = ID + components.
type Manager struct {
	AddressBook addr.Book
	ArchCatalog arch.Catalog
}

func (m *Manager) Init(cfg Config, onArchetypeCreated func(*arch.Archetype)) {
	m.AddressBook.Init(cfg.Cap, cfg.FreeCap)
	m.ArchCatalog.Init(func(a *arch.Archetype) {
		archID := a.Id
		a.Table.SetIDSeeder(func(dst []uid.UID64, pos colstore.Pos) {
			m.AddressBook.Seed(dst, archID, pos)
		})
		if onArchetypeCreated != nil {
			onArchetypeCreated(a)
		}
	})
}

// Remove validates the entity ID, removes it from its archetype table via
// swap-and-pop, clears its address entry, and recycles the ID.
func (m *Manager) Remove(id uid.UID64) bool {
	entry, ok := m.AddressBook.Get(id)
	if !ok {
		return false
	}
	m.removeFromArchetype(id, entry.ArchId, entry.Pos)
	return true
}

// CreateFactory resolves or creates the archetype from b and returns
// a reusable Factory ready for repeated Create/Next cycles.
func (m *Manager) CreateFactory(b comp.Blueprint) *Factory {
	var f Factory
	f.Init(m, b)
	return &f
}

// UpsertComp ensures the entity has the given component, migrating to a new
// archetype if necessary, and returns a pointer to the component's storage slot.
// If the component is a zero-size tag, returns (nil, nil).
func (m *Manager) UpsertComp(entityID uid.UID64, compDef comp.Def) (unsafe.Pointer, error) {
	entry, ok := m.AddressBook.Get(entityID)
	if !ok {
		return nil, errInvalidEntity
	}

	targetArchID := entry.ArchId
	targetPos := entry.Pos

	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(compDef.ID) {
		targetArchID = m.ArchCatalog.EnsureEdgeNext(compDef, entry.ArchId)
		targetPos = m.migrateEntity(entityID, entry.ArchId, entry.Pos, targetArchID)
	}

	if compDef.Size == 0 {
		return nil, nil
	}

	return m.ArchCatalog.Archetypes[targetArchID].Table.ComponentAt(targetPos, compDef.ID), nil
}

// RemoveComp removes the given component from the entity, migrating it to the
// appropriate archetype. If the entity would have no components remaining,
// it is unlinked from archetype storage entirely.
func (m *Manager) RemoveComp(entityID uid.UID64, compDef comp.Def) error {
	entry, ok := m.AddressBook.Get(entityID)
	if !ok {
		return errInvalidEntity
	}

	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(compDef.ID) {
		return nil
	}

	targetArchID, shouldUnlink := m.ArchCatalog.EnsureEdgePrev(compDef, entry.ArchId)
	if shouldUnlink {
		m.removeFromArchetype(entityID, entry.ArchId, entry.Pos)
		return nil
	}

	m.migrateEntity(entityID, entry.ArchId, entry.Pos, targetArchID)
	return nil
}

// GetComp returns an unsafe.Pointer to the entity's component data for the given
// component ID, or an error if the entity is invalid or lacks the component.
func (m *Manager) GetComp(entityID uid.UID64, compID comp.ID) (unsafe.Pointer, error) {
	entry, ok := m.AddressBook.Get(entityID)
	if !ok {
		return nil, errInvalidEntity
	}

	ptr := m.ArchCatalog.Archetypes[entry.ArchId].Table.ComponentAt(entry.Pos, compID)
	if ptr == nil {
		return nil, errComponentMissing
	}
	return ptr, nil
}

// Reset clears all entity state, returning the manager to its initial condition.
func (m *Manager) Reset() {
	m.ArchCatalog.Reset()
	m.AddressBook.Reset()
}

func (m *Manager) removeFromArchetype(id uid.UID64, archID arch.ID, pos colstore.Pos) {
	swappedEntity, swapped := m.ArchCatalog.RemoveEntity(archID, pos)
	if swapped {
		m.AddressBook.Move(swappedEntity, archID, pos)
	}
	m.AddressBook.Delete(id)
}

func (m *Manager) migrateEntity(id uid.UID64, srcArchID arch.ID, srcPos colstore.Pos, dstArchID arch.ID) colstore.Pos {
	newPos, swappedEntity, swapped := m.ArchCatalog.MigrateEntity(id, srcArchID, srcPos, dstArchID)
	if swapped {
		m.AddressBook.Move(swappedEntity, srcArchID, srcPos)
	}
	m.AddressBook.Move(id, dstArchID, newPos)
	return newPos
}
