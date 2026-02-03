package core

import "unsafe"

// -------------------------------------------------------------

// Command represents a deferred operation on an entity
type commandType int

const (
	cmdAssignComponent commandType = iota
	cmdRemoveComponent
	cmdRemoveEntity
	cmdCreateEntity
)

type systemCommand struct {
	cType    commandType
	entity   Entity
	compInfo ComponentInfo
	dataPtr  unsafe.Pointer
}

// -------------------------------------------------------------

const pageSize = 4096

// Linear Allocator
type SystemCommandBuffer struct {
	commands      []systemCommand
	pages         [][]byte
	pageIdx       int
	offset        int
	nextVirtualID uint32
}

func NewSystemCommandBuffer() *SystemCommandBuffer {
	return &SystemCommandBuffer{
		commands: make([]systemCommand, 0, 128),
		pages:    [][]byte{make([]byte, pageSize)},
	}
}

// AssignComponent safely copies component data into the buffer's pool
func AssignComponent[T any](cb *SystemCommandBuffer, e Entity, info ComponentInfo, value T) {
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

func (cb *SystemCommandBuffer) RemoveComponent(e Entity, compInfo ComponentInfo) {
	cb.commands = append(cb.commands, systemCommand{
		cType:    cmdRemoveComponent,
		entity:   e,
		compInfo: compInfo,
	})
}

func (cb *SystemCommandBuffer) CreateEntity() Entity {
	vEntity := NewVirtualEntity(cb.nextVirtualID)
	cb.nextVirtualID++

	cb.commands = append(cb.commands, systemCommand{
		cType:  cmdCreateEntity,
		entity: vEntity,
	})
	return vEntity
}

func (cb *SystemCommandBuffer) RemoveEntity(e Entity) {
	cb.commands = append(cb.commands, systemCommand{
		cType:  cmdRemoveEntity,
		entity: e,
	})
}

func (cb *SystemCommandBuffer) reset() {
	cb.commands = cb.commands[:0]
	cb.pageIdx = 0
	cb.offset = 0
	cb.nextVirtualID = 0
}

// reserveSpace ensures there is enough contiguous memory in the pages
// and returns a pointer to the start of the reserved block.
func (cb *SystemCommandBuffer) reserveSpace(size int, align int) unsafe.Pointer {
	// 1. Align the current offset
	// This moves the offset to the next multiple of 'align'
	cb.offset = (cb.offset + align - 1) &^ (align - 1)

	// 2. Check if it fits in the current page after alignment
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
