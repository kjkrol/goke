package comp

import (
	"iter"
	"math/bits"
)

// Mask encodes a set of component IDs as an array of MaskSize uint64 words.
// Supports up to MaxComponents component types.
type Mask [MaskSize]uint64

func (Mask) Build(b *Blueprint) Mask {
	var mask Mask
	for _, info := range b.CompInfos {
		mask = mask.Set(info.ID)
	}
	for _, id := range b.TagIDs {
		mask = mask.Set(id)
	}
	return mask
}

func NewMask(componentIDs ...ID) Mask {
	var mask Mask
	for _, id := range componentIDs {
		mask = mask.Set(id)
	}
	return mask
}

func (b Mask) Set(bit ID) Mask {
	word, pos := bit/64, bit%64
	if word < MaskSize {
		b[word] |= 1 << pos
	}
	return b
}

func (b Mask) Clear(bit ID) Mask {
	word, pos := bit/64, bit%64
	if word < MaskSize {
		b[word] &= ^(1 << pos)
	}
	return b
}

func (b Mask) Equals(other Mask) bool {
	return b == other
}

func (b Mask) Contains(subMask Mask) bool {
	for i := range MaskSize {
		if (b[i] & subMask[i]) != subMask[i] {
			return false
		}
	}
	return true
}

func (b Mask) AllSet() iter.Seq[ID] {
	return func(yield func(ID) bool) {
		for wordIdx := range MaskSize {
			word := b[wordIdx]
			for word != 0 {
				bitPos := bits.TrailingZeros64(word)
				id := ID(wordIdx*64 + bitPos)
				if !yield(id) {
					return
				}
				word &= word - 1
			}
		}
	}
}

func (b Mask) IsSet(bit ID) bool {
	word, pos := bit/64, bit%64
	if word >= MaskSize {
		return false
	}
	return (b[word] & (1 << pos)) != 0
}

func (b Mask) IsEmpty() bool {
	return b[0] == 0 && b[1] == 0
}

func (b Mask) Count() int {
	return bits.OnesCount64(b[0]) +
		bits.OnesCount64(b[1])
}

// Matches returns true if the mask contains all bits from include AND none from exclude.
func (b Mask) Matches(include, exclude Mask) bool {
	if (b[0]&include[0]) != include[0] ||
		(b[1]&include[1]) != include[1] {
		return false
	}
	if (b[0]&exclude[0]) != 0 ||
		(b[1]&exclude[1]) != 0 {
		return false
	}
	return true
}
