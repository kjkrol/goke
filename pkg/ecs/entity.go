package ecs

const (
	IndexMask       = 0xFFFFFFFF
	GenerationShift = 32
)

type Entity uint64

func IndexOf(e Entity) uint32 {
	return uint32(uint64(e) & IndexMask)
}

func GenerationOf(e Entity) uint32 {
	return uint32(uint64(e) >> GenerationShift)
}

func IndexWithGenOf(e Entity) (index uint32, gen uint32) {
	return IndexOf(e), GenerationOf(e)
}

func EntityFrom(gen, index uint32) Entity {
	return Entity(uint64(gen)<<GenerationShift | uint64(index))
}
