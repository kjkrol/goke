package orch

import (
	"reflect"
	"strconv"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
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

type migrateCmd struct {
	op   bulk.Migrator
	snap bulk.ChunkSnapshot
	ids  []uid.UID64
}

type migrateValueCmd struct {
	op      bulk.ValueMigrator
	snap    bulk.ChunkSnapshot
	ids     []uid.UID64
	payload unsafe.Pointer
}

type spawnCmd struct {
	spawner bulk.Spawner
	count   int
	outIDs  *[]uid.UID64
}

const allocBlockSize = 4096

// CmdBuf queues deferred commands, backing their payloads with a linear
// page allocator so registration never heap-allocates once warm.
type CmdBuf struct {
	cmds             []bufferedCmd
	migrateCmds      []migrateCmd
	migrateValueCmds []migrateValueCmd
	spawnCmds        []spawnCmd
	remover          bulk.Migrator
	defs             *comp.DefIndex
	plain            pagePool
}

// pagePool is a linear allocator over a list of reusable pages, holding only
// payloads that carry no pointers.
type pagePool struct {
	pages   [][]byte
	pageIdx int
	offset  int
}

// SetRemover installs the shared Remover that CmdBuf.Remove queues against.
// Called once by Scheduler.Register — not part of the per-tick hot path.
func (cb *CmdBuf) SetRemover(r bulk.Migrator) { cb.remover = r }

// Remover returns the shared Remover installed by SetRemover.
func (cb *CmdBuf) Remover() bulk.Migrator { return cb.remover }

// SetDefs installs the component registry that AddOne resolves comp.IDs
// against. Called once by Scheduler.Register.
func (cb *CmdBuf) SetDefs(defs *comp.DefIndex) { cb.defs = defs }

func (p *pagePool) clearPages() {
	for i := 0; i <= p.pageIdx; i++ {
		if i < len(p.pages) {
			clear(p.pages[i])
		}
	}
	p.pageIdx = 0
	p.offset = 0
}

func (cb *CmdBuf) Clear() {
	clear(cb.cmds)
	cb.cmds = cb.cmds[:0]
	cb.migrateCmds = cb.migrateCmds[:0]
	cb.migrateValueCmds = cb.migrateValueCmds[:0]
	cb.spawnCmds = cb.spawnCmds[:0]

	cb.plain.clearPages()
}

func NewCmdBuf() *CmdBuf {
	return &CmdBuf{
		cmds:  make([]bufferedCmd, 0, 128),
		plain: pagePool{pages: [][]byte{make([]byte, allocBlockSize)}},
	}
}

// AddOne queues an add-component command, staging value for Sync. compID is
// resolved against the registry and must name T.
func AddOne[T any](cb *CmdBuf, entityID uid.UID64, compID comp.ID, value T) {
	size := int(unsafe.Sizeof(value))

	scan := false
	if cb.defs != nil {
		def := cb.defs.ByID(compID)
		if def.Type != reflect.TypeFor[T]() {
			panic("goke: CmdBuf.AddOne: component ID " + strconv.Itoa(int(compID)) +
				" is registered as " + typeName(def.Type) + ", not " + reflect.TypeFor[T]().String())
		}
		scan = def.NeedsScan
	}

	var ptr unsafe.Pointer

	if size > 0 {
		align := int(unsafe.Alignof(value))
		if scan {
			ptr = unsafe.Pointer(new(T))
		} else {
			ptr = cb.reserveSpace(size, align)
		}
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

func (cb *CmdBuf) RemoveCompOne(entityID uid.UID64, compID comp.ID) {
	cb.cmds = append(cb.cmds, bufferedCmd{
		cType:    cmdRemoveComp,
		entityID: entityID,
		compID:   compID,
	})
}

func (cb *CmdBuf) RemoveOne(entityID uid.UID64) {
	cb.cmds = append(cb.cmds, bufferedCmd{
		cType:    cmdRemoveEntity,
		entityID: entityID,
	})
}

// enqueueMigrate copies ids into the page pool and appends a migrateCmd —
// shared by Migrate and Remove, which differ only in what op they carry
// (both satisfy bulk.Migrator).
func (cb *CmdBuf) enqueueMigrate(op bulk.Migrator, snap bulk.ChunkSnapshot, ids []uid.UID64) {
	n := len(ids)
	if n == 0 {
		return
	}
	snap.SlotAligned = unsafe.Pointer(&ids[0]) == snap.ChunkPtr
	var u uid.UID64
	ptr := cb.reserveSpace(n*int(unsafe.Sizeof(u)), int(unsafe.Alignof(u)))
	copied := unsafe.Slice((*uid.UID64)(ptr), n)
	copy(copied, ids)
	cb.migrateCmds = append(cb.migrateCmds, migrateCmd{op: op, snap: snap, ids: copied})
}

// Migrate records a bulk component add/remove of ids from the chunk
// described by snap, setting snap.SlotAligned when ids is a leading window
// of the chunk's entity column. ids is copied into the page pool — the
// caller may reuse it.
func (cb *CmdBuf) Migrate(op bulk.Migrator, snap bulk.ChunkSnapshot, ids []uid.UID64) {
	cb.enqueueMigrate(op, snap, ids)
}

// Remove records a bulk whole-entity removal of ids from the chunk
// described by snap, using the shared Remover installed by SetRemover — the
// caller never builds or holds a Remover, since it carries no per-call
// configuration.
func (cb *CmdBuf) Remove(snap bulk.ChunkSnapshot, ids []uid.UID64) {
	cb.enqueueMigrate(cb.remover, snap, ids)
}

// Spawn records a deferred entity-creation command: at Sync, spawner.Spawn
// is called once with count, and the resulting ids are written into
// *outIDs for a later system (running after this Sync) to pick up via a
// normal Query. Unlike Migrate/Remove/AddCompValue, there are no ids to
// protect from caller reuse — count is a plain value and outIDs is the
// caller's own field — so this never touches the page pool.
func (cb *CmdBuf) Spawn(spawner bulk.Spawner, count int, outIDs *[]uid.UID64) {
	cb.spawnCmds = append(cb.spawnCmds, spawnCmd{spawner: spawner, count: count, outIDs: outIDs})
}

// ReserveIDs reserves capacity for up to n ids in the page pool and returns
// a zero-length slice backed by that reservation (cap == n) — for callers
// that stage ids directly via append instead of building a separate scratch
// slice first, then hand the result to CommitReserved.
func (cb *CmdBuf) ReserveIDs(n int) []uid.UID64 {
	if n == 0 {
		return nil
	}
	var u uid.UID64
	ptr := cb.reserveSpace(n*int(unsafe.Sizeof(u)), int(unsafe.Alignof(u)))
	return unsafe.Slice((*uid.UID64)(ptr), n)[:0]
}

// CommitReserved queues op against ids without copying — ids must be a
// slice obtained from ReserveIDs on this same CmdBuf (already arena-backed,
// safe to store as-is).
func (cb *CmdBuf) CommitReserved(op bulk.Migrator, snap bulk.ChunkSnapshot, ids []uid.UID64) {
	if len(ids) == 0 {
		return
	}
	cb.migrateCmds = append(cb.migrateCmds, migrateCmd{op: op, snap: snap, ids: ids})
}

// AddCompValue records a bulk migration like Migrate, additionally
// reserving n*elemSize bytes (aligned to align) in the page pool for the
// caller to fill with per-id values for op's added component — the
// returned pointer is uninitialized, written into by the caller, and read
// back at Sync when op.MigrateWithValue runs.
func (cb *CmdBuf) AddCompValue(op bulk.ValueMigrator, snap bulk.ChunkSnapshot, ids []uid.UID64, elemSize, align uintptr, elemType reflect.Type, scan bool) unsafe.Pointer {
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
		if scan {
			payloadPtr = reflect.New(reflect.ArrayOf(n, elemType)).UnsafePointer()
		} else {
			payloadPtr = cb.reserveSpace(n*int(elemSize), int(align))
		}
	}

	cb.migrateValueCmds = append(cb.migrateValueCmds, migrateValueCmd{
		op: op, snap: snap, ids: copiedIDs, payload: payloadPtr,
	})
	return payloadPtr
}

func (cb *CmdBuf) reset() {
	cb.cmds = cb.cmds[:0]
	cb.migrateCmds = cb.migrateCmds[:0]
	cb.migrateValueCmds = cb.migrateValueCmds[:0]
	cb.spawnCmds = cb.spawnCmds[:0]
	cb.plain.pageIdx, cb.plain.offset = 0, 0
}

// reserveSpace returns a pointer to a contiguous block of arena space.
func (cb *CmdBuf) reserveSpace(size int, align int) unsafe.Pointer {
	p := &cb.plain

	p.offset = (p.offset + align - 1) &^ (align - 1)

	if p.offset+size > allocBlockSize {
		p.pageIdx++
		p.offset = 0

		if p.pageIdx >= len(p.pages) {
			blockSize := max(size, allocBlockSize)
			p.pages = append(p.pages, make([]byte, blockSize))
		} else if len(p.pages[p.pageIdx]) < size {
			p.pages[p.pageIdx] = make([]byte, size)
		}
	}

	ptr := unsafe.Pointer(&p.pages[p.pageIdx][p.offset])
	p.offset += size
	return ptr
}

// typeName renders a possibly-unregistered component type for an error message.
func typeName(t reflect.Type) string {
	if t == nil {
		return "(no component registered under that ID)"
	}
	return t.String()
}
