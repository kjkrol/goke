package collisions

import (
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/gokg/geom"
	"github.com/kjkrol/gokg/plane"
	"github.com/kjkrol/uid"
)

type contactSide struct {
	Entity   uid.UID64
	Pos      *kinematics.Position
	Vel      *kinematics.Velocity
	IsSensor bool // this side carries the Sensor tag — see NarrowPhase.solve
}

// Contact is one candidate pair from the broad phase, fully resolved on both
// sides before it ever enters NarrowPhase's buffer. Component pointers are
// valid only within the Update that produced them.
type Contact struct {
	A, B     contactSide
	resolved bool
}

// findActiveCollision returns the (mainBox + fragments) × (mainBox + fragments)
// combination with the strongest overlap (|pen.X| + |pen.Y|), recomputed from
// current geometry — solver push-apart between iterations can shift which
// combination is active. fragsA × fragsB combinations cannot occur and are
// deliberately skipped.
func (c *Contact) findActiveCollision() (boxA, boxB geom.AABB[uint32], pen geom.Vec[int32]) {
	mainA := c.A.Pos.AABB.AABB
	mainB := c.B.Pos.AABB.AABB

	hasFragsA := hasAnyFragment(&c.A.Pos.AABB)
	hasFragsB := hasAnyFragment(&c.B.Pos.AABB)

	if !hasFragsA && !hasFragsB {
		boxA = mainA
		boxB = mainB
		pen = c.penetration(mainA, mainB)
		return
	}

	bestArea := int32(0)

	tryCombo := func(bA, bB geom.AABB[uint32]) {
		p := c.penetration(bA, bB)
		if p.X == 0 && p.Y == 0 {
			return
		}
		area := abs32(p.X) + abs32(p.Y)
		if area > bestArea {
			bestArea = area
			boxA = bA
			boxB = bB
			pen = p
		}
	}

	tryCombo(mainA, mainB)

	if hasFragsB {
		c.B.Pos.AABB.VisitFragments(func(_ plane.FragPosition, b geom.AABB[uint32]) bool {
			tryCombo(mainA, b)
			return true
		})
	}

	if hasFragsA {
		c.A.Pos.AABB.VisitFragments(func(_ plane.FragPosition, b geom.AABB[uint32]) bool {
			tryCombo(b, mainB)
			return true
		})
	}

	return
}

func hasAnyFragment(ab *plane.AABB[uint32]) bool {
	has := false
	ab.VisitFragments(func(_ plane.FragPosition, _ geom.AABB[uint32]) bool {
		has = true
		return false
	})
	return has
}

func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func (c *Contact) calculateMtv(r1, r2 geom.AABB[uint32], isStaticB bool) (mtv1, mtv2 geom.Vec[uint32], res bool) {
	pen := c.penetration(r1, r2)

	if pen.X == 0 && pen.Y == 0 {
		return geom.Vec[uint32]{}, geom.Vec[uint32]{}, false
	}

	calculatePush := func(penetration int32) (int32, int32) {
		if isStaticB {
			return penetration, 0
		}
		pA := penetration / 2
		pB := -(penetration - pA) // no pixel lost on odd penetrations
		return pA, pB
	}

	var pushA, pushB geom.Vec[int32]

	if pen.X != 0 {
		p1, p2 := calculatePush(pen.X)
		pushA = geom.Vec[int32]{X: p1, Y: 0}
		pushB = geom.Vec[int32]{X: p2, Y: 0}
	} else {
		p1, p2 := calculatePush(pen.Y)
		pushA = geom.Vec[int32]{X: 0, Y: p1}
		pushB = geom.Vec[int32]{X: 0, Y: p2}
	}

	mtv1 = geom.NewVec(uint32(pushA.X), uint32(pushA.Y))
	mtv2 = geom.NewVec(uint32(pushB.X), uint32(pushB.Y))
	res = true
	return
}

// penetration returns the minimum translation vector separating r1 from r2,
// or a zero vector when they do not overlap.
func (c *Contact) penetration(r1, r2 geom.AABB[uint32]) geom.Vec[int32] {
	leftX := max(int32(r1.TopLeft.X), int32(r2.TopLeft.X))
	rightX := min(int32(r1.BottomRight.X), int32(r2.BottomRight.X))
	overlapX := rightX - leftX
	if overlapX <= 0 {
		return geom.Vec[int32]{}
	}

	topY := max(int32(r1.TopLeft.Y), int32(r2.TopLeft.Y))
	bottomY := min(int32(r1.BottomRight.Y), int32(r2.BottomRight.Y))
	overlapY := bottomY - topY
	if overlapY <= 0 {
		return geom.Vec[int32]{}
	}

	pushRight := int32(r2.BottomRight.X) - int32(r1.TopLeft.X)
	pushLeft := int32(r1.BottomRight.X) - int32(r2.TopLeft.X)
	pushDown := int32(r2.BottomRight.Y) - int32(r1.TopLeft.Y)
	pushUp := int32(r1.BottomRight.Y) - int32(r2.TopLeft.Y)

	minPush := pushRight
	mtv := geom.Vec[int32]{X: pushRight, Y: 0}

	if pushLeft < minPush {
		minPush = pushLeft
		mtv = geom.Vec[int32]{X: -pushLeft, Y: 0}
	}
	if pushDown < minPush {
		minPush = pushDown
		mtv = geom.Vec[int32]{X: 0, Y: pushDown}
	}
	if pushUp < minPush {
		mtv = geom.Vec[int32]{X: 0, Y: -pushUp}
	}

	return mtv
}
