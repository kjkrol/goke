package query

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/comp"
)

// catalogMatchers are how many matchers the catalog holds while an archetype is announced.
var catalogMatchers = []int{8, 56, 256, 1024}

// Benchmark_Catalog_OnArchetypeCreated measures announcing a new archetype to a catalog of n
// matchers, none of which it matches: the walk over the matchers and their mask test — what every
// new archetype costs, n times.
func Benchmark_Catalog_OnArchetypeCreated(b *testing.B) {
	type Seen struct{ V float32 }
	type Wanted struct{ V float32 }
	for _, n := range catalogMatchers {
		cat, cc, em := newQueryCatalog()
		seen, wanted := cc.Intern(reflect.TypeFor[Seen]()), cc.Intern(reflect.TypeFor[Wanted]())
		for range n {
			var spec comp.AccessSpec
			spec.Comp(wanted)
			cat.AddMatcher(&spec)
		}
		var spec comp.AccessSpec
		spec.Comp(seen)
		f := em.CreateFactory(spec)
		f.Create(1)
		f.Next()
		var archetype *arch.Archetype
		for id := arch.RootID; id < em.ArchCatalog.Len(); id++ {
			if a := &em.ArchCatalog.Archetypes[id]; !a.Mask().IsEmpty() {
				archetype = a
			}
		}
		b.Run(fmt.Sprintf("matchers=%d", n), func(b *testing.B) {
			for b.Loop() {
				cat.OnArchetypeCreated(archetype)
			}
		})
	}
}
