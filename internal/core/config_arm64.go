//go:build arm64

package core

// L1DataCacheSize: 128KB for ARM64.
// ARM64 architectures (Apple M1/M2/M3) have L1 Data Cache(64KB - 128KB)
const L1DataCacheSize = 96 * 1024
