package comp

import "reflect"

// Catalog is the central registry for component types.
// It maps Go types to component metadata and resolves Blueprints into Sets.
type Catalog struct {
	index metaIndex
}

func (c *Catalog) Init() {
	c.index.Init()
}

// Intern interns a Go type as a component and returns its Meta.
// Calling Intern twice for the same type returns the same Meta.
func (c *Catalog) Intern(t reflect.Type) Meta {
	return c.index.intern(t)
}

// ByType looks up a registered component by its Go type.
func (c *Catalog) ByType(t reflect.Type) (Meta, bool) {
	return c.index.byType(t)
}

// ByID looks up a registered component by its ID.
func (c *Catalog) ByID(id ID) Meta {
	return c.index.byID(id)
}

// Compose resolves a Blueprint into a fully populated Composition using registered metadata.
func (c *Catalog) Compose(b *Blueprint) Composition {
	mask := Mask{}.Build(b)
	var metas []Meta
	for id := range mask.AllSet() {
		m := c.index.byID(id)
		if m.Size > 0 {
			metas = append(metas, m)
		}
	}
	return Composition{Mask: mask, Metas: metas}
}

func (c *Catalog) Reset() {
	c.index.reset()
}
