package query

import (
	"fmt"

	"github.com/kjkrol/goke/internal/addr"
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
)

type Catalog struct {
	views       []View
	cc          *comp.DefIndex
	entityIndex *addr.Index
	archCatalog *arch.Catalog
}

func (c *Catalog) Init(cc *comp.DefIndex, entityIndex *addr.Index, archCatalog *arch.Catalog, cfg Config) {
	c.views = make([]View, 0, cfg.Cap)
	c.cc = cc
	c.entityIndex = entityIndex
	c.archCatalog = archCatalog
}

// Add allocates the next free View slot and returns a stable pointer to it.
// Panics if MaxViews is exceeded — increase MaxViews in const.go if needed.
func (c *Catalog) Add() *View {
	if len(c.views) == cap(c.views) {
		panic("query: view catalog capacity exceeded — increase MaxViews")
	}
	c.views = append(c.views, View{})
	return &c.views[len(c.views)-1]
}

// NewView creates a View using Track/Include/Exclude opts.
// Track[T]() opts register component data columns (accessible via Slice/At);
// Include[T]() opts add filter-only requirements; Exclude[T]() opts add exclusions.
func NewView(c *Catalog, opts ...comp.BlueprintOpt) *View {
	var b comp.Blueprint
	for _, opt := range opts {
		if err := opt(&b, c.cc); err != nil {
			panic(fmt.Sprintf("query: NewView option: %v", err))
		}
	}
	return c.AddView(&b)
}

func (c *Catalog) AddView(blueprint *comp.Blueprint) *View {
	view := c.Add()
	view.Init(c.entityIndex, blueprint)
	for archID := arch.RootID; archID < c.archCatalog.Len(); archID++ {
		view.BakeIfMatch(&c.archCatalog.Archetypes[archID])
	}
	return view
}

func (c *Catalog) OnArchetypeCreated(archetype *arch.Archetype) {
	for i := range c.views {
		c.views[i].BakeIfMatch(archetype)
	}
}

func (c *Catalog) Reset() {
	for i := range c.views {
		c.views[i].Clear()
	}
	c.views = c.views[:0]
}
