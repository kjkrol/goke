package ecs

import (
	"unsafe"
)

const initCapacity = 1024

type archetype struct {
	mask          ArchetypeMask
	entities      []Entity
	columns       map[ComponentID]*column
	entityToIndex map[Entity]int
	len           int
	cap           int
}

func newArchetype(mask ArchetypeMask) *archetype {
	return &archetype{
		mask:          mask,
		entities:      make([]Entity, initCapacity),
		columns:       make(map[ComponentID]*column),
		entityToIndex: make(map[Entity]int),
		len:           0,
		cap:           initCapacity,
	}
}

func (a *archetype) addEntity(entity Entity, compID ComponentID, data unsafe.Pointer) {
	a.ensureCapacity()

	newIdx := a.registerEntity(entity)

	for id, col := range a.columns {
		if id == compID {
			col.setData(newIdx, data)
		} else {
			col.zeroData(newIdx)
		}
	}
}

func (a *archetype) ensureCapacity() {
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

func (a *archetype) removeEntity(index int) {
	lastIdx := a.len - 1
	entityToRemove := a.entities[index]
	lastEntity := a.entities[lastIdx]

	for _, col := range a.columns {
		if index != lastIdx {
			col.copyData(index, lastIdx)
		}
		col.zeroData(lastIdx)
	}

	if index != lastIdx {
		a.entities[index] = lastEntity
		a.entityToIndex[lastEntity] = index
	}

	a.entities[lastIdx] = 0
	delete(a.entityToIndex, entityToRemove)
	a.len--
}

func (a *archetype) registerEntity(entity Entity) int {
	newIdx := a.len

	a.entities[newIdx] = entity
	a.entityToIndex[entity] = newIdx
	a.len++

	return newIdx
}
