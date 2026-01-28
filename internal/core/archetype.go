package core

import (
	"unsafe"
)

// Supports 256 unique component types
type Archetype struct {
	Mask     ArchetypeMask
	entities []Entity
	Columns  map[ComponentID]*Column
	len      int
	cap      int

	edgesNext map[ComponentID]*Archetype
	edgesPrev map[ComponentID]*Archetype
	initCap   int
}

type ArchRow uint32

type EntityArchLink struct {
	Arch *Archetype
	Row  ArchRow
}

func NewArchetype(mask ArchetypeMask, defaultArchetypeChunkSize int) *Archetype {
	return &Archetype{
		Mask:      mask,
		entities:  make([]Entity, defaultArchetypeChunkSize),
		Columns:   make(map[ComponentID]*Column),
		len:       0,
		cap:       defaultArchetypeChunkSize,
		edgesNext: make(map[ComponentID]*Archetype),
		edgesPrev: make(map[ComponentID]*Archetype),
		initCap:   defaultArchetypeChunkSize,
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
		col.len--
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

	for _, col := range a.Columns {
		col.len++
	}

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
		newCap = a.initCap
	}

	newEntities := make([]Entity, newCap)
	copy(newEntities, a.entities)
	a.entities = newEntities

	for _, col := range a.Columns {
		col.growTo(newCap)
	}

	a.cap = newCap
}
