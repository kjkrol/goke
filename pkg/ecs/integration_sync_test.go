/*
This test suite validates the "Deferred Command Pattern" and "System Synchronization".

It focuses on the lifecycle of component modifications within an ExecutionPlan:
 1. Snapshot Integrity: Ensures systems operate on a consistent state during their Update cycle.
 2. Modification Deferral: Verifies that changes made via SystemCommandBuffer are buffered
    and isolated from other systems in the same synchronization stage.
 3. Sync Point Enforcement: Confirms that ExecutionContext.Sync() correctly flushes the buffer,
    making all changes globally visible for subsequent stages.

This is critical for preventing race conditions and ensuring deterministic system execution.
*/
package ecs_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/kjkrol/goke/pkg/ecs"
	"github.com/kjkrol/goke/pkg/ecs/ecsq"
)

// --- Components ---

type Task struct{ Completed bool }
type Log struct{ Msg string }

var logInfo ecs.ComponentInfo

// --- Systems ---

type WorkerSystem struct {
	query *ecsq.Query1[Task]
}

func (s *WorkerSystem) Init(eng *ecs.Engine) {
	s.query = ecs.NewQuery1[Task](eng)
}

func (s *WorkerSystem) Update(reg ecs.Lookup, cb *ecs.Commands, d time.Duration) {
	for head := range s.query.All1() {
		msg := Log{Msg: "Done"}
		ecs.AssignComponent(cb, head.Entity, logInfo, msg)
	}
}

type LoggerSystem struct {
	query *ecsq.Query1[Log]
	Found bool
}

func (s *LoggerSystem) Init(eng *ecs.Engine) {
	s.query = ecs.NewQuery1[Log](eng)
}

func (s *LoggerSystem) Update(reg ecs.Lookup, cb *ecs.Commands, d time.Duration) {
	for range s.query.All1() {
		s.Found = true
	}
}

// --- TEST ---

// TestECS_SystemInteractions validates that the state remains isolated between systems
// until an explicit Sync point is reached, ensuring that concurrent-safe logic
// is maintained by the scheduler.
func TestECS_SystemInteractions(t *testing.T) {

	setupComponents := func(e *ecs.Engine) {
		logInfo = e.RegisterComponentType(reflect.TypeFor[Log]())
	}

	t.Run("Isolation: Logger should not see changes from Worker without Sync between them", func(t *testing.T) {
		engine := ecs.NewEngine()
		setupComponents(engine)

		e := engine.CreateEntity()
		task, _ := ecs.AllocateComponent[Task](engine, e)
		*task = Task{Completed: false}

		worker := &WorkerSystem{}
		logger := &LoggerSystem{}

		engine.RegisterSystem(worker)
		engine.RegisterSystem(logger)

		engine.SetExecutionPlan(func(s ecs.ExecutionContext, d time.Duration) {
			s.Run(worker, d)
			s.Run(logger, d)
			s.Sync()
		})

		engine.Tick(time.Millisecond)

		if logger.Found {
			t.Error("LoggerSystem found Log that should have been deferred until the end of the plan")
		}
	})

	t.Run("Synchronization: Logger should see changes from Worker due to explicit Sync in Plan", func(t *testing.T) {
		engine := ecs.NewEngine()
		setupComponents(engine)

		e := engine.CreateEntity()
		task, _ := ecs.AllocateComponent[Task](engine, e)
		*task = Task{Completed: false}

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

		engine.Tick(time.Millisecond)

		if !logger.Found {
			t.Error("LoggerSystem should have found Log due to explicit Sync call in the ExecutionPlan")
		}
	})
}
