package gokebiten

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Typy akcji dla zdarzeń dyskretnych
type KeyAction int

const (
	ActionPress   KeyAction = iota // Moment wciśnięcia
	ActionRelease                  // Moment puszczenia
)

// Zdarzenie naciśnięcia/puszczenia klawisza
type KeyEvent struct {
	Key    ebiten.Key
	Action KeyAction
}

// Zdarzenie interakcji myszy
type ClickEvent struct {
	Pos    image.Point
	Button ebiten.MouseButton
	Action KeyAction
}

// Struktura przechowująca pełny obraz klatki (Snapshot)
type InputEvents struct {
	// --- STAN CIĄGŁY (Aktualizowany co klatkę) ---
	MousePos    image.Point
	CursorDelta image.Point
	Modifiers   struct {
		Shift, Ctrl, Alt bool
	}

	// --- ZDARZENIA CHWILOWE (Czyszczone po zakończeniu wszystkich kroków fizyki) ---
	ClickQueue  []ClickEvent
	KeyEvents   []KeyEvent
	ScrollDelta float64
}

func (e *InputEvents) ResetTransient() {
	e.ClickQueue = e.ClickQueue[:0]
	e.KeyEvents = e.KeyEvents[:0]
	e.ScrollDelta = 0
	e.CursorDelta = image.Pt(0, 0)
}

func (e *InputEvents) AddKeyEvent(key ebiten.Key, action KeyAction) {
	e.KeyEvents = append(e.KeyEvents, KeyEvent{Key: key, Action: action})
}

func (e *InputEvents) AddClickEvent(x, y int, button ebiten.MouseButton, action KeyAction) {
	e.ClickQueue = append(e.ClickQueue, ClickEvent{
		Pos:    image.Pt(x, y),
		Button: button,
		Action: action,
	})
}

type InputAdapter interface {
	Capture(e *InputEvents)
}
