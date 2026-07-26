package collisions

import "github.com/kjkrol/uid"

// MaxTouching caps how many distinct neighbors Collision records in one
// tick — extras are silently dropped once full, not evicted.
const MaxTouching = 8

// Collision is always present on every entity BroadPhase probes: it holds
// this tick's spatial-index neighbors (Touching), written directly by
// BroadPhase's own space.Query and consumed by NarrowPhase, which clears it
// after resolving — so NarrowPhase never needs to query the spatial index
// itself.
type Collision struct {
	Touching      [MaxTouching]uid.UID64
	TouchingCount uint8
}

// addTouching records id as a neighbor this tick, deduplicated, up to
// MaxTouching — further neighbors past the cap are silently dropped.
func (c *Collision) addTouching(id uid.UID64) {
	for i := uint8(0); i < c.TouchingCount; i++ {
		if c.Touching[i] == id {
			return
		}
	}
	if c.TouchingCount < MaxTouching {
		c.Touching[c.TouchingCount] = id
		c.TouchingCount++
	}
}

// clear resets the touching list for the next tick.
func (c *Collision) clear() {
	c.TouchingCount = 0
}
