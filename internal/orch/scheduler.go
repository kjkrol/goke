package orch

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

type RunCtx interface {
	Run(Runnable, time.Duration)
	RunParallel(time.Duration, ...Runnable)
	Sync() error
}

type Plan func(RunCtx, time.Duration)

type Scheduler struct {
	lookup    Lookup
	mutator   Mutator
	runnables []Runnable
	buffers   map[Runnable]*CmdBuf
	plan      Plan
}

var _ RunCtx = (*Scheduler)(nil)

func (s *Scheduler) Reset() {
	clear(s.runnables)
	s.runnables = s.runnables[:0]
	for _, cb := range s.buffers {
		if cb != nil {
			cb.Clear()
		}
	}
	clear(s.buffers)
	s.plan = nil
}

func NewScheduler(lookup Lookup, mutator Mutator) Scheduler {
	return Scheduler{
		lookup:    lookup,
		mutator:   mutator,
		buffers:   make(map[Runnable]*CmdBuf),
		runnables: make([]Runnable, 0),
	}
}

func (s *Scheduler) SetPlan(plan Plan) {
	s.plan = plan
}

func (s *Scheduler) Register(runnable Runnable) {
	s.runnables = append(s.runnables, runnable)
	s.buffers[runnable] = NewCmdBuf()
}

func (s *Scheduler) Tick(duration time.Duration) {
	if s.plan == nil {
		panic("ECS Error: Plan is not defined! Use SetPlan() before starting the loop.")
	}
	s.plan(s, duration)
}

// -------------------------------------------------------------

func (s *Scheduler) Run(runnable Runnable, d time.Duration) {
	runnable.Update(s.lookup, s.buffers[runnable], d)
}

func (s *Scheduler) RunParallel(d time.Duration, runnables ...Runnable) {
	var wg sync.WaitGroup

	for _, runnable := range runnables {
		wg.Add(1)
		go func(r Runnable) {
			defer wg.Done()
			r.Update(s.lookup, s.buffers[r], d)
		}(runnable)
	}

	wg.Wait()
}

func (s *Scheduler) Sync() error {
	for _, cb := range s.buffers {
		if len(cb.cmds) > 0 {
			err := s.applyBufferCmds(cb)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scheduler) applyBufferCmds(cb *CmdBuf) error {
	for _, cmd := range cb.cmds {
		target := cmd.entityID

		switch cmd.cType {
		case cmdAssignComp:
			// TODD: need to be tested !!!
			ptr, err := s.mutator.UpsertComp(target, cmd.compMeta)
			if err != nil {
				return fmt.Errorf("failed to allocate ID for target %d: %w", target, err)
			}
			if ptr != nil {
				copyMemory(ptr, cmd.dataPtr, cmd.compMeta.Size)
			}
		case cmdRemoveComp:
			s.mutator.RemoveComp(target, cmd.compMeta)
		case cmdRemoveEntity:
			s.mutator.Remove(target)
		}
	}
	cb.reset()
	return nil
}

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}
