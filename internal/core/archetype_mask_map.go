package core

// -------------------------------------------------------------------------
// Configuration & Constants
// -------------------------------------------------------------------------

const (
	// HashSize defines the size of the lookup table.
	// CALCULATION: MaxArchetypeId * 2
	//
	// Why * 2?
	// Linear Probing Hash Maps perform best when the Load Factor is <= 50%.
	// By doubling the size, we ensure that even if we hit the MaxArchetypeId limit,
	// the map is only half full, minimizing collisions.
	//
	// CRITICAL: This value MUST result in a Power of 2 (e.g., 8192) for the
	// bitwise AND optimization (HashMask) to work as a modulo operator.
	HashSize = uint64(MaxArchetypeId) * 2
	// HashMask is used to wrap the index quickly (idx & HashMask).
	// Equivalent to (idx % HashSize) but much faster.
	HashMask = HashSize - 1
)

// -------------------------------------------------------------------------
// ArchetypeMaskMap Structure
// -------------------------------------------------------------------------
type ArchetypeMaskMap struct {
	// Keys are stored to verify matches (collision resolution).
	keys [HashSize]ArchetypeMask

	// Values store the Archetype ID.
	// value == 0 (NullArchetypeId) indicates an empty slot.
	values [HashSize]ArchetypeId
}

func (m *ArchetypeMaskMap) Reset() {
	clear(m.keys[:])
	clear(m.values[:])
}

// -------------------------------------------------------------------------
// Hash Function
// -------------------------------------------------------------------------

// hashMask mixes the bits of the 128-bit mask to produce a high-entropy 64-bit hash.
//
// Algorithm:
// It combines the two 64-bit words of the mask using a large prime multiplier
// and XOR-shifts to achieve the "Avalanche Effect" (changing 1 input bit changes many output bits).
func hashMask(m ArchetypeMask) uint64 {
	// Combine high and low parts with a prime mixer.
	h := m[0] ^ (m[1] * 0x517cc1b727220a95)

	// Mix bits thoroughly.
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33

	return h
}

// -------------------------------------------------------------------------
// Map Operations (Get / Put)
// -------------------------------------------------------------------------

// Get retrieves the ArchetypeId associated with the given mask.
// Returns:
//   - id: The found ArchetypeId (or 0 if not found).
//   - found: True if the mask exists in the map.
func (m *ArchetypeMaskMap) Get(mask ArchetypeMask) (ArchetypeId, bool) {
	idx := int(hashMask(mask) & HashMask)
	id := m.values[idx]

	if id == NullArchetypeId {
		return NullArchetypeId, false
	}

	if m.keys[idx] == mask {
		return id, true
	}

	idx = (idx + 1) & int(HashMask)

	for {
		id = m.values[idx]

		if id == NullArchetypeId {
			return NullArchetypeId, false
		}

		if m.keys[idx] == mask {
			return id, true
		}

		idx = (idx + 1) & int(HashMask)
	}
}

// Put inserts a mapping from mask to id.
// If the mask already exists, it updates the ID (idempotent).
//
// Panic:
//   - If id is NullArchetypeId (0).
//   - If the map is completely full (theoretical infinite loop, prevented by HashSize logic).
func (m *ArchetypeMaskMap) Put(mask ArchetypeMask, id ArchetypeId) {
	if id == NullArchetypeId {
		panic("ArchetypeMaskMap: Cannot store NullArchetypeId (0) as a value")
	}

	idx := hashMask(mask) & HashMask

	for {
		currentId := m.values[idx]

		// 1. Empty Slot -> Insert new mapping.
		if currentId == NullArchetypeId {
			m.keys[idx] = mask
			m.values[idx] = id
			return
		}

		// 2. Key Match -> Update existing (or simple return).
		if m.keys[idx] == mask {
			m.values[idx] = id
			return
		}

		// 3. Collision -> Linear Probe.
		idx = (idx + 1) & HashMask
	}
}
