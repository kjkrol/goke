package ent

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/colstore"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/iter"
)

// Factory bulk-spawns entities for a single archetype using a chunk-based iterator.
// Call Create to set the count, then loop with Next; access entities via IDs and
// components via col.Slice(&factory.Cursor).
type Factory struct {
	IDs       []uid.UID64
	Cursor    iter.Cursor
	arch      *arch.Archetype
	colBakes  []colstore.ColBake
	chunkCap  int
	remaining int
	pos       colstore.Pos
	available int
}

// Init resolves or creates the archetype from accessSpec and prepares
// the Factory for repeated Create/Next cycles.
func (f *Factory) Init(em *Manager, accessSpec comp.AccessSpec) {
	archID := em.ArchCatalog.Upsert(accessSpec.Compose())
	f.arch = &em.ArchCatalog.Archetypes[archID]
	f.colBakes = f.arch.Table.BakeColumns(accessSpec.CompInfos)
	f.Cursor = iter.Cursor{Offsets: make([]uintptr, len(accessSpec.CompInfos))}
}

// Create pre-allocates chunks for count entities and resets the iterator.
// Call Next in a loop to advance through each allocated batch.
func (f *Factory) Create(count int) {
	f.pos.Idx, f.available, f.chunkCap = f.arch.Table.ReserveSlots(count)
	f.remaining = count
}

// Next advances to the next batch, registers entity IDs, and populates IDs and Cursor.
// Returns false when all requested entities have been created.
func (f *Factory) Next() bool {
	if f.remaining < 1 {
		f.arch.Table.ReleaseSlots()
		f.IDs = nil
		return false
	}

	allocatedSlots := min(f.remaining, f.available)
	f.IDs, f.pos = f.arch.Table.SpawnCursor(&f.Cursor, f.pos.Idx, allocatedSlots, f.colBakes)

	f.remaining -= allocatedSlots
	if f.remaining > 0 {
		f.pos.Idx++
		f.available = f.chunkCap
	}

	return true
}
