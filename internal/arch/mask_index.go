package arch

import "github.com/kjkrol/goke/internal/comp"

// -------------------------------------------------------------------------
// Configuration & Constants
// -------------------------------------------------------------------------

const (
	// HashSize defines the size of the MaskIndex lookup table.
	// Sized at 2×MaxID so the load factor stays at or below 50%,
	// minimising collision chains under linear probing.
	// CRITICAL: Must be a power of 2 for the HashMask optimisation to work.
	HashSize = uint64(MaxID) * 2
	// HashMask wraps the probing index via bitwise AND instead of modulo.
	HashMask = HashSize - 1
)

// MaskIndex is a hash map from comp.Mask to ID
// backed by open addressing with linear probing.
type MaskIndex struct {
	keys   [HashSize]comp.Mask
	values [HashSize]ID
}

func (m *MaskIndex) Reset() {
	clear(m.keys[:])
	clear(m.values[:])
}

// -------------------------------------------------------------------------
// Hash Function
// -------------------------------------------------------------------------

// hashMask mixes both 64-bit words of the mask using a prime multiplier
// and XOR-shifts to achieve good avalanche behaviour.
func hashMask(m comp.Mask) uint64 {
	h := m[0] ^ (m[1] * 0x517cc1b727220a95)
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	return h
}

// -------------------------------------------------------------------------
// Map Operations (Get / Upsert)
// -------------------------------------------------------------------------

// Get retrieves the ID for the given mask.
// Returns (0, false) if the mask is not present.
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

// Upsert inserts or updates the mapping from mask to id.
// Panics if id is NullID.
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
