package goke

// ECSOption defines a function signature for configuring the ECS.
type ECSOption func(*ECSConfig)

// WithInitialEntityCap sets the starting capacity for the entity pool and metadata tracking.
// Increasing this prevents reallocations when the entity count grows significantly.
func WithInitialEntityCap(cap int) ECSOption {
	return func(c *ECSConfig) {
		c.InitialEntityCap = cap
	}
}

// WithDefaultArchetypeChunkSize sets the number of entities per memory chunk in an archetype.
// Larger values reduce the frequency of 'growTo' operations but increase memory footprint
// for sparsely populated archetypes.
func WithDefaultArchetypeChunkSize(size int) ECSOption {
	return func(c *ECSConfig) {
		c.DefaultArchetypeChunkSize = size
	}
}

// WithInitialArchetypeRegistryCap sets the initial capacity for the archetype storage map.
// Pre-allocating this prevents map resizing as new unique component combinations are discovered.
func WithInitialArchetypeRegistryCap(cap int) ECSOption {
	return func(c *ECSConfig) {
		c.InitialArchetypeRegistryCap = cap
	}
}

// WithFreeIndicesCap sets the capacity of the recycler for deleted entity IDs.
// This should ideally match InitialEntityCap to handle mass deletions without allocation.
func WithFreeIndicesCap(cap int) ECSOption {
	return func(c *ECSConfig) {
		c.FreeIndicesCap = cap
	}
}

// WithViewRegistryInitCap sets the initial capacity for the query/view cache.
// Optimization for scenarios with a high number of unique system views.
func WithViewRegistryInitCap(cap int) ECSOption {
	return func(c *ECSConfig) {
		c.ViewRegistryInitCap = cap
	}
}
