package kinematics

import (
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// Placement decides where the index-th of count entities starts. Used on
// its own for entities that need no Velocity at all (see Game.PopulateStatic).
// Components/Init mirror gokebiten.EntityExtras (duplicated here, not
// imported, to avoid a cycle) so any Placement can be passed directly as
// PopulateStatic's first populator.
type Placement interface {
	Place(index, count int) Position
	Components() []goke.Addable
	Init(cursor *goke.Cursor, i, index int, id uid.UID64)
}

// Motion decides the starting velocity of the index-th of count entities.
type Motion interface {
	InitialVelocity(index, count int) Velocity
}

// Spawner decides the starting Position and Velocity of the index-th of
// count entities together, since a moving entity always needs both.
// Components/Init mirror gokebiten.EntityExtras (see Placement) so any
// Spawner can be passed directly as Populate's first populator.
type Spawner interface {
	Spawn(index, count int) (Position, Velocity)
	Components() []goke.Addable
	Init(cursor *goke.Cursor, i, index int, id uid.UID64)
}

// NewSpawner composes an independently-named Placement and Motion into a
// Spawner, so each keeps a name describing what it does rather than what
// it's combined into.
func NewSpawner(placement Placement, motion Motion) Spawner {
	return &spawner{placement: placement, motion: motion}
}

type spawner struct {
	placement Placement
	motion    Motion
}

func (s *spawner) Spawn(index, count int) (Position, Velocity) {
	return s.placement.Place(index, count), s.motion.InitialVelocity(index, count)
}
func (s *spawner) Components() []goke.Addable             { return nil }
func (s *spawner) Init(*goke.Cursor, int, int, uid.UID64) {}
