// Package comp defines component primitives used across all internal packages.
// It contains no business logic — only types, constants, and the component registry.
//
// # ID and Meta
//
// Each registered Go type receives a unique [ID] (uint8).
// [Meta] carries the ID, memory size, alignment, and reflect.Type
// used during layout calculation.
//
// [MetaIndex] maps Go types to [Meta] at initialization time.
// Component registration is sequential and deterministic — the first registered
// type gets ID 0, the next gets ID 1, and so on.
//
// # Mask
//
// [Mask] is a bitset of component IDs, used to identify archetype composition.
//
// # BlueprintOpt
//
// [BlueprintOpt] is a functional option that configures a [Blueprint]:
//   - [Track][T] — registers T as a data column; sets Col[T].Idx to its position
//   - [Include][T] — adds T as a filter-only requirement (no data column)
//   - [Exclude][T] — adds T as an exclusion constraint
//
// # Blueprint
//
// [Blueprint] is a pure value object describing a set of component requirements:
// which component types to include ([Blueprint.Comp], [Blueprint.Tag]) and which
// to exclude ([Blueprint.Exclude]). Initialise one from opts via [Blueprint.Init].
// Call [Blueprint.Compose] to derive a [Composition] without a registry reference.
//
// # Composition
//
// [Composition] describes a fully resolved component composition: a [Mask] plus
// the [Meta] slice for all non-tag components. Derived from a [Blueprint] via
// [Blueprint.Compose].
//
// # Constants
//
//	MaskSize      = 2    // number of uint64 words in Mask
//	MaxComponents = 128  // 64 * MaskSize — max registered component types
package comp
