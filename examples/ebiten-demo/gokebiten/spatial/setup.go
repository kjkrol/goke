package spatial

import "github.com/kjkrol/goke/v2"

// Setup runs once at registration to populate the world; unlike Controller
// or a scheduled System, it has no per-tick Update and is never run again.
type Setup interface {
	Init(*goke.ECS)
}
