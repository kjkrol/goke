package ent

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/soa"
)

// Manager owns entity lifecycle and component composition.
// It combines ID allocation, location tracking, and archetype management
// into a single owner: entity = ID + components.
type Manager struct {
	pool        uid.UID64Pool
	Index       Index
	ArchCatalog arch.Catalog
}

func (m *Manager) Init(cfg Config, onArchetypeCreated func(*arch.Archetype)) {
	m.pool.Init(cfg.Cap, cfg.FreeCap)
	m.Index.Init(cfg.Cap)
	m.ArchCatalog.Init(onArchetypeCreated)
}

// Create allocates a new entity ID, places the entity in the root archetype,
// and registers its location. Standard single-entity creation path.
func (m *Manager) Create() uid.UID64 {
	id := m.pool.Next()
	pos := m.ArchCatalog.AddEntityToTable(arch.RootID, id)
	m.Index.Upsert(id, arch.RootID, pos)
	return id
}

// Remove validates the entity ID, removes it from its archetype table via
// swap-and-pop, clears its index entry, and recycles the ID.
func (m *Manager) Remove(id uid.UID64) bool {
	if !m.pool.IsValid(id) {
		return false
	}
	link, ok := m.Index.Get(id)
	if ok {
		swappedEntity, swapped := m.ArchCatalog.RemoveEntityFromTable(link.ArchId, link.Pos)
		if swapped {
			m.Index.Upsert(swappedEntity, link.ArchId, link.Pos)
		}
		m.Index.Clear(id)
	}
	m.pool.Release(id)
	return true
}

// NextID allocates a new entity ID without placing it in any archetype.
// Used by blueprints for bulk creation, which handle archetype placement directly.
func (m *Manager) NextID() uid.UID64 {
	return m.pool.Next()
}

// UpdateLocation registers or overwrites the entity's storage position.
// Used by blueprints after bulk placement into an archetype table.
func (m *Manager) UpdateLocation(id uid.UID64, archID arch.ID, pos soa.BlockPos) {
	m.Index.Upsert(id, archID, pos)
}

// UpsertComp ensures the entity has the given component, migrating to a new
// archetype if necessary, and returns a pointer to the component's storage slot.
// If the component is a zero-size tag, returns (nil, nil).
func (m *Manager) UpsertComp(entityID uid.UID64, compMeta comp.Meta) (unsafe.Pointer, error) {
	if !m.pool.IsValid(entityID) {
		return nil, errInvalidEntity
	}

	link, ok := m.Index.Get(entityID)
	if !ok {
		panic("entity: UpsertComp called with entityID not present in Index — broken invariant")
	}

	targetArchID := link.ArchId
	targetPos := link.Pos

	if !m.ArchCatalog.Archetypes[link.ArchId].Mask().IsSet(compMeta.ID) {
		targetArchID = m.ArchCatalog.EnsureEdgeNext(compMeta, link.ArchId)
		newPos, swappedEntity, swapped := m.ArchCatalog.MigrateEntity(entityID, link.ArchId, link.Pos, targetArchID)
		if swapped {
			m.Index.Upsert(swappedEntity, link.ArchId, link.Pos)
		}
		m.Index.Upsert(entityID, targetArchID, newPos)
		targetPos = newPos
	}

	if compMeta.Size == 0 {
		return nil, nil
	}

	targetArch := &m.ArchCatalog.Archetypes[targetArchID]
	column := targetArch.Table.GetColumn(compMeta.ID)
	chunk := targetArch.Table.GetChunk(targetPos.ChunkIdx)
	return column.At(chunk, targetPos.ChunkSlot), nil
}

// RemoveComp removes the given component from the entity, migrating to the
// appropriate archetype. If the entity would have no components remaining,
// it is unlinked from archetype storage entirely.
func (m *Manager) RemoveComp(entityID uid.UID64, compMeta comp.Meta) error {
	if !m.pool.IsValid(entityID) {
		return errInvalidEntity
	}

	link, ok := m.Index.Get(entityID)
	if !ok {
		return nil
	}

	if !m.ArchCatalog.Archetypes[link.ArchId].Mask().IsSet(compMeta.ID) {
		return nil
	}

	targetArchID, shouldUnlink := m.ArchCatalog.EnsureEdgePrev(compMeta, link.ArchId)
	if shouldUnlink {
		swappedEntity, swapped := m.ArchCatalog.RemoveEntityFromTable(link.ArchId, link.Pos)
		if swapped {
			m.Index.Upsert(swappedEntity, link.ArchId, link.Pos)
		}
		m.Index.Clear(entityID)
		return nil
	}

	newPos, swappedEntity, swapped := m.ArchCatalog.MigrateEntity(entityID, link.ArchId, link.Pos, targetArchID)
	if swapped {
		m.Index.Upsert(swappedEntity, link.ArchId, link.Pos)
	}
	m.Index.Upsert(entityID, targetArchID, newPos)
	return nil
}

// GetComp returns a pointer to the entity's component data for the given
// component ID, or an error if the entity is invalid or lacks the component.
func (m *Manager) GetComp(entityID uid.UID64, compID comp.ID) (unsafe.Pointer, error) {
	link, ok := m.Index.Get(entityID)
	if !ok {
		return nil, errInvalidEntity
	}

	archetype := &m.ArchCatalog.Archetypes[link.ArchId]
	col := archetype.Table.GetColumn(compID)
	if col == nil {
		return nil, errComponentMissing
	}
	chunk := archetype.Table.GetChunk(link.Pos.ChunkIdx)
	return col.At(chunk, link.Pos.ChunkSlot), nil
}

// Reset clears all entity state, returning the manager to its initial condition.
func (m *Manager) Reset() {
	m.ArchCatalog.Reset()
	m.Index.Reset()
	m.pool.Reset()
}
