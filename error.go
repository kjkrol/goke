package goke

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kjkrol/goke/internal/comp"
)

// ErrTypeMismatch is returned when the requested generic type T does not match
// the type registered for a given component ID.
var ErrTypeMismatch = errors.New("component type mismatch")

func errTypeMismatch(compID comp.ID, registered, requested reflect.Type) error {
	return fmt.Errorf("%w: component ID %d registered as %v, requested as %v",
		ErrTypeMismatch, compID, registered, requested)
}
