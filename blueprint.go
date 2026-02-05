package goke

import (
	"github.com/kjkrol/goke/internal/core"
)

type BlueprintOption = core.BlueprintOption

func Tag[T any]() BlueprintOption {
	return core.WithBlueprintTag[T]()
}
