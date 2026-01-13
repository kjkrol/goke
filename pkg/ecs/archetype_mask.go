package ecs

import "math/bits"

const MaskSize = 4

// ArchetypeMask Should handle 4 * 64 types of components
type ArchetypeMask [MaskSize]uint64

func (b ArchetypeMask) Set(bit ComponentID) ArchetypeMask {
	word, pos := bit/64, bit%64
	if word < MaskSize {
		b[word] |= 1 << pos
	}
	return b
}

func (b ArchetypeMask) Clear(bit ComponentID) ArchetypeMask {
	word, pos := bit/64, bit%64
	if word < MaskSize {
		b[word] &= ^(1 << pos)
	}
	return b
}

func (b ArchetypeMask) Equals(other ArchetypeMask) bool {
	return b == other
}

func (b ArchetypeMask) Contains(subMask ArchetypeMask) bool {
	for i := 0; i < MaskSize; i++ {
		if (b[i] & subMask[i]) != subMask[i] {
			return false
		}
	}
	return true
}

func (b ArchetypeMask) ForEachSet(fn func(id ComponentID)) {
	for wordIdx := 0; wordIdx < MaskSize; wordIdx++ {
		word := b[wordIdx]
		for word != 0 {
			bitPos := bits.TrailingZeros64(word)
			id := ComponentID(wordIdx*64 + bitPos)
			fn(id)
			word &= ^(1 << bitPos)
		}
	}
}

func (b ArchetypeMask) IsSet(bit ComponentID) bool {
	word, pos := bit/64, bit%64
	if word >= MaskSize {
		return false
	}
	return (b[word] & (1 << pos)) != 0
}
