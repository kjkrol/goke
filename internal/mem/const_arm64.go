//go:build arm64

package mem

// L1DataCacheSize: 128KB for ARM64.
// ARM64 architectures (Apple M1/M2/M3) have L1 Data Cache(64KB - 128KB)
const L1DataCacheSize = 96 * 1024

// L1DataCacheSets: number of cache sets in L1D (128KB / (8-way × 64B line)).
const L1DataCacheSets = 256
