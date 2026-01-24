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
	sync  bool
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

func (s *WorkerSystem) ShouldSync() bool { return s.sync }

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

func (s *LoggerSystem) ShouldSync() bool { return false }

// --- TEST ---

func TestECS_SystemInteractions(t *testing.T) {

	setupComponents := func(e *ecs.Engine) {
		logInfo = e.RegisterComponentType(reflect.TypeFor[Log]())
	}

	t.Run("Isolation: Logger should not see changes from Worker without Sync", func(t *testing.T) {
		engine := ecs.NewEngine()
		setupComponents(engine)

		e := engine.CreateEntity()
		ecs.Assign(engine, e, Task{Completed: false})

		worker := &WorkerSystem{sync: false}
		logger := &LoggerSystem{}

		engine.RegisterSystems([]ecs.System{worker, logger})

		engine.UpdateSystems(time.Millisecond)

		if logger.Found {
			t.Error("LoggerSystem found Log that should have been deferred!")
		}
	})

	t.Run("Synchronization: Logger should see changes from Worker due to ShouldSync", func(t *testing.T) {
		engine := ecs.NewEngine()
		setupComponents(engine)

		e := engine.CreateEntity()
		ecs.Assign(engine, e, Task{Completed: false})

		worker := &WorkerSystem{sync: true}
		logger := &LoggerSystem{}

		engine.RegisterSystems([]ecs.System{worker, logger})

		engine.UpdateSystems(time.Millisecond)

		if !logger.Found {
			t.Error("LoggerSystem should have found Log due to forced synchronization")
		}
	})
}
