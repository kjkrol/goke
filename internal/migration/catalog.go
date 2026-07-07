package migration

import "github.com/kjkrol/goke/v2/internal/arch"

// Catalog tracks registered Migrators and forwards new-archetype notifications
// to each of them so that destination lookups stay O(1) on the Apply hot path.
type Catalog struct {
	migrators   []*Migrator
	archCatalog *arch.Catalog
}

func (c *Catalog) Init(archCatalog *arch.Catalog) {
	c.archCatalog = archCatalog
}

// Add registers m and seeds it with all archetypes created since m was constructed
// (including destination archetypes created during New's seeding pass).
func (c *Catalog) Add(m *Migrator) {
	for archID := arch.RootID; archID < c.archCatalog.Len(); archID++ {
		m.OnNewArchetype(archID)
	}
	c.migrators = append(c.migrators, m)
}

// OnArchetypeCreated forwards the new archetype to every registered Migrator.
func (c *Catalog) OnArchetypeCreated(archetype *arch.Archetype) {
	for _, m := range c.migrators {
		m.OnNewArchetype(archetype.Id)
	}
}

func (c *Catalog) Reset() {
	c.migrators = c.migrators[:0]
}
