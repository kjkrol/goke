package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
)

type PreRenderSystem struct {
	*Resources
}

var _ (gokebiten.RenderSystem) = (*PreRenderSystem)(nil)

func NewPreRenderSystem(resources *Resources) gokebiten.RenderSystem {
	return &PreRenderSystem{
		Resources: resources,
	}
}

func (s *PreRenderSystem) Update(lookup goke.Lookup, sched *goke.CmdBuf, d time.Duration) {}
func (s *PreRenderSystem) Init(ecs *goke.ECS)                                             {}
func (s *PreRenderSystem) Draw(screen *ebiten.Image)                                      {}
