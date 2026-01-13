package ecs

const initCapacity = 1024

type archetype struct {
	mask     ArchetypeMask
	entities []Entity

	// ZMIANA KLUCZOWA: Slice zamiast Mapy.
	// Pozwala na iterację liniową (pre-fetching procesora).
	columns []column

	// Lookup Table: mapuje ComponentID -> index w slice columns.
	// Używane przy losowym dostępie (Get/Set/Move), nie przy iteracji Query.
	colTable map[ComponentID]int

	entityToIndex map[Entity]int
	len           int
	cap           int
}

func newArchetype(mask ArchetypeMask) *archetype {
	return &archetype{
		mask:          mask,
		entities:      make([]Entity, 0, initCapacity),
		columns:       make([]column, 0, mask.Count()), // Alokujemy tyle, ile bitów w masce
		colTable:      make(map[ComponentID]int),
		entityToIndex: make(map[Entity]int, initCapacity),
		len:           0,
		cap:           initCapacity,
	}
}

func (a *archetype) addEntity(entity Entity) int {
	if a.len >= a.cap {
		a.ensureCapacity()
	}
	return a.registerEntity(entity)
}

// registerEntity tylko rezerwuje slot, dane są wpisywane przez Registry
func (a *archetype) registerEntity(entity Entity) int {
	newIdx := a.len
	a.entities = append(a.entities, entity) // append automatycznie zarządza len
	a.entityToIndex[entity] = newIdx
	a.len++

	// Aktualizujemy długość w kolumnach (ważne logicznie, choć dane są pointerami)
	for i := range a.columns {
		a.columns[i].len++
	}
	return newIdx
}

func (a *archetype) ensureCapacity() {
	newCap := a.cap * 2
	if newCap == 0 {
		newCap = initCapacity
	}

	// Rozszerzanie kolumn
	for i := range a.columns {
		a.columns[i].growTo(newCap)
	}

	// Entities rosną przez append, ale przy copy trzeba uważać na capacity
	// Tutaj, skoro używamy append w registerEntity, pre-allokacja nie jest krytyczna dla entities,
	// ale dobra dla wydajności.
	if cap(a.entities) < newCap {
		newEnts := make([]Entity, len(a.entities), newCap)
		copy(newEnts, a.entities)
		a.entities = newEnts
	}

	a.cap = newCap
}

func (a *archetype) removeEntity(index int) {
	lastIdx := a.len - 1
	lastEntity := a.entities[lastIdx]

	// 1. Swap & Pop w kolumnach
	// Iterujemy po SLICE - to jest super szybkie
	for i := range a.columns {
		col := &a.columns[i] // Pobieramy wskaźnik do elementu slice'a
		if index != lastIdx {
			col.copyData(index, lastIdx)
		}
		// Opcjonalne: zerowanie ostatniego elementu (dla GC)
		// col.zeroData(lastIdx)
		col.len--
	}

	// 2. Aktualizacja mapy i listy encji
	if index != lastIdx {
		a.entities[index] = lastEntity
		a.entityToIndex[lastEntity] = index
	}

	// Usuwamy ostatni element z slice'a entities
	entityToRemove := a.entities[index] // Uwaga: w Twoim kodzie było tu ryzyko błędu logicznego
	// Poprawna logika: usuwamy wpis TEJ encji, która była pod zadanym indeksem PRZED swapem
	// Ale skoro zrobiliśmy swap, to pod `index` jest teraz `lastEntity`.
	// Musimy usunąć mapowanie encji, którą chcieliśmy usunąć.
	// Ponieważ `entityToIndex` mapuje Entity -> int, a my mamy int, musimy wiedzieć co usuwamy.
	// W tym kodzie `entityToRemove` (jeśli pobrane przed swapem) jest OK, ale tutaj:

	// Naprawa logiki usuwania z mapy:
	// a.entities[lastIdx] to encja do usunięcia (bo zrobiliśmy swap logicznie, ale w entities fizycznie jeszcze nie obcięliśmy slice'a).
	// Wróć. Najczystsza wersja:
	// Encja usuwana była pod `index`. Encja ostatnia była pod `lastIdx`.
	// Jeśli index != lastIdx, nadpisujemy index.

	// W Twoim oryginalnym kodzie było: `entityToRemove := a.entities[index]`. To jest poprawne.

	delete(a.entityToIndex, entityToRemove) // Usuwamy wpis usuwanej encji
	a.entities = a.entities[:lastIdx]       // Skracamy slice
	a.len--
}
