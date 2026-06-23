/*
This test suite validates the "Deferred Command Pattern" and "System Synchronization".

It focuses on the lifecycle of component modifications within an Plan:
 1. Snapshot Integrity: Ensures systems operate on a consistent state during their Update cycle.
 2. Modification Deferral: Verifies that changes made via SystemCommandBuffer are buffered
    and isolated from other systems in the same synchronization stage.
 3. Sync Point Enforcement: Confirms that RunCtx.Sync() correctly flushes the buffer,
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

var taskID goke.CompID
var logID goke.CompID

// --- Systems ---

type WorkerSystem struct {
	query *goke.Query
}

func (s *WorkerSystem) Init(eng *goke.ECS) {
	s.query = eng.NewQueryBuilder().Include(goke.Include[Task]()).Build()
}

func (s *WorkerSystem) Update(schedule *goke.CmdBuf, duration time.Duration) {
	s.query.All()
	for s.query.Next() {
		for _, entityID := range s.query.Cursor.IDs {
			msg := Log{Msg: "Done"}
			goke.CmdBufAddComp(schedule, entityID, logID, msg)
		}
	}
}

type LoggerSystem struct {
	query *goke.Query
	Found bool
}

func (s *LoggerSystem) Init(eng *goke.ECS) {
	s.query = eng.NewQueryBuilder().Include(goke.Include[Log]()).Build()
}

func (s *LoggerSystem) Update(schedule *goke.CmdBuf, duration time.Duration) {
	s.query.All()
	for s.query.Next() {
		s.Found = true
	}
}

// --- TEST ---

// TestECS_SystemInteractions validates that the state remains isolated between systems
// until an explicit Sync point is reached, ensuring that concurrent-safe logic
// is maintained by the scheduler.
func TestECS_SystemInteractions(t *testing.T) {

	setupComponents := func(e *goke.ECS) {
		taskID = goke.RegComp[Task](e)
		logID = goke.RegComp[Log](e)
	}

	t.Run("Isolation: Logger should not see changes from Worker without Sync between them", func(t *testing.T) {
		ecs := goke.New()
		setupComponents(ecs)

		var task goke.Comp[Task]
		factory := ecs.NewFactory(&task)
		factory.Create(1)
		factory.Next()
		task.Slice(&factory.Cursor)[0] = Task{Completed: false}

		worker := &WorkerSystem{}
		logger := &LoggerSystem{}

		ecs.RegSys(worker)
		ecs.RegSys(logger)

		ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
			s.Run(worker, d)
			s.Run(logger, d)
			s.Sync()
		})

		ecs.Tick(time.Millisecond)

		if logger.Found {
			t.Error("LoggerSystem found Log that should have been deferred until the end of the plan")
		}
	})

	t.Run("Synchronization: Logger should see changes from Worker due to explicit Sync in Plan", func(t *testing.T) {
		ecs := goke.New()
		setupComponents(ecs)

		var task goke.Comp[Task]
		factory := ecs.NewFactory(&task)
		factory.Create(1)
		factory.Next()
		task.Slice(&factory.Cursor)[0] = Task{Completed: false}

		worker := &WorkerSystem{}
		logger := &LoggerSystem{}

		ecs.RegSys(worker)
		ecs.RegSys(logger)

		ecs.SetPlan(func(s goke.RunCtx, d time.Duration) {
			s.Run(worker, d)
			s.Sync() // Force synchronization between systems
			s.Run(logger, d)
			s.Sync()
		})

		ecs.Tick(time.Millisecond)

		if !logger.Found {
			t.Error("LoggerSystem should have found Log due to explicit Sync call in the Plan")
		}
	})
}
