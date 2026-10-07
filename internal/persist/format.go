package persist

import (
	"fmt"
	"io"
	"math"

	"github.com/kjkrol/goke/v3/internal/comp"
)

// Magic identifies a goke save file; the first bytes written by Save and
// checked by Load.
const Magic = "GKSV"

// FormatVersion is the current save-file format version.
const FormatVersion uint32 = 1

func writeHeader(w io.Writer) error {
	if _, err := io.WriteString(w, Magic); err != nil {
		return err
	}
	return writeUint32(w, FormatVersion)
}

func readHeader(r io.Reader) error {
	magic := make([]byte, len(Magic))
	if _, err := io.ReadFull(r, magic); err != nil {
		return fmt.Errorf("persist: reading magic: %w", err)
	}
	if string(magic) != Magic {
		return fmt.Errorf("persist: not a goke save file (bad magic %q)", magic)
	}
	version, err := readUint32(r)
	if err != nil {
		return fmt.Errorf("persist: reading format version: %w", err)
	}
	if version != FormatVersion {
		return fmt.Errorf("persist: unsupported save file version %d (this build supports %d)", version, FormatVersion)
	}
	return nil
}

// compHeader is one entry in the save file's component directory: the
// recorded Go type name and layout, in comp.ID order.
type compHeader struct {
	Name  string
	Size  uint32
	Align uint32
}

func writeComponentHeader(w io.Writer, h compHeader) error {
	if err := writeBytes(w, []byte(h.Name)); err != nil {
		return err
	}
	if err := writeUint32(w, h.Size); err != nil {
		return err
	}
	return writeUint32(w, h.Align)
}

func readComponentHeader(r io.Reader) (compHeader, error) {
	name, err := readBytes(r)
	if err != nil {
		return compHeader{}, err
	}
	size, err := readUint32(r)
	if err != nil {
		return compHeader{}, err
	}
	align, err := readUint32(r)
	if err != nil {
		return compHeader{}, err
	}
	return compHeader{Name: string(name), Size: size, Align: align}, nil
}

// archHeader is one entry in the save file's archetype directory: its
// composition (as component IDs, including tags, in the archetype's own order —
// its data columns as it holds them, then its tags — the order its values are
// written in) and live entity count.
type archHeader struct {
	CompIDs     []comp.ID
	EntityCount uint32
}

// A component ID is written in two bytes: the build fails once MaxComponents outgrows them.
var _ [math.MaxUint16 + 1 - comp.MaxComponents]struct{}

func writeArchHeader(w io.Writer, h archHeader) error {
	if err := writeUint32(w, uint32(len(h.CompIDs))); err != nil {
		return err
	}
	for _, id := range h.CompIDs {
		if err := writeUint16(w, uint16(id)); err != nil {
			return err
		}
	}
	return writeUint32(w, h.EntityCount)
}

func readArchHeader(r io.Reader) (archHeader, error) {
	n, err := readUint32(r)
	if err != nil {
		return archHeader{}, err
	}
	ids := make([]comp.ID, n)
	for i := range ids {
		id, err := readUint16(r)
		if err != nil {
			return archHeader{}, err
		}
		ids[i] = comp.ID(id)
	}
	count, err := readUint32(r)
	if err != nil {
		return archHeader{}, err
	}
	return archHeader{CompIDs: ids, EntityCount: count}, nil
}
