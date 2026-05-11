package gokebiten

import (
	"time"
)

// TimeTracker is responsible for the fixed physics step (Fixed Time Step) and statistics.
type TimeTracker struct {
	lastUpdate    time.Time
	lastTPSUpdate time.Time
	accumulator   time.Duration
}

func NewTimeTracker() *TimeTracker {
	now := time.Now()
	return &TimeTracker{
		lastUpdate:    now,
		lastTPSUpdate: now,
	}
}

// CalculateSteps calculates how many physics ticks should be performed in the current frame.
func (t *TimeTracker) CalculateSteps(physicsStep time.Duration, maxSteps int) int {
	now := time.Now()

	if t.lastUpdate.IsZero() {
		t.lastUpdate = now
		t.lastTPSUpdate = now
	}

	elapsed := now.Sub(t.lastUpdate)
	t.lastUpdate = now
	t.accumulator += elapsed

	steps := 0
	for t.accumulator >= physicsStep && steps < maxSteps {
		t.accumulator -= physicsStep
		steps++
	}

	// Time debt clamping mechanism (forces slow motion instead of a death spiral).
	if t.accumulator > physicsStep {
		t.accumulator = 0
	}

	return steps
}

func (t *TimeTracker) ProcessStatsInterval() bool {
	duration := time.Since(t.lastTPSUpdate)
	if duration > 2*time.Second {
		t.lastTPSUpdate = time.Now()
		return true
	}
	if duration >= time.Second {
		t.lastTPSUpdate = t.lastTPSUpdate.Add(time.Second)
		return true
	}
	return false
}
