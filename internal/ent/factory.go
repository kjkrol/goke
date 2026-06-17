package ent

import (
	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/colstore"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/mem"
	"github.com/kjkrol/goke/iter"
)

// Factory bulk-spawns entities for a single archetype using a chunk-based iterator.
// Call Create to set the count, then loop with Next; access entities via IDs and
// components via col.Slice(&factory.Cursor).
type Factory struct {
	IDs       []uid.UID64
	Cursor    iter.Cursor
	arch      *arch.Archetype
	cols      []*colstore.Column
	remaining int
	pos       mem.BlockPos
	available int
}

// Init resolves or creates the archetype from b and prepares
// the Factory for repeated Create/Next cycles.
func (f *Factory) Init(em *Manager, b comp.Blueprint) {
	archID := em.ArchCatalog.Upsert(b.Compose())
	f.arch = &em.ArchCatalog.Archetypes[archID]
	f.cols = make([]*colstore.Column, len(b.CompInfos))
	for i, meta := range b.CompInfos {
		f.cols[i] = f.arch.Table.GetColumn(meta.ID)
	}
	f.Cursor = iter.Cursor{Offsets: make([]uintptr, len(b.CompInfos))}
}

// Create pre-allocates chunks for count entities and resets the iterator.
// Call Next in a loop to advance through each allocated batch.
func (f *Factory) Create(count int) {
	f.pos.ChunkIdx, f.available = f.arch.Table.PrepareSlots(count)
	f.remaining = count
}

// Next advances to the next batch, registers entity IDs, and populates IDs and Cursor.
// Returns false when all requested entities have been created.
func (f *Factory) Next() bool {
	if f.remaining < 1 {
		f.arch.Table.Reserved = 0
		f.IDs = nil
		return false
	}

	allocatedSlots := min(f.remaining, f.available)
	f.Cursor.Base, f.pos.ChunkSlot, f.IDs = f.arch.Table.SpawnEntitySlice(f.pos.ChunkIdx, allocatedSlots)

	for i, col := range f.cols {
		f.Cursor.Offsets[i] = col.Offset + uintptr(f.pos.ChunkSlot)*col.CompSize
	}
	f.Cursor.EntSlice = f.IDs

	f.remaining -= allocatedSlots
	if f.remaining > 0 {
		f.pos.ChunkIdx++
		f.available = int(f.arch.Table.Layout.ChunkCap)
	}

	return true
}
