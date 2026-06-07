package goke

import (
	"fmt"
	"iter"
	"reflect"
	"unsafe"

    "github.com/kjkrol/goke/internal/core"
)


// --------------- View1 ---------------

type View1[T1 any] struct {
    *core.View
}

func NewView1[T1 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View1[T1] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view1 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    mustAdd(info1)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view1 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View1[T1]{View: view}
}

func (v *View1[T1]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View1[T1]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View2 ---------------

type View2[T1 any, T2 any] struct {
    *core.View
}

func NewView2[T1 any, T2 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View2[T1, T2] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view2 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    mustAdd(info1)
    mustAdd(info2)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view2 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View2[T1, T2]{View: view}
}

func (v *View2[T1, T2]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View2[T1, T2]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View3 ---------------

type View3[T1 any, T2 any, T3 any] struct {
    *core.View
}

func NewView3[T1 any, T2 any, T3 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View3[T1, T2, T3] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view3 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view3 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, info3, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View3[T1, T2, T3]{View: view}
}

func (v *View3[T1, T2, T3]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
	Comp3 []T3
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
			Comp3 []T3
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
						Comp3 []T3
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3: unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View3[T1, T2, T3]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			c2Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[2]+(row*ma.CompSizes[2]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
				Comp3: (*T3)(c2Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View4 ---------------

type View4[T1 any, T2 any, T3 any, T4 any] struct {
    *core.View
}

func NewView4[T1 any, T2 any, T3 any, T4 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View4[T1, T2, T3, T4] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view4 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view4 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, info3, info4, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View4[T1, T2, T3, T4]{View: view}
}

func (v *View4[T1, T2, T3, T4]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
	Comp3 []T3
	Comp4 []T4
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
			Comp3 []T3
			Comp4 []T4
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
						Comp3 []T3
						Comp4 []T4
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3: unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4: unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View4[T1, T2, T3, T4]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
		Comp4  *T4
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			c2Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[2]+(row*ma.CompSizes[2]))
			c3Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[3]+(row*ma.CompSizes[3]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
				Comp3: (*T3)(c2Ptr),
				Comp4: (*T4)(c3Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View5 ---------------

type View5[T1 any, T2 any, T3 any, T4 any, T5 any] struct {
    *core.View
}

func NewView5[T1 any, T2 any, T3 any, T4 any, T5 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View5[T1, T2, T3, T4, T5] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view5 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view5 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View5[T1, T2, T3, T4, T5]{View: view}
}

func (v *View5[T1, T2, T3, T4, T5]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
	Comp3 []T3
	Comp4 []T4
	Comp5 []T5
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
			Comp3 []T3
			Comp4 []T4
			Comp5 []T5
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
						Comp3 []T3
						Comp4 []T4
						Comp5 []T5
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3: unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4: unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5: unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View5[T1, T2, T3, T4, T5]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
		Comp4  *T4
		Comp5  *T5
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			c2Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[2]+(row*ma.CompSizes[2]))
			c3Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[3]+(row*ma.CompSizes[3]))
			c4Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[4]+(row*ma.CompSizes[4]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
				Comp5  *T5
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
				Comp3: (*T3)(c2Ptr),
				Comp4: (*T4)(c3Ptr),
				Comp5: (*T5)(c4Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View6 ---------------

type View6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any] struct {
    *core.View
}

func NewView6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View6[T1, T2, T3, T4, T5, T6] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view6 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view6 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View6[T1, T2, T3, T4, T5, T6]{View: view}
}

func (v *View6[T1, T2, T3, T4, T5, T6]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
	Comp3 []T3
	Comp4 []T4
	Comp5 []T5
	Comp6 []T6
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
			Comp3 []T3
			Comp4 []T4
			Comp5 []T5
			Comp6 []T6
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
						Comp3 []T3
						Comp4 []T4
						Comp5 []T5
						Comp6 []T6
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3: unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4: unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5: unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6: unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View6[T1, T2, T3, T4, T5, T6]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
	Comp6  *T6
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
		Comp4  *T4
		Comp5  *T5
		Comp6  *T6
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			c2Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[2]+(row*ma.CompSizes[2]))
			c3Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[3]+(row*ma.CompSizes[3]))
			c4Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[4]+(row*ma.CompSizes[4]))
			c5Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[5]+(row*ma.CompSizes[5]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
				Comp5  *T5
				Comp6  *T6
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
				Comp3: (*T3)(c2Ptr),
				Comp4: (*T4)(c3Ptr),
				Comp5: (*T5)(c4Ptr),
				Comp6: (*T6)(c5Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View7 ---------------

type View7[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any] struct {
    *core.View
}

func NewView7[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View7[T1, T2, T3, T4, T5, T6, T7] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view7 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())
    info7 := componentsRegistry.GetOrRegister(reflect.TypeFor[T7]())
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)
    mustAdd(info7)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view7 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, info7, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View7[T1, T2, T3, T4, T5, T6, T7]{View: view}
}

func (v *View7[T1, T2, T3, T4, T5, T6, T7]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
	Comp3 []T3
	Comp4 []T4
	Comp5 []T5
	Comp6 []T6
	Comp7 []T7
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
			Comp3 []T3
			Comp4 []T4
			Comp5 []T5
			Comp6 []T6
			Comp7 []T7
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
						Comp3 []T3
						Comp4 []T4
						Comp5 []T5
						Comp6 []T6
						Comp7 []T7
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3: unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4: unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5: unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6: unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
						Comp7: unsafe.Slice((*T7)(unsafe.Add(base, ma.CompOffsets[6])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View7[T1, T2, T3, T4, T5, T6, T7]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
	Comp6  *T6
	Comp7  *T7
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
		Comp4  *T4
		Comp5  *T5
		Comp6  *T6
		Comp7  *T7
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			c2Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[2]+(row*ma.CompSizes[2]))
			c3Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[3]+(row*ma.CompSizes[3]))
			c4Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[4]+(row*ma.CompSizes[4]))
			c5Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[5]+(row*ma.CompSizes[5]))
			c6Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[6]+(row*ma.CompSizes[6]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
				Comp5  *T5
				Comp6  *T6
				Comp7  *T7
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
				Comp3: (*T3)(c2Ptr),
				Comp4: (*T4)(c3Ptr),
				Comp5: (*T5)(c4Ptr),
				Comp6: (*T6)(c5Ptr),
				Comp7: (*T7)(c6Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View8 ---------------

type View8[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any] struct {
    *core.View
}

func NewView8[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View8[T1, T2, T3, T4, T5, T6, T7, T8] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view8 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())
    info7 := componentsRegistry.GetOrRegister(reflect.TypeFor[T7]())
    info8 := componentsRegistry.GetOrRegister(reflect.TypeFor[T8]())
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)
    mustAdd(info7)
    mustAdd(info8)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view8 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, info7, info8, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View8[T1, T2, T3, T4, T5, T6, T7, T8]{View: view}
}

func (v *View8[T1, T2, T3, T4, T5, T6, T7, T8]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
	Comp3 []T3
	Comp4 []T4
	Comp5 []T5
	Comp6 []T6
	Comp7 []T7
	Comp8 []T8
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
			Comp3 []T3
			Comp4 []T4
			Comp5 []T5
			Comp6 []T6
			Comp7 []T7
			Comp8 []T8
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
						Comp3 []T3
						Comp4 []T4
						Comp5 []T5
						Comp6 []T6
						Comp7 []T7
						Comp8 []T8
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3: unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4: unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5: unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6: unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
						Comp7: unsafe.Slice((*T7)(unsafe.Add(base, ma.CompOffsets[6])), count),
						Comp8: unsafe.Slice((*T8)(unsafe.Add(base, ma.CompOffsets[7])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View8[T1, T2, T3, T4, T5, T6, T7, T8]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
	Comp6  *T6
	Comp7  *T7
	Comp8  *T8
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
		Comp4  *T4
		Comp5  *T5
		Comp6  *T6
		Comp7  *T7
		Comp8  *T8
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			c2Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[2]+(row*ma.CompSizes[2]))
			c3Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[3]+(row*ma.CompSizes[3]))
			c4Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[4]+(row*ma.CompSizes[4]))
			c5Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[5]+(row*ma.CompSizes[5]))
			c6Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[6]+(row*ma.CompSizes[6]))
			c7Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[7]+(row*ma.CompSizes[7]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
				Comp5  *T5
				Comp6  *T6
				Comp7  *T7
				Comp8  *T8
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
				Comp3: (*T3)(c2Ptr),
				Comp4: (*T4)(c3Ptr),
				Comp5: (*T5)(c4Ptr),
				Comp6: (*T6)(c5Ptr),
				Comp7: (*T7)(c6Ptr),
				Comp8: (*T8)(c7Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View9 ---------------

type View9[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any] struct {
    *core.View
}

func NewView9[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view9 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())
    info7 := componentsRegistry.GetOrRegister(reflect.TypeFor[T7]())
    info8 := componentsRegistry.GetOrRegister(reflect.TypeFor[T8]())
    info9 := componentsRegistry.GetOrRegister(reflect.TypeFor[T9]())
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)
    mustAdd(info7)
    mustAdd(info8)
    mustAdd(info9)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view9 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, info7, info8, info9, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]{View: view}
}

func (v *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
	Comp3 []T3
	Comp4 []T4
	Comp5 []T5
	Comp6 []T6
	Comp7 []T7
	Comp8 []T8
	Comp9 []T9
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
			Comp3 []T3
			Comp4 []T4
			Comp5 []T5
			Comp6 []T6
			Comp7 []T7
			Comp8 []T8
			Comp9 []T9
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
						Comp3 []T3
						Comp4 []T4
						Comp5 []T5
						Comp6 []T6
						Comp7 []T7
						Comp8 []T8
						Comp9 []T9
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3: unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4: unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5: unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6: unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
						Comp7: unsafe.Slice((*T7)(unsafe.Add(base, ma.CompOffsets[6])), count),
						Comp8: unsafe.Slice((*T8)(unsafe.Add(base, ma.CompOffsets[7])), count),
						Comp9: unsafe.Slice((*T9)(unsafe.Add(base, ma.CompOffsets[8])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
	Comp6  *T6
	Comp7  *T7
	Comp8  *T8
	Comp9  *T9
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
		Comp4  *T4
		Comp5  *T5
		Comp6  *T6
		Comp7  *T7
		Comp8  *T8
		Comp9  *T9
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			c2Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[2]+(row*ma.CompSizes[2]))
			c3Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[3]+(row*ma.CompSizes[3]))
			c4Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[4]+(row*ma.CompSizes[4]))
			c5Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[5]+(row*ma.CompSizes[5]))
			c6Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[6]+(row*ma.CompSizes[6]))
			c7Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[7]+(row*ma.CompSizes[7]))
			c8Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[8]+(row*ma.CompSizes[8]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
				Comp5  *T5
				Comp6  *T6
				Comp7  *T7
				Comp8  *T8
				Comp9  *T9
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
				Comp3: (*T3)(c2Ptr),
				Comp4: (*T4)(c3Ptr),
				Comp5: (*T5)(c4Ptr),
				Comp6: (*T6)(c5Ptr),
				Comp7: (*T7)(c6Ptr),
				Comp8: (*T8)(c7Ptr),
				Comp9: (*T9)(c8Ptr),
			}) {
				return
			}
		}
	}
}

// --------------- View10 ---------------

type View10[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any] struct {
    *core.View
}

func NewView10[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view10 init failed: %v", err))
        }
    }
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())
    info7 := componentsRegistry.GetOrRegister(reflect.TypeFor[T7]())
    info8 := componentsRegistry.GetOrRegister(reflect.TypeFor[T8]())
    info9 := componentsRegistry.GetOrRegister(reflect.TypeFor[T9]())
    info10 := componentsRegistry.GetOrRegister(reflect.TypeFor[T10]())
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)
    mustAdd(info7)
    mustAdd(info8)
    mustAdd(info9)
    mustAdd(info10)

    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view10 option failed: %v", err))
        }
    }

    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, info7, info8, info9, info10, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{View: view}
}

func (v *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) All() iter.Seq[struct {
	Entity []Entity
	Comp1 []T1
	Comp2 []T2
	Comp3 []T3
	Comp4 []T4
	Comp5 []T5
	Comp6 []T6
	Comp7 []T7
	Comp8 []T8
	Comp9 []T9
	Comp10 []T10
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1 []T1
			Comp2 []T2
			Comp3 []T3
			Comp4 []T4
			Comp5 []T5
			Comp6 []T6
			Comp7 []T7
			Comp8 []T8
			Comp9 []T9
			Comp10 []T10
		},
	) bool) {
		for _, ma := range v.Baked {
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}
				base := page.Ptr
				if !yield(
					struct {
						Entity []Entity
						Comp1 []T1
						Comp2 []T2
						Comp3 []T3
						Comp4 []T4
						Comp5 []T5
						Comp6 []T6
						Comp7 []T7
						Comp8 []T8
						Comp9 []T9
						Comp10 []T10
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1: unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2: unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3: unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4: unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5: unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6: unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
						Comp7: unsafe.Slice((*T7)(unsafe.Add(base, ma.CompOffsets[6])), count),
						Comp8: unsafe.Slice((*T8)(unsafe.Add(base, ma.CompOffsets[7])), count),
						Comp9: unsafe.Slice((*T9)(unsafe.Add(base, ma.CompOffsets[8])), count),
						Comp10: unsafe.Slice((*T10)(unsafe.Add(base, ma.CompOffsets[9])), count),
					}) {
					return
				}
			}
		}
	}
}

func (v *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Filter(selected []Entity) iter.Seq2[int, struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
	Comp6  *T6
	Comp7  *T7
	Comp8  *T8
	Comp9  *T9
	Comp10  *T10
}] {
	return func(yield func(int, struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
		Comp4  *T4
		Comp5  *T5
		Comp6  *T6
		Comp7  *T7
		Comp8  *T8
		Comp9  *T9
		Comp10  *T10
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for i, e := range selected {
			link, ok := store.Get(e)
			if !ok {
				continue
			}
			if link.ArchId != lastArchID {
				ma = v.GetMatchedArch(link.ArchId)
				lastArchID = link.ArchId
			}
			if ma == nil {
				continue
			}
			physPage := ma.Arch.Memory.Pages[link.PageIdx]
			row := uintptr(link.PageRow)
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			c2Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[2]+(row*ma.CompSizes[2]))
			c3Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[3]+(row*ma.CompSizes[3]))
			c4Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[4]+(row*ma.CompSizes[4]))
			c5Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[5]+(row*ma.CompSizes[5]))
			c6Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[6]+(row*ma.CompSizes[6]))
			c7Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[7]+(row*ma.CompSizes[7]))
			c8Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[8]+(row*ma.CompSizes[8]))
			c9Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[9]+(row*ma.CompSizes[9]))
			if !yield(i, struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
				Comp5  *T5
				Comp6  *T6
				Comp7  *T7
				Comp8  *T8
				Comp9  *T9
				Comp10  *T10
			}{
				Entity: e,
				Comp1: (*T1)(c0Ptr),
				Comp2: (*T2)(c1Ptr),
				Comp3: (*T3)(c2Ptr),
				Comp4: (*T4)(c3Ptr),
				Comp5: (*T5)(c4Ptr),
				Comp6: (*T6)(c5Ptr),
				Comp7: (*T7)(c6Ptr),
				Comp8: (*T8)(c7Ptr),
				Comp9: (*T9)(c8Ptr),
				Comp10: (*T10)(c9Ptr),
			}) {
				return
			}
		}
	}
}

