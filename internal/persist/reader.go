package persist

import (
	"compress/gzip"
	"fmt"
	"io"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/colstore"
	"github.com/kjkrol/goke/v3/internal/comp"
)

// Load reads a snapshot written by Save, registering components via comps
// (see [CompRequest]) and repopulating catalog/book with its archetypes and
// entities.
func Load(r io.Reader, defIndex *comp.DefIndex, book *addr.Book, catalog *arch.Catalog, comps []CompRequest) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("persist: %w", err)
	}
	defer gr.Close()

	if err := loadFrom(gr, defIndex, book, catalog, comps); err != nil {
		return err
	}

	// Force the gzip trailer (CRC32 + size) to be checked — our structured
	// read above stops exactly at the payload's logical end and may not
	// otherwise trigger it.
	var probe [1]byte
	n, err := gr.Read(probe[:])
	if err != nil && err != io.EOF {
		return fmt.Errorf("persist: %w", err)
	}
	if n > 0 {
		return fmt.Errorf("persist: unexpected trailing data after save file payload")
	}
	return nil
}

func loadFrom(r io.Reader, defIndex *comp.DefIndex, book *addr.Book, catalog *arch.Catalog, comps []CompRequest) error {
	if err := readHeader(r); err != nil {
		return err
	}

	nextIndex, err := readUint32(r)
	if err != nil {
		return err
	}
	generations, err := readUint32Slice(r)
	if err != nil {
		return err
	}
	freeIndices, err := readUint32Slice(r)
	if err != nil {
		return err
	}

	compCount, err := readUint32(r)
	if err != nil {
		return err
	}
	headers := make([]compHeader, compCount)
	for i := range headers {
		h, err := readComponentHeader(r)
		if err != nil {
			return err
		}
		headers[i] = h
	}

	if err := registerComponents(headers, comps); err != nil {
		return err
	}

	book.RestorePoolState(nextIndex, generations, freeIndices)

	archCount, err := readUint32(r)
	if err != nil {
		return err
	}
	archHeaders := make([]archHeader, archCount)
	for i := range archHeaders {
		h, err := readArchHeader(r)
		if err != nil {
			return err
		}
		archHeaders[i] = h
	}

	for _, ah := range archHeaders {
		if err := loadArchetype(r, defIndex, book, catalog, ah); err != nil {
			return err
		}
	}
	return nil
}

// registerComponents matches comps to headers by Name (not position — the
// header order, driven by the file, dictates registration order) and calls
// Register for each. Requests matching no header are registered afterward,
// for forward compatibility with a save made before that type existed.
func registerComponents(headers []compHeader, comps []CompRequest) error {
	byName := make(map[string]CompRequest, len(comps))
	for _, c := range comps {
		if _, dup := byName[c.Name]; dup {
			panic(fmt.Sprintf("persist: RegisterFor provided more than once for %q", c.Name))
		}
		byName[c.Name] = c
	}

	matched := make(map[string]bool, len(headers))
	for _, h := range headers {
		req, ok := byName[h.Name]
		if !ok {
			return fmt.Errorf("persist: save file needs component %q, but no matching RegisterFor was provided", h.Name)
		}
		size := h.Size
		if err := req.Register(&size); err != nil {
			return err
		}
		matched[h.Name] = true
	}
	for _, c := range comps {
		if !matched[c.Name] {
			if err := c.Register(nil); err != nil {
				return err
			}
		}
	}
	return nil
}

type slotBatch struct {
	ptr   unsafe.Pointer
	start colstore.Slot
	n     int
}

func loadArchetype(r io.Reader, defIndex *comp.DefIndex, book *addr.Book, catalog *arch.Catalog, ah archHeader) error {
	var composition comp.Composition
	for _, id := range ah.CompIDs {
		composition = composition.With(defIndex.ByID(comp.ID(id)))
	}
	archID := catalog.Upsert(composition)
	table := &catalog.Archetypes[archID].Table
	defs := composition.Defs

	batches := reserveBatches(table, int(ah.EntityCount))

	ids := make([]uid.UID64, ah.EntityCount)
	if err := readIDsInto(r, ids); err != nil {
		return err
	}
	offset := 0
	for _, b := range batches {
		table.SetEntityRange(b.ptr, b.start, ids[offset:offset+b.n])
		for i := range b.n {
			id := ids[offset+i]
			if !book.RestoreKnown(id, archID, b.ptr, b.start+colstore.Slot(i)) {
				return fmt.Errorf("persist: corrupt save file: entity %v not recognized by the restored id pool state", id)
			}
		}
		offset += b.n
	}

	for _, def := range defs {
		for _, b := range batches {
			for i := range b.n {
				ptr := table.ComponentAt(b.ptr, b.start+colstore.Slot(i), def.ID)
				if err := DecodeValue(r, def.Type, ptr); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// reserveBatches allocates count slots in table — possibly across several
// chunks — without seeding entity IDs, mirroring ent.Factory's Create/Next
// reserve-once, consume-incrementally pattern: ReserveSlots is called once,
// then subsequent chunks are walked by advancing idx and resetting
// available to chunkCap, not by reserving again.
func reserveBatches(table *colstore.Table, count int) []slotBatch {
	if count == 0 {
		return nil
	}
	idx, available, chunkCap := table.ReserveSlots(count)
	var batches []slotBatch
	remaining := count
	for remaining > 0 {
		n := min(remaining, available)
		ptr, startSlot := table.AllocSlots(idx, n)
		batches = append(batches, slotBatch{ptr: ptr, start: startSlot, n: n})
		remaining -= n
		if remaining > 0 {
			idx++
			available = chunkCap
		}
	}
	table.ReleaseSlots()
	return batches
}

func readIDsInto(r io.Reader, dst []uid.UID64) error {
	for i := range dst {
		n, err := readUint64(r)
		if err != nil {
			return err
		}
		dst[i] = uid.UID64(n)
	}
	return nil
}

func readUint32Slice(r io.Reader) ([]uint32, error) {
	n, err := readUint32(r)
	if err != nil {
		return nil, err
	}
	out := make([]uint32, n)
	for i := range out {
		v, err := readUint32(r)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}
