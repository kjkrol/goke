package ecs

import "unsafe"

// -------------------------------------------------------------

// Command represents a deferred operation on an entity
type commandType int

const (
	cmdAssignComponent commandType = iota
	cmdRemoveComponent
	cmdRemoveEntity
)

type systemCommand struct {
	cType   commandType
	entity  Entity
	compID  ComponentID
	dataPtr unsafe.Pointer
}

// -------------------------------------------------------------

const pageSize = 4096

type SystemCommandBuffer struct {
	commands []systemCommand
	pages    [][]byte
	pageIdx  int
	offset   int
}

func NewSystemCommandBuffer() *SystemCommandBuffer {
	return &SystemCommandBuffer{
		commands: make([]systemCommand, 0, 128),
		pages:    [][]byte{make([]byte, pageSize)},
	}
}

// RecordAssign safely copies component data into the buffer's pool
func (cb *SystemCommandBuffer) AssignComponent(e Entity, info ComponentInfo, data unsafe.Pointer) {
	size := int(info.Size)

	var dest unsafe.Pointer

	if size > 0 && data != nil {
		// 1. Get stable memory address
		dest = cb.reserveSpace(size)
		// 2. Copy the data
		copy(unsafe.Slice((*byte)(dest), size), unsafe.Slice((*byte)(data), size))
	}

	// 3. Queue the command
	cb.commands = append(cb.commands, systemCommand{
		cType:   cmdAssignComponent,
		entity:  e,
		compID:  info.ID,
		dataPtr: dest,
	})
}

func (cb *SystemCommandBuffer) RemoveComponent(e Entity, compID ComponentID) {
	cb.commands = append(cb.commands, systemCommand{
		cType:  cmdRemoveComponent,
		entity: e,
		compID: compID,
	})
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
}

// reserveSpace ensures there is enough contiguous memory in the pages
// and returns a pointer to the start of the reserved block.
func (cb *SystemCommandBuffer) reserveSpace(size int) unsafe.Pointer {
	// Check if the component fits in the remaining space of the current page
	if cb.offset+size > pageSize {
		cb.pageIdx++
		cb.offset = 0

		// Handle allocation of a new page or resizing if the component is huge
		if cb.pageIdx >= len(cb.pages) {
			newPageSize := pageSize
			if size > newPageSize {
				newPageSize = size
			}
			cb.pages = append(cb.pages, make([]byte, newPageSize))
		} else if len(cb.pages[cb.pageIdx]) < size {
			// Resize existing page if it's reused but too small for a large component
			cb.pages[cb.pageIdx] = make([]byte, size)
		}
	}

	ptr := unsafe.Pointer(&cb.pages[cb.pageIdx][cb.offset])
	cb.offset += size
	return ptr
}
