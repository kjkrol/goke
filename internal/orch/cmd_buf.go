package orch

import (
	"unsafe"

	"github.com/kjkrol/goke/v2/internal/bulk"
	"github.com/kjkrol/goke/v2/internal/comp"
	"github.com/kjkrol/uid"
)

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

type massMigrateCmd struct {
	migrator bulk.Migrator
	snap     bulk.ChunkSnapshot
	ids      []uid.UID64
}

type massMigrateValueCmd struct {
	migrator bulk.ValueMigrator
	snap     bulk.ChunkSnapshot
	ids      []uid.UID64
	payload  unsafe.Pointer
}

const allocBlockSize = 4096

// CmdBuf queues deferred commands, backing their payloads with a linear
// page allocator so registration never heap-allocates once warm.
type CmdBuf struct {
	cmds          []bufferedCmd
	massCmds      []massMigrateCmd
	massValueCmds []massMigrateValueCmd
	pages         [][]byte
	pageIdx       int
	offset        int
}

func (cb *CmdBuf) Clear() {
	clear(cb.cmds)
	cb.cmds = cb.cmds[:0]
	cb.massCmds = cb.massCmds[:0]
	cb.massValueCmds = cb.massValueCmds[:0]

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

// AddComp queues an add-component command, copying value into the page pool.
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

// MassMigrate records a bulk migration of ids from the chunk described by
// snap, setting snap.SlotAligned when ids is a leading window of the chunk's
// entity column. ids is copied into the page pool — the caller may reuse it.
func (cb *CmdBuf) MassMigrate(migrator bulk.Migrator, snap bulk.ChunkSnapshot, ids []uid.UID64) {
	n := len(ids)
	if n == 0 {
		return
	}
	snap.SlotAligned = unsafe.Pointer(&ids[0]) == snap.ChunkPtr
	var u uid.UID64
	ptr := cb.reserveSpace(n*int(unsafe.Sizeof(u)), int(unsafe.Alignof(u)))
	copied := unsafe.Slice((*uid.UID64)(ptr), n)
	copy(copied, ids)
	cb.massCmds = append(cb.massCmds, massMigrateCmd{migrator: migrator, snap: snap, ids: copied})
}

// MassMigrateInit records a bulk migration like MassMigrate, additionally
// reserving n*elemSize bytes (aligned to align) in the page pool for the
// caller to fill with per-id values for the migrator's added component — the
// returned pointer is uninitialized, written into by the caller, and read
// back at Sync when migrator.MigrateWithValue runs.
func (cb *CmdBuf) MassMigrateInit(migrator bulk.ValueMigrator, snap bulk.ChunkSnapshot, ids []uid.UID64, elemSize, align uintptr) unsafe.Pointer {
	n := len(ids)
	if n == 0 {
		return nil
	}
	snap.SlotAligned = unsafe.Pointer(&ids[0]) == snap.ChunkPtr
	var u uid.UID64
	idPtr := cb.reserveSpace(n*int(unsafe.Sizeof(u)), int(unsafe.Alignof(u)))
	copiedIDs := unsafe.Slice((*uid.UID64)(idPtr), n)
	copy(copiedIDs, ids)

	var payloadPtr unsafe.Pointer
	if elemSize > 0 {
		payloadPtr = cb.reserveSpace(n*int(elemSize), int(align))
	}

	cb.massValueCmds = append(cb.massValueCmds, massMigrateValueCmd{
		migrator: migrator, snap: snap, ids: copiedIDs, payload: payloadPtr,
	})
	return payloadPtr
}

func (cb *CmdBuf) reset() {
	cb.cmds = cb.cmds[:0]
	cb.massCmds = cb.massCmds[:0]
	cb.massValueCmds = cb.massValueCmds[:0]
	cb.pageIdx = 0
	cb.offset = 0
}

// reserveSpace returns a pointer to a contiguous block from the page pool.
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
