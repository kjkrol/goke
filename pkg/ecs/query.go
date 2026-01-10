package ecs

// --------------------------------------
type CachedQuery1[T1 any] struct {
	reg  *Registry
	mask Bitmask
	id1  ComponentID
}

func NewQuery1[T1 any](reg *Registry) *CachedQuery1[T1] {
	id1 := registerComponent[T1](reg)
	return &CachedQuery1[T1]{
		reg:  reg,
		id1:  id1,
		mask: Bitmask{}.Set(id1),
	}
}

func (q *CachedQuery1[T1]) All() func(yield func(Entity, *T1) bool) {
	s1 := q.reg.storages[q.id1].(map[Entity]*T1)
	return func(yield func(Entity, *T1) bool) {
		for e, m := range q.reg.masks {
			if m.Matches(q.mask) {
				if !yield(e, s1[e]) {
					return
				}
			}
		}
	}
}

// --------------------------------------
type CachedQuery2[T1, T2 any] struct {
	reg      *Registry
	mask     Bitmask
	id1, id2 ComponentID
}

type Row2[T1, T2 any] struct {
	V1 *T1
	V2 *T2
}

func NewQuery2[T1, T2 any](reg *Registry) *CachedQuery2[T1, T2] {
	id1, id2 := registerComponent[T1](reg), registerComponent[T2](reg)
	return &CachedQuery2[T1, T2]{
		reg: reg,
		id1: id1, id2: id2,
		mask: Bitmask{}.Set(id1).Set(id2),
	}
}

func (q *CachedQuery2[T1, T2]) All() func(yield func(Entity, Row2[T1, T2]) bool) {
	s1 := q.reg.storages[q.id1].(map[Entity]*T1)
	s2 := q.reg.storages[q.id2].(map[Entity]*T2)
	return func(yield func(Entity, Row2[T1, T2]) bool) {
		for e, m := range q.reg.masks {
			if m.Matches(q.mask) {
				if !yield(e, Row2[T1, T2]{s1[e], s2[e]}) {
					return
				}
			}
		}
	}
}

// --------------------------------------
type CachedQuery3[T1, T2, T3 any] struct {
	reg           *Registry
	mask          Bitmask
	id1, id2, id3 ComponentID
}

type Row3[T1, T2, T3 any] struct {
	V1 *T1
	V2 *T2
	V3 *T3
}

func NewQuery3[T1, T2, T3 any](reg *Registry) *CachedQuery3[T1, T2, T3] {
	id1, id2, id3 := registerComponent[T1](reg), registerComponent[T2](reg), registerComponent[T3](reg)
	return &CachedQuery3[T1, T2, T3]{
		reg: reg,
		id1: id1, id2: id2, id3: id3,
		mask: Bitmask{}.Set(id1).Set(id2).Set(id3),
	}
}

func (q *CachedQuery3[T1, T2, T3]) All() func(yield func(Entity, Row3[T1, T2, T3]) bool) {
	s1 := q.reg.storages[q.id1].(map[Entity]*T1)
	s2 := q.reg.storages[q.id2].(map[Entity]*T2)
	s3 := q.reg.storages[q.id3].(map[Entity]*T3)
	return func(yield func(Entity, Row3[T1, T2, T3]) bool) {
		for e, m := range q.reg.masks {
			if m.Matches(q.mask) {
				if !yield(e, Row3[T1, T2, T3]{s1[e], s2[e], s3[e]}) {
					return
				}
			}
		}
	}
}
