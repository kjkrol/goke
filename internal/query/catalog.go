package query

import (
	"fmt"

	"github.com/kjkrol/goke/v3/internal/addr"
	"github.com/kjkrol/goke/v3/internal/arch"
	"github.com/kjkrol/goke/v3/internal/comp"
)

type Catalog struct {
	// blocks hold the Matchers, each block allocated once with room for blockSize and never
	// grown: a query keeps a pointer to its Matcher, so no Matcher ever moves; a full block is
	// followed by a new one
	blocks      [][]Matcher
	blockSize   int
	cc          *comp.DefIndex
	entityIndex *addr.Index
	archCatalog *arch.Catalog
}

func (c *Catalog) Init(cc *comp.DefIndex, entityIndex *addr.Index, archCatalog *arch.Catalog, cfg Config) {
	c.blockSize = max(cfg.Cap, 1)
	c.blocks = [][]Matcher{make([]Matcher, 0, c.blockSize)}
	c.cc = cc
	c.entityIndex = entityIndex
	c.archCatalog = archCatalog
}

// Add allocates the next Matcher and returns a pointer to it, stable for the lifetime of the ECS
// world: past a block's room a new block is allocated, the Matchers already given staying put.
func (c *Catalog) Add() *Matcher {
	last := &c.blocks[len(c.blocks)-1]
	if len(*last) == cap(*last) {
		c.blocks = append(c.blocks, make([]Matcher, 0, c.blockSize))
		last = &c.blocks[len(c.blocks)-1]
	}
	*last = append(*last, Matcher{})
	return &(*last)[len(*last)-1]
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
	for _, block := range c.blocks {
		for i := range block {
			block[i].BakeIfMatch(archetype)
		}
	}
}

func (c *Catalog) Reset() {
	for _, block := range c.blocks {
		for i := range block {
			block[i].Clear()
		}
	}
	c.blocks = [][]Matcher{c.blocks[0][:0]}
}
