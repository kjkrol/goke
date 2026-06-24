package orch

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/v2/internal/comp"
)

// mockMutator is a controllable Mutator for testing Scheduler/CmdBuf
// dispatch in isolation, without pulling in a real reg.Registry.
type mockMutator struct {
	upsertErr      error
	removeCompCall struct {
		called bool
		id     uid.UID64
		comp   comp.ID
	}
	removeCall struct {
		called bool
		id     uid.UID64
	}
}

func (m *mockMutator) UpsertComp(uid.UID64, comp.ID) (unsafe.Pointer, error) {
	return nil, m.upsertErr
}

func (m *mockMutator) RemoveComp(id uid.UID64, c comp.ID) error {
	m.removeCompCall.called = true
	m.removeCompCall.id = id
	m.removeCompCall.comp = c
	return nil
}

func (m *mockMutator) Remove(id uid.UID64) bool {
	m.removeCall.called = true
	m.removeCall.id = id
	return true
}

// fnRunnable adapts a plain function to the Runnable interface.
type fnRunnable struct {
	fn func(cb *CmdBuf, d time.Duration)
}

func (r *fnRunnable) Update(cb *CmdBuf, d time.Duration) { r.fn(cb, d) }

func TestScheduler_Tick_PanicsWithoutPlan(t *testing.T) {
	sched := NewScheduler(&mockMutator{})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Tick to panic when no Plan is set")
		}
	}()
	sched.Tick(time.Millisecond)
}

func TestScheduler_Tick_RunsThePlan(t *testing.T) {
	sched := NewScheduler(&mockMutator{})
	ran := false
	sched.SetPlan(func(ctx RunCtx, d time.Duration) {
		ran = true
	})

	sched.Tick(time.Millisecond)

	if !ran {
		t.Error("expected Tick to invoke the configured Plan")
	}
}

func TestScheduler_RunParallel_RunsAllRunnablesConcurrently(t *testing.T) {
	sched := NewScheduler(&mockMutator{})

	var counter atomic.Int32
	var mu sync.Mutex
	seenBufs := make(map[*CmdBuf]bool)

	makeRunnable := func() *fnRunnable {
		return &fnRunnable{fn: func(cb *CmdBuf, d time.Duration) {
			counter.Add(1)
			mu.Lock()
			seenBufs[cb] = true
			mu.Unlock()
		}}
	}
	r1, r2, r3 := makeRunnable(), makeRunnable(), makeRunnable()
	sched.Register(r1)
	sched.Register(r2)
	sched.Register(r3)

	sched.RunParallel(time.Millisecond, r1, r2, r3)

	if counter.Load() != 3 {
		t.Errorf("expected all 3 runnables to run, got %d", counter.Load())
	}
	if len(seenBufs) != 3 {
		t.Errorf("expected each runnable to receive its own distinct CmdBuf, got %d distinct buffers", len(seenBufs))
	}
}

func TestScheduler_Reset_ClearsStateAndPlan(t *testing.T) {
	sched := NewScheduler(&mockMutator{})
	r := &fnRunnable{fn: func(cb *CmdBuf, d time.Duration) {
		cb.RemoveEntity(uid.UID64(1))
	}}
	sched.Register(r)
	sched.SetPlan(func(ctx RunCtx, d time.Duration) {})
	sched.Run(r, 0) // queue a command so the buffer has state to clear

	sched.Reset()

	if len(sched.runnables) != 0 {
		t.Errorf("expected runnables to be cleared, got %d", len(sched.runnables))
	}
	if len(sched.buffers) != 0 {
		t.Errorf("expected buffers to be cleared, got %d", len(sched.buffers))
	}

	defer func() {
		if rec := recover(); rec == nil {
			t.Error("expected Tick to panic again after Reset clears the Plan")
		}
	}()
	sched.Tick(time.Millisecond)
}

func TestScheduler_Sync_PropagatesUpsertCompError(t *testing.T) {
	wantErr := errors.New("boom")
	mut := &mockMutator{upsertErr: wantErr}
	sched := NewScheduler(mut)
	r := &fnRunnable{fn: func(cb *CmdBuf, d time.Duration) {
		AddComp(cb, uid.UID64(1), comp.ID(0), 42)
	}}
	sched.Register(r)
	sched.Run(r, 0)

	err := sched.Sync()
	if err == nil {
		t.Fatal("expected Sync to propagate the Mutator's error")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("expected wrapped sentinel error, got %v", err)
	}
}

func TestScheduler_Sync_DispatchesRemoveEntity(t *testing.T) {
	mut := &mockMutator{}
	sched := NewScheduler(mut)
	target := uid.UID64(7)
	r := &fnRunnable{fn: func(cb *CmdBuf, d time.Duration) {
		cb.RemoveEntity(target)
	}}
	sched.Register(r)
	sched.Run(r, 0)

	if err := sched.Sync(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mut.removeCall.called || mut.removeCall.id != target {
		t.Errorf("expected Mutator.Remove to be called with %v, got called=%v id=%v",
			target, mut.removeCall.called, mut.removeCall.id)
	}
}

func TestScheduler_Sync_DispatchesRemoveComp(t *testing.T) {
	mut := &mockMutator{}
	sched := NewScheduler(mut)
	target := uid.UID64(9)
	wantComp := comp.ID(3)
	r := &fnRunnable{fn: func(cb *CmdBuf, d time.Duration) {
		cb.RemoveComp(target, wantComp)
	}}
	sched.Register(r)
	sched.Run(r, 0)

	if err := sched.Sync(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mut.removeCompCall.called || mut.removeCompCall.id != target || mut.removeCompCall.comp != wantComp {
		t.Errorf("expected RemoveComp(%v,%v), got called=%v id=%v comp=%v",
			target, wantComp, mut.removeCompCall.called, mut.removeCompCall.id, mut.removeCompCall.comp)
	}
}
