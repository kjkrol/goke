package orch

import (
	"unsafe"

	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/uid"
)

// -------------------------------------------------------------

// cmdType represents the kind of deferred operation on an entity
type cmdType int

const (
	cmdAssignComp cmdType = iota
	cmdRemoveComp
	cmdRemoveEntity
)

type bufferedCmd struct {
	cType    cmdType
	entityID uid.UID64
	compID   comp.ID
	size     uintptr
	dataPtr  unsafe.Pointer
}

// -------------------------------------------------------------

// BulkMigrator is satisfied by any type that can apply a bulk archetype
// migration to a batch of entity IDs from a single source chunk.
type BulkMigrator interface {
	ApplyChunk(ctx arch.ChunkCtx, ids []uid.UID64)
}

type massMigrateCmd struct {
	migrator BulkMigrator
	ctx      arch.ChunkCtx
	ids      []uid.UID64
}

// -------------------------------------------------------------

const allocBlockSize = 4096

// CmdBuf as Linear Allocator
type CmdBuf struct {
	cmds     []bufferedCmd
	massCmds []massMigrateCmd
	pages    [][]byte
	pageIdx  int
	offset   int
}

func (cb *CmdBuf) Clear() {
	clear(cb.cmds)
	cb.cmds = cb.cmds[:0]
	cb.massCmds = cb.massCmds[:0]

	for i := 0; i <= cb.pageIdx; i++ {
		if i < len(cb.pages) {
			clear(cb.pages[i])
		}
	}

	cb.pageIdx = 0
	cb.offset = 0
}

func NewCmdBuf() *CmdBuf {
	return &CmdBuf{
		cmds:  make([]bufferedCmd, 0, 128),
		pages: [][]byte{make([]byte, allocBlockSize)},
	}
}

// AddComp safely copies component data into the buffer's pool
func AddComp[T any](cb *CmdBuf, entityID uid.UID64, compID comp.ID, value T) {
	size := int(unsafe.Sizeof(value))

	var ptr unsafe.Pointer

	if size > 0 {
		align := int(unsafe.Alignof(value))
		ptr = cb.reserveSpace(size, align)
		*(*T)(ptr) = value
	} else {
		ptr = nil
	}

	cb.cmds = append(cb.cmds, bufferedCmd{
		cType:    cmdAssignComp,
		entityID: entityID,
		compID:   compID,
		size:     uintptr(size),
		dataPtr:  ptr,
	})
}

func (cb *CmdBuf) RemoveComp(entityID uid.UID64, compID comp.ID) {
	cb.cmds = append(cb.cmds, bufferedCmd{
		cType:    cmdRemoveComp,
		entityID: entityID,
		compID:   compID,
	})
}

func (cb *CmdBuf) RemoveEntity(entityID uid.UID64) {
	cb.cmds = append(cb.cmds, bufferedCmd{
		cType:    cmdRemoveEntity,
		entityID: entityID,
	})
}

// MassMigrate records a bulk migration of ids from a single source chunk.
// ctx must come from Matcher.ChunkCtx() called immediately after Next() returns
// true — it captures the chunk's ArchID, Ptr, and Idx so the Migrator can skip
// per-entity addr.Book lookups for those fields.
// The ids slice is copied into the buffer's page pool; the caller may reuse or
// modify it after this call. No heap allocation once the pool pages are warm.
func (cb *CmdBuf) MassMigrate(migrator BulkMigrator, ctx arch.ChunkCtx, ids []uid.UID64) {
	n := len(ids)
	if n == 0 {
		return
	}
	var u uid.UID64
	ptr := cb.reserveSpace(n*int(unsafe.Sizeof(u)), int(unsafe.Alignof(u)))
	copied := unsafe.Slice((*uid.UID64)(ptr), n)
	copy(copied, ids)
	cb.massCmds = append(cb.massCmds, massMigrateCmd{migrator: migrator, ctx: ctx, ids: copied})
}

func (cb *CmdBuf) reset() {
	cb.cmds = cb.cmds[:0]
	cb.massCmds = cb.massCmds[:0]
	cb.pageIdx = 0
	cb.offset = 0
}

// reserveSpace ensures there is enough contiguous memory in the pages
// and returns a pointer to the start of the reserved block.
func (cb *CmdBuf) reserveSpace(size int, align int) unsafe.Pointer {
	cb.offset = (cb.offset + align - 1) &^ (align - 1)

	if cb.offset+size > allocBlockSize {
		cb.pageIdx++
		cb.offset = 0

		if cb.pageIdx >= len(cb.pages) {
			blockSize := max(size, allocBlockSize)
			cb.pages = append(cb.pages, make([]byte, blockSize))
		} else if len(cb.pages[cb.pageIdx]) < size {
			cb.pages[cb.pageIdx] = make([]byte, size)
		}
	}

	ptr := unsafe.Pointer(&cb.pages[cb.pageIdx][cb.offset])
	cb.offset += size
	return ptr
}
