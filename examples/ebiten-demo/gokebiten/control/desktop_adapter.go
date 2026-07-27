package control

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type DesktopAdapter struct{}

func (a *DesktopAdapter) Capture(e *InputEvents) {
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		e.AddKeyEvent(k, ActionPress)
	}
	for _, k := range inpututil.AppendJustReleasedKeys(nil) {
		e.AddKeyEvent(k, ActionRelease)
	}

	currX, currY := ebiten.CursorPosition()
	e.CursorDelta.X = currX - e.MousePos.X
	e.CursorDelta.Y = currY - e.MousePos.Y
	e.MousePos.X = currX
	e.MousePos.Y = currY

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		e.AddClickEvent(currX, currY, ebiten.MouseButtonLeft, ActionPress)
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		e.AddClickEvent(currX, currY, ebiten.MouseButtonLeft, ActionRelease)
	}

	e.Modifiers.Shift = ebiten.IsKeyPressed(ebiten.KeyShift)
	e.Modifiers.Ctrl = ebiten.IsKeyPressed(ebiten.KeyControl)
}
