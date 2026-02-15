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

// WithFreeIndicesCap sets the capacity of the recycler for deleted entity IDs.
// This should ideally match InitialEntityCap to handle mass deletions without allocation.
func WithFreeIndicesCap(cap int) ECSOption {
	return func(c *ECSConfig) {
		c.FreeIndicesCap = cap
	}
}
