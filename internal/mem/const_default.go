//go:build !arm64

package mem

// L1DataCacheSize: 32 KB for typical x86-64 (Intel/AMD desktop/laptop).
const L1DataCacheSize = 32 * 1024

// L1DataCacheSets: 32 KB / (8-way × 64 B line) = 64 sets.
const L1DataCacheSets = 64
