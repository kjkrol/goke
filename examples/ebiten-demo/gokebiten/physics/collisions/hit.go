package collisions

import "time"

// Hit is a temporary tag applied to entities involved in a collision.
// BroadPhase batch-adds it (empty value) to every candidate; NarrowPhase
// either confirms it (sets ExpiresAt) once a real contact resolves, or
// withdraws it (removes the component) if the candidate turned out to be a
// broad-phase false positive. TagExpirySystem removes it later once
// ExpiresAt passes, independent of this confirm/reject step.
type Hit struct {
	ExpiresAt time.Time
}
