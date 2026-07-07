package arch

import (
	"unsafe"

	"github.com/kjkrol/goke/v2/internal/colstore"
)

// ChunkCtx captures the iteration context for a single table chunk.
// Obtained from Matcher.ChunkCtx() during Query.All() iteration;
// passed to CmdBuf.MassMigrate so the Migrator can skip per-entity
// addr.Book lookups for ArchID, Ptr, and Idx.
type ChunkCtx struct {
	ArchID ID
	Ptr    unsafe.Pointer
	Idx    colstore.Idx
}
