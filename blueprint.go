package goke

import (
	"github.com/kjkrol/goke/internal/core"
)

type BlueprintOption func(*core.Blueprint)

func Include[T any]() BlueprintOption {
	return func(b *core.Blueprint) {
		compInfo := core.EnsureComponentRegistered[T](b.Reg.ComponentsRegistry)
		b.WithTag(compInfo.ID)
	}
}

func Exclude[T any]() BlueprintOption {
	return func(b *core.Blueprint) {
		compInfo := core.EnsureComponentRegistered[T](b.Reg.ComponentsRegistry)
		b.ExcludeType(compInfo.ID)
	}
}
