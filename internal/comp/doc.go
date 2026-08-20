// Package comp defines component primitives shared across all internal packages.
// It contains no business logic — only types, constants, and the component registry.
//
// # ID and Def
//
// Each registered Go type receives a unique [ID] (uint8).
// [ID] is the identity of a component. [Def] additionally carries the layout
// information (memory size, alignment, reflect.Type) needed only when storage
// is laid out; identity alone suffices otherwise.
// [DefIndex] maps Go types to [Def] in O(1) and resolves an [ID] back to its
// [Def] via [DefIndex.ByID].
//
// # Mask
//
// [Mask] is a fixed-size bitset of component IDs.
//
// # AccessOpt
//
// [AccessOpt] is a functional option that configures an [AccessSpec]:
//   - [Track][T] — registers T as a data column; sets ArrayRef[T].Idx to its position
//   - [Include][T] — adds T as a filter-only requirement (no data column)
//   - [Exclude][T] — adds T as an exclusion constraint
//
// # Encoding constraints
//
// [DefIndex.Intern] rejects (panics) a type that cannot be encoded:
// recursively, through structs and fixed-size arrays, every field must be a
// bool, a numeric kind, a string, or a type implementing
// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler — checked first,
// at every level, before a type's own fields are inspected. Ptr,
// UnsafePointer, Slice, Map, Interface, Chan, and Func are rejected unless
// covered by that escape hatch. See [ValidateEncodable] and
// [OffChunkFields].
//
// # Constants
//
//	MaskSize      = 2    // number of uint64 words in Mask
//	MaxComponents = 128  // 64 * MaskSize — max registered component types
package comp
