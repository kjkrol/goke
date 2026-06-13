package reg

import "errors"

var (
	errInvalidEntity    = errors.New("invalid entity")
	errComponentMissing = errors.New("component not found in archetype")
)
