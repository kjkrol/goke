package ecs

import (
	"github.com/kjkrol/goke/internal/core"
)

// ViewOption defines criteria for filtering entities in a query beyond
// the required components.
type ViewOption = core.ViewOption

// WithTag restricts the query to entities that possess the specified component T,
// even if T is not requested in the query's returned data.
func WithTag[T any]() ViewOption {
	return core.WithTag[T]()
}

// Without restricts the query to entities that do not possess the specified component T.
func Without[T any]() ViewOption {
	return core.Without[T]()
}
