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
	blocks [][]Matcher
	// rules hold, at the Matchers' places, what each matches: a new archetype is tried against
	// the rules alone
	rules       [][]rule
	blockSize   int
	cc          *comp.DefIndex
	entityIndex *addr.Index
	archCatalog *arch.Catalog
}

func (c *Catalog) Init(cc *comp.DefIndex, entityIndex *addr.Index, archCatalog *arch.Catalog, cfg Config) {
	c.blockSize = max(cfg.Cap, 1)
	c.blocks = [][]Matcher{make([]Matcher, 0, c.blockSize)}
	c.rules = [][]rule{make([]rule, 0, c.blockSize)}
	c.cc = cc
	c.entityIndex = entityIndex
	c.archCatalog = archCatalog
}

// add allocates the next Matcher, stable for the lifetime of the ECS world, and its rule: past a
// block's room a new block is allocated, the Matchers already given staying put.
func (c *Catalog) add() (*Matcher, *rule) {
	last := len(c.blocks) - 1
	if len(c.blocks[last]) == cap(c.blocks[last]) {
		c.blocks = append(c.blocks, make([]Matcher, 0, c.blockSize))
		c.rules = append(c.rules, make([]rule, 0, c.blockSize))
		last++
	}
	c.blocks[last] = append(c.blocks[last], Matcher{})
	c.rules[last] = append(c.rules[last], rule{})
	n := len(c.blocks[last]) - 1
	return &c.blocks[last][n], &c.rules[last][n]
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
	r := newRule(accessSpec)
	matcher, place := c.add()
	*place = r
	matcher.Init(c.entityIndex, c.archCatalog, accessSpec)
	for archID := arch.RootID; archID < c.archCatalog.Len(); archID++ {
		matcher.bake(&c.archCatalog.Archetypes[archID], place)
	}
	return matcher
}

func (c *Catalog) OnArchetypeCreated(archetype *arch.Archetype) {
	mask := archetype.Mask()
	for b, rules := range c.rules {
		for i := range rules {
			if r := &rules[i]; r.wants(mask) {
				c.blocks[b][i].bake(archetype, r)
			}
		}
	}
}

func (c *Catalog) Reset() {
	for _, block := range c.blocks {
		for i := range block {
			block[i].Clear()
		}
	}
	for _, rules := range c.rules {
		clear(rules)
	}
	c.blocks = [][]Matcher{c.blocks[0][:0]}
	c.rules = [][]rule{c.rules[0][:0]}
}
