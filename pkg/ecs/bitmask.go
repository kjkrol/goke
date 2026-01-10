package ecs

import "math/bits"

type Bitmask []uint64

func (b Bitmask) Set(bit ComponentID) Bitmask {
	word, pos := bit/64, bit%64
	for len(b) <= int(word) {
		b = append(b, 0)
	}
	b[word] |= (1 << pos)
	return b
}

func (b Bitmask) Matches(required Bitmask) bool {
	if len(b) < len(required) {
		return false
	}
	for i := range required {
		if (b[i] & required[i]) != required[i] {
			return false
		}
	}
	return true
}

func (b Bitmask) ForEachSet(fn func(id ComponentID)) {
	for wordIdx, word := range b {
		if word == 0 {
			continue
		}

		for word != 0 {
			bitPos := bits.TrailingZeros64(word)

			id := ComponentID(wordIdx*64 + bitPos)
			fn(id)

			word &= ^(1 << bitPos)
		}
	}
}

func (b Bitmask) Clear(bit ComponentID) Bitmask {
	word, pos := bit/64, bit%64
	if len(b) <= int(word) {
		return b
	}
	b[word] &= ^(1 << pos)
	return b
}

func (b Bitmask) IsSet(bit ComponentID) bool {
	word, pos := bit/64, bit%64
	if len(b) <= int(word) {
		return false
	}
	return (b[word] & (1 << pos)) != 0
}
