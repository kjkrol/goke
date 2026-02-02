package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kjkrol/goke/ecs"
	"github.com/kjkrol/gokg/pkg/geom"
	"github.com/kjkrol/gokg/pkg/plane"
	"github.com/kjkrol/gokg/pkg/spatial"
)

// --- Configuration ---
const (
	ScreenWidth    = 1024
	ScreenHeight   = 1024
	EntityCount    = 2048
	BucketCapacity = 1024
	RectSize       = 7

	// 2. Define a fixed physics time step (e.g., 120Hz for high precision)
	// This decouples physics simulation from the rendering framerate
	TargetTPS   = 120
	PhysicsStep = time.Second / TargetTPS
)

// --- Components ---

type Position struct {
	plane.AABB[uint32]
	// Accumulators for sub-pixel movement
	accX float64
	accY float64
}
type Velocity struct{ geom.Vec[int32] }
type Appearance struct {
	Color color.RGBA
}
type CollisionPair struct {
	With ecs.Entity
}

// --- Game Loop (Ebitengine Adapter) ---

type Game struct {
	engine           *ecs.Engine
	renderView       *ecs.View2[Position, Appearance]
	accumulator      time.Duration
	lastUpdate       time.Time
	collisionCounter float64
	ticks            int       // Raw counter for physics steps
	physicsTPS       int       // Final value to display
	lastTPSUpdate    time.Time // Timer for resetting the counter
}

func (g *Game) Update() error {
	// 1. Calculate the real time elapsed since the last update (Delta Time)
	now := time.Now()
	if g.lastUpdate.IsZero() {
		g.lastUpdate = now
		g.lastTPSUpdate = now
	}
	elapsed := time.Since(g.lastUpdate)
	g.lastUpdate = now

	// 3. Add the elapsed time to the accumulator ("time bank")
	g.accumulator += elapsed

	// 4. Consume the accumulated time in fixed increments
	// If the frame rate drops, this loop will "catch up" by running multiple ticks
	for g.accumulator >= PhysicsStep {
		g.engine.Tick(PhysicsStep)
		g.accumulator -= PhysicsStep
		g.ticks++
	}
	// --- CALCULATE ACTUAL PHYSICS TPS ONCE PER SECOND ---
	if time.Since(g.lastTPSUpdate) >= time.Second {
		g.physicsTPS = g.ticks
		g.ticks = 0
		g.collisionCounter = 0
		g.lastTPSUpdate = time.Now()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 50, 50, 255})
	for head := range g.renderView.All() {
		aabb, app := head.V1, head.V2

		x := float32(aabb.TopLeft.X)
		y := float32(aabb.TopLeft.Y)
		w := float32(aabb.BottomRight.X - aabb.TopLeft.X)
		h := float32(aabb.BottomRight.Y - aabb.TopLeft.Y)

		vector.FillRect(screen, x, y, w, h, app.Color, true)

		aabb.AABB.VisitFragments(func(pos plane.FragPosition, box geom.AABB[uint32]) bool {
			x := float32(box.TopLeft.X)
			y := float32(box.TopLeft.Y)
			w := float32(box.BottomRight.X - box.TopLeft.X)
			h := float32(box.BottomRight.Y - box.TopLeft.Y)

			vector.FillRect(screen, x, y, w, h, app.Color, true)
			return true
		})
	}

	avgCollisionsPerTick := float64(0)
	if g.physicsTPS > 0 {
		avgCollisionsPerTick = float64(g.collisionCounter) / float64(g.physicsTPS)
	}
	debugMsg := fmt.Sprintf(
		"FPS: %0.2f\nTPS (Ebiten): %0.2f\nTPS (Physics): %d\nEntities: %d\nCollisions/Tick: %0.2f",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.physicsTPS,
		EntityCount,
		avgCollisionsPerTick,
	)
	ebitenutil.DebugPrint(screen, debugMsg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	// 1. Initialize GOKe Engine
	engine := ecs.NewEngine()

	game := &Game{
		engine:     engine,
		renderView: ecs.NewView2[Position, Appearance](engine),
	}

	// 2. Register Components
	ecs.RegisterComponent[Position](engine)
	ecs.RegisterComponent[Velocity](engine)
	ecs.RegisterComponent[Appearance](engine)
	collisionPairType := ecs.RegisterComponent[CollisionPair](engine)

	space := plane.NewToroidal2D[uint32](ScreenWidth, ScreenHeight)

	spatialIndex, err := spatial.NewGridIndexManager(space, spatial.GridIndexConfig{
		Resolution:       spatial.Size1024x1024,
		BucketResolution: spatial.Size32x32,
		BucketCapacity:   BucketCapacity,
	})
	if err != nil {
		log.Fatalf("Failed to create bucket grid: %v", err)
	}

	// 3. Define Systems

	// System: Movement (Torus Topology)
	moveView := ecs.NewView3[Position, Velocity, Appearance](engine)
	moveSystem := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		dt := d.Seconds()
		for head := range moveView.All() {
			pos, vel := head.V1, head.V2
			app := head.V3
			app.Color = color.RGBA{255, 255, 255, 255}
			pos.accX += float64(vel.X) * dt
			pos.accY += float64(vel.Y) * dt

			dx := int32(pos.accX)
			dy := int32(pos.accY)

			if dx != 0 {
				pos.accX -= float64(dx)
			}
			if dy != 0 {
				pos.accY -= float64(dy)
			}

			if dx != 0 || dy != 0 {
				delta := geom.NewVec[uint32](uint32(dx), uint32(dy))
				space.Translate(&pos.AABB, delta)
				spatialIndex.QueueUpdate(uint64(head.Entity), pos.AABB.AABB, true)
			}
		}
		spatialIndex.Flush(func(a spatial.AABB) {})
	})

	collisionView := ecs.NewView1[Position](engine)
	// System: Collision (Broad-phase using BucketGrid)
	detectSystem := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for head := range collisionView.All() {
			pos := head.V1
			spatialIndex.QueryRange(pos.AABB.AABB, func(otherID uint64) {
				other := ecs.Entity(otherID / 4) // TODO: fix gokg!!
				if head.Entity.Index() < other.Index() {
					ecs.AssignComponent(cb, head.Entity, collisionPairType, CollisionPair{With: other})
				}
			})
		}
	})

	collisionProcessView := ecs.NewView4[Position, Velocity, Appearance, CollisionPair](engine)
	resolveSystem := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for head, tail := range collisionProcessView.All() {
			pos1, vel1, app, collision := head.V1, head.V2, head.V3, tail.V4
			entity2 := collision.With

			pos2, _ := ecs.GetComponent[Position](engine, entity2)
			vel2, _ := ecs.GetComponent[Velocity](engine, entity2)
			app2, _ := ecs.GetComponent[Appearance](engine, entity2)

			app.Color = color.RGBA{255, 0, 0, 255}
			app2.Color = color.RGBA{255, 0, 0, 255}

			resolveCollision(pos1, vel1, pos2, vel2, space)

			cb.RemoveComponent(head.Entity, collisionPairType)
			game.collisionCounter++
		}
	})

	// 4. Execution Plan
	engine.RegisterSystem(moveSystem)
	engine.RegisterSystem(detectSystem)
	engine.RegisterSystem(resolveSystem)
	engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		ctx.Run(moveSystem, d)
		ctx.Sync()
		ctx.Run(detectSystem, d)
		ctx.Sync()
		ctx.Run(resolveSystem, d)
		ctx.Sync()
	})

	// 5. Spawn paricles
	spawnEntities(engine, spatialIndex)

	// 6. Run Ebitengine
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("GOKe + GOKg + Ebiten Integration")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func spawnEntities(engine *ecs.Engine, spatialIndex *spatial.GridIndexManager) {
	gridSize := math.Ceil(math.Sqrt(float64(EntityCount)))
	cols := uint32(gridSize)

	// 2. Calculate dynamic spacing to fill the whole ScreenWidth/Height
	cellWidth := uint32(ScreenWidth / cols)
	cellHeight := uint32(ScreenHeight / cols)

	for i := 0; i < EntityCount; i++ {
		e := engine.CreateEntity()

		row := uint32(i) / cols
		col := uint32(i) % cols

		// 3. Center the entity within its allocated cell
		// Cell center minus half of RectSize
		startX := (col * cellWidth) + (cellWidth / 2) - (RectSize / 2)
		startY := (row * cellHeight) + (cellHeight / 2) - (RectSize / 2)

		startPos := geom.NewVec(startX, startY)
		aabb := plane.NewAABB(startPos, RectSize, RectSize)

		ecs.SetComponent(engine, e, Position{
			AABB: aabb,
			accX: 0,
			accY: 0,
		})

		// Velocity initialization
		dx := int32(rand.Int32N(401) - 200)
		dy := int32(rand.Int32N(401) - 200)

		// Unikaj bardzo małych prędkości, żeby nie stały w miejscu
		if dx >= 0 && dx < 50 {
			dx = 10
		} else if dx < 0 && dx > -50 {
			dx = -10
		}

		ecs.SetComponent(engine, e, Velocity{
			Vec: geom.NewVec[int32](dx, dy),
		})

		ecs.SetComponent(engine, e, Appearance{
			Color: color.RGBA{255, 255, 255, 255},
		})

		spatialIndex.QueueInsert(uint64(e.Index()), aabb.AABB)
	}
	spatialIndex.Flush(func(a spatial.AABB) {})
}

func resolveCollision(
	pos1 *Position, vel1 *Velocity,
	pos2 *Position, vel2 *Velocity,
	space plane.Space2D[uint32],
) {
	// 1. Elastic momentum exchange
	tempVel := vel1.Vec
	vel1.Vec = vel2.Vec
	vel2.Vec = tempVel

	// 2. Calculate object centers
	// (Simplified for flat space; for a torus topology, edge wrapping should ideally
	// be considered, but for small objects and GridIndex, local difference is usually sufficient).

	// Get bounding boxes
	r1 := pos1.AABB.AABB
	r2 := pos2.AABB.AABB

	// Calculate penetration depth
	// (We assume Intersects returned true, so they definitely overlap)

	// Centers
	c1x := int32(r1.TopLeft.X) + int32(RectSize)/2
	c1y := int32(r1.TopLeft.Y) + int32(RectSize)/2
	c2x := int32(r2.TopLeft.X) + int32(RectSize)/2
	c2y := int32(r2.TopLeft.Y) + int32(RectSize)/2

	// Distance along axes
	dx := c1x - c2x
	dy := c1y - c2y

	// Calculate the amount needed to separate (HalfSize + HalfSize = RectSize)
	// Penetration X = RectSize - abs(dx)
	penX := int32(RectSize) - int32(math.Abs(float64(dx)))
	penY := int32(RectSize) - int32(math.Abs(float64(dy)))

	// Separate along the axis of LEAST penetration (Minimum Translation Vector)
	// It's "cheaper" to push them apart the shortest way out.
	if penX < penY {
		// Push along X axis
		push := penX / 2
		if push == 0 {
			push = 1
		} // Safety

		if dx > 0 {
			// Object 1 is to the right, so push it further right
			space.Translate(&pos1.AABB, geom.NewVec[uint32](uint32(push), 0))
			// Casting -push to uint32 in Go acts like modulo subtraction (torus safe)
			space.Translate(&pos2.AABB, geom.NewVec[uint32](uint32(-push), 0))
		} else {
			space.Translate(&pos1.AABB, geom.NewVec[uint32](uint32(-push), 0))
			space.Translate(&pos2.AABB, geom.NewVec[uint32](uint32(push), 0))
		}
	} else {
		// Push along Y axis
		push := penY / 2
		if push == 0 {
			push = 1
		}

		if dy > 0 {
			space.Translate(&pos1.AABB, geom.NewVec[uint32](0, uint32(push)))
			space.Translate(&pos2.AABB, geom.NewVec[uint32](0, uint32(-push)))
		} else {
			space.Translate(&pos1.AABB, geom.NewVec[uint32](0, uint32(-push)))
			space.Translate(&pos2.AABB, geom.NewVec[uint32](0, uint32(push)))
		}
	}
}

func getShift(v, s int32) int32 {
	if v > 0 {
		return s
	}
	if v < 0 {
		return -s
	}
	return 0
}
