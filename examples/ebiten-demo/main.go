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
	"github.com/kjkrol/goke"
	"github.com/kjkrol/gokg/pkg/geom"
	"github.com/kjkrol/gokg/pkg/plane"
	"github.com/kjkrol/gokg/pkg/spatial"
)

// --- Configuration ---
const (
	ScreenWidth    = 1024
	ScreenHeight   = 1024
	EntityCount    = 2048 * 4
	BucketCapacity = 32
	RectSize       = 2

	// TargetTPS Define a fixed physics time step (e.g., 120Hz for high precision)
	// This decouples physics simulation from the rendering framerate
	TargetTPS   = 60
	PhysicsStep = time.Second / TargetTPS
)

var spatialGridConfig = spatial.GridIndexConfig{
	Resolution:       spatial.Size1024x1024,
	BucketResolution: spatial.Size32x32,
	BucketCapacity:   BucketCapacity,
	OpsBufferSize:    EntityCount,
}

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

// --- Game Loop (Ebitengine Adapter) ---

type Game struct {
	ecs              *goke.ECS
	renderView       *goke.View2[Position, Appearance]
	accumulator      time.Duration
	lastUpdate       time.Time
	collisionCounter float64
	ticks            int       // Raw counter for physics steps
	physicsTPS       int       // Final value to display
	lastTPSUpdate    time.Time // Timer for resetting the counter
}

var pixelImage *ebiten.Image

func init() {
	pixelImage = ebiten.NewImage(RectSize, RectSize)
	pixelImage.Fill(color.White)
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
	maxSteps := 5 // Nie symuluj więcej niż 5 kroków na klatkę, nawet jak laguje
	steps := 0
	for g.accumulator >= PhysicsStep && steps < maxSteps {
		goke.Tick(g.ecs, PhysicsStep)
		g.accumulator -= PhysicsStep
		g.ticks++
		steps++
	}
	// Jeśli po 5 krokach nadal mamy "dług", po prostu go odpuszczamy,
	// żeby gra zwolniła (slow-motion), a nie klatkowała.
	if g.accumulator > PhysicsStep {
		g.accumulator = 0
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

	screen.Fill(color.RGBA{R: 50, G: 50, B: 50, A: 255})

	// Opcje rysowania, alokujemy raz, żeby nie śmiecić pamięci
	op := &ebiten.DrawImageOptions{}

	for head := range g.renderView.All() {
		aabb, app := head.V1, head.V2

		// Resetujemy opcje (geoM - macierz transformacji)
		op.GeoM.Reset()

		// Skalowanie (width, height)
		w := float64(aabb.BottomRight.X - aabb.TopLeft.X)
		h := float64(aabb.BottomRight.Y - aabb.TopLeft.Y)
		op.GeoM.Scale(w, h)

		// Przesunięcie (pozycja X, Y)
		op.GeoM.Translate(float64(aabb.TopLeft.X), float64(aabb.TopLeft.Y))

		// Kolorowanie
		op.ColorScale.Reset()
		op.ColorScale.ScaleWithColor(app.Color)

		// To jest BARDZO szybkie (batching na GPU)
		screen.DrawImage(pixelImage, op)

		// Uwaga: Obsługę "fragmentów" torusa pomijam dla czytelności,
		// ale analogicznie używasz DrawImage zamiast FillRect.
		aabb.AABB.VisitFragments(func(pos plane.FragPosition, box geom.AABB[uint32]) bool {
			return true
		})
	}

	// screen.Fill(color.RGBA{R: 50, G: 50, B: 50, A: 255})
	// for head := range g.renderView.All() {
	// 	aabb, app := head.V1, head.V2

	// 	x := float32(aabb.TopLeft.X)
	// 	y := float32(aabb.TopLeft.Y)
	// 	w := float32(aabb.BottomRight.X - aabb.TopLeft.X)
	// 	h := float32(aabb.BottomRight.Y - aabb.TopLeft.Y)

	// 	vector.FillRect(screen, x, y, w, h, app.Color, true)

	// 	aabb.AABB.VisitFragments(func(pos plane.FragPosition, box geom.AABB[uint32]) bool {
	// 		x := float32(box.TopLeft.X)
	// 		y := float32(box.TopLeft.Y)
	// 		w := float32(box.BottomRight.X - box.TopLeft.X)
	// 		h := float32(box.BottomRight.Y - box.TopLeft.Y)

	// 		vector.FillRect(screen, x, y, w, h, app.Color, true)
	// 		return true
	// 	})
	// }

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

var (
	posDesc goke.ComponentDesc
	velDesc goke.ComponentDesc
	appDesc goke.ComponentDesc
)

func main() {
	ecs := goke.New()

	game := &Game{
		ecs:        ecs,
		renderView: goke.NewView2[Position, Appearance](ecs),
	}

	// 2. Register Components
	posDesc = goke.RegisterComponent[Position](ecs)
	velDesc = goke.RegisterComponent[Velocity](ecs)
	appDesc = goke.RegisterComponent[Appearance](ecs)

	space := plane.NewToroidal2D[uint32](ScreenWidth, ScreenHeight)

	spatialIndex, err := spatial.NewGridIndexManager(space, spatialGridConfig)
	if err != nil {
		log.Fatalf("Failed to create bucket grid: %v", err)
	}

	// 3. Define Systems

	// System: Movement (Torus Topology)
	moveView := goke.NewView3[Position, Velocity, Appearance](ecs)
	moveSystem := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		dt := d.Seconds()
		for head := range moveView.All() {
			pos, vel := head.V1, head.V2
			app := head.V3
			app.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
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
				delta := geom.NewVec(uint32(dx), uint32(dy))
				space.Translate(&pos.AABB, delta)
				spatialIndex.QueueUpdate(uint64(head.Entity), pos.AABB.AABB, true)
			}
		}
		spatialIndex.Flush(func(a spatial.AABB) {})
	})

	collisionView := goke.NewView3[Position, Velocity, Appearance](ecs)
	// System: Collision (Broad-phase using BucketGrid)
	detectSystem := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		for head := range collisionView.All() {
			pos, vel, app := head.V1, head.V2, head.V3
			spatialIndex.QueryRange(pos.AABB.AABB, func(otherID uint64) {
				entity2 := goke.Entity(otherID / 4) // TODO: fix gokg!!
				if head.Entity.Index() < entity2.Index() {
					pos2, _ := goke.GetComponent[Position](ecs, entity2, posDesc)
					vel2, _ := goke.GetComponent[Velocity](ecs, entity2, velDesc)
					app2, _ := goke.GetComponent[Appearance](ecs, entity2, appDesc)

					app.Color = color.RGBA{R: 255, A: 255}
					app2.Color = color.RGBA{R: 255, A: 255}

					resolveCollision(pos, vel, pos2, vel2, space)

					game.collisionCounter++
				}
			})
		}
	})

	// 4. Execution Plan
	goke.RegisterSystem(ecs, moveSystem)
	goke.RegisterSystem(ecs, detectSystem)
	goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(moveSystem, d)
		ctx.Sync()
		ctx.Run(detectSystem, d)
		ctx.Sync()
	})

	// 5. Spawn paricles
	spawnEntities(ecs, spatialIndex)

	// 6. Run Ebitengine
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("GOKe + GOKg + Ebiten Integration")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func spawnEntities(ecs *goke.ECS, spatialIndex *spatial.GridIndexManager) {
	gridSize := math.Ceil(math.Sqrt(float64(EntityCount)))
	cols := uint32(gridSize)

	// 2. Calculate dynamic spacing to fill the whole ScreenWidth/Height
	cellWidth := uint32(ScreenWidth / cols)
	cellHeight := uint32(ScreenHeight / cols)

	blueprint := goke.NewBlueprint3[Position, Velocity, Appearance](ecs)
	for i := 0; i < EntityCount; i++ {
		entity, position, velocity, appearance := blueprint.Create()

		row := uint32(i) / cols
		col := uint32(i) % cols

		// 3. Center the entity within its allocated cell
		// Cell center minus half of RectSize
		startX := (col * cellWidth) + (cellWidth / 2) - (RectSize / 2)
		startY := (row * cellHeight) + (cellHeight / 2) - (RectSize / 2)

		startPos := geom.NewVec(startX, startY)
		aabb := plane.NewAABB(startPos, RectSize, RectSize)

		*position = Position{
			AABB: aabb,
			accX: 0,
			accY: 0,
		}

		// Velocity initialization
		dx := rand.Int32N(401) - 200
		dy := rand.Int32N(401) - 200

		if dx >= 0 && dx < 50 {
			dx = 10
		} else if dx < 0 && dx > -50 {
			dx = -10
		}

		*velocity = Velocity{
			Vec: geom.NewVec(dx, dy),
		}

		*appearance = Appearance{
			Color: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		}

		spatialIndex.QueueInsert(uint64(entity.Index()), aabb.AABB)
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
			space.Translate(&pos1.AABB, geom.NewVec(uint32(push), 0))
			// Casting -push to uint32 in Go acts like modulo subtraction (torus safe)
			space.Translate(&pos2.AABB, geom.NewVec(uint32(-push), 0))
		} else {
			space.Translate(&pos1.AABB, geom.NewVec(uint32(-push), 0))
			space.Translate(&pos2.AABB, geom.NewVec(uint32(push), 0))
		}
	} else {
		// Push along Y axis
		push := penY / 2
		if push == 0 {
			push = 1
		}

		if dy > 0 {
			space.Translate(&pos1.AABB, geom.NewVec(0, uint32(push)))
			space.Translate(&pos2.AABB, geom.NewVec(0, uint32(-push)))
		} else {
			space.Translate(&pos1.AABB, geom.NewVec(0, uint32(-push)))
			space.Translate(&pos2.AABB, geom.NewVec(0, uint32(push)))
		}
	}
}
