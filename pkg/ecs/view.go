package ecs

// -------------- Query 1 -----------------
type View1[T1 any] struct {
	reg  *Registry
	mask Bitmask
	id1  ComponentID
	s1   *ComponentStorage[T1]
}

func NewView1[T1 any](reg *Registry) *View1[T1] {
	id1, _ := componentId[T1](reg.componentsRegistry)
	bitmask := Bitmask{}.Set(id1)
	s1 := reg.componentsRegistry.storages[id1].(*ComponentStorage[T1])
	return &View1[T1]{reg, bitmask, id1, s1}
}

func (q *View1[T1]) All() func(func(Entity, *T1) bool) {
	return q.Filtered(nil)
}

func (q *View1[T1]) Filtered(entities []Entity) func(func(Entity, *T1) bool) {
	return func(yield func(Entity, *T1) bool) {
		if entities == nil {
			for e, m := range q.reg.entitiesRegistry.active() {
				if m.Matches(q.mask) {
					if !yield(e, q.s1.Get(e)) {
						return
					}
				}
			}
		} else {
			for _, e := range entities {
				if m, ok := q.reg.entitiesRegistry.mask(e); ok && m.Matches(q.mask) {
					if !yield(e, q.s1.Get(e)) {
						return
					}
				}
			}
		}
	}
}

// -------------- Query 2 --------------
type View2[T1, T2 any] struct {
	reg      *Registry
	mask     Bitmask
	id1, id2 ComponentID
	s1       *ComponentStorage[T1]
	s2       *ComponentStorage[T2]
}

type Row2[T1, T2 any] struct {
	V1 *T1
	V2 *T2
}

func NewView2[T1, T2 any](reg *Registry) *View2[T1, T2] {
	id1, _ := componentId[T1](reg.componentsRegistry)
	id2, _ := componentId[T2](reg.componentsRegistry)
	bitmask := Bitmask{}.Set(id1).Set(id2)
	s1 := reg.componentsRegistry.storages[id1].(*ComponentStorage[T1])
	s2 := reg.componentsRegistry.storages[id2].(*ComponentStorage[T2])
	return &View2[T1, T2]{reg, bitmask, id1, id2, s1, s2}
}

func (q *View2[T1, T2]) All() func(func(Entity, Row2[T1, T2]) bool) {
	return q.Filtered(nil)
}

func (q *View2[T1, T2]) Filtered(entities []Entity) func(func(Entity, Row2[T1, T2]) bool) {
	return func(yield func(Entity, Row2[T1, T2]) bool) {
		if entities == nil {
			for e, m := range q.reg.entitiesRegistry.active() {
				if m.Matches(q.mask) {
					if !yield(e, newRow2(q.s1.Get(e), q.s2.Get(e))) {
						return
					}
				}
			}
		} else {
			for _, e := range entities {
				if m, ok := q.reg.entitiesRegistry.mask(e); ok && m.Matches(q.mask) {
					if !yield(e, newRow2(q.s1.Get(e), q.s2.Get(e))) {
						return
					}
				}
			}
		}
	}
}

func newRow2[T1, T2 any](v1 *T1, v2 *T2) Row2[T1, T2] {
	return Row2[T1, T2]{v1, v2}
}

// -------------- Query 3 --------------
type View3[T1, T2, T3 any] struct {
	reg           *Registry
	mask          Bitmask
	id1, id2, id3 ComponentID
	s1            *ComponentStorage[T1]
	s2            *ComponentStorage[T2]
	s3            *ComponentStorage[T3]
}

type Row3[T1, T2, T3 any] struct {
	V1 *T1
	V2 *T2
	V3 *T3
}

func NewView3[T1, T2, T3 any](reg *Registry) *View3[T1, T2, T3] {
	id1, _ := componentId[T1](reg.componentsRegistry)
	id2, _ := componentId[T2](reg.componentsRegistry)
	id3, _ := componentId[T3](reg.componentsRegistry)

	bitmask := Bitmask{}.Set(id1).Set(id2).Set(id3)
	s1 := reg.componentsRegistry.storages[id1].(*ComponentStorage[T1])
	s2 := reg.componentsRegistry.storages[id2].(*ComponentStorage[T2])
	s3 := reg.componentsRegistry.storages[id3].(*ComponentStorage[T3])
	return &View3[T1, T2, T3]{reg, bitmask, id1, id2, id3, s1, s2, s3}
}

func (q *View3[T1, T2, T3]) All() func(func(Entity, Row3[T1, T2, T3]) bool) {
	return q.Filtered(nil)
}

func (q *View3[T1, T2, T3]) Filtered(entities []Entity) func(func(Entity, Row3[T1, T2, T3]) bool) {
	return func(yield func(Entity, Row3[T1, T2, T3]) bool) {
		if entities == nil {
			for e, m := range q.reg.entitiesRegistry.active() {
				if m.Matches(q.mask) {
					if !yield(e, newRow3(q.s1.Get(e), q.s2.Get(e), q.s3.Get(e))) {
						return
					}
				}
			}
		} else {
			for _, e := range entities {
				if m, ok := q.reg.entitiesRegistry.mask(e); ok && m.Matches(q.mask) {
					if !yield(e, newRow3(q.s1.Get(e), q.s2.Get(e), q.s3.Get(e))) {
						return
					}
				}
			}
		}
	}
}

func newRow3[T1, T2, T3 any](v1 *T1, v2 *T2, v3 *T3) Row3[T1, T2, T3] {
	return Row3[T1, T2, T3]{v1, v2, v3}
}
