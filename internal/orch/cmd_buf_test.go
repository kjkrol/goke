package orch

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/bulk"
	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/internal/query"
	"github.com/kjkrol/goke/v3/internal/reg"
	"github.com/kjkrol/goke/v3/iter"
	"github.com/kjkrol/uid"
)

type mockCompA struct {
	Val int
}

type mockCompB struct {
	Msg string
}

type modifyTestSystem struct {
	compA  comp.ID
	compB  comp.ID
	target uid.UID64
}

func (s *modifyTestSystem) Update(cb *CmdBuf, d time.Duration) {
	cb.RemoveCompOne(s.target, s.compA)
	AddOne(cb, s.target, s.compB, mockCompB{Msg: "added"})
}

// AddOne must not try to copy any bytes for a zero-size (tag) component —
// it should queue a command with a nil dataPtr instead of dereferencing a
// zero-size value's address.
func TestAddOne_ZeroSizeComponent(t *testing.T) {
	type tag struct{}
	cb := NewCmdBuf()

	AddOne(cb, uid.UID64(1), comp.ID(5), tag{})

	if len(cb.cmds) != 1 {
		t.Fatalf("expected 1 queued command, got %d", len(cb.cmds))
	}
	cmd := cb.cmds[0]
	if cmd.dataPtr != nil {
		t.Errorf("expected nil dataPtr for a zero-size component, got %v", cmd.dataPtr)
	}
	if cmd.size != 0 {
		t.Errorf("expected size 0, got %d", cmd.size)
	}
}

func TestCmdBuf_ComponentCmds(t *testing.T) {
	var registry reg.Registry
	registry.Init(reg.Config{
		Entity:  ent.Config{Cap: 100, FreeCap: 100},
		Matcher: query.Config{Cap: 10},
	})

	compA := registry.RegComp(reflect.TypeFor[mockCompA]())
	compB := registry.RegComp(reflect.TypeFor[mockCompB]())

	sched := NewScheduler(&registry)

	var colA iter.ArrayRef[mockCompA]
	f := registry.CreateFactory(comp.Add(&colA), comp.Add(new(iter.ArrayRef[mockCompB])))
	f.Create(1)
	f.Next()
	e := f.IDs[0]
	*colA.At(&f.Cursor) = mockCompA{Val: 100}

	sys := &modifyTestSystem{
		compA:  compA,
		compB:  compB,
		target: e,
	}

	sched.Register(sys, NewCmdBuf())
	sched.Run(sys, 0)

	err := sched.Sync()
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// compA should be gone: a matcher requiring it must not match e
	matcherA := registry.AddMatcher(comp.Include[mockCompA]())
	if matcherA.Pick([]uid.UID64{e}).Next() {
		t.Errorf("Expected compA to be removed")
	}

	// compB should be present with the assigned value
	var colB iter.ArrayRef[mockCompB]
	matcherB := registry.AddMatcher(comp.Track(&colB))
	if !matcherB.Pick([]uid.UID64{e}).Next() {
		t.Fatalf("Expected compB to be present")
	}
	if colB.At(&matcherB.Cursor).Msg != "added" {
		t.Errorf("Expected compB.Msg to be 'added'")
	}
}

func TestCmdBuf_Clear(t *testing.T) {
	cb := NewCmdBuf()

	cb.RemoveOne(uid.UID64(1))

	if len(cb.cmds) == 0 {
		t.Fatalf("Expected commands to not be empty")
	}

	cb.Clear()

	if len(cb.cmds) != 0 {
		t.Errorf("Expected commands to be empty")
	}
}

func TestCmdBuf_ReserveSpace(t *testing.T) {
	cb := NewCmdBuf()

	p1 := cb.reserveSpace(10, 1)
	if p1 == nil {
		t.Errorf("Expected non-nil pointer")
	}

	p2 := cb.reserveSpace(8192, 1)
	if p2 == nil {
		t.Errorf("Expected non-nil pointer for large alloc")
	}
}

func TestCmdBuf_Remover(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubMigrator{}

	cb.SetRemover(m)

	if got := cb.Remover(); got != bulk.Migrator(m) {
		t.Errorf("expected Remover() to return the value set via SetRemover, got %v", got)
	}
}

func TestCmdBuf_ReserveIDs_Zero(t *testing.T) {
	cb := NewCmdBuf()

	got := cb.ReserveIDs(0)
	if got != nil {
		t.Errorf("expected nil for n=0, got %v", got)
	}
}

func TestCmdBuf_ReserveIDs_ThenCommitReserved(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubMigrator{}

	ids := cb.ReserveIDs(3)
	if len(ids) != 0 || cap(ids) < 3 {
		t.Fatalf("expected zero-length slice with cap >= 3, got len=%d cap=%d", len(ids), cap(ids))
	}
	ids = append(ids, 1, 2, 3)

	cb.CommitReserved(m, bulk.ChunkSnapshot{}, ids)

	if len(cb.migrateCmds) != 1 {
		t.Fatalf("expected 1 migrateCmd, got %d", len(cb.migrateCmds))
	}
	if len(cb.migrateCmds[0].ids) != 3 {
		t.Errorf("expected 3 ids in queued cmd, got %d", len(cb.migrateCmds[0].ids))
	}
}

func TestCmdBuf_CommitReserved_Empty_NoCommand(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubMigrator{}

	cb.CommitReserved(m, bulk.ChunkSnapshot{}, nil)
	cb.CommitReserved(m, bulk.ChunkSnapshot{}, []uid.UID64{})

	if len(cb.migrateCmds) != 0 {
		t.Errorf("expected no commands for empty ids, got %d", len(cb.migrateCmds))
	}
}

// --- Migrate ---

type stubMigrator struct {
	applyCalls int
	got        []uid.UID64
}

func (s *stubMigrator) Migrate(_ bulk.ChunkSnapshot, ids []uid.UID64) {
	s.applyCalls++
	s.got = append(s.got, ids...)
}

func TestCmdBuf_Migrate_QueuesCommand(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubMigrator{}

	cb.Migrate(m, bulk.ChunkSnapshot{}, []uid.UID64{1, 2, 3})

	if len(cb.migrateCmds) != 1 {
		t.Fatalf("expected 1 migrateCmd, got %d", len(cb.migrateCmds))
	}
	if len(cb.migrateCmds[0].ids) != 3 {
		t.Errorf("expected 3 ids in queued cmd, got %d", len(cb.migrateCmds[0].ids))
	}
}

func TestCmdBuf_Migrate_CopiesIDs(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubMigrator{}
	ids := []uid.UID64{10, 20, 30}

	cb.Migrate(m, bulk.ChunkSnapshot{}, ids)
	ids[0] = 99

	if cb.migrateCmds[0].ids[0] != 10 {
		t.Errorf("Migrate must copy ids; mutated original changed queued value to %v", cb.migrateCmds[0].ids[0])
	}
}

func TestCmdBuf_Migrate_Empty_NoCommand(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubMigrator{}

	cb.Migrate(m, bulk.ChunkSnapshot{}, nil)
	cb.Migrate(m, bulk.ChunkSnapshot{}, []uid.UID64{})

	if len(cb.migrateCmds) != 0 {
		t.Errorf("expected no commands for empty id slices, got %d", len(cb.migrateCmds))
	}
}

func TestCmdBuf_Clear_ResetsMigrateCmds(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubMigrator{}
	cb.Migrate(m, bulk.ChunkSnapshot{}, []uid.UID64{1, 2})

	cb.Clear()

	if len(cb.migrateCmds) != 0 {
		t.Errorf("expected migrateCmds empty after Clear, got %d", len(cb.migrateCmds))
	}
}

// --- Remove ---

func TestCmdBuf_Remove_QueuesAgainstSetRemover(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubMigrator{}
	cb.SetRemover(m)

	cb.Remove(bulk.ChunkSnapshot{}, []uid.UID64{1, 2, 3})

	if len(cb.migrateCmds) != 1 {
		t.Fatalf("expected 1 migrateCmd, got %d", len(cb.migrateCmds))
	}
	if cb.migrateCmds[0].op != bulk.Migrator(m) {
		t.Error("expected queued cmd's op to be the remover installed via SetRemover")
	}
	if len(cb.migrateCmds[0].ids) != 3 {
		t.Errorf("expected 3 ids in queued cmd, got %d", len(cb.migrateCmds[0].ids))
	}
}

func TestCmdBuf_Remove_Empty_NoCommand(t *testing.T) {
	cb := NewCmdBuf()
	cb.SetRemover(&stubMigrator{})

	cb.Remove(bulk.ChunkSnapshot{}, nil)

	if len(cb.migrateCmds) != 0 {
		t.Errorf("expected no commands for an empty id slice, got %d", len(cb.migrateCmds))
	}
}

// --- Spawn ---

type stubSpawner struct {
	gotCount int
	ids      []uid.UID64
}

func (s *stubSpawner) Spawn(count int) []uid.UID64 {
	s.gotCount = count
	return s.ids
}

func TestCmdBuf_Spawn_QueuesCommand(t *testing.T) {
	cb := NewCmdBuf()
	sp := &stubSpawner{}
	var outIDs []uid.UID64

	cb.Spawn(sp, 5, &outIDs)

	if len(cb.spawnCmds) != 1 {
		t.Fatalf("expected 1 spawnCmd, got %d", len(cb.spawnCmds))
	}
	if cb.spawnCmds[0].count != 5 {
		t.Errorf("expected count 5, got %d", cb.spawnCmds[0].count)
	}
}

func TestCmdBuf_Clear_ResetsSpawnCmds(t *testing.T) {
	cb := NewCmdBuf()
	var outIDs []uid.UID64
	cb.Spawn(&stubSpawner{}, 1, &outIDs)

	cb.Clear()

	if len(cb.spawnCmds) != 0 {
		t.Errorf("expected spawnCmds empty after Clear, got %d", len(cb.spawnCmds))
	}
}

func TestScheduler_Sync_AppliesSpawnCmds(t *testing.T) {
	var registry reg.Registry
	registry.Init(reg.Config{
		Entity:  ent.Config{Cap: 10, FreeCap: 10},
		Matcher: query.Config{Cap: 4},
	})
	sched := NewScheduler(&registry)

	sp := &stubSpawner{ids: []uid.UID64{7, 8, 9}}
	var outIDs []uid.UID64
	sys := &fnRunnable{fn: func(cb *CmdBuf, d time.Duration) {
		cb.Spawn(sp, 3, &outIDs)
	}}
	sched.Register(sys, NewCmdBuf())
	sched.Run(sys, 0)

	if err := sched.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if sp.gotCount != 3 {
		t.Errorf("expected Spawn called with count 3, got %d", sp.gotCount)
	}
	if len(outIDs) != 3 || outIDs[0] != 7 || outIDs[1] != 8 || outIDs[2] != 9 {
		t.Errorf("expected outIDs to be filled with spawner's result, got %v", outIDs)
	}
}

// --- AddCompValue ---

type stubValueMigrator struct {
	applyCalls int
	got        []uid.UID64
	gotPayload unsafe.Pointer
}

func (s *stubValueMigrator) MigrateWithValue(_ bulk.ChunkSnapshot, ids []uid.UID64, payload unsafe.Pointer) {
	s.applyCalls++
	s.got = append(s.got, ids...)
	s.gotPayload = payload
}

func TestCmdBuf_AddCompValue_QueuesCommand(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubValueMigrator{}

	ptr := cb.AddCompValue(m, bulk.ChunkSnapshot{}, []uid.UID64{1, 2, 3}, unsafe.Sizeof(int32(0)), unsafe.Alignof(int32(0)))

	if len(cb.migrateValueCmds) != 1 {
		t.Fatalf("expected 1 massMigrateValueCmd, got %d", len(cb.migrateValueCmds))
	}
	if len(cb.migrateValueCmds[0].ids) != 3 {
		t.Errorf("expected 3 ids in queued cmd, got %d", len(cb.migrateValueCmds[0].ids))
	}
	if ptr == nil {
		t.Error("expected a non-nil reserved payload pointer for a positive elemSize")
	}
	if cb.migrateValueCmds[0].payload != ptr {
		t.Error("queued cmd's payload must be the same pointer returned to the caller")
	}
}

func TestCmdBuf_AddCompValue_CopiesIDs(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubValueMigrator{}
	ids := []uid.UID64{10, 20, 30}

	cb.AddCompValue(m, bulk.ChunkSnapshot{}, ids, unsafe.Sizeof(int32(0)), unsafe.Alignof(int32(0)))
	ids[0] = 99

	if cb.migrateValueCmds[0].ids[0] != 10 {
		t.Errorf("AddCompValue must copy ids; mutated original changed queued value to %v", cb.migrateValueCmds[0].ids[0])
	}
}

func TestCmdBuf_AddCompValue_Empty_NoCommand(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubValueMigrator{}

	if ptr := cb.AddCompValue(m, bulk.ChunkSnapshot{}, nil, unsafe.Sizeof(int32(0)), unsafe.Alignof(int32(0))); ptr != nil {
		t.Error("expected nil payload pointer for a nil id slice")
	}
	cb.AddCompValue(m, bulk.ChunkSnapshot{}, []uid.UID64{}, unsafe.Sizeof(int32(0)), unsafe.Alignof(int32(0)))

	if len(cb.migrateValueCmds) != 0 {
		t.Errorf("expected no commands for empty id slices, got %d", len(cb.migrateValueCmds))
	}
}

func TestCmdBuf_AddCompValue_ZeroElemSize_NoPayloadReserved(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubValueMigrator{}

	ptr := cb.AddCompValue(m, bulk.ChunkSnapshot{}, []uid.UID64{1, 2}, 0, 1)

	if ptr != nil {
		t.Error("expected nil payload pointer when elemSize is 0 (zero-sized added component)")
	}
	if len(cb.migrateValueCmds) != 1 {
		t.Fatalf("expected the command to still be queued (ids matter even with no payload), got %d", len(cb.migrateValueCmds))
	}
}

func TestCmdBuf_Clear_ResetsMigrateValueCmds(t *testing.T) {
	cb := NewCmdBuf()
	m := &stubValueMigrator{}
	cb.AddCompValue(m, bulk.ChunkSnapshot{}, []uid.UID64{1, 2}, unsafe.Sizeof(int32(0)), unsafe.Alignof(int32(0)))

	cb.Clear()

	if len(cb.migrateValueCmds) != 0 {
		t.Errorf("expected migrateValueCmds empty after Clear, got %d", len(cb.migrateValueCmds))
	}
}

type addCompValueTestSystem struct {
	migrator bulk.ValueMigrator
	batches  [][]uid.UID64
}

func (s *addCompValueTestSystem) Update(cb *CmdBuf, d time.Duration) {
	for _, batch := range s.batches {
		cb.AddCompValue(s.migrator, bulk.ChunkSnapshot{}, batch, unsafe.Sizeof(int32(0)), unsafe.Alignof(int32(0)))
	}
}

func TestScheduler_Sync_AppliesAddCompValueCmds(t *testing.T) {
	var registry reg.Registry
	registry.Init(reg.Config{
		Entity:  ent.Config{Cap: 10, FreeCap: 10},
		Matcher: query.Config{Cap: 4},
	})
	sched := NewScheduler(&registry)

	m := &stubValueMigrator{}
	sys := &addCompValueTestSystem{migrator: m, batches: [][]uid.UID64{{1, 2, 3}}}
	sched.Register(sys, NewCmdBuf())
	sched.Run(sys, 0)

	if err := sched.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if m.applyCalls != 1 {
		t.Errorf("expected MigrateWithValue called once, got %d", m.applyCalls)
	}
	if len(m.got) != 3 {
		t.Errorf("expected 3 ids passed to MigrateWithValue, got %d", len(m.got))
	}
	if m.gotPayload == nil {
		t.Error("expected a non-nil payload pointer to reach MigrateWithValue")
	}
}

type migrateTestSystem struct {
	migrator bulk.Migrator
	batches  [][]uid.UID64
}

func (s *migrateTestSystem) Update(cb *CmdBuf, d time.Duration) {
	for _, batch := range s.batches {
		cb.Migrate(s.migrator, bulk.ChunkSnapshot{}, batch)
	}
}

func newMigrateSystem(m bulk.Migrator, batches ...[]uid.UID64) *migrateTestSystem {
	return &migrateTestSystem{migrator: m, batches: batches}
}

func TestScheduler_Sync_AppliesMigrateCmds(t *testing.T) {
	var registry reg.Registry
	registry.Init(reg.Config{
		Entity:  ent.Config{Cap: 10, FreeCap: 10},
		Matcher: query.Config{Cap: 4},
	})
	sched := NewScheduler(&registry)

	m := &stubMigrator{}
	sys := newMigrateSystem(m, []uid.UID64{1, 2, 3})
	sched.Register(sys, NewCmdBuf())
	sched.Run(sys, 0)

	if err := sched.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if m.applyCalls != 1 {
		t.Errorf("expected Apply called once, got %d", m.applyCalls)
	}
	if len(m.got) != 3 {
		t.Errorf("expected 3 ids passed to Apply, got %d", len(m.got))
	}
}

func TestScheduler_Sync_EachChunkAppliedSeparately(t *testing.T) {
	// Two Migrate calls (two chunks) result in two separate Migrate calls.
	var registry reg.Registry
	registry.Init(reg.Config{
		Entity:  ent.Config{Cap: 10, FreeCap: 10},
		Matcher: query.Config{Cap: 4},
	})
	sched := NewScheduler(&registry)

	m := &stubMigrator{}
	sys := newMigrateSystem(m,
		[]uid.UID64{1, 2},
		[]uid.UID64{3, 4, 5},
	)
	sched.Register(sys, NewCmdBuf())
	sched.Run(sys, 0)

	if err := sched.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if m.applyCalls != 2 {
		t.Errorf("expected Migrate called twice (one per chunk), got %d", m.applyCalls)
	}
	if len(m.got) != 5 {
		t.Errorf("expected 5 ids total across both calls, got %d", len(m.got))
	}
}

func TestScheduler_Sync_TwoDistinctMigrators_PerChunkApply(t *testing.T) {
	// Three Migrate calls: mA twice, mB once. Each call maps to one chunk,
	// so mA gets two Migrate calls and mB gets one — no merging.
	mA := &stubMigrator{}
	mB := &stubMigrator{}

	cb := NewCmdBuf()
	cb.Migrate(mA, bulk.ChunkSnapshot{}, []uid.UID64{1, 2})
	cb.Migrate(mB, bulk.ChunkSnapshot{}, []uid.UID64{3})
	cb.Migrate(mA, bulk.ChunkSnapshot{}, []uid.UID64{4, 5})

	var registry reg.Registry
	registry.Init(reg.Config{
		Entity:  ent.Config{Cap: 10, FreeCap: 10},
		Matcher: query.Config{Cap: 4},
	})
	sched := NewScheduler(&registry)
	if err := sched.applyBufferCmds(cb); err != nil {
		t.Fatalf("applyBufferCmds failed: %v", err)
	}

	if mA.applyCalls != 2 {
		t.Errorf("mA: expected 2 Migrate calls (one per chunk), got %d", mA.applyCalls)
	}
	if len(mA.got) != 4 {
		t.Errorf("mA: expected 4 ids total, got %d", len(mA.got))
	}
	if mB.applyCalls != 1 {
		t.Errorf("mB: expected 1 Migrate call, got %d", mB.applyCalls)
	}
	if len(mB.got) != 1 {
		t.Errorf("mB: expected 1 id, got %d", len(mB.got))
	}
}

// reserveSpace must replace an existing-but-undersized page when advancing
// onto it, not just when appending a brand new one.
func TestCmdBuf_ReserveSpace_GrowsExistingUndersizedPage(t *testing.T) {
	cb := NewCmdBuf()

	cb.reserveSpace(4000, 1) // fills most of page 0 (4096 bytes)
	cb.reserveSpace(200, 1)  // spills onto a freshly appended page 1 (4096 bytes)
	if len(cb.pages) != 2 {
		t.Fatalf("setup error: expected 2 pages, got %d", len(cb.pages))
	}

	cb.Clear() // pageIdx/offset reset to 0, but both pages (and their sizes) survive

	cb.reserveSpace(4000, 1)      // fills page 0 again
	p := cb.reserveSpace(5000, 1) // page 1 (4096 bytes) is now too small for 5000
	if p == nil {
		t.Fatal("expected a non-nil pointer after growing the undersized page")
	}
	if len(cb.pages[1]) < 5000 {
		t.Errorf("expected page 1 to grow to at least 5000 bytes, got %d", len(cb.pages[1]))
	}
}
