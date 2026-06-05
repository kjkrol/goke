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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view1.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        
//    }
func (v *View1[T1]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		
	},
	struct {
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			
		},
		struct {
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
					}

					// Remaining components spill over to Tail
					tail := struct { 
					}{ 
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view1.Each(func(entities []core.Entity, c1s []T1) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             
//         }
//     })
func (v *View1[T1]) Each(fn func([]core.Entity, []T1)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view2.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        
//    }
func (v *View2[T1, T2]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		
	},
	struct {
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			
		},
		struct {
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
					}

					// Remaining components spill over to Tail
					tail := struct { 
					}{ 
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view2.Each(func(entities []core.Entity, c1s []T1, c2s []T2) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             
//         }
//     })
func (v *View2[T1, T2]) Each(fn func([]core.Entity, []T1, []T2)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view3.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        v3 := head.Comp3
//        
//    }
func (v *View3[T1, T2, T3]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		Comp3 *T3
		
	},
	struct {
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			Comp3 *T3
			
		},
		struct {
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))
		stride2 := unsafe.Sizeof(*new(T3))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]
			offset2 := ma.FieldsOffsets[2]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)
				ptr2 := unsafe.Add(base, offset2)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
						Comp3 *T3
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
						Comp3: (*T3)(ptr2),
					}

					// Remaining components spill over to Tail
					tail := struct { 
					}{ 
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)
					ptr2 = unsafe.Add(ptr2, stride2)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view3.Each(func(entities []core.Entity, c1s []T1, c2s []T2, c3s []T3) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             v3 := c3s[i]
//             
//         }
//     })
func (v *View3[T1, T2, T3]) Each(fn func([]core.Entity, []T1, []T2, []T3)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)
			c3 := unsafe.Slice((*T3)(unsafe.Add(base, ma.FieldsOffsets[2])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2, c3)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view4.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        v3 := head.Comp3
//        v4 := head.Comp4
//        
//    }
func (v *View4[T1, T2, T3, T4]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		Comp3 *T3
		
	},
	struct {
		Comp4 *T4
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			Comp3 *T3
			
		},
		struct {
			Comp4 *T4
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))
		stride2 := unsafe.Sizeof(*new(T3))
		stride3 := unsafe.Sizeof(*new(T4))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]
			offset2 := ma.FieldsOffsets[2]
			offset3 := ma.FieldsOffsets[3]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)
				ptr2 := unsafe.Add(base, offset2)
				ptr3 := unsafe.Add(base, offset3)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
						Comp3 *T3
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
						Comp3: (*T3)(ptr2),
					}

					// Remaining components spill over to Tail
					tail := struct { 
						Comp4 *T4
					}{ 
					    Comp4: (*T4)(ptr3),
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)
					ptr2 = unsafe.Add(ptr2, stride2)
					ptr3 = unsafe.Add(ptr3, stride3)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view4.Each(func(entities []core.Entity, c1s []T1, c2s []T2, c3s []T3, c4s []T4) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             v3 := c3s[i]
//             v4 := c4s[i]
//             
//         }
//     })
func (v *View4[T1, T2, T3, T4]) Each(fn func([]core.Entity, []T1, []T2, []T3, []T4)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)
			c3 := unsafe.Slice((*T3)(unsafe.Add(base, ma.FieldsOffsets[2])), count)
			c4 := unsafe.Slice((*T4)(unsafe.Add(base, ma.FieldsOffsets[3])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2, c3, c4)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view5.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        v3 := head.Comp3
//        v4 := head.Comp4
//        v5 := tail.V5
//        
//    }
func (v *View5[T1, T2, T3, T4, T5]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		Comp3 *T3
		
	},
	struct {
		Comp4 *T4
		Comp5 *T5
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			Comp3 *T3
			
		},
		struct {
			Comp4 *T4
			Comp5 *T5
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))
		stride2 := unsafe.Sizeof(*new(T3))
		stride3 := unsafe.Sizeof(*new(T4))
		stride4 := unsafe.Sizeof(*new(T5))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]
			offset2 := ma.FieldsOffsets[2]
			offset3 := ma.FieldsOffsets[3]
			offset4 := ma.FieldsOffsets[4]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)
				ptr2 := unsafe.Add(base, offset2)
				ptr3 := unsafe.Add(base, offset3)
				ptr4 := unsafe.Add(base, offset4)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
						Comp3 *T3
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
						Comp3: (*T3)(ptr2),
					}

					// Remaining components spill over to Tail
					tail := struct { 
						Comp4 *T4
						Comp5 *T5
					}{ 
					    Comp4: (*T4)(ptr3),
					    Comp5: (*T5)(ptr4),
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)
					ptr2 = unsafe.Add(ptr2, stride2)
					ptr3 = unsafe.Add(ptr3, stride3)
					ptr4 = unsafe.Add(ptr4, stride4)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view5.Each(func(entities []core.Entity, c1s []T1, c2s []T2, c3s []T3, c4s []T4, c5s []T5) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             v3 := c3s[i]
//             v4 := c4s[i]
//             v5 := c5s[i]
//             
//         }
//     })
func (v *View5[T1, T2, T3, T4, T5]) Each(fn func([]core.Entity, []T1, []T2, []T3, []T4, []T5)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)
			c3 := unsafe.Slice((*T3)(unsafe.Add(base, ma.FieldsOffsets[2])), count)
			c4 := unsafe.Slice((*T4)(unsafe.Add(base, ma.FieldsOffsets[3])), count)
			c5 := unsafe.Slice((*T5)(unsafe.Add(base, ma.FieldsOffsets[4])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2, c3, c4, c5)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view6.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        v3 := head.Comp3
//        v4 := head.Comp4
//        v5 := tail.V5
//        v6 := tail.V6
//        
//    }
func (v *View6[T1, T2, T3, T4, T5, T6]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		Comp3 *T3
		
	},
	struct {
		Comp4 *T4
		Comp5 *T5
		Comp6 *T6
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			Comp3 *T3
			
		},
		struct {
			Comp4 *T4
			Comp5 *T5
			Comp6 *T6
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))
		stride2 := unsafe.Sizeof(*new(T3))
		stride3 := unsafe.Sizeof(*new(T4))
		stride4 := unsafe.Sizeof(*new(T5))
		stride5 := unsafe.Sizeof(*new(T6))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]
			offset2 := ma.FieldsOffsets[2]
			offset3 := ma.FieldsOffsets[3]
			offset4 := ma.FieldsOffsets[4]
			offset5 := ma.FieldsOffsets[5]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)
				ptr2 := unsafe.Add(base, offset2)
				ptr3 := unsafe.Add(base, offset3)
				ptr4 := unsafe.Add(base, offset4)
				ptr5 := unsafe.Add(base, offset5)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
						Comp3 *T3
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
						Comp3: (*T3)(ptr2),
					}

					// Remaining components spill over to Tail
					tail := struct { 
						Comp4 *T4
						Comp5 *T5
						Comp6 *T6
					}{ 
					    Comp4: (*T4)(ptr3),
					    Comp5: (*T5)(ptr4),
					    Comp6: (*T6)(ptr5),
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)
					ptr2 = unsafe.Add(ptr2, stride2)
					ptr3 = unsafe.Add(ptr3, stride3)
					ptr4 = unsafe.Add(ptr4, stride4)
					ptr5 = unsafe.Add(ptr5, stride5)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view6.Each(func(entities []core.Entity, c1s []T1, c2s []T2, c3s []T3, c4s []T4, c5s []T5, c6s []T6) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             v3 := c3s[i]
//             v4 := c4s[i]
//             v5 := c5s[i]
//             v6 := c6s[i]
//             
//         }
//     })
func (v *View6[T1, T2, T3, T4, T5, T6]) Each(fn func([]core.Entity, []T1, []T2, []T3, []T4, []T5, []T6)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)
			c3 := unsafe.Slice((*T3)(unsafe.Add(base, ma.FieldsOffsets[2])), count)
			c4 := unsafe.Slice((*T4)(unsafe.Add(base, ma.FieldsOffsets[3])), count)
			c5 := unsafe.Slice((*T5)(unsafe.Add(base, ma.FieldsOffsets[4])), count)
			c6 := unsafe.Slice((*T6)(unsafe.Add(base, ma.FieldsOffsets[5])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2, c3, c4, c5, c6)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view7.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        v3 := head.Comp3
//        v4 := head.Comp4
//        v5 := tail.V5
//        v6 := tail.V6
//        v7 := tail.V7
//        
//    }
func (v *View7[T1, T2, T3, T4, T5, T6, T7]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		Comp3 *T3
		
	},
	struct {
		Comp4 *T4
		Comp5 *T5
		Comp6 *T6
		Comp7 *T7
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			Comp3 *T3
			
		},
		struct {
			Comp4 *T4
			Comp5 *T5
			Comp6 *T6
			Comp7 *T7
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))
		stride2 := unsafe.Sizeof(*new(T3))
		stride3 := unsafe.Sizeof(*new(T4))
		stride4 := unsafe.Sizeof(*new(T5))
		stride5 := unsafe.Sizeof(*new(T6))
		stride6 := unsafe.Sizeof(*new(T7))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]
			offset2 := ma.FieldsOffsets[2]
			offset3 := ma.FieldsOffsets[3]
			offset4 := ma.FieldsOffsets[4]
			offset5 := ma.FieldsOffsets[5]
			offset6 := ma.FieldsOffsets[6]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)
				ptr2 := unsafe.Add(base, offset2)
				ptr3 := unsafe.Add(base, offset3)
				ptr4 := unsafe.Add(base, offset4)
				ptr5 := unsafe.Add(base, offset5)
				ptr6 := unsafe.Add(base, offset6)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
						Comp3 *T3
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
						Comp3: (*T3)(ptr2),
					}

					// Remaining components spill over to Tail
					tail := struct { 
						Comp4 *T4
						Comp5 *T5
						Comp6 *T6
						Comp7 *T7
					}{ 
					    Comp4: (*T4)(ptr3),
					    Comp5: (*T5)(ptr4),
					    Comp6: (*T6)(ptr5),
					    Comp7: (*T7)(ptr6),
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)
					ptr2 = unsafe.Add(ptr2, stride2)
					ptr3 = unsafe.Add(ptr3, stride3)
					ptr4 = unsafe.Add(ptr4, stride4)
					ptr5 = unsafe.Add(ptr5, stride5)
					ptr6 = unsafe.Add(ptr6, stride6)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view7.Each(func(entities []core.Entity, c1s []T1, c2s []T2, c3s []T3, c4s []T4, c5s []T5, c6s []T6, c7s []T7) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             v3 := c3s[i]
//             v4 := c4s[i]
//             v5 := c5s[i]
//             v6 := c6s[i]
//             v7 := c7s[i]
//             
//         }
//     })
func (v *View7[T1, T2, T3, T4, T5, T6, T7]) Each(fn func([]core.Entity, []T1, []T2, []T3, []T4, []T5, []T6, []T7)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)
			c3 := unsafe.Slice((*T3)(unsafe.Add(base, ma.FieldsOffsets[2])), count)
			c4 := unsafe.Slice((*T4)(unsafe.Add(base, ma.FieldsOffsets[3])), count)
			c5 := unsafe.Slice((*T5)(unsafe.Add(base, ma.FieldsOffsets[4])), count)
			c6 := unsafe.Slice((*T6)(unsafe.Add(base, ma.FieldsOffsets[5])), count)
			c7 := unsafe.Slice((*T7)(unsafe.Add(base, ma.FieldsOffsets[6])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2, c3, c4, c5, c6, c7)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view8.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        v3 := head.Comp3
//        v4 := head.Comp4
//        v5 := tail.V5
//        v6 := tail.V6
//        v7 := tail.V7
//        v8 := tail.V8
//        
//    }
func (v *View8[T1, T2, T3, T4, T5, T6, T7, T8]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		Comp3 *T3
		
	},
	struct {
		Comp4 *T4
		Comp5 *T5
		Comp6 *T6
		Comp7 *T7
		Comp8 *T8
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			Comp3 *T3
			
		},
		struct {
			Comp4 *T4
			Comp5 *T5
			Comp6 *T6
			Comp7 *T7
			Comp8 *T8
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))
		stride2 := unsafe.Sizeof(*new(T3))
		stride3 := unsafe.Sizeof(*new(T4))
		stride4 := unsafe.Sizeof(*new(T5))
		stride5 := unsafe.Sizeof(*new(T6))
		stride6 := unsafe.Sizeof(*new(T7))
		stride7 := unsafe.Sizeof(*new(T8))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]
			offset2 := ma.FieldsOffsets[2]
			offset3 := ma.FieldsOffsets[3]
			offset4 := ma.FieldsOffsets[4]
			offset5 := ma.FieldsOffsets[5]
			offset6 := ma.FieldsOffsets[6]
			offset7 := ma.FieldsOffsets[7]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)
				ptr2 := unsafe.Add(base, offset2)
				ptr3 := unsafe.Add(base, offset3)
				ptr4 := unsafe.Add(base, offset4)
				ptr5 := unsafe.Add(base, offset5)
				ptr6 := unsafe.Add(base, offset6)
				ptr7 := unsafe.Add(base, offset7)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
						Comp3 *T3
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
						Comp3: (*T3)(ptr2),
					}

					// Remaining components spill over to Tail
					tail := struct { 
						Comp4 *T4
						Comp5 *T5
						Comp6 *T6
						Comp7 *T7
						Comp8 *T8
					}{ 
					    Comp4: (*T4)(ptr3),
					    Comp5: (*T5)(ptr4),
					    Comp6: (*T6)(ptr5),
					    Comp7: (*T7)(ptr6),
					    Comp8: (*T8)(ptr7),
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)
					ptr2 = unsafe.Add(ptr2, stride2)
					ptr3 = unsafe.Add(ptr3, stride3)
					ptr4 = unsafe.Add(ptr4, stride4)
					ptr5 = unsafe.Add(ptr5, stride5)
					ptr6 = unsafe.Add(ptr6, stride6)
					ptr7 = unsafe.Add(ptr7, stride7)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view8.Each(func(entities []core.Entity, c1s []T1, c2s []T2, c3s []T3, c4s []T4, c5s []T5, c6s []T6, c7s []T7, c8s []T8) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             v3 := c3s[i]
//             v4 := c4s[i]
//             v5 := c5s[i]
//             v6 := c6s[i]
//             v7 := c7s[i]
//             v8 := c8s[i]
//             
//         }
//     })
func (v *View8[T1, T2, T3, T4, T5, T6, T7, T8]) Each(fn func([]core.Entity, []T1, []T2, []T3, []T4, []T5, []T6, []T7, []T8)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)
			c3 := unsafe.Slice((*T3)(unsafe.Add(base, ma.FieldsOffsets[2])), count)
			c4 := unsafe.Slice((*T4)(unsafe.Add(base, ma.FieldsOffsets[3])), count)
			c5 := unsafe.Slice((*T5)(unsafe.Add(base, ma.FieldsOffsets[4])), count)
			c6 := unsafe.Slice((*T6)(unsafe.Add(base, ma.FieldsOffsets[5])), count)
			c7 := unsafe.Slice((*T7)(unsafe.Add(base, ma.FieldsOffsets[6])), count)
			c8 := unsafe.Slice((*T8)(unsafe.Add(base, ma.FieldsOffsets[7])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2, c3, c4, c5, c6, c7, c8)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view9.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        v3 := head.Comp3
//        v4 := head.Comp4
//        v5 := tail.V5
//        v6 := tail.V6
//        v7 := tail.V7
//        v8 := tail.V8
//        v9 := tail.V9
//        
//    }
func (v *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		Comp3 *T3
		
	},
	struct {
		Comp4 *T4
		Comp5 *T5
		Comp6 *T6
		Comp7 *T7
		Comp8 *T8
		Comp9 *T9
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			Comp3 *T3
			
		},
		struct {
			Comp4 *T4
			Comp5 *T5
			Comp6 *T6
			Comp7 *T7
			Comp8 *T8
			Comp9 *T9
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))
		stride2 := unsafe.Sizeof(*new(T3))
		stride3 := unsafe.Sizeof(*new(T4))
		stride4 := unsafe.Sizeof(*new(T5))
		stride5 := unsafe.Sizeof(*new(T6))
		stride6 := unsafe.Sizeof(*new(T7))
		stride7 := unsafe.Sizeof(*new(T8))
		stride8 := unsafe.Sizeof(*new(T9))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]
			offset2 := ma.FieldsOffsets[2]
			offset3 := ma.FieldsOffsets[3]
			offset4 := ma.FieldsOffsets[4]
			offset5 := ma.FieldsOffsets[5]
			offset6 := ma.FieldsOffsets[6]
			offset7 := ma.FieldsOffsets[7]
			offset8 := ma.FieldsOffsets[8]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)
				ptr2 := unsafe.Add(base, offset2)
				ptr3 := unsafe.Add(base, offset3)
				ptr4 := unsafe.Add(base, offset4)
				ptr5 := unsafe.Add(base, offset5)
				ptr6 := unsafe.Add(base, offset6)
				ptr7 := unsafe.Add(base, offset7)
				ptr8 := unsafe.Add(base, offset8)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
						Comp3 *T3
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
						Comp3: (*T3)(ptr2),
					}

					// Remaining components spill over to Tail
					tail := struct { 
						Comp4 *T4
						Comp5 *T5
						Comp6 *T6
						Comp7 *T7
						Comp8 *T8
						Comp9 *T9
					}{ 
					    Comp4: (*T4)(ptr3),
					    Comp5: (*T5)(ptr4),
					    Comp6: (*T6)(ptr5),
					    Comp7: (*T7)(ptr6),
					    Comp8: (*T8)(ptr7),
					    Comp9: (*T9)(ptr8),
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)
					ptr2 = unsafe.Add(ptr2, stride2)
					ptr3 = unsafe.Add(ptr3, stride3)
					ptr4 = unsafe.Add(ptr4, stride4)
					ptr5 = unsafe.Add(ptr5, stride5)
					ptr6 = unsafe.Add(ptr6, stride6)
					ptr7 = unsafe.Add(ptr7, stride7)
					ptr8 = unsafe.Add(ptr8, stride8)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view9.Each(func(entities []core.Entity, c1s []T1, c2s []T2, c3s []T3, c4s []T4, c5s []T5, c6s []T6, c7s []T7, c8s []T8, c9s []T9) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             v3 := c3s[i]
//             v4 := c4s[i]
//             v5 := c5s[i]
//             v6 := c6s[i]
//             v7 := c7s[i]
//             v8 := c8s[i]
//             v9 := c9s[i]
//             
//         }
//     })
func (v *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Each(fn func([]core.Entity, []T1, []T2, []T3, []T4, []T5, []T6, []T7, []T8, []T9)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)
			c3 := unsafe.Slice((*T3)(unsafe.Add(base, ma.FieldsOffsets[2])), count)
			c4 := unsafe.Slice((*T4)(unsafe.Add(base, ma.FieldsOffsets[3])), count)
			c5 := unsafe.Slice((*T5)(unsafe.Add(base, ma.FieldsOffsets[4])), count)
			c6 := unsafe.Slice((*T6)(unsafe.Add(base, ma.FieldsOffsets[5])), count)
			c7 := unsafe.Slice((*T7)(unsafe.Add(base, ma.FieldsOffsets[6])), count)
			c8 := unsafe.Slice((*T8)(unsafe.Add(base, ma.FieldsOffsets[7])), count)
			c9 := unsafe.Slice((*T9)(unsafe.Add(base, ma.FieldsOffsets[8])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2, c3, c4, c5, c6, c7, c8, c9)
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

// All returns an iterator (iter.Seq2) that yields components split into two structures:
// head (containing the unique Entity identifier and components V1-V4) and tail (containing
// remaining components V5+). This split optimizes register usage and prevents heap allocation.
//
// The iteration is performed archetype by archetype, ensuring that data is
// accessed contiguously in memory, which significantly reduces CPU cache misses.
//
// Performance Note:
// For views with 8 or more components (N >= 8), it is recommended to use the
// non-iterator-based Each() method instead. At this scale, the tail structure
// grows large enough that the batch-slicing approach of Each() significantly
// outperforms the iterator overhead of All().
//
// Example usage:
//    for head, tail := range view10.All() {
//        entity := head.Entity
//        v1 := head.Comp1
//        v2 := head.Comp2
//        v3 := head.Comp3
//        v4 := head.Comp4
//        v5 := tail.V5
//        v6 := tail.V6
//        v7 := tail.V7
//        v8 := tail.V8
//        v9 := tail.V9
//        v10 := tail.V10
//        
//    }
func (v *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) All() iter.Seq2[
	struct {
		Entity core.Entity
		Comp1 *T1
		Comp2 *T2
		Comp3 *T3
		
	},
	struct {
		Comp4 *T4
		Comp5 *T5
		Comp6 *T6
		Comp7 *T7
		Comp8 *T8
		Comp9 *T9
		Comp10 *T10
		
	},
] {
	return func(yield func(
		struct {
			Entity core.Entity
			Comp1 *T1
			Comp2 *T2
			Comp3 *T3
			
		},
		struct {
			Comp4 *T4
			Comp5 *T5
			Comp6 *T6
			Comp7 *T7
			Comp8 *T8
			Comp9 *T9
			Comp10 *T10
			
		},
	) bool) {
		// 1. Pre-calculate Strides (Invariant)
		stride0 := unsafe.Sizeof(*new(T1))
		stride1 := unsafe.Sizeof(*new(T2))
		stride2 := unsafe.Sizeof(*new(T3))
		stride3 := unsafe.Sizeof(*new(T4))
		stride4 := unsafe.Sizeof(*new(T5))
		stride5 := unsafe.Sizeof(*new(T6))
		stride6 := unsafe.Sizeof(*new(T7))
		stride7 := unsafe.Sizeof(*new(T8))
		stride8 := unsafe.Sizeof(*new(T9))
		stride9 := unsafe.Sizeof(*new(T10))

		// Loop over matched archetypes
		for _, ma := range v.Baked {
			// 2. Load Offsets from Cache
			offsetEntity := ma.EntityPageOffset
			offset0 := ma.FieldsOffsets[0]
			offset1 := ma.FieldsOffsets[1]
			offset2 := ma.FieldsOffsets[2]
			offset3 := ma.FieldsOffsets[3]
			offset4 := ma.FieldsOffsets[4]
			offset5 := ma.FieldsOffsets[5]
			offset6 := ma.FieldsOffsets[6]
			offset7 := ma.FieldsOffsets[7]
			offset8 := ma.FieldsOffsets[8]
			offset9 := ma.FieldsOffsets[9]

			// 3. Loop over Physical Memory Pages
			for _, page := range ma.Arch.Memory.Pages {
				count := page.Len
				if count == 0 {
					continue
				}

				// 4. Resolve Base Pointers
				base := page.Ptr
				ptrEntity := unsafe.Add(base, offsetEntity)
				ptr0 := unsafe.Add(base, offset0)
				ptr1 := unsafe.Add(base, offset1)
				ptr2 := unsafe.Add(base, offset2)
				ptr3 := unsafe.Add(base, offset3)
				ptr4 := unsafe.Add(base, offset4)
				ptr5 := unsafe.Add(base, offset5)
				ptr6 := unsafe.Add(base, offset6)
				ptr7 := unsafe.Add(base, offset7)
				ptr8 := unsafe.Add(base, offset8)
				ptr9 := unsafe.Add(base, offset9)

				// 5. Hot Loop
				for count > 0 {
					// Max 3 components in Head to stay in CPU Registers
					head := struct {
						Entity core.Entity
						Comp1 *T1
						Comp2 *T2
						Comp3 *T3
					}{
						Entity: *(*core.Entity)(ptrEntity),
						Comp1: (*T1)(ptr0),
						Comp2: (*T2)(ptr1),
						Comp3: (*T3)(ptr2),
					}

					// Remaining components spill over to Tail
					tail := struct { 
						Comp4 *T4
						Comp5 *T5
						Comp6 *T6
						Comp7 *T7
						Comp8 *T8
						Comp9 *T9
						Comp10 *T10
					}{ 
					    Comp4: (*T4)(ptr3),
					    Comp5: (*T5)(ptr4),
					    Comp6: (*T6)(ptr5),
					    Comp7: (*T7)(ptr6),
					    Comp8: (*T8)(ptr7),
					    Comp9: (*T9)(ptr8),
					    Comp10: (*T10)(ptr9),
					}

					if !yield(head, tail) {
						return
					}

					// 6. Pointer Arithmetic
					ptrEntity = unsafe.Add(ptrEntity, core.EntitySize)
					ptr0 = unsafe.Add(ptr0, stride0)
					ptr1 = unsafe.Add(ptr1, stride1)
					ptr2 = unsafe.Add(ptr2, stride2)
					ptr3 = unsafe.Add(ptr3, stride3)
					ptr4 = unsafe.Add(ptr4, stride4)
					ptr5 = unsafe.Add(ptr5, stride5)
					ptr6 = unsafe.Add(ptr6, stride6)
					ptr7 = unsafe.Add(ptr7, stride7)
					ptr8 = unsafe.Add(ptr8, stride8)
					ptr9 = unsafe.Add(ptr9, stride9)

					count--
				}
			}
		}
	}
}

// Each executes the provided callback function across all matching archetypes,
// passing components as contiguous slices (`[]T`) alongside their corresponding `[]core.Entity`.
//
// Unlike the iterative All() approach, this method avoids per-element iterator overhead
// by processing data in bulk, archetype by archetype. This ensures optimal CPU cache
// locality and allows the compiler to better optimize dense memory loops.
//
// Performance Note:
// For views with fewer than 8 components (N < 8), it is recommended to use the
// iterator-based All() method (iter.Seq2), as Each() incurs a higher setup and
// slicing overhead that is only amortized with larger tail structures (N >= 8).
//
// Example usage:
//     view10.Each(func(entities []core.Entity, c1s []T1, c2s []T2, c3s []T3, c4s []T4, c5s []T5, c6s []T6, c7s []T7, c8s []T8, c9s []T9, c10s []T10) {
//         for i := range entities {
//             entity := entities[i]
//             v1 := c1s[i]
//             v2 := c2s[i]
//             v3 := c3s[i]
//             v4 := c4s[i]
//             v5 := c5s[i]
//             v6 := c6s[i]
//             v7 := c7s[i]
//             v8 := c8s[i]
//             v9 := c9s[i]
//             v10 := c10s[i]
//             
//         }
//     })
func (v *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Each(fn func([]core.Entity, []T1, []T2, []T3, []T4, []T5, []T6, []T7, []T8, []T9, []T10)) {
	for _, ma := range v.Baked {

		// Loop over Physical Memory Pages
		for _, page := range ma.Arch.Memory.Pages {
			count := page.Len
			if count == 0 {
				continue
			}

			// 3. Resolve Base Pointer for this Page
			base := page.Ptr

			// 4. Map raw memory pages directly to Go slices (Zero Heap Allocation)
			entities := unsafe.Slice((*core.Entity)(unsafe.Add(base, ma.EntityPageOffset)), count)
			c1 := unsafe.Slice((*T1)(unsafe.Add(base, ma.FieldsOffsets[0])), count)
			c2 := unsafe.Slice((*T2)(unsafe.Add(base, ma.FieldsOffsets[1])), count)
			c3 := unsafe.Slice((*T3)(unsafe.Add(base, ma.FieldsOffsets[2])), count)
			c4 := unsafe.Slice((*T4)(unsafe.Add(base, ma.FieldsOffsets[3])), count)
			c5 := unsafe.Slice((*T5)(unsafe.Add(base, ma.FieldsOffsets[4])), count)
			c6 := unsafe.Slice((*T6)(unsafe.Add(base, ma.FieldsOffsets[5])), count)
			c7 := unsafe.Slice((*T7)(unsafe.Add(base, ma.FieldsOffsets[6])), count)
			c8 := unsafe.Slice((*T8)(unsafe.Add(base, ma.FieldsOffsets[7])), count)
			c9 := unsafe.Slice((*T9)(unsafe.Add(base, ma.FieldsOffsets[8])), count)
			c10 := unsafe.Slice((*T10)(unsafe.Add(base, ma.FieldsOffsets[9])), count)

			// 5. Bulk Callback Execution (Once per page)
			fn(entities, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10)
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

