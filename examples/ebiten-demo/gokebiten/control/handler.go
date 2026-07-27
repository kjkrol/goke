package control

// EventHandler reacts to this tick's input events. Game calls HandleEvents
// once per tick, right after capture and before the user's Loop.
type EventHandler interface {
	HandleEvents(events *InputEvents)
}
