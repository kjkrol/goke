package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
)

type InputSystem struct {
	*Resources
}

func NewInputSystem(resources *Resources) goke.System {
	return &InputSystem{
		Resources: resources,
	}
}

func (s *InputSystem) Init(ecs *goke.ECS) {}

func (s *InputSystem) Update(sched *goke.CmdBuf, d time.Duration) {
	events := s.GetInputEvents()

	for _, keyEvent := range events.KeyEvents {
		if keyEvent.Key == ebiten.KeySpace && keyEvent.Action == gokebiten.ActionPress {
			s.gamePause = !s.gamePause
		}
	}

	for _, mouseEvent := range events.ClickQueue {
		fmt.Printf("mouseEvent=%v\n", mouseEvent)
	}
}
