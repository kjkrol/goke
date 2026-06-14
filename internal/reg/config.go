package reg

// Config defines the initial sizing of internal data structures.
type Config struct {
	// InitialEntityCap sets the pre-allocated capacity for arrays indexed by Entity ID.
	InitialEntityCap int

	// InitialArchetypeRegistryCap sets the initial capacity for the unique archetype combinations map.
	InitialArchetypeRegistryCap int

	// FreeIndicesCap defines the initial capacity of the recycled entity ID stack.
	FreeIndicesCap int

	// ViewRegistryInitCap sets the initial capacity for the query/view cache.
	ViewRegistryInitCap int
}

// DefaultConfig returns a configuration with sensible defaults for most use cases.
func DefaultConfig() Config {
	return Config{
		InitialEntityCap:            1000,
		InitialArchetypeRegistryCap: 64,
		FreeIndicesCap:              1024,
		ViewRegistryInitCap:         32,
	}
}
