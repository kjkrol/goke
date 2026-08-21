package reg

import (
	"fmt"
	"reflect"
)

// CompToken is a component-type token for Load, produced by LoadComp[T]().
// Load matches tokens against the save file's component directory by name,
// not by position — pass them to Load in any order.
type CompToken struct {
	// Name is the real Go type's name (reflect.Type.String()), used to
	// match this token against a save file's component directory entry.
	Name string

	// Register validates and registers the type. wantSize is the matching
	// directory entry's recorded size, or nil if this token's type is not
	// present in the save file (registered anyway, for forward
	// compatibility with saves made before it existed).
	Register func(r *Registry, wantSize *uint32) error
}

// LoadComp declares that a Load call may need to register component type
// T — see [Registry.Load]. The order tokens are passed to Load does not matter.
func LoadComp[T any]() CompToken {
	t := reflect.TypeFor[T]()
	return CompToken{
		Name: t.String(),
		Register: func(r *Registry, wantSize *uint32) error {
			if wantSize != nil && uint32(t.Size()) != *wantSize {
				return fmt.Errorf("goke: Load: component %q size mismatch: save file has %d bytes, current type has %d", t, *wantSize, t.Size())
			}
			r.RegComp(t)
			return nil
		},
	}
}

// CompProvider is implemented by systems or modules that register their
// own components, exposing them for callers assembling a Load call
// without needing to know a module's internal component types by name.
// Optional — not part of the System interface, so a system with no
// components of its own implements nothing extra. See [ProvidedComps].
type CompProvider interface {
	LoadComps() []CompToken
}

// ProvidedComps collects LoadComps from every value that implements
// CompProvider, in order — values that don't implement it are skipped.
// Convenience for assembling a Load call from a mix of systems, e.g.
// ProvidedComps(movementSystem, collisionSystem).
func ProvidedComps(values ...any) []CompToken {
	var out []CompToken
	for _, v := range values {
		if p, ok := v.(CompProvider); ok {
			out = append(out, p.LoadComps()...)
		}
	}
	return out
}
