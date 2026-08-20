package persist

// CompRequest resolves and registers one component type for Load, in
// whatever order Load's caller provides — order does not matter, because
// Load matches requests to the save file's component directory by Name, not
// by position, and drives registration itself in the file's recorded order.
// Produced by the goke package's RegisterFor[T]().
type CompRequest struct {
	// Name is the real Go type's name (reflect.Type.String()), used to
	// match this request against a save file's component directory entry.
	Name string

	// Register validates and registers the type. wantSize is the matching
	// directory entry's recorded size, or nil if this request did not match
	// any entry (a type not present in the save file — registered anyway,
	// for forward compatibility with saves made before it existed).
	Register func(wantSize *uint32) error
}
