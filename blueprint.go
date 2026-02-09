package goke

import (
	"github.com/kjkrol/goke/internal/core"
)

type BlueprintOption func(*core.Blueprint) error

func Include[T any]() BlueprintOption {
	return func(b *core.Blueprint) error {
		compInfo := core.EnsureComponentRegistered[T](&b.Reg.ComponentsRegistry)
		return b.WithTag(compInfo.ID)
	}
}

func Exclude[T any]() BlueprintOption {
	return func(b *core.Blueprint) error {
		compInfo := core.EnsureComponentRegistered[T](&b.Reg.ComponentsRegistry)
		return b.ExcludeType(compInfo.ID)
	}
}
