package goke

import (
	"fmt"
	"iter"
	"reflect"
	"unsafe"

    "github.com/kjkrol/goke/internal/core"
)

// --------------- View1 ---------------

// View1 provides a type-safe iterator and access layer for entities that
// possess exactly 1 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View1 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View1[T1 any] struct {
    *core.View
}

// NewView1 initializes a query for exactly 1 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView1[T1 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View1[T1] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view1 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view1 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View1[T1]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view1.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item1
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view1.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//	}
func (v *View1[T1]) Filter(selected []Entity) iter.Seq[Item1[T1]] {
	return func(yield func(Item1[T1]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item1[T1]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View2 ---------------

// View2 provides a type-safe iterator and access layer for entities that
// possess exactly 2 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View2 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View2[T1 any, T2 any] struct {
    *core.View
}

// NewView2 initializes a query for exactly 2 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView2[T1 any, T2 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View2[T1, T2] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view2 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)
    mustAdd(info2)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view2 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View2[T1, T2]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view2.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item2
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view2.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//	}
func (v *View2[T1, T2]) Filter(selected []Entity) iter.Seq[Item2[T1, T2]] {
	return func(yield func(Item2[T1, T2]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item2[T1, T2]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View3 ---------------

// View3 provides a type-safe iterator and access layer for entities that
// possess exactly 3 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View3 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View3[T1 any, T2 any, T3 any] struct {
    *core.View
}

// NewView3 initializes a query for exactly 3 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView3[T1 any, T2 any, T3 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View3[T1, T2, T3] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view3 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view3 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, info3, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View3[T1, T2, T3]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view3.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item3
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view3.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//		comp3 := item.Comp3
//	}
func (v *View3[T1, T2, T3]) Filter(selected []Entity) iter.Seq[Item3[T1, T2, T3]] {
	return func(yield func(Item3[T1, T2, T3]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column
		var col2 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)
				col2 = currentArch.GetColumn(v.Layout[2].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item3[T1, T2, T3]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
				Comp3: (*T3)(col2.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View4 ---------------

// View4 provides a type-safe iterator and access layer for entities that
// possess exactly 4 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View4 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View4[T1 any, T2 any, T3 any, T4 any] struct {
    *core.View
}

// NewView4 initializes a query for exactly 4 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView4[T1 any, T2 any, T3 any, T4 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View4[T1, T2, T3, T4] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view4 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view4 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, info3, info4, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View4[T1, T2, T3, T4]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view4.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item4
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view4.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//		comp3 := item.Comp3
//		comp4 := item.Comp4
//	}
func (v *View4[T1, T2, T3, T4]) Filter(selected []Entity) iter.Seq[Item4[T1, T2, T3, T4]] {
	return func(yield func(Item4[T1, T2, T3, T4]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column
		var col2 *core.Column
		var col3 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)
				col2 = currentArch.GetColumn(v.Layout[2].ID)
				col3 = currentArch.GetColumn(v.Layout[3].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item4[T1, T2, T3, T4]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
				Comp3: (*T3)(col2.GetPointer(chunk, link.PageRow)),
				Comp4: (*T4)(col3.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View5 ---------------

// View5 provides a type-safe iterator and access layer for entities that
// possess exactly 5 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View5 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View5[T1 any, T2 any, T3 any, T4 any, T5 any] struct {
    *core.View
}

// NewView5 initializes a query for exactly 5 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView5[T1 any, T2 any, T3 any, T4 any, T5 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View5[T1, T2, T3, T4, T5] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view5 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view5 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View5[T1, T2, T3, T4, T5]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view5.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item5
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view5.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//		comp3 := item.Comp3
//		comp4 := item.Comp4
//		comp5 := item.Comp5
//	}
func (v *View5[T1, T2, T3, T4, T5]) Filter(selected []Entity) iter.Seq[Item5[T1, T2, T3, T4, T5]] {
	return func(yield func(Item5[T1, T2, T3, T4, T5]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column
		var col2 *core.Column
		var col3 *core.Column
		var col4 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)
				col2 = currentArch.GetColumn(v.Layout[2].ID)
				col3 = currentArch.GetColumn(v.Layout[3].ID)
				col4 = currentArch.GetColumn(v.Layout[4].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item5[T1, T2, T3, T4, T5]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
				Comp3: (*T3)(col2.GetPointer(chunk, link.PageRow)),
				Comp4: (*T4)(col3.GetPointer(chunk, link.PageRow)),
				Comp5: (*T5)(col4.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View6 ---------------

// View6 provides a type-safe iterator and access layer for entities that
// possess exactly 6 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View6 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any] struct {
    *core.View
}

// NewView6 initializes a query for exactly 6 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View6[T1, T2, T3, T4, T5, T6] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view6 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view6 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View6[T1, T2, T3, T4, T5, T6]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view6.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item6
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view6.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//		comp3 := item.Comp3
//		comp4 := item.Comp4
//		comp5 := item.Comp5
//		comp6 := item.Comp6
//	}
func (v *View6[T1, T2, T3, T4, T5, T6]) Filter(selected []Entity) iter.Seq[Item6[T1, T2, T3, T4, T5, T6]] {
	return func(yield func(Item6[T1, T2, T3, T4, T5, T6]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column
		var col2 *core.Column
		var col3 *core.Column
		var col4 *core.Column
		var col5 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)
				col2 = currentArch.GetColumn(v.Layout[2].ID)
				col3 = currentArch.GetColumn(v.Layout[3].ID)
				col4 = currentArch.GetColumn(v.Layout[4].ID)
				col5 = currentArch.GetColumn(v.Layout[5].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item6[T1, T2, T3, T4, T5, T6]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
				Comp3: (*T3)(col2.GetPointer(chunk, link.PageRow)),
				Comp4: (*T4)(col3.GetPointer(chunk, link.PageRow)),
				Comp5: (*T5)(col4.GetPointer(chunk, link.PageRow)),
				Comp6: (*T6)(col5.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View7 ---------------

// View7 provides a type-safe iterator and access layer for entities that
// possess exactly 7 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View7 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View7[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any] struct {
    *core.View
}

// NewView7 initializes a query for exactly 7 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView7[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View7[T1, T2, T3, T4, T5, T6, T7] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view7 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())
    info7 := componentsRegistry.GetOrRegister(reflect.TypeFor[T7]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)
    mustAdd(info7)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view7 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, info7, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View7[T1, T2, T3, T4, T5, T6, T7]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view7.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			c7 := &page.Comp7[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item7
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view7.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//		comp3 := item.Comp3
//		comp4 := item.Comp4
//		comp5 := item.Comp5
//		comp6 := item.Comp6
//		comp7 := item.Comp7
//	}
func (v *View7[T1, T2, T3, T4, T5, T6, T7]) Filter(selected []Entity) iter.Seq[Item7[T1, T2, T3, T4, T5, T6, T7]] {
	return func(yield func(Item7[T1, T2, T3, T4, T5, T6, T7]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column
		var col2 *core.Column
		var col3 *core.Column
		var col4 *core.Column
		var col5 *core.Column
		var col6 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)
				col2 = currentArch.GetColumn(v.Layout[2].ID)
				col3 = currentArch.GetColumn(v.Layout[3].ID)
				col4 = currentArch.GetColumn(v.Layout[4].ID)
				col5 = currentArch.GetColumn(v.Layout[5].ID)
				col6 = currentArch.GetColumn(v.Layout[6].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item7[T1, T2, T3, T4, T5, T6, T7]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
				Comp3: (*T3)(col2.GetPointer(chunk, link.PageRow)),
				Comp4: (*T4)(col3.GetPointer(chunk, link.PageRow)),
				Comp5: (*T5)(col4.GetPointer(chunk, link.PageRow)),
				Comp6: (*T6)(col5.GetPointer(chunk, link.PageRow)),
				Comp7: (*T7)(col6.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View8 ---------------

// View8 provides a type-safe iterator and access layer for entities that
// possess exactly 8 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View8 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View8[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any] struct {
    *core.View
}

// NewView8 initializes a query for exactly 8 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView8[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View8[T1, T2, T3, T4, T5, T6, T7, T8] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view8 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())
    info7 := componentsRegistry.GetOrRegister(reflect.TypeFor[T7]())
    info8 := componentsRegistry.GetOrRegister(reflect.TypeFor[T8]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)
    mustAdd(info7)
    mustAdd(info8)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view8 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, info7, info8, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View8[T1, T2, T3, T4, T5, T6, T7, T8]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view8.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			c7 := &page.Comp7[i]
//			c8 := &page.Comp8[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item8
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view8.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//		comp3 := item.Comp3
//		comp4 := item.Comp4
//		comp5 := item.Comp5
//		comp6 := item.Comp6
//		comp7 := item.Comp7
//		comp8 := item.Comp8
//	}
func (v *View8[T1, T2, T3, T4, T5, T6, T7, T8]) Filter(selected []Entity) iter.Seq[Item8[T1, T2, T3, T4, T5, T6, T7, T8]] {
	return func(yield func(Item8[T1, T2, T3, T4, T5, T6, T7, T8]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column
		var col2 *core.Column
		var col3 *core.Column
		var col4 *core.Column
		var col5 *core.Column
		var col6 *core.Column
		var col7 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)
				col2 = currentArch.GetColumn(v.Layout[2].ID)
				col3 = currentArch.GetColumn(v.Layout[3].ID)
				col4 = currentArch.GetColumn(v.Layout[4].ID)
				col5 = currentArch.GetColumn(v.Layout[5].ID)
				col6 = currentArch.GetColumn(v.Layout[6].ID)
				col7 = currentArch.GetColumn(v.Layout[7].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item8[T1, T2, T3, T4, T5, T6, T7, T8]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
				Comp3: (*T3)(col2.GetPointer(chunk, link.PageRow)),
				Comp4: (*T4)(col3.GetPointer(chunk, link.PageRow)),
				Comp5: (*T5)(col4.GetPointer(chunk, link.PageRow)),
				Comp6: (*T6)(col5.GetPointer(chunk, link.PageRow)),
				Comp7: (*T7)(col6.GetPointer(chunk, link.PageRow)),
				Comp8: (*T8)(col7.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View9 ---------------

// View9 provides a type-safe iterator and access layer for entities that
// possess exactly 9 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View9 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View9[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any] struct {
    *core.View
}

// NewView9 initializes a query for exactly 9 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView9[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view9 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
    info1 := componentsRegistry.GetOrRegister(reflect.TypeFor[T1]())
    info2 := componentsRegistry.GetOrRegister(reflect.TypeFor[T2]())
    info3 := componentsRegistry.GetOrRegister(reflect.TypeFor[T3]())
    info4 := componentsRegistry.GetOrRegister(reflect.TypeFor[T4]())
    info5 := componentsRegistry.GetOrRegister(reflect.TypeFor[T5]())
    info6 := componentsRegistry.GetOrRegister(reflect.TypeFor[T6]())
    info7 := componentsRegistry.GetOrRegister(reflect.TypeFor[T7]())
    info8 := componentsRegistry.GetOrRegister(reflect.TypeFor[T8]())
    info9 := componentsRegistry.GetOrRegister(reflect.TypeFor[T9]())

    // 2. Add to Blueprint (Build Mask)
    mustAdd(info1)
    mustAdd(info2)
    mustAdd(info3)
    mustAdd(info4)
    mustAdd(info5)
    mustAdd(info6)
    mustAdd(info7)
    mustAdd(info8)
    mustAdd(info9)

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view9 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, info7, info8, info9, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view9.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			c7 := &page.Comp7[i]
//			c8 := &page.Comp8[i]
//			c9 := &page.Comp9[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item9
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view9.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//		comp3 := item.Comp3
//		comp4 := item.Comp4
//		comp5 := item.Comp5
//		comp6 := item.Comp6
//		comp7 := item.Comp7
//		comp8 := item.Comp8
//		comp9 := item.Comp9
//	}
func (v *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Filter(selected []Entity) iter.Seq[Item9[T1, T2, T3, T4, T5, T6, T7, T8, T9]] {
	return func(yield func(Item9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column
		var col2 *core.Column
		var col3 *core.Column
		var col4 *core.Column
		var col5 *core.Column
		var col6 *core.Column
		var col7 *core.Column
		var col8 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)
				col2 = currentArch.GetColumn(v.Layout[2].ID)
				col3 = currentArch.GetColumn(v.Layout[3].ID)
				col4 = currentArch.GetColumn(v.Layout[4].ID)
				col5 = currentArch.GetColumn(v.Layout[5].ID)
				col6 = currentArch.GetColumn(v.Layout[6].ID)
				col7 = currentArch.GetColumn(v.Layout[7].ID)
				col8 = currentArch.GetColumn(v.Layout[8].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item9[T1, T2, T3, T4, T5, T6, T7, T8, T9]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
				Comp3: (*T3)(col2.GetPointer(chunk, link.PageRow)),
				Comp4: (*T4)(col3.GetPointer(chunk, link.PageRow)),
				Comp5: (*T5)(col4.GetPointer(chunk, link.PageRow)),
				Comp6: (*T6)(col5.GetPointer(chunk, link.PageRow)),
				Comp7: (*T7)(col6.GetPointer(chunk, link.PageRow)),
				Comp8: (*T8)(col7.GetPointer(chunk, link.PageRow)),
				Comp9: (*T9)(col8.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}
// --------------- View10 ---------------

// View10 provides a type-safe iterator and access layer for entities that
// possess exactly 10 specific stateful components. It acts as a specialized
// window into the ECS world, filtering archetypes that satisfy the required
// component mask and any additional constraints defined via BlueprintOptions.
//
// By leveraging pre-calculated component offsets, View10 enables
// O(1) access to component data during iteration, making it the primary
// tool for implementing high-performance systems and logic loops.
type View10[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any] struct {
    *core.View
}

// NewView10 initializes a query for exactly 10 components.
// It panics if the component registration fails, if there are duplicate
// components, or if options (like Exclude) create a logical contradiction.
//
// This ensures that the View is valid and ready for high-performance
// iteration immediately after creation.
func NewView10[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any](
    ecs *ECS,
    opts ...BlueprintOption,
) *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10] {
    registry := ecs.registry
    blueprint := core.NewBlueprint(registry)
    componentsRegistry := &registry.ComponentsRegistry

    // Helper: Validates that the required component can be part of the view.
    mustAdd := func(info core.ComponentInfo) {
        if err := blueprint.WithComp(info); err != nil {
            panic(fmt.Sprintf("goke: view10 init failed: %v", err))
        }
    }

    // 1. Resolve Component Infos (Type -> ID)
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

    // 2. Add to Blueprint (Build Mask)
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

    // 3. Apply dynamic options (Include/Exclude)
    for _, opt := range opts {
        if err := opt(blueprint); err != nil {
            panic(fmt.Sprintf("goke: view10 option failed: %v", err))
        }
    }

    // 4. Define Rigid Layout (Slice Literal - Zero Allocation Overhead)
    // This guarantees that T1 is at index 0, T2 at index 1, etc.
    layout := []core.ComponentInfo{
        info1, info2, info3, info4, info5, info6, info7, info8, info9, info10, 
    }

    view := core.NewView(blueprint, layout, registry)
    return &View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{View: view}
}

// All returns an iterator (iter.Seq) that yields physical memory pages as batches of slices.
// Each yielded struct represents a contiguous block of memory for the matched entities
// and their corresponding components, mapped directly to Go slices (Zero Heap Allocation) to
// preserves the Structure of Arrays (SoA) layout.
// By shifting the inner loop to the caller side, it guarantees optimal CPU cache coherence
// and allows the Go compiler to easily apply advanced optimizations such as loop unrolling,
// bounds-check elimination, and SIMD vectorization.
//
// The iteration is performed archetype by archetype, and yields page by page.
//
// Example usage:
//	for page := range view10.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			c7 := &page.Comp7[i]
//			c8 := &page.Comp8[i]
//			c9 := &page.Comp9[i]
//			c10 := &page.Comp10[i]
//			// Apply domain logic here...
//		}
//	}
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

// Filter iterates over a provided slice of entities, yielding an Item10
// for each entity that satisfies the View's component constraints.
//
// This is highly optimized for targeted queries where the entity subset
// is already known (e.g., from spatial partitioning, sorted lists, or event
// payloads), allowing direct, high-speed access to their component data.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	selected := []Entity{e1, e5, e10}
//	for item := range view10.Filter(selected) {
//		entity := item.Entity
//		comp1 := item.Comp1
//		comp2 := item.Comp2
//		comp3 := item.Comp3
//		comp4 := item.Comp4
//		comp5 := item.Comp5
//		comp6 := item.Comp6
//		comp7 := item.Comp7
//		comp8 := item.Comp8
//		comp9 := item.Comp9
//		comp10 := item.Comp10
//	}
func (v *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Filter(selected []Entity) iter.Seq[Item10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]] {
	return func(yield func(Item10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) bool) {
		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var currentArch *core.Archetype

		// Column descriptor cache
		var col0 *core.Column
		var col1 *core.Column
		var col2 *core.Column
		var col3 *core.Column
		var col4 *core.Column
		var col5 *core.Column
		var col6 *core.Column
		var col7 *core.Column
		var col8 *core.Column
		var col9 *core.Column

		registry := v.Reg.ArchetypeRegistry

		for _, entity := range selected {
			link, ok := registry.EntityLinkStore.Get(entity)
			if !ok {
				continue
			}

			// 1. Archetype Change Detection (Cache descriptors)
			if link.ArchId != lastArchID {
				currentArch = &registry.Archetypes[link.ArchId]

				if !v.View.Matches(currentArch.Mask) {
					lastArchID = core.NullArchetypeId
					currentArch = nil
					continue
				}

				// Cache all column descriptors for this archetype
				col0 = currentArch.GetColumn(v.Layout[0].ID)
				col1 = currentArch.GetColumn(v.Layout[1].ID)
				col2 = currentArch.GetColumn(v.Layout[2].ID)
				col3 = currentArch.GetColumn(v.Layout[3].ID)
				col4 = currentArch.GetColumn(v.Layout[4].ID)
				col5 = currentArch.GetColumn(v.Layout[5].ID)
				col6 = currentArch.GetColumn(v.Layout[6].ID)
				col7 = currentArch.GetColumn(v.Layout[7].ID)
				col8 = currentArch.GetColumn(v.Layout[8].ID)
				col9 = currentArch.GetColumn(v.Layout[9].ID)

				lastArchID = link.ArchId
			}

			if currentArch == nil {
				continue
			}

			// 2. Resolve Page
			// Access the physical page using the index from the link
			chunk := currentArch.Memory.Pages[link.PageIdx]

			// 3. Construct Result
			if !yield(Item10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{
				Entity: entity,
				Comp1: (*T1)(col0.GetPointer(chunk, link.PageRow)),
				Comp2: (*T2)(col1.GetPointer(chunk, link.PageRow)),
				Comp3: (*T3)(col2.GetPointer(chunk, link.PageRow)),
				Comp4: (*T4)(col3.GetPointer(chunk, link.PageRow)),
				Comp5: (*T5)(col4.GetPointer(chunk, link.PageRow)),
				Comp6: (*T6)(col5.GetPointer(chunk, link.PageRow)),
				Comp7: (*T7)(col6.GetPointer(chunk, link.PageRow)),
				Comp8: (*T8)(col7.GetPointer(chunk, link.PageRow)),
				Comp9: (*T9)(col8.GetPointer(chunk, link.PageRow)),
				Comp10: (*T10)(col9.GetPointer(chunk, link.PageRow)),
			}) {
				return
			}
		}
	}
}

