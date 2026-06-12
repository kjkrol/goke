package exec

import (
	"unsafe"

	"github.com/kjkrol/goke/internal/core"
	"github.com/kjkrol/uid"
)

// -------------------------------------------------------------

// Command represents a deferred operation on an entity
type commandType int

const (
	cmdAssignComponent commandType = iota
	cmdRemoveComponent
	cmdRemoveEntity
)

type systemCommand struct {
	cType    commandType
	entity   uid.UID64
	compInfo core.ComponentInfo
	dataPtr  unsafe.Pointer
}

// -------------------------------------------------------------

const pageSize = 4096

// CommandBuf as Linear Allocator
type CommandBuf struct {
	commands []systemCommand
	pages    [][]byte
	pageIdx  int
	offset   int
}

func (cb *CommandBuf) Clear() {
	clear(cb.commands)
	cb.commands = cb.commands[:0]

	for i := 0; i <= cb.pageIdx; i++ {
		if i < len(cb.pages) {
			clear(cb.pages[i])
		}
	}

	cb.pageIdx = 0
	cb.offset = 0
}

func NewCommandBuf() *CommandBuf {
	return &CommandBuf{
		commands: make([]systemCommand, 0, 128),
		pages:    [][]byte{make([]byte, pageSize)},
	}
}

// AddComponent safely copies component data into the buffer's pool
func AddComponent[T any](cb *CommandBuf, e uid.UID64, info core.ComponentInfo, value T) {
	size := int(unsafe.Sizeof(value))

	var ptr unsafe.Pointer

	if size > 0 {
		align := int(unsafe.Alignof(value))
		ptr = cb.reserveSpace(size, align)
		*(*T)(ptr) = value
	} else {
		ptr = nil
	}

	cb.commands = append(cb.commands, systemCommand{
		cType:    cmdAssignComponent,
		entity:   e,
		compInfo: info,
		dataPtr:  ptr,
	})
}

func RemoveComponent(cb *CommandBuf, e uid.UID64, compInfo core.ComponentInfo) {
	cb.commands = append(cb.commands, systemCommand{
		cType:    cmdRemoveComponent,
		entity:   e,
		compInfo: compInfo,
	})
}

func RemoveEntity(cb *CommandBuf, e uid.UID64) {
	cb.commands = append(cb.commands, systemCommand{
		cType:  cmdRemoveEntity,
		entity: e,
	})
}

func (cb *CommandBuf) reset() {
	cb.commands = cb.commands[:0]
	cb.pageIdx = 0
	cb.offset = 0
}

// reserveSpace ensures there is enough contiguous memory in the pages
// and returns a pointer to the start of the reserved block.
func (cb *CommandBuf) reserveSpace(size int, align int) unsafe.Pointer {
	cb.offset = (cb.offset + align - 1) &^ (align - 1)

	if cb.offset+size > pageSize {
		cb.pageIdx++
		cb.offset = 0

		if cb.pageIdx >= len(cb.pages) {
			newPageSize := pageSize
			if size > newPageSize {
				newPageSize = size
			}
			cb.pages = append(cb.pages, make([]byte, newPageSize))
		} else if len(cb.pages[cb.pageIdx]) < size {
			cb.pages[cb.pageIdx] = make([]byte, size)
		}
	}

	ptr := unsafe.Pointer(&cb.pages[cb.pageIdx][cb.offset])
	cb.offset += size
	return ptr
}
