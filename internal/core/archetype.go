package core

import (
	"unsafe"
)

const initCapacity = 1024

// Supports 256 unique component types
type Archetype struct {
	Mask     ArchetypeMask
	entities []Entity
	Columns  map[ComponentID]*Column
	len      int
	cap      int

	edgesNext map[ComponentID]*Archetype
	edgesPrev map[ComponentID]*Archetype
}

type ArchRow uint32

type EntityArchLink struct {
	Arch *Archetype
	Row  ArchRow
}

func NewArchetype(mask ArchetypeMask) *Archetype {
	return &Archetype{
		Mask:      mask,
		entities:  make([]Entity, initCapacity),
		Columns:   make(map[ComponentID]*Column),
		len:       0,
		cap:       initCapacity,
		edgesNext: make(map[ComponentID]*Archetype),
		edgesPrev: make(map[ComponentID]*Archetype),
	}
}

func (a *Archetype) AddEntity(entity Entity, compID ComponentID, data unsafe.Pointer) EntityArchLink {
	row := a.registerEntity(entity)

	for id, col := range a.Columns {
		if id == compID {
			col.setData(row, data)
		} else {
			col.zeroData(row)
		}
	}
	return EntityArchLink{Arch: a, Row: row}
}

func (a *Archetype) SwapRemoveEntity(row ArchRow) (swapedEntity Entity, swaped bool) {
	lastRow := ArchRow(a.len - 1)
	entityToMove := a.entities[lastRow]

	// 1. Swap data in all columns
	for _, col := range a.Columns {
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

	for _, col := range a.Columns {
		col.growTo(newCap)
	}

	a.cap = newCap
}
