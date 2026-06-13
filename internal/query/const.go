package query

const (
	// MaxViews is the maximum number of Views that can be registered in a Catalog.
	// The Catalog pre-allocates this capacity once at Init time so that pointers
	// to individual View slots remain stable for the lifetime of the ECS world.
	MaxViews = 64
)
