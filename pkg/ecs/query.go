package ecs

// --- Query 1 ---
type CachedQuery1[T1 any] struct {
	reg  *Registry
	mask Bitmask
	id1  ComponentID
}

// --- Query 1 ---

func NewQuery1[T1 any](r *Registry) *CachedQuery1[T1] {
	id1 := registerComponent[T1](r)
	return &CachedQuery1[T1]{r, Bitmask{}.Set(id1), id1}
}

func (q *CachedQuery1[T1]) All() func(func(Entity, *T1) bool) {
	return q.Filtered(nil)
}

func (q *CachedQuery1[T1]) Filtered(ents []Entity) func(func(Entity, *T1) bool) {
	s1 := q.reg.storages[q.id1].(map[Entity]*T1)
	return func(yield func(Entity, *T1) bool) {
		matchAndIterate(q.reg, q.mask, ents, yield, func(e Entity) *T1 { return s1[e] })
	}
}

// --- Query 2 ---
type CachedQuery2[T1, T2 any] struct {
	reg      *Registry
	mask     Bitmask
	id1, id2 ComponentID
}

func NewQuery2[T1, T2 any](r *Registry) *CachedQuery2[T1, T2] {
	id1, id2 := registerComponent[T1](r), registerComponent[T2](r)
	return &CachedQuery2[T1, T2]{r, Bitmask{}.Set(id1).Set(id2), id1, id2}
}

func (q *CachedQuery2[T1, T2]) All() func(func(Entity, Row2[T1, T2]) bool) {
	return q.Filtered(nil)
}

func (q *CachedQuery2[T1, T2]) Filtered(ents []Entity) func(func(Entity, Row2[T1, T2]) bool) {
	s1, s2 := q.reg.storages[q.id1].(map[Entity]*T1), q.reg.storages[q.id2].(map[Entity]*T2)

	return func(yield func(Entity, Row2[T1, T2]) bool) {
		matchAndIterate(q.reg, q.mask, ents, yield, func(e Entity) Row2[T1, T2] {
			return newRow2(s1[e], s2[e]) // Czysto i prosto
		})
	}
}

// --- Query 3 ---
type CachedQuery3[T1, T2, T3 any] struct {
	reg           *Registry
	mask          Bitmask
	id1, id2, id3 ComponentID
}

func NewQuery3[T1, T2, T3 any](r *Registry) *CachedQuery3[T1, T2, T3] {
	id1, id2, id3 := registerComponent[T1](r), registerComponent[T2](r), registerComponent[T3](r)
	return &CachedQuery3[T1, T2, T3]{r, Bitmask{}.Set(id1).Set(id2).Set(id3), id1, id2, id3}
}

func (q *CachedQuery3[T1, T2, T3]) All() func(func(Entity, Row3[T1, T2, T3]) bool) {
	return q.Filtered(nil)
}

func (q *CachedQuery3[T1, T2, T3]) Filtered(ents []Entity) func(func(Entity, Row3[T1, T2, T3]) bool) {
	s1, s2, s3 := q.reg.storages[q.id1].(map[Entity]*T1), q.reg.storages[q.id2].(map[Entity]*T2), q.reg.storages[q.id3].(map[Entity]*T3)

	return func(yield func(Entity, Row3[T1, T2, T3]) bool) {
		matchAndIterate(q.reg, q.mask, ents, yield, func(e Entity) Row3[T1, T2, T3] {
			return newRow3(s1[e], s2[e], s3[e]) // Żadnego zagnieżdżania!
		})
	}
}

func matchAndIterate[R any](r *Registry, mask Bitmask, entities []Entity, yield func(Entity, R) bool, mapper func(Entity) R) {
	if entities == nil {
		for e, m := range r.masks {
			if m.Matches(mask) {
				if !yield(e, mapper(e)) {
					return
				}
			}
		}
	} else {
		for _, e := range entities {
			if m, ok := r.masks[e]; ok && m.Matches(mask) {
				if !yield(e, mapper(e)) {
					return
				}
			}
		}
	}
}

type Row2[T1, T2 any] struct {
	V1 *T1
	V2 *T2
}

type Row3[T1, T2, T3 any] struct {
	V1 *T1
	V2 *T2
	V3 *T3
}

// Konstruktory (skracają kod wewnątrz mapera)
func newRow2[T1, T2 any](v1 *T1, v2 *T2) Row2[T1, T2] {
	return Row2[T1, T2]{v1, v2}
}

func newRow3[T1, T2, T3 any](v1 *T1, v2 *T2, v3 *T3) Row3[T1, T2, T3] {
	return Row3[T1, T2, T3]{v1, v2, v3}
}
