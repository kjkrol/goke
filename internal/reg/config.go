package reg

import (
	"github.com/kjkrol/goke/internal/ent"
	"github.com/kjkrol/goke/internal/query"
)

// Config defines the initial sizing of internal data structures.
type Config struct {
	Entity ent.Config
	View   query.Config
}

// DefaultConfig returns a configuration with sensible defaults for most use cases.
func DefaultConfig() Config {
	return Config{
		Entity: ent.DefaultConfig(),
		View:   query.DefaultConfig(),
	}
}
