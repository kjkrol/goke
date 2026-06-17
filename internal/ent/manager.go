package ent

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/addr"
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/mem"
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
		a.Table.SetIDSeeder(func(dst []uid.UID64, chunkIdx mem.ChunkIdx, startSlot mem.ChunkSlot) {
			m.AddressBook.Seed(dst, archID, chunkIdx, startSlot)
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
	swappedEntity, swapped := m.ArchCatalog.RemoveEntityFromTable(entry.ArchId, entry.Pos)
	if swapped {
		m.AddressBook.Move(swappedEntity, entry.ArchId, entry.Pos)
	}
	m.AddressBook.Delete(id)
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
func (m *Manager) UpsertComp(entityID uid.UID64, compMeta comp.Meta) (unsafe.Pointer, error) {
	entry, ok := m.AddressBook.Get(entityID)
	if !ok {
		return nil, errInvalidEntity
	}

	targetArchID := entry.ArchId
	targetPos := entry.Pos

	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(compMeta.ID) {
		targetArchID = m.ArchCatalog.EnsureEdgeNext(compMeta, entry.ArchId)
		newPos, swappedEntity, swapped := m.ArchCatalog.MigrateEntity(entityID, entry.ArchId, entry.Pos, targetArchID)
		if swapped {
			m.AddressBook.Move(swappedEntity, entry.ArchId, entry.Pos)
		}
		m.AddressBook.Move(entityID, targetArchID, newPos)
		targetPos = newPos
	}

	if compMeta.Size == 0 {
		return nil, nil
	}

	targetArch := &m.ArchCatalog.Archetypes[targetArchID]
	column := targetArch.Table.GetColumn(compMeta.ID)
	return column.At(targetArch.Table.ChunkPtr(targetPos.ChunkIdx), targetPos.ChunkSlot), nil
}

// RemoveComp removes the given component from the entity, migrating it to the
// appropriate archetype. If the entity would have no components remaining,
// it is unlinked from archetype storage entirely.
func (m *Manager) RemoveComp(entityID uid.UID64, compMeta comp.Meta) error {
	entry, ok := m.AddressBook.Get(entityID)
	if !ok {
		return errInvalidEntity
	}

	if !m.ArchCatalog.Archetypes[entry.ArchId].Mask().IsSet(compMeta.ID) {
		return nil
	}

	targetArchID, shouldUnlink := m.ArchCatalog.EnsureEdgePrev(compMeta, entry.ArchId)
	if shouldUnlink {
		swappedEntity, swapped := m.ArchCatalog.RemoveEntityFromTable(entry.ArchId, entry.Pos)
		if swapped {
			m.AddressBook.Move(swappedEntity, entry.ArchId, entry.Pos)
		}
		m.AddressBook.Delete(entityID)
		return nil
	}

	newPos, swappedEntity, swapped := m.ArchCatalog.MigrateEntity(entityID, entry.ArchId, entry.Pos, targetArchID)
	if swapped {
		m.AddressBook.Move(swappedEntity, entry.ArchId, entry.Pos)
	}
	m.AddressBook.Move(entityID, targetArchID, newPos)
	return nil
}

// GetComp returns an unsafe.Pointer to the entity's component data for the given
// component ID, or an error if the entity is invalid or lacks the component.
func (m *Manager) GetComp(entityID uid.UID64, compID comp.ID) (unsafe.Pointer, error) {
	entry, ok := m.AddressBook.Get(entityID)
	if !ok {
		return nil, errInvalidEntity
	}

	archetype := &m.ArchCatalog.Archetypes[entry.ArchId]
	col := archetype.Table.GetColumn(compID)
	if col == nil {
		return nil, errComponentMissing
	}
	return col.At(archetype.Table.ChunkPtr(entry.Pos.ChunkIdx), entry.Pos.ChunkSlot), nil
}

// Reset clears all entity state, returning the manager to its initial condition.
func (m *Manager) Reset() {
	m.ArchCatalog.Reset()
	m.AddressBook.Reset()
}
