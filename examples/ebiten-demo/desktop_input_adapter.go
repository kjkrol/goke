package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
)

type DesktopInputAdapter struct{}

func (a *DesktopInputAdapter) Capture(e *gokebiten.InputEvents) {
	// 1. Klawiatura
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		e.AddKeyEvent(k, gokebiten.ActionPress)
	}
	for _, k := range inpututil.AppendJustReleasedKeys(nil) {
		e.AddKeyEvent(k, gokebiten.ActionRelease)
	}

	// 2. Mysz
	currX, currY := ebiten.CursorPosition()
	e.CursorDelta.X = currX - e.MousePos.X
	e.CursorDelta.Y = currY - e.MousePos.Y
	e.MousePos.X = currX
	e.MousePos.Y = currY

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		e.AddClickEvent(currX, currY, ebiten.MouseButtonLeft, gokebiten.ActionPress)
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		e.AddClickEvent(currX, currY, ebiten.MouseButtonLeft, gokebiten.ActionRelease)
	}

	// 3. Modyfikatory
	e.Modifiers.Shift = ebiten.IsKeyPressed(ebiten.KeyShift)
	e.Modifiers.Ctrl = ebiten.IsKeyPressed(ebiten.KeyControl)
}
