package core

import (
	"errors"
	"fmt"
	"unsafe"
)

type ArchetypeRegistry struct {
	archetypeMaskMap          ArchetypeMaskMap
	Archetypes                [MaxArchetypeId]Archetype
	lastArchetypeId           ArchetypeId
	EntityLinkStore           EntityLinkStore
	componentsRegistry        *ComponentsRegistry
	viewRegistry              *ViewRegistry
	defaultArchetypeChunkSize int
	sharedColsInfos           []ComponentInfo
}

func NewArchetypeRegistry(
	componentsRegistry *ComponentsRegistry,
	viewRegistry *ViewRegistry,
	cfg RegistryConfig,
) *ArchetypeRegistry {
	reg := ArchetypeRegistry{
		EntityLinkStore:           NewEntityLinkStore(cfg.InitialEntityCap),
		componentsRegistry:        componentsRegistry,
		viewRegistry:              viewRegistry,
		defaultArchetypeChunkSize: cfg.DefaultArchetypeChunkSize,
		lastArchetypeId:           RootArchetypeId,
	}

	rootMask := ArchetypeMask{}
	reg.InitArchetype(rootMask, reg.defaultArchetypeChunkSize)

	return &reg
}

func (r *ArchetypeRegistry) InitArchetype(mask ArchetypeMask, initCapacity int) ArchetypeId {
	if r.lastArchetypeId >= MaxArchetypeId {
		panic(fmt.Sprintf("Max archetype number exceeded: %d", MaxArchetypeId))
	}

	r.sharedColsInfos = r.sharedColsInfos[:0]

	defer func() {
		clear(r.sharedColsInfos)
		r.sharedColsInfos = r.sharedColsInfos[:0]
	}()

	for id := range mask.AllSet() {
		info := r.componentsRegistry.idToInfo[id]
		// tags should not have columns
		if info.Size == 0 {
			continue
		}
		r.sharedColsInfos = append(r.sharedColsInfos, info)
	}

	archId := r.lastArchetypeId
	arch := &r.Archetypes[archId]
	arch.InitArchetype(archId, mask, r.sharedColsInfos)

	r.archetypeMaskMap.Put(mask, archId)

	r.lastArchetypeId++

	return archId
}

func (r *ArchetypeRegistry) Get(mask ArchetypeMask) ArchetypeId {
	if archId, ok := r.archetypeMaskMap.Get(mask); ok {
		return archId
	}
	return NullArchetypeId
}

func (r *ArchetypeRegistry) AddEntity(entity Entity, archId ArchetypeId) (ChunkIdx, ChunkRow) {
	arch := &r.Archetypes[archId]
	chunkIdx, chunkRow := arch.AddEntity(entity)
	r.EntityLinkStore.Update(entity, archId, chunkIdx, chunkRow)
	return chunkIdx, chunkRow
}

func (r *ArchetypeRegistry) UnlinkEntity(entity Entity) {
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}

	linkArch := &r.Archetypes[link.ArchId]
	swappedEntity, swapped := linkArch.SwapRemoveEntity(link.ChunkIdx, link.ChunkRow)

	if swapped {
		r.EntityLinkStore.Update(swappedEntity, link.ArchId, link.ChunkIdx, link.ChunkRow)
	}

	r.EntityLinkStore.Clear(entity)
}

var (
	ErrEntityNotFound = errors.New("entity not found in registry")
)

// AllocateComponentMemory ensures the entity has the component and returns a pointer to its memory.
func (r *ArchetypeRegistry) AllocateComponentMemory(
	entity Entity,
	compInfo ComponentInfo,
) (unsafe.Pointer, error) {
	compID := compInfo.ID

	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return nil, ErrEntityNotFound
	}

	currentArch := &r.Archetypes[link.ArchId]

	var targetArch *Archetype
	var targetChunkIdx ChunkIdx
	var targetRow ChunkRow

	// 2. Ścieżka Szybka: Komponent już istnieje
	if currentArch.Mask.IsSet(compID) {
		targetArch = currentArch
		targetChunkIdx = link.ChunkIdx
		targetRow = link.ChunkRow
	} else {
		// 3. Ścieżka Wolna: Strukturalna zmiana (Tranzycja)
		nextArchID := r.ensureNextEdgeId(compID, currentArch)

		// Przenosimy encję i dostajemy nowe koordynaty
		targetChunkIdx, targetRow = r.moveEntity(entity, link, nextArchID)
		targetArch = &r.Archetypes[nextArchID]
	}

	// Jeśli komponent jest znacznikiem (Tag, size 0), nie zwracamy wskaźnika
	if compInfo.Size == 0 {
		return nil, nil
	}

	// 4. Obliczamy i zwracamy wskaźnik
	// Teraz jest to dużo bezpieczniejsze dzięki Chunk Architecture.

	// Pobieramy kolumnę z docelowego archetypu
	column := targetArch.GetColumn(compID)

	// Pobieramy fizyczny chunk
	chunk := targetArch.Memory.GetChunk(targetChunkIdx)

	// Kolumna sama wylicza adres
	return column.GetPointer(chunk, targetRow), nil
}

func (r *ArchetypeRegistry) ensureNextEdgeId(compID ComponentID, oldArch *Archetype) ArchetypeId {
	if nextEdgeId := oldArch.graph.edgesNext[compID]; nextEdgeId != NullArchetypeId {
		return nextEdgeId
	}

	// Slow path: create or get new archetype
	newMask := oldArch.Mask.Set(compID)
	nextArchId := r.getOrRegister(newMask)

	// Link in the graph
	nextArch := &r.Archetypes[nextArchId]

	oldArch.linkNextArch(nextArch, compID)

	return nextArchId
}

func (r *ArchetypeRegistry) UnAssign(entity Entity, compInfo ComponentInfo) {
	link, ok := r.EntityLinkStore.Get(entity)
	if !ok {
		return
	}

	oldArchId := link.ArchId
	compID := compInfo.ID
	oldArch := &r.Archetypes[oldArchId]

	// -------------------------------------------------------------------------
	// KROK 1: Pobranie fizycznego chunka (Resolve Once)
	// Musimy mieć dostęp do pamięci, aby wyzerować usuwany komponent.
	// -------------------------------------------------------------------------
	srcChunk := oldArch.Memory.GetChunk(link.ChunkIdx)

	// Helper do zerowania danych konkretnego komponentu
	zeroComponent := func(arch *Archetype, c *chunk, r ChunkRow, cID ComponentID) {
		col := arch.GetColumn(cID)
		ptr := col.GetPointer(c, r)
		// Helper zeroMemory (clear) - musisz go mieć zdefiniowanego w pakiecie
		zeroMemory(ptr, col.ItemSize)
	}

	// -------------------------------------------------------------------------
	// FAST PATH (Graph Edge Exists)
	// -------------------------------------------------------------------------
	if prevArchId := oldArch.graph.edgesPrev[compID]; prevArchId != NullArchetypeId {
		prevArch := &r.Archetypes[prevArchId]

		// Jeśli nowy archetyp jest pusty (Root), usuwamy encję całkowicie
		if prevArch.Mask.IsEmpty() {
			r.UnlinkEntity(entity)
			return
		}

		// Wyzeruj usuwany komponent (dla bezpieczeństwa GC przed przeniesieniem reszty)
		zeroComponent(oldArch, srcChunk, link.ChunkRow, compID)

		// Przenieś encję (moveEntity sam skopiuje pozostałe komponenty i zrobi SwapRemove)
		r.moveEntity(entity, link, prevArchId)
		return
	}

	// -------------------------------------------------------------------------
	// SLOW PATH (Calculate Mask & Create Edge)
	// -------------------------------------------------------------------------
	newMask := oldArch.Mask.Clear(compID)

	// Jeśli maska się nie zmieniła, to znaczy, że encja nie miała tego komponentu
	if oldArch.Mask == newMask {
		return
	}

	// Jeśli nowa maska jest pusta -> Unlink
	if newMask.IsEmpty() {
		r.UnlinkEntity(entity)
		return
	}

	// Znajdź lub stwórz nowy archetyp
	newPrevArchId := r.getOrRegister(newMask)
	newPrevArch := &r.Archetypes[newPrevArchId]

	// Zaktualizuj Graf (Edges)
	oldArch.linkPrevArch(newPrevArch, compID)

	// Wyzeruj i Przenieś
	zeroComponent(oldArch, srcChunk, link.ChunkRow, compID)
	r.moveEntity(entity, link, newPrevArchId)
}

func (r *ArchetypeRegistry) Reset() {
	for i := range int(r.lastArchetypeId) {
		r.Archetypes[i].Reset()
	}
	clear(r.Archetypes[:])

	r.archetypeMaskMap.Reset()
	r.EntityLinkStore.Reset()

	clear(r.sharedColsInfos)
	r.sharedColsInfos = r.sharedColsInfos[:0]

	r.lastArchetypeId = RootArchetypeId
	rootMask := ArchetypeMask{}
	r.InitArchetype(rootMask, r.defaultArchetypeChunkSize)
}

// --------------------------------------------------------------

func (r *ArchetypeRegistry) getOrRegister(mask ArchetypeMask) ArchetypeId {
	if found, ok := r.archetypeMaskMap.Get(mask); ok {
		return found
	}
	archId := r.InitArchetype(mask, r.defaultArchetypeChunkSize)
	r.viewRegistry.OnArchetypeCreated(&r.Archetypes[archId])
	return archId
}

// --------------------------------------------------------------

func (r *ArchetypeRegistry) moveEntity(
	entity Entity,
	link EntityArchLink,
	archId ArchetypeId,
) (ChunkIdx, ChunkRow) {
	oldArch := &r.Archetypes[link.ArchId]
	newArch := &r.Archetypes[archId]

	newChunkIdx, newRow := newArch.AddEntity(entity)

	srcChunk := oldArch.Memory.GetChunk(link.ChunkIdx)
	dstChunk := newArch.Memory.GetChunk(newChunkIdx)

	// 3. Kopiuj wspólne komponenty
	// Iterujemy po kolumnach NOWEGO archetypu
	for i := range newArch.Columns {
		dstCol := &newArch.Columns[i]

		// Pomijamy EntityID (Index 0), bo AddEntity już to obsłużyło
		if dstCol.CompID == 0 {
			continue
		}

		// Jeśli stary archetyp też miał ten komponent -> Kopiujemy dane
		if srcCol := oldArch.GetColumn(dstCol.CompID); srcCol != nil {
			// Używamy helperów do wyciągnięcia wskaźników
			srcPtr := srcCol.GetPointer(srcChunk, link.ChunkRow)
			dstPtr := dstCol.GetPointer(dstChunk, newRow)

			// Kopiowanie pamięci
			copyMemory(dstPtr, srcPtr, dstCol.ItemSize)
		}
	}

	// 4. Usuń ze starego miejsca (Swap Remove)
	swappedEntity, swapped := oldArch.SwapRemoveEntity(link.ChunkIdx, link.ChunkRow)

	// 5. Aktualizacja LinkStore
	// A. Jeśli coś wskoczyło w dziurę po starej encji -> aktualizujemy to
	if swapped {
		r.EntityLinkStore.Update(swappedEntity, link.ArchId, link.ChunkIdx, link.ChunkRow)
	}

	// B. Aktualizujemy przenoszoną encję (nowy adres)
	r.EntityLinkStore.Update(entity, newArch.Id, newChunkIdx, newRow)

	return newChunkIdx, newRow
}
