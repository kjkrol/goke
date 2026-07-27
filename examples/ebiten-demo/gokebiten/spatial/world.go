package spatial

import (
	"fmt"
	"log"
	"math"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/kinematics"
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/gokg"
	"github.com/kjkrol/gokg/spatial"
	"github.com/kjkrol/uid"
)

type Config struct {
	Width, Height uint32
	Toroidal      bool
}

type Population struct {
	MaxCount int
	MinSize  uint32
	MaxSize  uint32
}

// EntityExtras supplies the components and per-entity values a game needs
// beyond Position/Velocity — Populate/PopulateStatic own entity creation,
// but have no way to know a game's own component types on their own. index
// is the entity's position in the overall spawn (matching Spawner/Placement);
// i is its position within the current internal batch, for slicing components.
type EntityExtras interface {
	Components() []goke.Addable
	Init(cursor *goke.Cursor, i, index int, id uid.UID64)
}

// World owns the spatial index and population bookkeeping for one game.
type World struct {
	ecs          *goke.ECS
	space        *gokg.Space
	population   Population
	spawnedCount int
}

// buildSpace picks a bucket resolution and capacity from world/population
// density: capacity = round(1/sqrt(density)) clamped to [2, 8]; sparser
// worlds get bigger buckets, denser worlds get finer ones.
func buildSpace(cfg Config, pop Population) *gokg.Space {
	const minCapacity, maxCapacity = 2.0, 8.0

	worldArea := uint64(cfg.Width) * uint64(cfg.Height)
	entityArea := uint64(pop.MaxSize) * uint64(pop.MaxSize)
	density := float64(uint64(pop.MaxCount)*entityArea) / float64(worldArea)

	raw := math.Round(1.0 / math.Sqrt(density))
	capacity := uint32(math.Max(minCapacity, math.Min(maxCapacity, raw)))
	bucketResolution := spatial.ResolutionFrom(pop.MaxSize * capacity)

	log.Printf("[spatial] maxEntities=%d, density=%.2f%%, capacity=%d → bucket=%dx%d, bucketCap=%d, opsBuffer=%d",
		pop.MaxCount, density*100, capacity,
		bucketResolution.Side(), bucketResolution.Side(),
		capacity*capacity, pop.MaxCount*8)

	space, err := gokg.NewSpace(gokg.Config{
		Width:          cfg.Width,
		Height:         cfg.Height,
		Toroidal:       cfg.Toroidal,
		BucketSize:     bucketResolution,
		BucketCapacity: int(capacity * capacity),
		OpsBufferSize:  pop.MaxCount * 8,
	})
	if err != nil {
		panic(fmt.Sprintf("spatial: invalid space configuration: %v", err))
	}
	return space
}

// NewWorld builds the spatial index from cfg, provisioned for pop.
func NewWorld(ecs *goke.ECS, cfg Config, pop Population) *World {
	return &World{ecs: ecs, space: buildSpace(cfg, pop), population: pop}
}

func (w *World) Space() *gokg.Space     { return w.space }
func (w *World) Population() Population { return w.population }

// reserve checks and accounts for count more entities against the
// MaxCount declared to NewWorld, shared between Populate and PopulateStatic.
func (w *World) reserve(count int) {
	if w.spawnedCount+count > w.population.MaxCount {
		panic(fmt.Sprintf("spatial: spawning %d more would exceed Population.MaxCount %d (already spawned %d)",
			count, w.population.MaxCount, w.spawnedCount))
	}
	w.spawnedCount += count
}

func (w *World) validateSize(id uid.UID64, pos kinematics.Position) {
	if pos.Size.X < w.population.MinSize || pos.Size.X > w.population.MaxSize ||
		pos.Size.Y < w.population.MinSize || pos.Size.Y > w.population.MaxSize {
		panic(fmt.Sprintf("spatial: entity %d size %dx%d outside declared population bounds [%d, %d]",
			id, pos.Size.X, pos.Size.Y, w.population.MinSize, w.population.MaxSize))
	}
}

// Populate spawns count moving entities: Position and Velocity from
// populators[0] (which must implement kinematics.Spawner), everything else
// from the rest of populators.
func (w *World) Populate(count int, telemetry *kinematics.Telemetry, populators ...EntityExtras) {
	if len(populators) == 0 {
		panic("spatial: Populate requires a kinematics.Spawner as its first populator")
	}
	spawner, ok := populators[0].(kinematics.Spawner)
	if !ok {
		panic("spatial: Populate's first populator must implement kinematics.Spawner")
	}
	w.reserve(count)
	telemetry.DynamicCount = count

	var posComp goke.Comp[kinematics.Position]
	var velComp goke.Comp[kinematics.Velocity]
	var collisionComp goke.Comp[collisions.Collision]
	comps := []goke.Addable{&posComp, &velComp, &collisionComp}
	for _, p := range populators {
		comps = append(comps, p.Components()...)
	}
	factory := w.ecs.NewFactory(comps...)

	factory.Create(count)
	index := 0
	for factory.Next() {
		positions := posComp.Slice(&factory.Cursor)
		velocities := velComp.Slice(&factory.Cursor)
		for i, id := range factory.IDs {
			pos, vel := spawner.Spawn(index, count)
			w.validateSize(id, pos)
			positions[i] = pos
			velocities[i] = vel
			for _, p := range populators {
				p.Init(&factory.Cursor, i, index, id)
			}

			w.space.Insert(uint64(id), positions[i].AABB)
			index++
		}
	}
	w.space.Flush(nil)
}

// PopulateStatic spawns count entities with Position only — e.g. level
// geometry that never moves — from populators[0] (which must implement
// kinematics.Placement) and the rest of populators.
func (w *World) PopulateStatic(count int, telemetry *kinematics.Telemetry, populators ...EntityExtras) {
	if len(populators) == 0 {
		panic("spatial: PopulateStatic requires a kinematics.Placement as its first populator")
	}
	placement, ok := populators[0].(kinematics.Placement)
	if !ok {
		panic("spatial: PopulateStatic's first populator must implement kinematics.Placement")
	}
	w.reserve(count)
	telemetry.StaticCount = count

	var posComp goke.Comp[kinematics.Position]
	comps := []goke.Addable{&posComp}
	for _, p := range populators {
		comps = append(comps, p.Components()...)
	}
	factory := w.ecs.NewFactory(comps...)

	factory.Create(count)
	index := 0
	for factory.Next() {
		positions := posComp.Slice(&factory.Cursor)
		for i, id := range factory.IDs {
			pos := placement.Place(index, count)
			w.validateSize(id, pos)
			positions[i] = pos
			for _, p := range populators {
				p.Init(&factory.Cursor, i, index, id)
			}

			w.space.Insert(uint64(id), positions[i].AABB)
			index++
		}
	}
	w.space.Flush(nil)
}
