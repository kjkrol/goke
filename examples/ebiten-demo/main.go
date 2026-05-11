package main

import (
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
)

func main() {
	resources := NewResources()
	game := gokebiten.NewGame(resources)

	game.RegisterSystem(NewEntitiesInitSystem)

	moveSystem := game.RegisterScheduledSystem(NewMoveSystem)
	collisionSystem := game.RegisterScheduledSystem(NewCollisionSystem)
	game.LogicPlan(func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(moveSystem, d)
		ctx.Sync()
		ctx.Run(collisionSystem, d)
		ctx.Sync()
	})

	game.RenderSequence(NewEntitiesRendererSystem, NewStatsRendererSystem)

	game.Run()
}
