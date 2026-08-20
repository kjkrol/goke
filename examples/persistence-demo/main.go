package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kjkrol/goke/v3"
)

type Position struct{ X, Y float32 }
type Score struct{ Value int }

// movementModule stands in for a self-contained module (e.g. a physics
// package) that owns its own components. A caller wiring it up for
// persistence doesn't need to know Position's name — ComponentProvider
// exposes it instead.
type movementModule struct {
	pos goke.Comp[Position]
}

func (m *movementModule) Init(si *goke.SysInit)                   { si.ECS().RegComp[Position]() }
func (m *movementModule) Update(cb *goke.CmdBuf, d time.Duration) {}
func (m *movementModule) LoaderComps() []goke.LoaderComp {
	return []goke.LoaderComp{goke.RegisterFor[Position]()}
}

var (
	_ goke.System            = (*movementModule)(nil)
	_ goke.ComponentProvider = (*movementModule)(nil)
)

func main() {
	path := tempSavePath()
	defer os.Remove(path)

	saveWorld(path)
	loadWorld(path)
}

func tempSavePath() string {
	f, err := os.CreateTemp("", "goke-persistence-demo-*.save")
	if err != nil {
		panic(err)
	}
	f.Close()
	return f.Name()
}

func saveWorld(path string) {
	ecs := goke.New()
	module := &movementModule{}
	ecs.RegSys(module)
	_ = ecs.RegComp[Score]()

	var score goke.Comp[Score]
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&module.pos, &score)
		factory.Create(3)
		for factory.Next() {
			positions := module.pos.Slice(&factory.Cursor)
			scores := score.Slice(&factory.Cursor)
			for i := range positions {
				positions[i] = Position{X: float32(i), Y: float32(i) * 2}
				scores[i] = Score{Value: i * 100}
			}
		}
	}})

	// Nothing may mutate the world while Save writes it — Pause is the
	// required precondition.
	ecs.Pause()
	if err := ecs.Save(path); err != nil {
		panic(err)
	}
	fmt.Println("Saved 3 entities.")
}

func loadWorld(path string) {
	ecs := goke.New()
	module := &movementModule{}

	// Load runs before Setup, before any other registration — it registers
	// components itself, in the file's recorded order. ProvidedComps pulls
	// Position from the module without naming it; Score is named directly,
	// since it belongs to this program, not a module. Argument order
	// doesn't matter — Load matches each token by name, not by position.
	comps := append(goke.ProvidedComps(module), goke.RegisterFor[Score]())
	if err := ecs.Load(path, comps...); err != nil {
		panic(err)
	}

	var score goke.Comp[Score]
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		query = si.NewQueryBuilder(&module.pos, &score).Build()
	}})

	query.All()
	for query.Next() {
		cursor := query.Cursor()
		positions := module.pos.Slice(cursor)
		scores := score.Slice(cursor)
		for i, id := range cursor.IDs {
			fmt.Printf("Restored entity %v: %+v %+v\n", id, positions[i], scores[i])
		}
	}
}
