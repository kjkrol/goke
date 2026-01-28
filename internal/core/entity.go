package core

const (
	IndexMask         = 0xFFFFFFFF
	GenerationShift   = 32
	VirtualEntityMask = uint32(1 << 31)
)

type Entity uint64

func (e Entity) Index() uint32 {
	return uint32(uint64(e) & IndexMask)
}

func (e Entity) Generation() uint32 {
	return uint32(uint64(e) >> GenerationShift)
}

func (e Entity) Unpack() (index uint32, gen uint32) {
	return e.Index(), e.Generation()
}

func NewEntity(gen, index uint32) Entity {
	return Entity(uint64(gen)<<GenerationShift | uint64(index))
}

func NewVirtualEntity(id uint32) Entity {
	return NewEntity(0, id|VirtualEntityMask)
}

func (e Entity) IsVirtual() bool {
	return (e.Index() & VirtualEntityMask) != 0
}
