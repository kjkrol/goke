package goke

import (
	"github.com/kjkrol/goke/internal/core"
)

// ComponentRegister ensures a component of type T is registered in the engine
// and returns its metadata.
func ComponentRegister[T any](eng *Engine) ComponentInfo {
	return core.EnsureComponentRegistered[T](eng.registry.ComponentsRegistry)
}
