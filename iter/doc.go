// Package iter provides the low-level column-access primitives for SoA iteration.
//
// # Cursor
//
// [Cursor] is a value that describes the current position within a column iteration:
//
//   - Base      — raw pointer to the start of the current memory block
//   - Offsets   — per-component byte offsets from Base, one entry per tracked component
//   - Slot      — index into IDs (and into each tracked array) for single-entity access
//   - IDs       — the entity IDs currently addressable via Base/Offsets
//
// # ArrayRef[T]
//
// [ArrayRef][T] locates the array for component type T within the memory
// Base points to. It holds a single int field — Idx — which is the index
// into Cursor.Offsets that corresponds to T. Given a Cursor, ArrayRef[T]
// performs pointer arithmetic to produce either a typed slice
// ([ArrayRef.Slice]) or a typed pointer ([ArrayRef.At]) with no allocation
// and no type switch.
//
// Idx is a plain field, not computed — the caller is responsible for setting
// it to the correct index before calling Slice or At. Its zero value (0) is
// only correct if T's array happens to be the first one in Cursor.Offsets;
// otherwise it must be assigned explicitly first.
package iter
