package reg

import (
	"github.com/kjkrol/goke/v3/internal/ent"
	"github.com/kjkrol/goke/v3/internal/query"
)

// Config defines the initial sizing of internal data structures.
type Config struct {
	Entity  ent.Config
	Matcher query.Config
	// MutCheck makes every Sync check that no chunk changed past its change counter.
	MutCheck bool
}

// DefaultConfig returns a configuration with sensible defaults for most use cases.
func DefaultConfig() Config {
	return Config{
		Entity:  ent.DefaultConfig(),
		Matcher: query.DefaultConfig(),
	}
}
