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
//
//	for page := range view1.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			// Apply domain logic here...
//		}
//	}
func (v *View1[T1]) All() iter.Seq[struct {
	Entity []Entity
	Comp1  []T1
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
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
						Comp1  []T1
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View1 (≤ 2 stateful components) uses an inline per-entity loop.
//     For thin views the overhead of materializing the SoA cache, sorting it
//     and yielding chunked pages exceeds the savings — so this adapter
//     yields a 1-row page for every matching entity. The single-row Indices
//     slice aliases cache.SingleIdx, keeping the path allocation-free.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view1.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			_ = originalIdx
//		}
//	}
func (v *View1[T1]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Indices []int
	}) bool) {
		// Inline per-entity path (avoids WalkFilteredPages callback overhead).
		// Cheaper than the chunked algorithm for thin views regardless of
		// input ordering, because each entity touches only K=1..2 columns
		// and there is no resolve/sort phase to amortize.
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore
		idx := cache.SingleIdx[:1]

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
			ePtr := unsafe.Add(physPage.Ptr, ma.EntityPageOffset+(row*core.EntitySize))
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			idx[0] = i
			if !yield(struct {
				Entity  []Entity
				Comp1   []T1
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), 1),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), 1),
				Indices: idx,
			}) {
				return
			}
		}
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view1.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//	}
func (v *View1[T1]) FilterEach(selected []Entity) iter.Seq[struct {
	Entity Entity
	Comp1  *T1
}] {
	return func(yield func(struct {
		Entity Entity
		Comp1  *T1
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for _, e := range selected {
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
			if !yield(struct {
				Entity Entity
				Comp1  *T1
			}{
				Entity: e,
				Comp1:  (*T1)(c0Ptr),
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
//
//	for page := range view2.All() {
//		for i, entity := range page.Entity {
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			// Apply domain logic here...
//		}
//	}
func (v *View2[T1, T2]) All() iter.Seq[struct {
	Entity []Entity
	Comp1  []T1
	Comp2  []T2
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
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
						Comp1  []T1
						Comp2  []T2
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View2 (≤ 2 stateful components) uses an inline per-entity loop.
//     For thin views the overhead of materializing the SoA cache, sorting it
//     and yielding chunked pages exceeds the savings — so this adapter
//     yields a 1-row page for every matching entity. The single-row Indices
//     slice aliases cache.SingleIdx, keeping the path allocation-free.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view2.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			_ = originalIdx
//		}
//	}
func (v *View2[T1, T2]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Indices []int
	}) bool) {
		// Inline per-entity path (avoids WalkFilteredPages callback overhead).
		// Cheaper than the chunked algorithm for thin views regardless of
		// input ordering, because each entity touches only K=1..2 columns
		// and there is no resolve/sort phase to amortize.
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore
		idx := cache.SingleIdx[:1]

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
			ePtr := unsafe.Add(physPage.Ptr, ma.EntityPageOffset+(row*core.EntitySize))
			c0Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[0]+(row*ma.CompSizes[0]))
			c1Ptr := unsafe.Add(physPage.Ptr, ma.CompOffsets[1]+(row*ma.CompSizes[1]))
			idx[0] = i
			if !yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), 1),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), 1),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), 1),
				Indices: idx,
			}) {
				return
			}
		}
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view2.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//	}
func (v *View2[T1, T2]) FilterEach(selected []Entity) iter.Seq[struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
}] {
	return func(yield func(struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for _, e := range selected {
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
			if !yield(struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
			}{
				Entity: e,
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
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
//
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
	Comp1  []T1
	Comp2  []T2
	Comp3  []T3
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
			Comp3  []T3
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
						Comp1  []T1
						Comp2  []T2
						Comp3  []T3
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3:  unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2, Comp3 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View3 (3 stateful components) uses the chunked algorithm in
//     core.WalkFilteredPages: Resolve → optional Sort (skipped for already
//     ordered inputs) → group contiguous runs within one archetype page.
//     The K column-offset calculations are amortized across the page, which
//     more than compensates for the resolve+sort cost on ordered inputs.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view3.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			_ = originalIdx
//		}
//	}
func (v *View3[T1, T2, T3]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Comp3   []T3
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Comp3   []T3
		Indices []int
	}) bool) {
		// Chunked path via core.WalkFilteredPages — pays off because the
		// K=3 column-offset calculations are computed once per page
		// instead of K times per row.
		core.WalkFilteredPages(v.View, selected, cache, func(p core.FilterPage) bool {
			ePtr := unsafe.Add(p.BasePtr, p.Matched.EntityPageOffset+(p.StartRow*core.EntitySize))
			c0Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[0]+(p.StartRow*p.Matched.CompSizes[0]))
			c1Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[1]+(p.StartRow*p.Matched.CompSizes[1]))
			c2Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[2]+(p.StartRow*p.Matched.CompSizes[2]))
			return yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Comp3   []T3
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), p.Count),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), p.Count),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), p.Count),
				Comp3:   unsafe.Slice((*T3)(c2Ptr), p.Count),
				Indices: p.Indices,
			})
		})
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view3.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//		c3 := item.Comp3
//	}
func (v *View3[T1, T2, T3]) FilterEach(selected []Entity) iter.Seq[struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
}] {
	return func(yield func(struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for _, e := range selected {
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
			if !yield(struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
			}{
				Entity: e,
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
				Comp3:  (*T3)(c2Ptr),
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
//
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
	Comp1  []T1
	Comp2  []T2
	Comp3  []T3
	Comp4  []T4
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
			Comp3  []T3
			Comp4  []T4
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
						Comp1  []T1
						Comp2  []T2
						Comp3  []T3
						Comp4  []T4
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3:  unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4:  unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2, Comp3, Comp4 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View4 (4 stateful components) uses the chunked algorithm in
//     core.WalkFilteredPages: Resolve → optional Sort (skipped for already
//     ordered inputs) → group contiguous runs within one archetype page.
//     The K column-offset calculations are amortized across the page, which
//     more than compensates for the resolve+sort cost on ordered inputs.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view4.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			_ = originalIdx
//		}
//	}
func (v *View4[T1, T2, T3, T4]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Comp3   []T3
	Comp4   []T4
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Comp3   []T3
		Comp4   []T4
		Indices []int
	}) bool) {
		// Chunked path via core.WalkFilteredPages — pays off because the
		// K=4 column-offset calculations are computed once per page
		// instead of K times per row.
		core.WalkFilteredPages(v.View, selected, cache, func(p core.FilterPage) bool {
			ePtr := unsafe.Add(p.BasePtr, p.Matched.EntityPageOffset+(p.StartRow*core.EntitySize))
			c0Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[0]+(p.StartRow*p.Matched.CompSizes[0]))
			c1Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[1]+(p.StartRow*p.Matched.CompSizes[1]))
			c2Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[2]+(p.StartRow*p.Matched.CompSizes[2]))
			c3Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[3]+(p.StartRow*p.Matched.CompSizes[3]))
			return yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Comp3   []T3
				Comp4   []T4
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), p.Count),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), p.Count),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), p.Count),
				Comp3:   unsafe.Slice((*T3)(c2Ptr), p.Count),
				Comp4:   unsafe.Slice((*T4)(c3Ptr), p.Count),
				Indices: p.Indices,
			})
		})
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view4.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//		c3 := item.Comp3
//		c4 := item.Comp4
//	}
func (v *View4[T1, T2, T3, T4]) FilterEach(selected []Entity) iter.Seq[struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
}] {
	return func(yield func(struct {
		Entity Entity
		Comp1  *T1
		Comp2  *T2
		Comp3  *T3
		Comp4  *T4
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for _, e := range selected {
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
			if !yield(struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
			}{
				Entity: e,
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
				Comp3:  (*T3)(c2Ptr),
				Comp4:  (*T4)(c3Ptr),
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
//
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
	Comp1  []T1
	Comp2  []T2
	Comp3  []T3
	Comp4  []T4
	Comp5  []T5
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
			Comp3  []T3
			Comp4  []T4
			Comp5  []T5
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
						Comp1  []T1
						Comp2  []T2
						Comp3  []T3
						Comp4  []T4
						Comp5  []T5
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3:  unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4:  unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5:  unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2, Comp3, Comp4, Comp5 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View5 (5 stateful components) uses the chunked algorithm in
//     core.WalkFilteredPages: Resolve → optional Sort (skipped for already
//     ordered inputs) → group contiguous runs within one archetype page.
//     The K column-offset calculations are amortized across the page, which
//     more than compensates for the resolve+sort cost on ordered inputs.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view5.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			_ = originalIdx
//		}
//	}
func (v *View5[T1, T2, T3, T4, T5]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Comp3   []T3
	Comp4   []T4
	Comp5   []T5
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Comp3   []T3
		Comp4   []T4
		Comp5   []T5
		Indices []int
	}) bool) {
		// Chunked path via core.WalkFilteredPages — pays off because the
		// K=5 column-offset calculations are computed once per page
		// instead of K times per row.
		core.WalkFilteredPages(v.View, selected, cache, func(p core.FilterPage) bool {
			ePtr := unsafe.Add(p.BasePtr, p.Matched.EntityPageOffset+(p.StartRow*core.EntitySize))
			c0Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[0]+(p.StartRow*p.Matched.CompSizes[0]))
			c1Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[1]+(p.StartRow*p.Matched.CompSizes[1]))
			c2Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[2]+(p.StartRow*p.Matched.CompSizes[2]))
			c3Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[3]+(p.StartRow*p.Matched.CompSizes[3]))
			c4Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[4]+(p.StartRow*p.Matched.CompSizes[4]))
			return yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Comp3   []T3
				Comp4   []T4
				Comp5   []T5
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), p.Count),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), p.Count),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), p.Count),
				Comp3:   unsafe.Slice((*T3)(c2Ptr), p.Count),
				Comp4:   unsafe.Slice((*T4)(c3Ptr), p.Count),
				Comp5:   unsafe.Slice((*T5)(c4Ptr), p.Count),
				Indices: p.Indices,
			})
		})
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view5.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//		c3 := item.Comp3
//		c4 := item.Comp4
//		c5 := item.Comp5
//	}
func (v *View5[T1, T2, T3, T4, T5]) FilterEach(selected []Entity) iter.Seq[struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
}] {
	return func(yield func(struct {
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
		for _, e := range selected {
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
			if !yield(struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
				Comp5  *T5
			}{
				Entity: e,
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
				Comp3:  (*T3)(c2Ptr),
				Comp4:  (*T4)(c3Ptr),
				Comp5:  (*T5)(c4Ptr),
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
//
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
	Comp1  []T1
	Comp2  []T2
	Comp3  []T3
	Comp4  []T4
	Comp5  []T5
	Comp6  []T6
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
			Comp3  []T3
			Comp4  []T4
			Comp5  []T5
			Comp6  []T6
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
						Comp1  []T1
						Comp2  []T2
						Comp3  []T3
						Comp4  []T4
						Comp5  []T5
						Comp6  []T6
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3:  unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4:  unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5:  unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6:  unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2, Comp3, Comp4, Comp5, Comp6 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View6 (6 stateful components) uses the chunked algorithm in
//     core.WalkFilteredPages: Resolve → optional Sort (skipped for already
//     ordered inputs) → group contiguous runs within one archetype page.
//     The K column-offset calculations are amortized across the page, which
//     more than compensates for the resolve+sort cost on ordered inputs.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view6.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			_ = originalIdx
//		}
//	}
func (v *View6[T1, T2, T3, T4, T5, T6]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Comp3   []T3
	Comp4   []T4
	Comp5   []T5
	Comp6   []T6
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Comp3   []T3
		Comp4   []T4
		Comp5   []T5
		Comp6   []T6
		Indices []int
	}) bool) {
		// Chunked path via core.WalkFilteredPages — pays off because the
		// K=6 column-offset calculations are computed once per page
		// instead of K times per row.
		core.WalkFilteredPages(v.View, selected, cache, func(p core.FilterPage) bool {
			ePtr := unsafe.Add(p.BasePtr, p.Matched.EntityPageOffset+(p.StartRow*core.EntitySize))
			c0Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[0]+(p.StartRow*p.Matched.CompSizes[0]))
			c1Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[1]+(p.StartRow*p.Matched.CompSizes[1]))
			c2Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[2]+(p.StartRow*p.Matched.CompSizes[2]))
			c3Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[3]+(p.StartRow*p.Matched.CompSizes[3]))
			c4Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[4]+(p.StartRow*p.Matched.CompSizes[4]))
			c5Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[5]+(p.StartRow*p.Matched.CompSizes[5]))
			return yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Comp3   []T3
				Comp4   []T4
				Comp5   []T5
				Comp6   []T6
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), p.Count),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), p.Count),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), p.Count),
				Comp3:   unsafe.Slice((*T3)(c2Ptr), p.Count),
				Comp4:   unsafe.Slice((*T4)(c3Ptr), p.Count),
				Comp5:   unsafe.Slice((*T5)(c4Ptr), p.Count),
				Comp6:   unsafe.Slice((*T6)(c5Ptr), p.Count),
				Indices: p.Indices,
			})
		})
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view6.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//		c3 := item.Comp3
//		c4 := item.Comp4
//		c5 := item.Comp5
//		c6 := item.Comp6
//	}
func (v *View6[T1, T2, T3, T4, T5, T6]) FilterEach(selected []Entity) iter.Seq[struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
	Comp6  *T6
}] {
	return func(yield func(struct {
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
		for _, e := range selected {
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
			if !yield(struct {
				Entity Entity
				Comp1  *T1
				Comp2  *T2
				Comp3  *T3
				Comp4  *T4
				Comp5  *T5
				Comp6  *T6
			}{
				Entity: e,
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
				Comp3:  (*T3)(c2Ptr),
				Comp4:  (*T4)(c3Ptr),
				Comp5:  (*T5)(c4Ptr),
				Comp6:  (*T6)(c5Ptr),
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
//
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
	Comp1  []T1
	Comp2  []T2
	Comp3  []T3
	Comp4  []T4
	Comp5  []T5
	Comp6  []T6
	Comp7  []T7
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
			Comp3  []T3
			Comp4  []T4
			Comp5  []T5
			Comp6  []T6
			Comp7  []T7
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
						Comp1  []T1
						Comp2  []T2
						Comp3  []T3
						Comp4  []T4
						Comp5  []T5
						Comp6  []T6
						Comp7  []T7
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3:  unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4:  unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5:  unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6:  unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
						Comp7:  unsafe.Slice((*T7)(unsafe.Add(base, ma.CompOffsets[6])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2, Comp3, Comp4, Comp5, Comp6, Comp7 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View7 (7 stateful components) uses the chunked algorithm in
//     core.WalkFilteredPages: Resolve → optional Sort (skipped for already
//     ordered inputs) → group contiguous runs within one archetype page.
//     The K column-offset calculations are amortized across the page, which
//     more than compensates for the resolve+sort cost on ordered inputs.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view7.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			c7 := &page.Comp7[i]
//			_ = originalIdx
//		}
//	}
func (v *View7[T1, T2, T3, T4, T5, T6, T7]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Comp3   []T3
	Comp4   []T4
	Comp5   []T5
	Comp6   []T6
	Comp7   []T7
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Comp3   []T3
		Comp4   []T4
		Comp5   []T5
		Comp6   []T6
		Comp7   []T7
		Indices []int
	}) bool) {
		// Chunked path via core.WalkFilteredPages — pays off because the
		// K=7 column-offset calculations are computed once per page
		// instead of K times per row.
		core.WalkFilteredPages(v.View, selected, cache, func(p core.FilterPage) bool {
			ePtr := unsafe.Add(p.BasePtr, p.Matched.EntityPageOffset+(p.StartRow*core.EntitySize))
			c0Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[0]+(p.StartRow*p.Matched.CompSizes[0]))
			c1Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[1]+(p.StartRow*p.Matched.CompSizes[1]))
			c2Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[2]+(p.StartRow*p.Matched.CompSizes[2]))
			c3Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[3]+(p.StartRow*p.Matched.CompSizes[3]))
			c4Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[4]+(p.StartRow*p.Matched.CompSizes[4]))
			c5Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[5]+(p.StartRow*p.Matched.CompSizes[5]))
			c6Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[6]+(p.StartRow*p.Matched.CompSizes[6]))
			return yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Comp3   []T3
				Comp4   []T4
				Comp5   []T5
				Comp6   []T6
				Comp7   []T7
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), p.Count),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), p.Count),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), p.Count),
				Comp3:   unsafe.Slice((*T3)(c2Ptr), p.Count),
				Comp4:   unsafe.Slice((*T4)(c3Ptr), p.Count),
				Comp5:   unsafe.Slice((*T5)(c4Ptr), p.Count),
				Comp6:   unsafe.Slice((*T6)(c5Ptr), p.Count),
				Comp7:   unsafe.Slice((*T7)(c6Ptr), p.Count),
				Indices: p.Indices,
			})
		})
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view7.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//		c3 := item.Comp3
//		c4 := item.Comp4
//		c5 := item.Comp5
//		c6 := item.Comp6
//		c7 := item.Comp7
//	}
func (v *View7[T1, T2, T3, T4, T5, T6, T7]) FilterEach(selected []Entity) iter.Seq[struct {
	Entity Entity
	Comp1  *T1
	Comp2  *T2
	Comp3  *T3
	Comp4  *T4
	Comp5  *T5
	Comp6  *T6
	Comp7  *T7
}] {
	return func(yield func(struct {
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
		for _, e := range selected {
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
			if !yield(struct {
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
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
				Comp3:  (*T3)(c2Ptr),
				Comp4:  (*T4)(c3Ptr),
				Comp5:  (*T5)(c4Ptr),
				Comp6:  (*T6)(c5Ptr),
				Comp7:  (*T7)(c6Ptr),
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
//
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
	Comp1  []T1
	Comp2  []T2
	Comp3  []T3
	Comp4  []T4
	Comp5  []T5
	Comp6  []T6
	Comp7  []T7
	Comp8  []T8
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
			Comp3  []T3
			Comp4  []T4
			Comp5  []T5
			Comp6  []T6
			Comp7  []T7
			Comp8  []T8
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
						Comp1  []T1
						Comp2  []T2
						Comp3  []T3
						Comp4  []T4
						Comp5  []T5
						Comp6  []T6
						Comp7  []T7
						Comp8  []T8
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3:  unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4:  unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5:  unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6:  unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
						Comp7:  unsafe.Slice((*T7)(unsafe.Add(base, ma.CompOffsets[6])), count),
						Comp8:  unsafe.Slice((*T8)(unsafe.Add(base, ma.CompOffsets[7])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2, Comp3, Comp4, Comp5, Comp6, Comp7, Comp8 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View8 (8 stateful components) uses the chunked algorithm in
//     core.WalkFilteredPages: Resolve → optional Sort (skipped for already
//     ordered inputs) → group contiguous runs within one archetype page.
//     The K column-offset calculations are amortized across the page, which
//     more than compensates for the resolve+sort cost on ordered inputs.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view8.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			c7 := &page.Comp7[i]
//			c8 := &page.Comp8[i]
//			_ = originalIdx
//		}
//	}
func (v *View8[T1, T2, T3, T4, T5, T6, T7, T8]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Comp3   []T3
	Comp4   []T4
	Comp5   []T5
	Comp6   []T6
	Comp7   []T7
	Comp8   []T8
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Comp3   []T3
		Comp4   []T4
		Comp5   []T5
		Comp6   []T6
		Comp7   []T7
		Comp8   []T8
		Indices []int
	}) bool) {
		// Chunked path via core.WalkFilteredPages — pays off because the
		// K=8 column-offset calculations are computed once per page
		// instead of K times per row.
		core.WalkFilteredPages(v.View, selected, cache, func(p core.FilterPage) bool {
			ePtr := unsafe.Add(p.BasePtr, p.Matched.EntityPageOffset+(p.StartRow*core.EntitySize))
			c0Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[0]+(p.StartRow*p.Matched.CompSizes[0]))
			c1Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[1]+(p.StartRow*p.Matched.CompSizes[1]))
			c2Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[2]+(p.StartRow*p.Matched.CompSizes[2]))
			c3Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[3]+(p.StartRow*p.Matched.CompSizes[3]))
			c4Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[4]+(p.StartRow*p.Matched.CompSizes[4]))
			c5Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[5]+(p.StartRow*p.Matched.CompSizes[5]))
			c6Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[6]+(p.StartRow*p.Matched.CompSizes[6]))
			c7Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[7]+(p.StartRow*p.Matched.CompSizes[7]))
			return yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Comp3   []T3
				Comp4   []T4
				Comp5   []T5
				Comp6   []T6
				Comp7   []T7
				Comp8   []T8
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), p.Count),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), p.Count),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), p.Count),
				Comp3:   unsafe.Slice((*T3)(c2Ptr), p.Count),
				Comp4:   unsafe.Slice((*T4)(c3Ptr), p.Count),
				Comp5:   unsafe.Slice((*T5)(c4Ptr), p.Count),
				Comp6:   unsafe.Slice((*T6)(c5Ptr), p.Count),
				Comp7:   unsafe.Slice((*T7)(c6Ptr), p.Count),
				Comp8:   unsafe.Slice((*T8)(c7Ptr), p.Count),
				Indices: p.Indices,
			})
		})
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view8.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//		c3 := item.Comp3
//		c4 := item.Comp4
//		c5 := item.Comp5
//		c6 := item.Comp6
//		c7 := item.Comp7
//		c8 := item.Comp8
//	}
func (v *View8[T1, T2, T3, T4, T5, T6, T7, T8]) FilterEach(selected []Entity) iter.Seq[struct {
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
	return func(yield func(struct {
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
		for _, e := range selected {
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
			if !yield(struct {
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
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
				Comp3:  (*T3)(c2Ptr),
				Comp4:  (*T4)(c3Ptr),
				Comp5:  (*T5)(c4Ptr),
				Comp6:  (*T6)(c5Ptr),
				Comp7:  (*T7)(c6Ptr),
				Comp8:  (*T8)(c7Ptr),
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
//
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
	Comp1  []T1
	Comp2  []T2
	Comp3  []T3
	Comp4  []T4
	Comp5  []T5
	Comp6  []T6
	Comp7  []T7
	Comp8  []T8
	Comp9  []T9
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
			Comp3  []T3
			Comp4  []T4
			Comp5  []T5
			Comp6  []T6
			Comp7  []T7
			Comp8  []T8
			Comp9  []T9
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
						Comp1  []T1
						Comp2  []T2
						Comp3  []T3
						Comp4  []T4
						Comp5  []T5
						Comp6  []T6
						Comp7  []T7
						Comp8  []T8
						Comp9  []T9
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3:  unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4:  unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5:  unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6:  unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
						Comp7:  unsafe.Slice((*T7)(unsafe.Add(base, ma.CompOffsets[6])), count),
						Comp8:  unsafe.Slice((*T8)(unsafe.Add(base, ma.CompOffsets[7])), count),
						Comp9:  unsafe.Slice((*T9)(unsafe.Add(base, ma.CompOffsets[8])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2, Comp3, Comp4, Comp5, Comp6, Comp7, Comp8, Comp9 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View9 (9 stateful components) uses the chunked algorithm in
//     core.WalkFilteredPages: Resolve → optional Sort (skipped for already
//     ordered inputs) → group contiguous runs within one archetype page.
//     The K column-offset calculations are amortized across the page, which
//     more than compensates for the resolve+sort cost on ordered inputs.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view9.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
//			c1 := &page.Comp1[i]
//			c2 := &page.Comp2[i]
//			c3 := &page.Comp3[i]
//			c4 := &page.Comp4[i]
//			c5 := &page.Comp5[i]
//			c6 := &page.Comp6[i]
//			c7 := &page.Comp7[i]
//			c8 := &page.Comp8[i]
//			c9 := &page.Comp9[i]
//			_ = originalIdx
//		}
//	}
func (v *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Comp3   []T3
	Comp4   []T4
	Comp5   []T5
	Comp6   []T6
	Comp7   []T7
	Comp8   []T8
	Comp9   []T9
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Comp3   []T3
		Comp4   []T4
		Comp5   []T5
		Comp6   []T6
		Comp7   []T7
		Comp8   []T8
		Comp9   []T9
		Indices []int
	}) bool) {
		// Chunked path via core.WalkFilteredPages — pays off because the
		// K=9 column-offset calculations are computed once per page
		// instead of K times per row.
		core.WalkFilteredPages(v.View, selected, cache, func(p core.FilterPage) bool {
			ePtr := unsafe.Add(p.BasePtr, p.Matched.EntityPageOffset+(p.StartRow*core.EntitySize))
			c0Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[0]+(p.StartRow*p.Matched.CompSizes[0]))
			c1Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[1]+(p.StartRow*p.Matched.CompSizes[1]))
			c2Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[2]+(p.StartRow*p.Matched.CompSizes[2]))
			c3Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[3]+(p.StartRow*p.Matched.CompSizes[3]))
			c4Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[4]+(p.StartRow*p.Matched.CompSizes[4]))
			c5Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[5]+(p.StartRow*p.Matched.CompSizes[5]))
			c6Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[6]+(p.StartRow*p.Matched.CompSizes[6]))
			c7Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[7]+(p.StartRow*p.Matched.CompSizes[7]))
			c8Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[8]+(p.StartRow*p.Matched.CompSizes[8]))
			return yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Comp3   []T3
				Comp4   []T4
				Comp5   []T5
				Comp6   []T6
				Comp7   []T7
				Comp8   []T8
				Comp9   []T9
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), p.Count),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), p.Count),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), p.Count),
				Comp3:   unsafe.Slice((*T3)(c2Ptr), p.Count),
				Comp4:   unsafe.Slice((*T4)(c3Ptr), p.Count),
				Comp5:   unsafe.Slice((*T5)(c4Ptr), p.Count),
				Comp6:   unsafe.Slice((*T6)(c5Ptr), p.Count),
				Comp7:   unsafe.Slice((*T7)(c6Ptr), p.Count),
				Comp8:   unsafe.Slice((*T8)(c7Ptr), p.Count),
				Comp9:   unsafe.Slice((*T9)(c8Ptr), p.Count),
				Indices: p.Indices,
			})
		})
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view9.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//		c3 := item.Comp3
//		c4 := item.Comp4
//		c5 := item.Comp5
//		c6 := item.Comp6
//		c7 := item.Comp7
//		c8 := item.Comp8
//		c9 := item.Comp9
//	}
func (v *View9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) FilterEach(selected []Entity) iter.Seq[struct {
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
	return func(yield func(struct {
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
		for _, e := range selected {
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
			if !yield(struct {
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
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
				Comp3:  (*T3)(c2Ptr),
				Comp4:  (*T4)(c3Ptr),
				Comp5:  (*T5)(c4Ptr),
				Comp6:  (*T6)(c5Ptr),
				Comp7:  (*T7)(c6Ptr),
				Comp8:  (*T8)(c7Ptr),
				Comp9:  (*T9)(c8Ptr),
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
//
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
	Comp1  []T1
	Comp2  []T2
	Comp3  []T3
	Comp4  []T4
	Comp5  []T5
	Comp6  []T6
	Comp7  []T7
	Comp8  []T8
	Comp9  []T9
	Comp10 []T10
}] {
	return func(yield func(
		struct {
			Entity []Entity
			Comp1  []T1
			Comp2  []T2
			Comp3  []T3
			Comp4  []T4
			Comp5  []T5
			Comp6  []T6
			Comp7  []T7
			Comp8  []T8
			Comp9  []T9
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
						Comp1  []T1
						Comp2  []T2
						Comp3  []T3
						Comp4  []T4
						Comp5  []T5
						Comp6  []T6
						Comp7  []T7
						Comp8  []T8
						Comp9  []T9
						Comp10 []T10
					}{
						Entity: unsafe.Slice((*Entity)(unsafe.Add(base, ma.EntityPageOffset)), count),
						Comp1:  unsafe.Slice((*T1)(unsafe.Add(base, ma.CompOffsets[0])), count),
						Comp2:  unsafe.Slice((*T2)(unsafe.Add(base, ma.CompOffsets[1])), count),
						Comp3:  unsafe.Slice((*T3)(unsafe.Add(base, ma.CompOffsets[2])), count),
						Comp4:  unsafe.Slice((*T4)(unsafe.Add(base, ma.CompOffsets[3])), count),
						Comp5:  unsafe.Slice((*T5)(unsafe.Add(base, ma.CompOffsets[4])), count),
						Comp6:  unsafe.Slice((*T6)(unsafe.Add(base, ma.CompOffsets[5])), count),
						Comp7:  unsafe.Slice((*T7)(unsafe.Add(base, ma.CompOffsets[6])), count),
						Comp8:  unsafe.Slice((*T8)(unsafe.Add(base, ma.CompOffsets[7])), count),
						Comp9:  unsafe.Slice((*T9)(unsafe.Add(base, ma.CompOffsets[8])), count),
						Comp10: unsafe.Slice((*T10)(unsafe.Add(base, ma.CompOffsets[9])), count),
					}) {
					return
				}
			}
		}
	}
}

// Filter iterates `selected` entities and yields page-shaped views over the
// matching ones. Each yielded value exposes Entity, Comp1, Comp2, Comp3, Comp4, Comp5, Comp6, Comp7, Comp8, Comp9, Comp10 as Go
// slices over native memory (Zero Heap Allocation) plus an Indices slice that
// maps each row back to its position in `selected` (useful for correlating
// results with side-tables built during a resolve phase).
//
// The `cache` parameter is a reusable scratchpad — store it in a system
// field (or sync.Pool) to eliminate per-call allocations. Call cache.Grow(n)
// once at startup to pre-size the buffers to your expected working set.
//
// Two implementation strategies are dispatched at code-generation time based
// on the view's component count:
//
//   - View10 (10 stateful components) uses the chunked algorithm in
//     core.WalkFilteredPages: Resolve → optional Sort (skipped for already
//     ordered inputs) → group contiguous runs within one archetype page.
//     The K column-offset calculations are amortized across the page, which
//     more than compensates for the resolve+sort cost on ordered inputs.
//
// Safety guarantee: Filter dynamically verifies that each entity's current
// composition (archetype) still matches the View before yielding. This safely
// prevents invalid memory access if an entity was mutated (components added
// or removed) after the 'selected' slice was built.
//
// Example usage:
//
//	for page := range view10.Filter(selected, &cache) {
//		for i, originalIdx := range page.Indices {
//			entity := page.Entity[i]
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
//			_ = originalIdx
//		}
//	}
func (v *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Filter(selected []Entity, cache *FilterCache) iter.Seq[struct {
	Entity  []Entity
	Comp1   []T1
	Comp2   []T2
	Comp3   []T3
	Comp4   []T4
	Comp5   []T5
	Comp6   []T6
	Comp7   []T7
	Comp8   []T8
	Comp9   []T9
	Comp10  []T10
	Indices []int
}] {
	return func(yield func(struct {
		Entity  []Entity
		Comp1   []T1
		Comp2   []T2
		Comp3   []T3
		Comp4   []T4
		Comp5   []T5
		Comp6   []T6
		Comp7   []T7
		Comp8   []T8
		Comp9   []T9
		Comp10  []T10
		Indices []int
	}) bool) {
		// Chunked path via core.WalkFilteredPages — pays off because the
		// K=10 column-offset calculations are computed once per page
		// instead of K times per row.
		core.WalkFilteredPages(v.View, selected, cache, func(p core.FilterPage) bool {
			ePtr := unsafe.Add(p.BasePtr, p.Matched.EntityPageOffset+(p.StartRow*core.EntitySize))
			c0Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[0]+(p.StartRow*p.Matched.CompSizes[0]))
			c1Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[1]+(p.StartRow*p.Matched.CompSizes[1]))
			c2Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[2]+(p.StartRow*p.Matched.CompSizes[2]))
			c3Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[3]+(p.StartRow*p.Matched.CompSizes[3]))
			c4Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[4]+(p.StartRow*p.Matched.CompSizes[4]))
			c5Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[5]+(p.StartRow*p.Matched.CompSizes[5]))
			c6Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[6]+(p.StartRow*p.Matched.CompSizes[6]))
			c7Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[7]+(p.StartRow*p.Matched.CompSizes[7]))
			c8Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[8]+(p.StartRow*p.Matched.CompSizes[8]))
			c9Ptr := unsafe.Add(p.BasePtr, p.Matched.CompOffsets[9]+(p.StartRow*p.Matched.CompSizes[9]))
			return yield(struct {
				Entity  []Entity
				Comp1   []T1
				Comp2   []T2
				Comp3   []T3
				Comp4   []T4
				Comp5   []T5
				Comp6   []T6
				Comp7   []T7
				Comp8   []T8
				Comp9   []T9
				Comp10  []T10
				Indices []int
			}{
				Entity:  unsafe.Slice((*Entity)(ePtr), p.Count),
				Comp1:   unsafe.Slice((*T1)(c0Ptr), p.Count),
				Comp2:   unsafe.Slice((*T2)(c1Ptr), p.Count),
				Comp3:   unsafe.Slice((*T3)(c2Ptr), p.Count),
				Comp4:   unsafe.Slice((*T4)(c3Ptr), p.Count),
				Comp5:   unsafe.Slice((*T5)(c4Ptr), p.Count),
				Comp6:   unsafe.Slice((*T6)(c5Ptr), p.Count),
				Comp7:   unsafe.Slice((*T7)(c6Ptr), p.Count),
				Comp8:   unsafe.Slice((*T8)(c7Ptr), p.Count),
				Comp9:   unsafe.Slice((*T9)(c8Ptr), p.Count),
				Comp10:  unsafe.Slice((*T10)(c9Ptr), p.Count),
				Indices: p.Indices,
			})
		})
	}
}

// FilterEach iterates `selected` entities and yields one struct per matching
// entity — a per-entity counterpart to Filter, modelled after View0.Filter.
//
// Unlike Filter, FilterEach:
//   - takes no FilterCache (no resolve/sort/grouping phases)
//   - yields *pointers* to the live component memory (mutate in place)
//   - returns no Indices (the caller already has each entity's position via
//     the enumerated for-loop over the returned sequence)
//
// It is the lowest-overhead path for scanning a known list of entity handles
// when the caller does not need page-shaped slices for vectorized loops.
//
// Safety guarantee: FilterEach dynamically verifies that each entity's
// current composition (archetype) still matches the View before yielding,
// so it remains safe even if entities were mutated after `selected` was built.
//
// Example usage:
//
//	for item := range view10.FilterEach(selected) {
//		entity := item.Entity
//		c1 := item.Comp1
//		c2 := item.Comp2
//		c3 := item.Comp3
//		c4 := item.Comp4
//		c5 := item.Comp5
//		c6 := item.Comp6
//		c7 := item.Comp7
//		c8 := item.Comp8
//		c9 := item.Comp9
//		c10 := item.Comp10
//	}
func (v *View10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) FilterEach(selected []Entity) iter.Seq[struct {
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
	Comp10 *T10
}] {
	return func(yield func(struct {
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
		Comp10 *T10
	}) bool) {
		store := &v.Reg.ArchetypeRegistry.EntityLinkStore

		var lastArchID core.ArchetypeId = core.NullArchetypeId
		var ma *core.MatchedArch
		for _, e := range selected {
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
			if !yield(struct {
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
				Comp10 *T10
			}{
				Entity: e,
				Comp1:  (*T1)(c0Ptr),
				Comp2:  (*T2)(c1Ptr),
				Comp3:  (*T3)(c2Ptr),
				Comp4:  (*T4)(c3Ptr),
				Comp5:  (*T5)(c4Ptr),
				Comp6:  (*T6)(c5Ptr),
				Comp7:  (*T7)(c6Ptr),
				Comp8:  (*T8)(c7Ptr),
				Comp9:  (*T9)(c8Ptr),
				Comp10: (*T10)(c9Ptr),
			}) {
				return
			}
		}
	}
}
