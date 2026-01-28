package core

// RegistryConfig defines the initial sizing of internal data structures.
type RegistryConfig struct {
	// InitialEntityCap sets the pre-allocated capacity for arrays indexed by Entity ID.
	// This affects EntityPool.generations and ArchetypeRegistry.EntityArchLinks.
	InitialEntityCap int

	// DefaultArchetypeChunkSize sets the default capacity for component columns
	// when a new Archetype is created.
	DefaultArchetypeChunkSize int

	// InitialArchetypeRegistryCap sets the initial capacity for the unique
	// archetype combinations map.
	InitialArchetypeRegistryCap int

	// FreeIndicesCap defines the initial capacity of the recycled entity ID stack.
	FreeIndicesCap int

	// ViewRegistryInitCap sets the initial capacity for the query/view cache.
	ViewRegistryInitCap int
}

// DefaultRegistryConfig returns a configuration with sensible defaults for most games.
func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		InitialEntityCap:            1000,
		DefaultArchetypeChunkSize:   100,
		InitialArchetypeRegistryCap: 64,
		FreeIndicesCap:              1024,
		ViewRegistryInitCap:         32,
	}
}
