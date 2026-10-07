package goke

// ECSOption defines a function signature for configuring the ECS.
type ECSOption func(*Config)

// WithEntityCap sets the starting capacity for the entity pool and metadata tracking.
func WithEntityCap(cap int) ECSOption {
	return func(c *Config) {
		c.Entity.Cap = cap
	}
}

// WithEntityFreeCap sets the capacity of the recycler for deleted entity IDs.
func WithEntityFreeCap(cap int) ECSOption {
	return func(c *Config) {
		c.Entity.FreeCap = cap
	}
}

// WithMatcherCap sets how many queries the world has room for before its catalog of matchers
// grows: a hint for a world that builds many queries, never a limit.
func WithMatcherCap(cap int) ECSOption {
	return func(c *Config) {
		c.Matcher.Cap = cap
	}
}
