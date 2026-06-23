package query

const (
	// MaxMatchers is the maximum number of Matchers that can be registered in a Catalog.
	// The Catalog pre-allocates this capacity once at Init time so that pointers
	// to individual Matcher slots remain stable for the lifetime of the ECS world.
	MaxMatchers = 64
)
