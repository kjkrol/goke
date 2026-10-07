package comp

import (
	"iter"
	"math/bits"
)

// Mask encodes a set of component IDs as an array of MaskSize uint64 words.
// Supports up to MaxComponents component types.
type Mask [MaskSize]uint64

func NewMask(s *AccessSpec) Mask {
	var mask Mask
	for _, info := range s.CompInfos {
		mask = mask.Set(info.ID)
	}
	for _, id := range s.TagIDs {
		mask = mask.Set(id)
	}
	return mask
}

func (s Mask) Set(bit ID) Mask {
	word, pos := bit/64, bit%64
	if word < MaskSize {
		s[word] |= 1 << pos
	}
	return s
}

func (s Mask) Clear(bit ID) Mask {
	word, pos := bit/64, bit%64
	if word < MaskSize {
		s[word] &= ^(1 << pos)
	}
	return s
}

func (s Mask) Equals(other Mask) bool {
	return s == other
}

func (s Mask) Contains(subMask Mask) bool {
	for i := range MaskSize {
		if (s[i] & subMask[i]) != subMask[i] {
			return false
		}
	}
	return true
}

func (s Mask) AllSet() iter.Seq[ID] {
	return func(yield func(ID) bool) {
		for wordIdx := range MaskSize {
			word := s[wordIdx]
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

func (s Mask) IsSet(bit ID) bool {
	word, pos := bit/64, bit%64
	if word >= MaskSize {
		return false
	}
	return (s[word] & (1 << pos)) != 0
}

func (s Mask) IsEmpty() bool {
	for i := range MaskSize {
		if s[i] != 0 {
			return false
		}
	}
	return true
}

func (s Mask) Count() int {
	n := 0
	for i := range MaskSize {
		n += bits.OnesCount64(s[i])
	}
	return n
}

// Matches returns true if the mask contains all bits from include AND none from exclude.
func (s Mask) Matches(include, exclude Mask) bool {
	for i := range MaskSize {
		if s[i]&include[i] != include[i] || s[i]&exclude[i] != 0 {
			return false
		}
	}
	return true
}
