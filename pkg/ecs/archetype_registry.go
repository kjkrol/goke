package ecs

import (
	"reflect"
	"unsafe"
)

const initArchetypesCapacity = 64

type archetypeRegistry struct {
	archetypeMap       map[ArchetypeMask]*archetype
	archetypes         []*archetype
	componentsRegistry *componentsRegistry
}

func newArchetypeRegistry(componentsRegistry *componentsRegistry) *archetypeRegistry {
	return &archetypeRegistry{
		archetypeMap:       make(map[ArchetypeMask]*archetype),
		archetypes:         make([]*archetype, 0, initArchetypesCapacity),
		componentsRegistry: componentsRegistry,
	}
}

func (r *archetypeRegistry) All() []*archetype {
	return r.archetypes
}

// Get zwraca archetyp dla konkretnej maski (używane w Filtered)
func (r *archetypeRegistry) Get(mask ArchetypeMask) *archetype {
	return r.archetypeMap[mask]
}

func (r *archetypeRegistry) GetOrRegister(mask ArchetypeMask) *archetype {
	if arch, ok := r.archetypeMap[mask]; ok {
		return arch
	}

	arch := newArchetype(mask)

	// Wypełnianie kolumn na podstawie maski
	mask.ForEachSet(func(id ComponentID) {
		info := r.componentsRegistry.idToInfo[id]

		// Tworzymy bufor danych
		slice := reflect.MakeSlice(reflect.SliceOf(info.Type), initCapacity, initCapacity)

		// Dodajemy kolumnę do SLICE'a
		newCol := column{
			id:       id,
			data:     slice.UnsafePointer(),
			dataType: info.Type,
			itemSize: info.Size,
			len:      0,
			cap:      initCapacity,
		}

		// Rejestrujemy indeks kolumny w mapie lookup
		arch.colTable[id] = len(arch.columns)
		arch.columns = append(arch.columns, newCol)
	})

	r.archetypeMap[mask] = arch
	r.archetypes = append(r.archetypes, arch)
	return arch
}

// MoveEntity przenosi encję między archetypami kopiując dane
func (r *archetypeRegistry) MoveEntity(entity Entity, oldArch, newArch *archetype, newCompID ComponentID, newData unsafe.Pointer) {
	// 1. Zarejestruj encję w nowym archetypie
	if newArch.len >= newArch.cap {
		newArch.ensureCapacity()
	}
	newIdx := newArch.registerEntity(entity)

	// 2. Jeśli nie było starego archetypu (pierwsze dodanie), tylko ustawiamy nowy komponent
	if oldArch == nil {
		if colIdx, ok := newArch.colTable[newCompID]; ok {
			newArch.columns[colIdx].setData(newIdx, newData)
		}
		return
	}

	oldIdx := oldArch.entityToIndex[entity]

	// 3. Kopiowanie wspólnych komponentów
	// Iterujemy po mniejszym zbiorze lub po prostu po kolumnach nowego archetypu
	// Użycie colTable pozwala na szybkie parowanie
	for i := range newArch.columns {
		col := &newArch.columns[i] // cel

		if col.id == newCompID {
			// To jest ten nowy komponent, wpisujemy dane z argumentu
			col.setData(newIdx, newData)
		} else {
			// Sprawdzamy, czy stary archetyp też miał ten komponent
			if oldColIdx, exists := oldArch.colTable[col.id]; exists {
				oldCol := &oldArch.columns[oldColIdx]
				// Kopiujemy pamięć ze starej kolumny do nowej
				src := oldCol.GetElement(oldIdx)
				col.setData(newIdx, src)
			}
		}
	}

	// 4. Usunięcie ze starego archetypu
	oldArch.removeEntity(oldIdx)
}

func (r *archetypeRegistry) MoveEntityOnly(entity Entity, oldArch, newArch *archetype) {
	oldIdx, ok := oldArch.entityToIndex[entity]
	if !ok {
		return
	}

	if newArch.len >= newArch.cap {
		newArch.ensureCapacity()
	}
	newIdx := newArch.registerEntity(entity)

	// Kopiowanie danych
	for i := range newArch.columns {
		newCol := &newArch.columns[i]
		// Szukamy odpowiadającej kolumny w starym archetypie
		if oldColIdx, exists := oldArch.colTable[newCol.id]; exists {
			oldCol := &oldArch.columns[oldColIdx]
			src := oldCol.GetElement(oldIdx)
			newCol.setData(newIdx, src)
		}
	}

	oldArch.removeEntity(oldIdx)
}
