package main

import (
	"time"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
)

func main() {
	resources := NewResources()
	inputAdapter := &DesktopInputAdapter{}
	game := gokebiten.NewGame(resources, inputAdapter)

	game.RegisterSystem(NewEntitiesInitSystem)

	inputSystem := game.RegisterScheduledSystem(NewInputSystem)
	moveSystem := game.RegisterScheduledSystem(NewMoveSystem)
	collisionSystem := game.RegisterScheduledSystem(NewCollisionSystem)
	game.LogicPlan(func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(inputSystem, d)
		if resources.gamePause {
			return
		}
		ctx.Sync()
		ctx.Run(moveSystem, d)
		ctx.Sync()
		ctx.Run(collisionSystem, d)
		ctx.Sync()
	})

	game.RenderSequence(
		NewPreRenderSystem,
		NewEntitiesRendererSystem,
		NewStatsRendererSystem,
	)

	game.Run()
}
