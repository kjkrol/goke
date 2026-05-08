package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
)

func main() {
	resource := NewResource()
	game := gokebiten.NewGame(
		resource.gameProps,
		gokebiten.WithExtraFnPerSec(func(gs gokebiten.GameStats) {
			resource.collisionCounter = 0
			resource.tps = gs.Ticks
		}),
	)
	game.RegisterComponents(InitBaseComponents)

	initSystem := NewEntitiesInitSystem(resource)
	game.RegisterSystem(initSystem)

	moveSystem := NewMoveSystem(resource)
	collisionSystem := NewCollisionSystem(resource)
	game.RegisterScheduledSystem(moveSystem)
	game.RegisterScheduledSystem(collisionSystem)
	game.LogicPlan(func(ctx goke.ExecutionContext, d time.Duration) {
		ctx.Run(moveSystem, d)
		ctx.Sync()
		ctx.Run(collisionSystem, d)
		ctx.Sync()
	})

	entitiesRenderSystem := NewEntitiesRendererSystem(resource)
	statsRendererSystem := NewStatsRendererSystem(resource)
	game.RegisterRenderSystem(entitiesRenderSystem)
	game.RegisterRenderSystem(statsRendererSystem)
	game.RenderPlan(func(screen *ebiten.Image) {
		entitiesRenderSystem.Draw(screen)
		statsRendererSystem.Draw(screen)
	})

	game.Run()
}
