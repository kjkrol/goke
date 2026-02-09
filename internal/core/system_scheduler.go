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

type SystemScheduler struct {
	register *Registry
	systems  []System
	buffers  map[System]*SystemCommandBuffer
	plan     ExecutionPlan
}

func NewScheduler(register *Registry) SystemScheduler {
	return SystemScheduler{
		register: register,
		buffers:  make(map[System]*SystemCommandBuffer),
		systems:  make([]System, 0),
	}
}

var _ ExecutionContext = (*SystemScheduler)(nil)

func (s *SystemScheduler) SetExecutionPlan(plan ExecutionPlan) {
	s.plan = plan
}

func (s *SystemScheduler) RegisterSystem(sys System) {
	s.systems = append(s.systems, sys)
	s.buffers[sys] = NewSystemCommandBuffer()
}

func (s *SystemScheduler) Tick(duration time.Duration) {
	if s.plan == nil {
		panic("ECS Error: ExecutionPlan is not defined! Use SetPlan() before starting the loop.")
	}
	s.plan(s, duration)
}

// -------------------------------------------------------------

func (s *SystemScheduler) Run(sys System, d time.Duration) {
	cb := s.getBuffer(sys)
	sys.Update(s.register, cb, d)
}

func (s *SystemScheduler) RunParallel(d time.Duration, systems ...System) {
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

func (s *SystemScheduler) Sync() error {
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

func (s *SystemScheduler) getBuffer(sys System) *SystemCommandBuffer {
	return s.buffers[sys]
}

func (s *SystemScheduler) applyBufferCommands(cb *SystemCommandBuffer) error {
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
