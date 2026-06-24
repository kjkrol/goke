package query

import (
	"fmt"

	"github.com/kjkrol/goke/v2/internal/addr"
	"github.com/kjkrol/goke/v2/internal/arch"
	"github.com/kjkrol/goke/v2/internal/comp"
)

type Catalog struct {
	matchers    []Matcher
	cc          *comp.DefIndex
	entityIndex *addr.Index
	archCatalog *arch.Catalog
}

func (c *Catalog) Init(cc *comp.DefIndex, entityIndex *addr.Index, archCatalog *arch.Catalog, cfg Config) {
	c.matchers = make([]Matcher, 0, cfg.Cap)
	c.cc = cc
	c.entityIndex = entityIndex
	c.archCatalog = archCatalog
}

// Add allocates the next free Matcher slot and returns a stable pointer to it.
// Panics if MaxMatchers is exceeded — increase MaxMatchers in const.go if needed.
func (c *Catalog) Add() *Matcher {
	if len(c.matchers) == cap(c.matchers) {
		panic("query: matcher catalog capacity exceeded — increase MaxMatchers")
	}
	c.matchers = append(c.matchers, Matcher{})
	return &c.matchers[len(c.matchers)-1]
}

// NewMatcher creates a Matcher using Track/Include/Exclude opts.
// Track[T]() opts register component data columns (accessible via Slice/At);
// Include[T]() opts add filter-only requirements; Exclude[T]() opts add exclusions.
func NewMatcher(c *Catalog, opts ...comp.AccessOpt) *Matcher {
	var accessSpec comp.AccessSpec
	for _, opt := range opts {
		if err := opt(&accessSpec, c.cc); err != nil {
			panic(fmt.Sprintf("query: NewMatcher option: %v", err))
		}
	}
	return c.AddMatcher(&accessSpec)
}

func (c *Catalog) AddMatcher(accessSpec *comp.AccessSpec) *Matcher {
	matcher := c.Add()
	matcher.Init(c.entityIndex, c.archCatalog, accessSpec)
	for archID := arch.RootID; archID < c.archCatalog.Len(); archID++ {
		matcher.BakeIfMatch(&c.archCatalog.Archetypes[archID])
	}
	return matcher
}

func (c *Catalog) OnArchetypeCreated(archetype *arch.Archetype) {
	for i := range c.matchers {
		c.matchers[i].BakeIfMatch(archetype)
	}
}

func (c *Catalog) Reset() {
	for i := range c.matchers {
		c.matchers[i].Clear()
	}
	c.matchers = c.matchers[:0]
}
