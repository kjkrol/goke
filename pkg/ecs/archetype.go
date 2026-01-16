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

type ArchRow uint32

type EntityArchLink struct {
	arch *Archetype
	row  ArchRow
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
	row := a.registerEntity(entity)

	for id, col := range a.columns {
		if id == compID {
			col.setData(row, data)
		} else {
			col.zeroData(row)
		}
	}
	return EntityArchLink{arch: a, row: row}
}

func (a *Archetype) RemoveEntity(row ArchRow) (swapedEntity Entity, swaped bool) {
	lastRow := ArchRow(a.len - 1)
	entityToMove := a.entities[lastRow]

	// 1. Swap data in all columns
	for _, col := range a.columns {
		if row != lastRow {
			col.copyData(row, lastRow)
		}
		col.zeroData(lastRow)
	}

	// 2. Swap entity ID in the entities slice
	a.entities[row] = entityToMove
	a.entities[lastRow] = 0
	a.len--
	// 3. Return the entity that was moved to the new position
	// If we removed the last one, no entity was moved to 'index'
	if row == lastRow {
		return 0, false
	}
	return entityToMove, true
}

func (a *Archetype) registerEntity(entity Entity) ArchRow {
	a.ensureCapacity()
	newIdx := a.len

	a.entities[newIdx] = entity
	a.len++

	return ArchRow(newIdx)
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
