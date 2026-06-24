package orch

import (
	"unsafe"

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

const allocBlockSize = 4096

// CmdBuf as Linear Allocator
type CmdBuf struct {
	cmds    []bufferedCmd
	pages   [][]byte
	pageIdx int
	offset  int
}

func (cb *CmdBuf) Clear() {
	clear(cb.cmds)
	cb.cmds = cb.cmds[:0]

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

func (cb *CmdBuf) reset() {
	cb.cmds = cb.cmds[:0]
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
