// Package persist encodes and decodes a world snapshot as a self-contained
// byte stream: entity ID pool bookkeeping, component type definitions,
// archetype compositions, and per-entity component data.
//
// # Format
//
// A save opens with Magic and FormatVersion. Version 2 writes an archetype's
// composition as component IDs of two bytes — up to 512 component types; a
// save of version 1, its IDs in one byte, is refused by its version, never
// read wrong.
//
// # Value encoding
//
// [EncodeValue] and [DecodeValue] read and write a single value of a given
// type at a given memory address, recursively:
//
//   - a type implementing encoding.BinaryMarshaler and
//     encoding.BinaryUnmarshaler is written as a length-prefixed opaque
//     blob (checked first, at every level — its own fields are never
//     inspected);
//   - a string is written as a length-prefixed UTF-8 blob;
//   - a struct or fixed-size array is written field-by-field or
//     element-by-element;
//   - a bool or numeric kind (including complex64/complex128) is written
//     directly, big-endian; the platform-width int and uint kinds are
//     always widened to 8 bytes for portability across builds.
//
// This mirrors, byte for byte, the rule comp.ValidateEncodable enforces at
// registration time — by the time a value reaches this package, its type is
// already known to be encodable.
package persist
