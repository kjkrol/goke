package main

import (
	"fmt"
	"os"

	"github.com/kjkrol/goke/v3"
)

type Position struct{ X, Y float32 }
type Score struct{ Value int }

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
	_ = ecs.RegComp[Position]()
	_ = ecs.RegComp[Score]()

	var pos goke.Comp[Position]
	var score goke.Comp[Score]
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&pos, &score)
		factory.Create(3)
		for factory.Next() {
			positions := pos.Slice(&factory.Cursor)
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

	// Load runs before Setup, before any other registration — it registers
	// each named component itself, in the file's recorded order. Argument
	// order doesn't matter — Load matches each token by name, not position.
	if err := ecs.Load(path, goke.LoadComp[Position](), goke.LoadComp[Score]()); err != nil {
		panic(err)
	}

	var pos goke.Comp[Position]
	var score goke.Comp[Score]
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		query = si.NewQueryBuilder(&pos, &score).Build()
	}})

	query.All()
	for query.Next() {
		cursor := query.Cursor()
		positions := pos.Slice(cursor)
		scores := score.Slice(cursor)
		for i, id := range cursor.IDs {
			fmt.Printf("Restored entity %v: %+v %+v\n", id, positions[i], scores[i])
		}
	}
}
