package exec

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/internal/reg"
)

type RunCtx interface {
	Run(Runnable, time.Duration)
	RunParallel(time.Duration, ...Runnable)
	Sync() error
}

type Plan func(RunCtx, time.Duration)

type Scheduler struct {
	register *reg.Registry
	systems  []Runnable
	buffers  map[Runnable]*CommandBuf
	plan     Plan
}

var _ RunCtx = (*Scheduler)(nil)

func (s *Scheduler) Reset() {
	clear(s.systems)
	s.systems = s.systems[:0]
	for _, scb := range s.buffers {
		if scb != nil {
			scb.Clear()
		}
	}
	clear(s.buffers)
	s.plan = nil
}

func NewScheduler(register *reg.Registry) Scheduler {
	return Scheduler{
		register: register,
		buffers:  make(map[Runnable]*CommandBuf),
		systems:  make([]Runnable, 0),
	}
}

func (s *Scheduler) SetExecutionPlan(plan Plan) {
	s.plan = plan
}

func (s *Scheduler) RegisterSystem(sys Runnable) {
	s.systems = append(s.systems, sys)
	s.buffers[sys] = NewCommandBuf()
}

func (s *Scheduler) Tick(duration time.Duration) {
	if s.plan == nil {
		panic("ECS Error: Plan is not defined! Use Plan() before starting the loop.")
	}
	s.plan(s, duration)
}

// -------------------------------------------------------------

func (s *Scheduler) Run(sys Runnable, d time.Duration) {
	cb := s.getBuffer(sys)
	sys.Update(s.register, cb, d)
}

func (s *Scheduler) RunParallel(d time.Duration, systems ...Runnable) {
	var wg sync.WaitGroup

	for _, sys := range systems {
		wg.Add(1)
		go func(currSys Runnable) {
			defer wg.Done()
			cb := s.getBuffer(currSys)
			currSys.Update(s.register, cb, d)
		}(sys)
	}

	wg.Wait()
}

func (s *Scheduler) Sync() error {
	for _, cb := range s.buffers {
		if len(cb.commands) > 0 {
			err := s.applyBufferCommands(cb)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// -------------------------------------------------------------

func (s *Scheduler) getBuffer(sys Runnable) *CommandBuf {
	return s.buffers[sys]
}

func (s *Scheduler) applyBufferCommands(cb *CommandBuf) error {
	for _, cmd := range cb.commands {
		target := cmd.entity

		switch cmd.cType {
		case cmdAssignComponent:
			// TODD: need to be tested !!!
			ptr, err := s.register.AllocateByID(target, cmd.compInfo)
			if err != nil {
				return fmt.Errorf("failed to allocate ID for target %d: %w", target, err)
			}
			if ptr != nil {
				copyMemory(ptr, cmd.dataPtr, cmd.compInfo.Size)
			}
		case cmdRemoveComponent:
			s.register.UnassignByID(target, cmd.compInfo)
		case cmdRemoveEntity:
			s.register.RemoveEntity(target)
		}
	}
	cb.reset()
	return nil
}

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}
