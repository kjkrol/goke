// Package comp defines component primitives shared across all internal packages.
// It contains no business logic — only types, constants, and the component registry.
//
// # ID and Def
//
// Each registered Go type receives a unique [ID] (uint8).
// [Def] carries the ID, memory size, alignment, and reflect.Type
// used during layout calculation. [DefIndex] maps Go types to [Def] in O(1).
//
// # Mask
//
// [Mask] is a fixed-size bitset of component IDs that identifies archetype composition.
//
// # BlueprintOpt
//
// [BlueprintOpt] is a functional option that configures a [Blueprint]:
//   - [Track][T] — registers T as a data column; sets Col[T].Idx to its position
//   - [Include][T] — adds T as a filter-only requirement (no data column)
//   - [Exclude][T] — adds T as an exclusion constraint
//
// # Constants
//
//	MaskSize      = 2    // number of uint64 words in Mask
//	MaxComponents = 128  // 64 * MaskSize — max registered component types
package comp
