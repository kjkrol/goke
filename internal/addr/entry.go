package addr

import (
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/colstore"
)

// Entry is the full storage address of an entity; Gen guards against stale
// access after ID recycling. ChunkPtr addresses the chunk's backing memory
// directly and stays valid across SwapChunks — hence no chunk index is stored.
type Entry struct {
	ArchID   arch.ID
	ChunkPtr unsafe.Pointer
	Slot     colstore.Slot
	Gen      uint32
}
