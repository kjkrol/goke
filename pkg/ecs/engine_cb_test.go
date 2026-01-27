package ecs_test

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/pkg/ecs"
)

// --- Components ---

type Task struct{ Completed bool }
type Log struct{ Msg string }

var logInfo ecs.ComponentInfo

// --- Systems ---

type WorkerSystem struct {
	query *ecs.Query1[Task]
}

func (s *WorkerSystem) Init(reg *ecs.Registry) {
	s.query = ecs.NewQuery1[Task](reg)
}

func (s *WorkerSystem) Update(reg ecs.ReadOnlyRegistry, cb *ecs.SystemCommandBuffer, d time.Duration) {
	for head := range s.query.All1() {
		msg := Log{Msg: "Done"}
		cb.AssignComponent(head.Entity, logInfo, unsafe.Pointer(&msg))
	}
}

type LoggerSystem struct {
	query *ecs.Query1[Log]
	Found bool
}

func (s *LoggerSystem) Init(reg *ecs.Registry) {
	s.query = ecs.NewQuery1[Log](reg)
}

func (s *LoggerSystem) Update(reg ecs.ReadOnlyRegistry, cb *ecs.SystemCommandBuffer, d time.Duration) {
	for range s.query.All1() {
		s.Found = true
	}
}

// --- TEST ---

func TestECS_SystemInteractions(t *testing.T) {

	setupComponents := func(e *ecs.Engine) {
		logInfo = e.RegisterComponentType(reflect.TypeFor[Log]())
	}

	t.Run("Isolation: Logger should not see changes from Worker without Sync between them", func(t *testing.T) {
		engine := ecs.NewEngine()
		setupComponents(engine)

		e := engine.CreateEntity()
		ecs.Assign(engine, e, Task{Completed: false})

		worker := &WorkerSystem{}
		logger := &LoggerSystem{}

		engine.RegisterSystem(worker)
		engine.RegisterSystem(logger)

		engine.SetExecutionPlan(func(s ecs.ExecutionContext, d time.Duration) {
			s.Run(worker, d)
			s.Run(logger, d)
			s.Sync()
		})

		engine.Run(time.Millisecond)

		if logger.Found {
			t.Error("LoggerSystem found Log that should have been deferred until the end of the plan")
		}
	})

	t.Run("Synchronization: Logger should see changes from Worker due to explicit Sync in Plan", func(t *testing.T) {
		engine := ecs.NewEngine()
		setupComponents(engine)

		e := engine.CreateEntity()
		ecs.Assign(engine, e, Task{Completed: false})

		worker := &WorkerSystem{}
		logger := &LoggerSystem{}

		engine.RegisterSystem(worker)
		engine.RegisterSystem(logger)

		engine.SetExecutionPlan(func(s ecs.ExecutionContext, d time.Duration) {
			s.Run(worker, d)
			s.Sync() // Force synchronization between systems
			s.Run(logger, d)
			s.Sync()
		})

		engine.Run(time.Millisecond)

		if !logger.Found {
			t.Error("LoggerSystem should have found Log due to explicit Sync call in the ExecutionPlan")
		}
	})
}
