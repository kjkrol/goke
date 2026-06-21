// Package iter provides the low-level column-access primitives for SoA iteration.
//
// # Cursor
//
// [Cursor] is a value that describes the current position within a column iteration:
//
//   - Base      — raw pointer to the start of a Chunk's data allocation
//   - Offsets   — per-component byte offsets from Base, one entry per tracked component
//   - Slot      — index of the current entity within the Chunk (filter-mode only)
//   - IDs  — entity ID slice for the current chunk or batch (all-mode and factory-mode)
//
// Cursor carries no type information. Both Offsets and IDs are slice fields;
// Offsets is allocated once at construction time, IDs is set on each Next() call.
//
// # Col[T]
//
// [Col][T] is a typed column handle. It holds a single int field — Idx — which is
// the index into Cursor.Offsets that corresponds to component type T. Given a Cursor,
// Col[T] performs pointer arithmetic to produce either a typed slice ([Col.Slice]) or
// a typed pointer ([Col.At]) with no allocation and no type switch.
//
// Idx is set by Track at blueprint construction time; each Track call overwrites it,
// so the last registration wins. A Col[T] that has never been passed to Track must
// not be used.
package iter
