//go:build !arm64

package core

// DefaultChunkSize: 32KB for x86/AMD64 and others.
// Standard x86 processors usually have smaller L1 Data Cache (32KB - 48KB).
// 32KB is a safe bet to leave room for stack and other data.
const L1DataCacheSize = 16 * 1024
