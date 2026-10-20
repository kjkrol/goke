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

// WithMatcherCap sets how many matchers — one a query — the world keeps in one block of memory
// before it adds another: a hint for a world that builds many queries, never a limit.
func WithMatcherCap(cap int) ECSOption {
	return func(c *Config) {
		c.Matcher.Cap = cap
	}
}

// WithMutCheck makes every Sync panic, naming the archetype, on a chunk whose bytes changed since
// the last Sync while its change counter did not: a write through Comp where Mut belongs. For
// tests: each Sync hashes every chunk.
func WithMutCheck() ECSOption {
	return func(c *Config) {
		c.MutCheck = true
	}
}
