// Package orch implements a plan-based task orchestrator with deferred
// side-effect synchronization.
//
// # Runnable
//
// Any task must implement [Runnable]:
//
//	type Runnable interface {
//	    Update(*CmdBuf, time.Duration)
//	}
//
// Each call to Update receives a [*CmdBuf] for queuing deferred mutations and a
// time.Duration representing the elapsed tick interval. Read access to external
// state is the task's own responsibility, arranged during its initialization.
//
// # Plan and Scheduler
//
// Execution order is defined by a Plan — a user-provided function that controls
// which Runnables execute and when. [Scheduler] runs the Plan each tick and
// provides two execution modes:
//
//   - Run         — sequential execution
//   - RunParallel — concurrent execution via goroutines
//
// # CmdBuf
//
// Each Runnable owns a dedicated [CmdBuf] — a buffer that queues mutations
// without applying them immediately. This guarantees that all Runnables within
// a stage observe a consistent snapshot of external state.
//
// # Sync
//
// Sync drains all CmdBufs and applies the queued mutations through [Mutator].
// It is the only moment where external state changes. Calling Sync between stages
// defines explicit synchronization points within the plan:
//
//	Runnable A ──┐
//	Runnable B ──┤──► Sync ──► mutations become visible
//	Runnable C ──┘
//
// # Mutator
//
// [Mutator] is an interface defined in this package. Any external state that
// implements it can be orchestrated — for example, an ECS component storage.
package orch
