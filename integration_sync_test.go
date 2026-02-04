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
package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke"
)

// --- Components ---

type Task struct{ Completed bool }
type Log struct{ Msg string }

var taskDesc goke.ComponentDesc
var logDesc goke.ComponentDesc

// --- Systems ---

type WorkerSystem struct {
	view *goke.View1[Task]
}

func (s *WorkerSystem) Init(eng *goke.ECS) {
	s.view = goke.NewView1[Task](eng)
}

func (s *WorkerSystem) Update(lookup goke.Lookup, schedule *goke.Schedule, duration time.Duration) {
	for head := range s.view.All() {
		msg := Log{Msg: "Done"}
		goke.ScheduleAddComponent(schedule, head.Entity, logDesc, msg)
	}
}

type LoggerSystem struct {
	view  *goke.View1[Log]
	Found bool
}

func (s *LoggerSystem) Init(eng *goke.ECS) {
	s.view = goke.NewView1[Log](eng)
}

func (s *LoggerSystem) Update(lookup goke.Lookup, schedule *goke.Schedule, duration time.Duration) {
	for range s.view.All() {
		s.Found = true
	}
}

// --- TEST ---

// TestECS_SystemInteractions validates that the state remains isolated between systems
// until an explicit Sync point is reached, ensuring that concurrent-safe logic
// is maintained by the scheduler.
func TestECS_SystemInteractions(t *testing.T) {

	setupComponents := func(e *goke.ECS) {
		taskDesc = goke.RegisterComponent[Task](e)
		logDesc = goke.RegisterComponent[Log](e)
	}

	t.Run("Isolation: Logger should not see changes from Worker without Sync between them", func(t *testing.T) {
		ecs := goke.New()
		setupComponents(ecs)

		e := goke.CreateEntity(ecs)
		*goke.EnsureComponent[Task](ecs, e, taskDesc) = Task{Completed: false}

		worker := &WorkerSystem{}
		logger := &LoggerSystem{}

		goke.RegisterSystem(ecs, worker)
		goke.RegisterSystem(ecs, logger)

		goke.Plan(ecs, func(s goke.ExecutionContext, d time.Duration) {
			s.Run(worker, d)
			s.Run(logger, d)
			s.Sync()
		})

		goke.Tick(ecs, time.Millisecond)

		if logger.Found {
			t.Error("LoggerSystem found Log that should have been deferred until the end of the plan")
		}
	})

	t.Run("Synchronization: Logger should see changes from Worker due to explicit Sync in Plan", func(t *testing.T) {
		ecs := goke.New()
		setupComponents(ecs)

		e := goke.CreateEntity(ecs)
		*goke.EnsureComponent[Task](ecs, e, taskDesc) = Task{Completed: false}

		worker := &WorkerSystem{}
		logger := &LoggerSystem{}

		goke.RegisterSystem(ecs, worker)
		goke.RegisterSystem(ecs, logger)

		goke.Plan(ecs, func(s goke.ExecutionContext, d time.Duration) {
			s.Run(worker, d)
			s.Sync() // Force synchronization between systems
			s.Run(logger, d)
			s.Sync()
		})

		goke.Tick(ecs, time.Millisecond)

		if !logger.Found {
			t.Error("LoggerSystem should have found Log due to explicit Sync call in the ExecutionPlan")
		}
	})
}
