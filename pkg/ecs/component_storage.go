package ecs

type storageInterface interface {
	remove(e Entity)
}

type ComponentStorage[T any] struct {
	data map[Entity]*T
}

func (s *ComponentStorage[T]) remove(e Entity) {
	delete(s.data, e)
}

func (s *ComponentStorage[T]) Set(e Entity, val T) {
	v := val
	s.data[e] = &v
}

func (s *ComponentStorage[T]) Get(e Entity) *T {
	return s.data[e]
}
