package core

import (
	"fmt"
	"sync"
	"time"
)

type ExecutionContext interface {
	Run(System, time.Duration)
	RunParallel(time.Duration, ...System)
	Sync() error
}

type ExecutionPlan func(ExecutionContext, time.Duration)

type Scheduler struct {
	register *Registry
	systems  []System
	buffers  map[System]*SystemCommandBuffer
	plan     ExecutionPlan
}

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

func NewScheduler(register *Registry) Scheduler {
	return Scheduler{
		register: register,
		buffers:  make(map[System]*SystemCommandBuffer),
		systems:  make([]System, 0),
	}
}

var _ ExecutionContext = (*Scheduler)(nil)

func (s *Scheduler) SetExecutionPlan(plan ExecutionPlan) {
	s.plan = plan
}

func (s *Scheduler) RegisterSystem(sys System) {
	s.systems = append(s.systems, sys)
	s.buffers[sys] = NewSystemCommandBuffer()
}

func (s *Scheduler) Tick(duration time.Duration) {
	if s.plan == nil {
		panic("ECS Error: ExecutionPlan is not defined! Use SetPlan() before starting the loop.")
	}
	s.plan(s, duration)
}

// -------------------------------------------------------------

func (s *Scheduler) Run(sys System, d time.Duration) {
	cb := s.getBuffer(sys)
	sys.Update(s.register, cb, d)
}

func (s *Scheduler) RunParallel(d time.Duration, systems ...System) {
	var wg sync.WaitGroup

	for _, sys := range systems {
		wg.Add(1)
		go func(currSys System) {
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

func (s *Scheduler) getBuffer(sys System) *SystemCommandBuffer {
	return s.buffers[sys]
}

func (s *Scheduler) applyBufferCommands(cb *SystemCommandBuffer) error {
	vMap := make(map[Entity]Entity)
	for _, cmd := range cb.commands {

		target := cmd.entity
		if target.IsVirtual() {
			if realID, ok := vMap[target]; ok {
				target = realID
			}
		}

		switch cmd.cType {
		case cmdAssignComponent:
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
			s.register.RemoveEntity(cmd.entity)
		}
	}
	cb.reset()
	return nil
}
