package goke

import (
	"reflect"

	"github.com/kjkrol/goke/internal/core"
)

type BlueprintOption func(*core.Blueprint) error

func Include[T any]() BlueprintOption {
	return func(b *core.Blueprint) error {
		componentType := reflect.TypeFor[T]()
		compInfo := b.Reg.ComponentsRegistry.GetOrRegister(componentType)
		return b.WithTag(compInfo.ID)
	}
}

func Exclude[T any]() BlueprintOption {
	return func(b *core.Blueprint) error {
		componentType := reflect.TypeFor[T]()
		compInfo := b.Reg.ComponentsRegistry.GetOrRegister(componentType)
		return b.ExcludeType(compInfo.ID)
	}
}
