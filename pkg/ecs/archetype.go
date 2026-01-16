package ecs

import (
	"unsafe"
)

const initCapacity = 1024

type Archetype struct {
	mask     ArchetypeMask
	entities []Entity
	columns  map[ComponentID]*column
	len      int
	cap      int
}

func NewArchetype(mask ArchetypeMask) *Archetype {
	return &Archetype{
		mask:     mask,
		entities: make([]Entity, initCapacity),
		columns:  make(map[ComponentID]*column),
		len:      0,
		cap:      initCapacity,
	}
}

func (a *Archetype) AddEntity(entity Entity, compID ComponentID, data unsafe.Pointer) EntityArchLink {
	newIdx := a.registerEntity(entity)

	for id, col := range a.columns {
		if id == compID {
			col.setData(newIdx, data)
		} else {
			col.zeroData(newIdx)
		}
	}
	return EntityArchLink{arch: a, columnIndex: newIdx}
}

func (a *Archetype) RemoveEntity(index int) (swapedEntity Entity, swaped bool) {
	lastIdx := a.len - 1
	entityToMove := a.entities[lastIdx]

	// 1. Swap data in all columns
	for _, col := range a.columns {
		if index != lastIdx {
			col.copyData(index, lastIdx)
		}
		col.zeroData(lastIdx)
	}

	// 2. Swap entity ID in the entities slice
	a.entities[index] = entityToMove
	a.entities[lastIdx] = 0
	a.len--
	// 3. Return the entity that was moved to the new position
	// If we removed the last one, no entity was moved to 'index'
	if index == lastIdx {
		return 0, false
	}
	return entityToMove, true
}

func (a *Archetype) registerEntity(entity Entity) int {
	a.ensureCapacity()
	newIdx := a.len

	a.entities[newIdx] = entity
	a.len++

	return newIdx
}

func (a *Archetype) ensureCapacity() {
	if a.len < a.cap {
		return
	}

	newCap := a.cap * 2
	if newCap == 0 {
		newCap = initCapacity
	}

	newEntities := make([]Entity, newCap)
	copy(newEntities, a.entities)
	a.entities = newEntities

	for _, col := range a.columns {
		col.growTo(newCap)
	}

	a.cap = newCap
}
