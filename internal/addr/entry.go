package addr

import (
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/mem"
)

// Entry is the full storage address of an entity: which archetype table it
// belongs to, its position within that table, and the generation that guards
// against stale access after the ID is recycled.
type Entry struct {
	ArchId     arch.ID
	Pos        mem.BlockPos
	Generation uint32
}
