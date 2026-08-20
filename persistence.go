package goke

import (
	"fmt"
	"os"
	"reflect"

	"github.com/kjkrol/goke/v3/internal/persist"
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
	Register func(ecs *ECS, wantSize *uint32) error
}

// LoadComp declares that a Load call may need to register component type
// T — see [ECS.Load]. The order tokens are passed to Load does not matter.
func LoadComp[T any]() CompToken {
	t := reflect.TypeFor[T]()
	return CompToken{
		Name: t.String(),
		Register: func(ecs *ECS, wantSize *uint32) error {
			if wantSize != nil && uint32(t.Size()) != *wantSize {
				return fmt.Errorf("goke: Load: component %q size mismatch: save file has %d bytes, current type has %d", t, *wantSize, t.Size())
			}
			ecs.RegComp[T]()
			return nil
		},
	}
}

// CompProvider is implemented by systems or modules that register their
// own components, exposing them for callers assembling an ECS.Load call
// without needing to know a module's internal component types by name.
// Optional — not part of the System interface, so a system with no
// components of its own implements nothing extra. See [ProvidedComps].
type CompProvider interface {
	LoadComps() []CompToken
}

// ProvidedComps collects LoadComps from every value that implements
// CompProvider, in order — values that don't implement it are skipped.
// Convenience for assembling an ECS.Load call from a mix of systems, e.g.
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

// Save writes a full snapshot of the world — every entity and its
// components, with their original IDs preserved — to the file at path.
// Requires a prior [ECS.Pause] (panics otherwise), since nothing may mutate
// the world while the snapshot is being written.
func (ecs *ECS) Save(path string) error {
	if !ecs.paused {
		panic("goke: Save called without a prior Pause()")
	}
	ecs.saving = true
	defer func() { ecs.saving = false }()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return ecs.registry.Save(f)
}

// Load reads a snapshot written by Save, rebuilding its components,
// archetypes, and entities — with their original IDs — into ecs. Must be
// called before Setup and before any other component registration (panics
// otherwise) — Load registers components itself, in the file's recorded
// order, matching the given comps by name (see [LoadComp], [CompProvider]).
func (ecs *ECS) Load(path string, comps ...CompToken) error {
	if ecs.registry.CompDefIndex.Count() > 0 {
		panic("goke: Load called after other components were already registered — Load must run first, before Setup and before any RegComp call")
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	requests := make([]persist.CompRequest, len(comps))
	for i, c := range comps {
		c := c
		requests[i] = persist.CompRequest{
			Name: c.Name,
			Register: func(wantSize *uint32) error {
				return c.Register(ecs, wantSize)
			},
		}
	}

	return ecs.registry.Load(f, requests)
}
