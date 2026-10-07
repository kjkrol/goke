package query

const (
	// InitialMatchers is how many Matchers a Catalog keeps in one block of memory: past it a new
	// block follows — a hint, never a limit.
	InitialMatchers = 64
)
