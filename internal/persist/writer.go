package persist

import (
	"compress/gzip"
	"io"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/colstore"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/iter"
)

// Save writes a full snapshot of the world — entity ID pool bookkeeping,
// component definitions, archetype compositions, and per-entity component
// data — to w, gzip-compressed.
func Save(w io.Writer, defIndex *comp.DefIndex, book *addr.Book, catalog *arch.Catalog) error {
	gw := gzip.NewWriter(w)
	if err := saveTo(gw, defIndex, book, catalog); err != nil {
		return err
	}
	return gw.Close()
}

func saveTo(w io.Writer, defIndex *comp.DefIndex, book *addr.Book, catalog *arch.Catalog) error {
	if err := writeHeader(w); err != nil {
		return err
	}

	nextIndex, generations, freeIndices := book.PoolState()
	if err := writeUint32(w, nextIndex); err != nil {
		return err
	}
	if err := writeUint32Slice(w, generations); err != nil {
		return err
	}
	if err := writeUint32Slice(w, freeIndices); err != nil {
		return err
	}

	if err := writeComponentDirectory(w, defIndex); err != nil {
		return err
	}

	lives := liveArchetypes(catalog)
	if err := writeUint32(w, uint32(len(lives))); err != nil {
		return err
	}
	for _, archID := range lives {
		a := &catalog.Archetypes[archID]
		var ids []uint8
		for id := range a.Mask().AllSet() {
			ids = append(ids, uint8(id))
		}
		h := archHeader{CompIDs: ids, EntityCount: uint32(a.Len())}
		if err := writeArchHeader(w, h); err != nil {
			return err
		}
	}

	for _, archID := range lives {
		if err := saveArchetypeData(w, &catalog.Archetypes[archID]); err != nil {
			return err
		}
	}

	return nil
}

func writeComponentDirectory(w io.Writer, defIndex *comp.DefIndex) error {
	count := defIndex.Count()
	if err := writeUint32(w, uint32(count)); err != nil {
		return err
	}
	for i := range count {
		def := defIndex.ByID(comp.ID(i))
		h := compHeader{Name: def.Type.String(), Size: uint32(def.Size), Align: uint32(def.Align)}
		if err := writeComponentHeader(w, h); err != nil {
			return err
		}
	}
	return nil
}

// liveArchetypes returns the IDs of every archetype currently holding at
// least one entity, in ascending ID order.
func liveArchetypes(catalog *arch.Catalog) []arch.ID {
	var lives []arch.ID
	for archID := arch.RootID; archID < catalog.Len(); archID++ {
		if catalog.Archetypes[archID].Len() > 0 {
			lives = append(lives, archID)
		}
	}
	return lives
}

// saveArchetypeData writes a's entities as two flat, entity-count-long
// passes over its table — first every id, then every data-bearing
// component's values, one component at a time — decoupled from a's own
// chunk boundaries so a freshly-created table on Load need not replicate
// them.
func saveArchetypeData(w io.Writer, a *arch.Archetype) error {
	table := &a.Table
	defs := a.Composition().Defs

	if err := walkTableChunks(table, func(cur *iter.Cursor) error {
		return writeIDs(w, cur.IDs)
	}); err != nil {
		return err
	}

	for _, def := range defs {
		if err := walkTableChunks(table, func(cur *iter.Cursor) error {
			for slot := range len(cur.IDs) {
				ptr := table.ComponentAt(cur.Base, colstore.Slot(slot), def.ID)
				if err := EncodeValue(w, def.Type, ptr); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// walkTableChunks calls fn once per non-empty chunk of table, in order.
func walkTableChunks(table *colstore.Table, fn func(cur *iter.Cursor) error) error {
	var cur iter.Cursor
	for from := 0; ; {
		idx, ok := table.FillCursorNext(&cur, from, nil)
		if !ok {
			return nil
		}
		if err := fn(&cur); err != nil {
			return err
		}
		from = idx + 1
	}
}

func writeIDs(w io.Writer, ids []uid.UID64) error {
	for _, id := range ids {
		if err := writeUint64(w, uint64(id)); err != nil {
			return err
		}
	}
	return nil
}

func writeUint32Slice(w io.Writer, s []uint32) error {
	if err := writeUint32(w, uint32(len(s))); err != nil {
		return err
	}
	for _, v := range s {
		if err := writeUint32(w, v); err != nil {
			return err
		}
	}
	return nil
}
