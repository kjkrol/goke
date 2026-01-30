package core

import (
	"iter"
	"math/bits"
)

const (
	MaskSize      = 2
	MaxComponents = 128
)

// ArchetypeMask Should handle 2 * 64 = 128 types of components
type ArchetypeMask [MaskSize]uint64

func NewArchetypeMask(componentIDs ...ComponentID) ArchetypeMask {
	var mask ArchetypeMask
	for _, id := range componentIDs {
		mask = mask.Set(id)
	}
	return mask
}

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

func (b ArchetypeMask) AllSet() iter.Seq[ComponentID] {
	return func(yield func(ComponentID) bool) {
		for wordIdx := 0; wordIdx < MaskSize; wordIdx++ {
			word := b[wordIdx]
			for word != 0 {
				bitPos := bits.TrailingZeros64(word)
				id := ComponentID(wordIdx*64 + bitPos)

				if !yield(id) {
					return
				}

				word &= word - 1
			}
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

func (b ArchetypeMask) IsEmpty() bool {
	return b[0] == 0 && b[1] == 0
}

func (b ArchetypeMask) Count() int {
	return bits.OnesCount64(b[0]) +
		bits.OnesCount64(b[1])
}

// Matches returns true if the mask contains all bits from include AND none from exclude.
func (b ArchetypeMask) Matches(include, exclude ArchetypeMask) bool {
	// Inclusion check - must contain ALL bits from includeMask
	if (b[0]&include[0]) != include[0] ||
		(b[1]&include[1]) != include[1] {
		return false
	}

	// Exclusion check - must contain NONE of the bits from excludeMask
	if (b[0]&exclude[0]) != 0 ||
		(b[1]&exclude[1]) != 0 {
		return false
	}

	return true
}
