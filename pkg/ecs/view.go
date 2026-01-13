package ecs

type queryCore struct {
	reg        *Registry
	mask       ArchetypeMask
	archetypes []*archetype
}

func newQueryCore(reg *Registry, mask ArchetypeMask) queryCore {
	core := queryCore{
		reg:        reg,
		mask:       mask,
		archetypes: make([]*archetype, 0, 16),
	}

	for _, arch := range reg.archetypeRegistry.All() {
		if arch.mask.Contains(mask) {
			core.archetypes = append(core.archetypes, arch)
		}
	}

	return core
}

//
//import (
//	"iter"
//	"unsafe"
//)
//
//// -------------- Query 1 -----------------
//type View1[T1 any] struct {
//	reg  *Registry
//	mask ArchetypeMask
//	id1  ComponentID
//	s1   *ComponentStorage[T1]
//}
//
//func NewView1[T1 any](reg *Registry) *View1[T1] {
//	id1, _ := componentId[T1](reg.componentsRegistry)
//	bitmask := ArchetypeMask{}.Set(id1)
//	s1 := reg.componentsRegistry.storages[id1]
//	return &View1[T1]{reg, bitmask, id1, (*ComponentStorage[T1])(s1)}
//}
//
//func (q *View1[T1]) All() func(func(Entity, *T1) bool) {
//	return q.Filtered(nil)
//}
//
//func (q *View1[T1]) Filtered(entities []Entity) func(func(Entity, *T1) bool) {
//	return func(yield func(Entity, *T1) bool) {
//		if entities == nil {
//			for e, m := range q.reg.entitiesRegistry.active() {
//				if m.Equals(q.mask) {
//					if !yield(e, q.s1.Get(e)) {
//						return
//					}
//				}
//			}
//		} else {
//			for _, e := range entities {
//				if m, ok := q.reg.entitiesRegistry.GetMask(e); ok && m.Equals(q.mask) {
//					if !yield(e, q.s1.Get(e)) {
//						return
//					}
//				}
//			}
//		}
//	}
//}
//
//// -------------- Query 2 --------------
//type View2[T1, T2 any] struct {
//	reg      *Registry
//	mask     ArchetypeMask
//	id1, id2 ComponentID
//	s1       *ComponentStorage[T1]
//	s2       *ComponentStorage[T2]
//}
//
//type Row2[T1, T2 any] struct {
//	V1 *T1
//	V2 *T2
//}
//
//func NewView2[T1, T2 any](reg *Registry) *View2[T1, T2] {
//	id1, _ := componentId[T1](reg.componentsRegistry)
//	id2, _ := componentId[T2](reg.componentsRegistry)
//	bitmask := ArchetypeMask{}.Set(id1).Set(id2)
//	s1 := reg.componentsRegistry.storages[id1]
//	s2 := reg.componentsRegistry.storages[id2]
//	return &View2[T1, T2]{reg, bitmask, id1, id2, (*ComponentStorage[T1])(s1), (*ComponentStorage[T2])(s2)}
//}
//
//func (q *View2[T1, T2]) All() func(func(Entity, Row2[T1, T2]) bool) {
//	return q.Filtered(nil)
//}
//
//func (q *View2[T1, T2]) Filtered(entities []Entity) func(func(Entity, Row2[T1, T2]) bool) {
//	return func(yield func(Entity, Row2[T1, T2]) bool) {
//		if entities == nil {
//			for e, m := range q.reg.entitiesRegistry.active() {
//				if m.Equals(q.mask) {
//					if !yield(e, newRow2(q.s1.Get(e), q.s2.Get(e))) {
//						return
//					}
//				}
//			}
//		} else {
//			for _, e := range entities {
//				if m, ok := q.reg.entitiesRegistry.GetMask(e); ok && m.Equals(q.mask) {
//					if !yield(e, newRow2(q.s1.Get(e), q.s2.Get(e))) {
//						return
//					}
//				}
//			}
//		}
//	}
//}
//
//func newRow2[T1, T2 any](v1 *T1, v2 *T2) Row2[T1, T2] {
//	return Row2[T1, T2]{v1, v2}
//}
//
//// -------------- Query 3 --------------
//type View3[T1, T2, T3 any] struct {
//	reg           *Registry
//	mask          ArchetypeMask
//	id1, id2, id3 ComponentID
//	archetypes    []*archetype
//}
//
//type Row3[T1, T2, T3 any] struct {
//	V1 *T1
//	V2 *T2
//	V3 *T3
//}
//
//func NewView3[T1, T2, T3 any](reg *Registry) *View3[T1, T2, T3] {
//	id1 := ensureComponentRegistered[T1](reg.componentsRegistry)
//	id2 := ensureComponentRegistered[T2](reg.componentsRegistry)
//	id3 := ensureComponentRegistered[T3](reg.componentsRegistry)
//
//	bitmask := ArchetypeMask{}.Set(id1).Set(id2).Set(id3)
//
//	view := &View3[T1, T2, T3]{reg: reg, mask: bitmask, id1: id1, id2: id2, id3: id3}
//
//	for _, arch := range reg.archetypeRegistry.All() {
//		if arch.mask.Contains(bitmask) {
//			view.archetypes = append(view.archetypes, arch)
//		}
//	}
//
//	return view
//}
//
//func (v *View3[T1, T2, T3]) All() iter.Seq2[Entity, Row3[T1, T2, T3]] {
//	return func(yield func(Entity, Row3[T1, T2, T3]) bool) {
//		for _, arch := range v.archetypes {
//			ptr1 := arch.columns[v.id1].data
//			ptr2 := arch.columns[v.id2].data
//			ptr3 := arch.columns[v.id3].data
//
//			size1 := uintptr(arch.columns[v.id1].itemSize)
//			size2 := uintptr(arch.columns[v.id2].itemSize)
//			size3 := uintptr(arch.columns[v.id3].itemSize)
//
//			for i := 0; i < arch.len; i++ {
//				entity := arch.entities[i]
//				row := newRow3(
//					(*T1)(ptr1),
//					(*T2)(ptr2),
//					(*T3)(ptr3),
//				)
//
//				if !yield(entity, row) {
//					return
//				}
//
//				ptr1 = unsafe.Add(ptr1, size1)
//				ptr2 = unsafe.Add(ptr2, size2)
//				ptr3 = unsafe.Add(ptr3, size3)
//			}
//		}
//	}
//}
//
//func (v *View3[T1, T2, T3]) Filtered(entities []Entity) iter.Seq2[Entity, Row3[T1, T2, T3]] {
//	return func(yield func(Entity, Row3[T1, T2, T3]) bool) {
//		for _, e := range entities {
//			mask, ok := v.reg.entitiesRegistry.GetMask(e)
//			if !ok {
//				continue
//			}
//
//			if !mask.Contains(v.mask) {
//				continue
//			}
//
//			arch := v.reg.archetypeRegistry.Get(mask)
//			if arch == nil {
//				continue
//			}
//
//			idx, exists := arch.entityToIndex[e]
//			if !exists {
//				continue
//			}
//
//			row := newRow3(
//				(*T1)(arch.columns[v.id1].GetElement(idx)),
//				(*T2)(arch.columns[v.id2].GetElement(idx)),
//				(*T3)(arch.columns[v.id3].GetElement(idx)),
//			)
//
//			if !yield(e, row) {
//				return
//			}
//		}
//	}
//}
//
//func newRow3[T1, T2, T3 any](v1 *T1, v2 *T2, v3 *T3) Row3[T1, T2, T3] {
//	return Row3[T1, T2, T3]{v1, v2, v3}
//}
