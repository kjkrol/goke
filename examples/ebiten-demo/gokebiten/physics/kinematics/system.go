package kinematics

import (
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/gokg"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/spatial"
)

var _ goke.System = (*System)(nil)

type System struct {
	space     *gokg.Space
	maxSpeed  int32 // world units/sec; <= 0 means no limit
	moveQuery *goke.Query
	pos       goke.Comp[Position]
	vel       goke.Comp[Velocity]
}

// NewSystem builds the kinematics integrator. maxSpeed clamps Velocity each
// tick — pass 0 for no limit. Without a cap, a large enough velocity moves
// an entity further than its own size in one tick, skipping past whatever
// it should have collided with (tunneling) since collision detection only
// checks where entities currently are, not their path between ticks.
func NewSystem(space *gokg.Space, maxSpeed int32) *System {
	return &System{space: space, maxSpeed: maxSpeed}
}

func clampSpeed(v, max int32) int32 {
	if max <= 0 {
		return v
	}
	if v > max {
		return max
	}
	if v < -max {
		return -max
	}
	return v
}

func (s *System) Init(si *goke.SysInit) {
	s.moveQuery = si.NewQueryBuilder(&s.pos, &s.vel).Build()
}

func (s *System) Update(_ *goke.CmdBuf, d time.Duration) {
	dt := d.Seconds()
	s.moveQuery.All()
	for s.moveQuery.Next() {
		cursor := s.moveQuery.Cursor()
		pos := s.pos.Slice(cursor)
		vel := s.vel.Slice(cursor)
		for i, entityID := range cursor.IDs {
			vel[i].X = clampSpeed(vel[i].X, s.maxSpeed)
			vel[i].Y = clampSpeed(vel[i].Y, s.maxSpeed)

			pos[i].AccX += float64(vel[i].X) * dt
			pos[i].AccY += float64(vel[i].Y) * dt

			dx := int32(pos[i].AccX)
			dy := int32(pos[i].AccY)

			if dx != 0 {
				pos[i].AccX -= float64(dx)
			}
			if dy != 0 {
				pos[i].AccY -= float64(dy)
			}

			if dx != 0 || dy != 0 {
				delta := geom.NewVec(uint32(dx), uint32(dy))
				s.space.Translate(uint64(entityID), &pos[i].AABB, delta)
			}
		}
	}
	s.space.Flush(func(a spatial.AABB) {})
}
