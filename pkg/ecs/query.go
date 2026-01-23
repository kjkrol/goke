//go:generate go run ./gen/main.go
package ecs

type QueryOption func(b *ViewBuilder)

func WithTag[T any]() QueryOption {
	return func(b *ViewBuilder) {
		OnTagType[T](b)
	}
}
