package render

import (
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/spatial"
	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

var _ spatial.EntityExtras = (*AppearanceExtras)(nil)

// AppearanceExtras is a ready-made EntityExtras: it adds Appearance to each
// spawned entity, calling look(index) to decide each one's initial value —
// the game owns the actual color/sprite policy, this just wires it in.
type AppearanceExtras struct {
	fAppear goke.Comp[Appearance]
	look    func(index int) Appearance
}

func NewAppearanceExtras(look func(index int) Appearance) *AppearanceExtras {
	return &AppearanceExtras{look: look}
}

func (e *AppearanceExtras) Components() []goke.Addable {
	return []goke.Addable{&e.fAppear}
}

func (e *AppearanceExtras) Init(cursor *goke.Cursor, i, index int, _ uid.UID64) {
	e.fAppear.Slice(cursor)[i] = e.look(index)
}
