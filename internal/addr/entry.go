package addr

import (
	"unsafe"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/colstore"
)

// Entry is the full storage address of an entity: which archetype table it
// belongs to, its slot within that table's chunk memory, and the generation
// that guards against stale access after the ID is recycled.
// Ptr points directly to the chunk's backing memory; it remains valid across
// SwapChunks operations, which is why Idx is not stored.
type Entry struct {
	ArchId arch.ID
	Ptr    unsafe.Pointer
	Slot   colstore.Slot
	Gen    uint32
}
