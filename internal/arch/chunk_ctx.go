package arch

import (
	"unsafe"

	"github.com/kjkrol/goke/v2/internal/colstore"
)

// ChunkCtx captures the iteration context for a single table chunk.
// Obtained from Matcher.ChunkCtx() during Query.All() iteration;
// passed to CmdBuf.MassMigrate so the Migrator can skip per-entity
// addr.Book lookups for ArchID, Ptr, and Idx.
//
// Ver is the source table's structural version at capture time. When it still
// matches at apply time, no entity of that table was removed or relocated in
// between — every captured id is alive at its captured position, so the
// Migrator can skip per-entity address-book validation entirely.
//
// Prefix is set by CmdBuf.MassMigrate when the ids slice is a leading window
// of the chunk's entity column (ids[i] lives at slot i). Combined with a Ver
// match it removes even the per-entity Slot lookups.
type ChunkCtx struct {
	ArchID ID
	Ptr    unsafe.Pointer
	Idx    colstore.Idx
	Ver    uint32
	Prefix bool
}
