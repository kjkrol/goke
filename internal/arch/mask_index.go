package arch

import "github.com/kjkrol/goke/v2/internal/comp"

const (
	// HashSize is 2×MaxID (load factor ≤50%) and must stay a power of 2 —
	// HashMask relies on it to wrap probes with AND instead of modulo.
	HashSize = uint64(MaxID) * 2
	HashMask = HashSize - 1
)

// MaskIndex maps comp.Mask to ID via open addressing with linear probing.
type MaskIndex struct {
	keys   [HashSize]comp.Mask
	values [HashSize]ID
}

func (m *MaskIndex) Reset() {
	clear(m.keys[:])
	clear(m.values[:])
}

func hashMask(m comp.Mask) uint64 {
	h := m[0] ^ (m[1] * 0x517cc1b727220a95)
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	return h
}

func (m *MaskIndex) Get(mask comp.Mask) (ID, bool) {
	idx := int(hashMask(mask) & HashMask)
	id := m.values[idx]

	if id == NullID {
		return NullID, false
	}

	if m.keys[idx] == mask {
		return id, true
	}

	idx = (idx + 1) & int(HashMask)

	for {
		id = m.values[idx]
		if id == NullID {
			return NullID, false
		}
		if m.keys[idx] == mask {
			return id, true
		}
		idx = (idx + 1) & int(HashMask)
	}
}

// Upsert inserts or updates mask → id; panics on NullID.
func (m *MaskIndex) Upsert(mask comp.Mask, id ID) {
	if id == NullID {
		panic("MaskIndex: Cannot store NullID (0) as a value")
	}

	idx := hashMask(mask) & HashMask

	for {
		currentId := m.values[idx]

		if currentId == NullID {
			m.keys[idx] = mask
			m.values[idx] = id
			return
		}

		if m.keys[idx] == mask {
			m.values[idx] = id
			return
		}

		idx = (idx + 1) & HashMask
	}
}
