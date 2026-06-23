package chunk

// hasCacheSetConflict reports whether any two column start offsets map to the
// same L1D cache set. Conflicting columns thrash each other even when total
// data fits in L1D, causing super-linear slowdown as component count grows.
func hasCacheSetConflict(offsets []uintptr) bool {
	for i := range offsets {
		si := int(offsets[i]/64) % L1DataCacheSets
		for j := i + 1; j < len(offsets); j++ {
			if si == int(offsets[j]/64)%L1DataCacheSets {
				return true
			}
		}
	}
	return false
}
