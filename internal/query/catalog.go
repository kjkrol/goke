package query

import (
	"github.com/kjkrol/goke/internal/arch"
	"github.com/kjkrol/goke/internal/comp"
	"github.com/kjkrol/goke/internal/ent"
)

type Catalog struct {
	views []View
	cc    *comp.Catalog
	em    *ent.Manager
}

func (c *Catalog) Init(cc *comp.Catalog, em *ent.Manager, cfg Config) {
	c.views = make([]View, 0, cfg.Cap)
	c.cc = cc
	c.em = em
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
	var s comp.Spec0
	s.Init(c.cc, opts...)
	return c.AddView(&s.Blueprint)
}

func (c *Catalog) AddView(blueprint *comp.Blueprint) *View {
	view := c.Add()
	view.Init(&c.em.Index, blueprint)
	for archID := arch.RootID; archID < c.em.ArchCatalog.Len(); archID++ {
		view.BakeIfMatch(&c.em.ArchCatalog.Archetypes[archID])
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
