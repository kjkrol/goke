package mem

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
)

// TestChunkLayout_CacheSetConflictsDuringIteration simulates the iteration
// access pattern for the 10-component archetype used in bench/view_all_test.go
// and reports every entity index at which two or more columns land in the same
// L1D cache set.
//
// This test checks the full per-entity access pattern to expose any stride-based
// cache set conflicts that remain after ChunkLayout.Init (which only eliminates
// start-offset conflicts). Use it to understand the residual thrashing profile
// of a given archetype layout.
func TestChunkLayout_CacheSetConflictsDuringIteration(t *testing.T) {
	// Mirror of bench/shared_test.go component types:
	//   Pos, Vel, Acc  — {X, Y float32}  8 bytes, align 4
	//   T04, T05, T06  — {V float32}     4 bytes, align 4
	//   T07..T10       — {V float64}     8 bytes, align 8
	compMetas := []comp.Meta{
		{ID: 1, Size: 8, Align: 4},  // Pos
		{ID: 2, Size: 8, Align: 4},  // Vel
		{ID: 3, Size: 8, Align: 4},  // Acc
		{ID: 4, Size: 4, Align: 4},  // T04
		{ID: 5, Size: 4, Align: 4},  // T05
		{ID: 6, Size: 4, Align: 4},  // T06
		{ID: 7, Size: 8, Align: 8},  // T07
		{ID: 8, Size: 8, Align: 8},  // T08
		{ID: 9, Size: 8, Align: 8},  // T09
		{ID: 10, Size: 8, Align: 8}, // T10
	}
	names := []string{"uid", "Pos", "Vel", "Acc", "T04", "T05", "T06", "T07", "T08", "T09", "T10"}

	var layout ChunkLayout
	layout.Init(compMetas)

	// strides: entity column first, then component columns
	strides := make([]uintptr, len(compMetas)+1)
	strides[0] = unsafe.Sizeof(uid.UID64(0))
	for i, m := range compMetas {
		strides[i+1] = m.Size
	}

	t.Logf("ChunkCap=%d  ChunkBytes=%d  L1DataCacheSize=%d  L1DataCacheSets=%d",
		layout.ChunkCap, layout.ChunkBytes, L1DataCacheSize, L1DataCacheSets)

	// Assert: Init must place every column start in a distinct L1D cache set.
	startSets := make(map[int]int, len(layout.Offsets)) // cache set → column index
	for j, off := range layout.Offsets {
		set := int((off / 64) % uintptr(L1DataCacheSets))
		if prev, ok := startSets[set]; ok {
			t.Errorf("start-offset conflict: col[%d](%s) and col[%d](%s) both map to cache set %d",
				prev, names[prev], j, names[j], set)
		}
		startSets[set] = j
	}

	// Informational: stride-based conflicts during iteration are structural
	// and cannot be eliminated by Init — log them without failing.
	t.Log("\nColumn layout (start offsets and cache sets):")
	for j, off := range layout.Offsets {
		startSet := (off / 64) % uintptr(L1DataCacheSets)
		t.Logf("  col[%2d] %-4s  stride=%d  offset=%6d  start_set=%3d",
			j, names[j], strides[j], off, startSet)
	}

	type collision struct {
		entity   int
		colA     int
		colB     int
		cacheSet int
	}
	var collisions []collision

	n := len(layout.Offsets)
	pairCount := make([][]int, n)
	for i := range pairCount {
		pairCount[i] = make([]int, n)
	}

	for i := 0; i < int(layout.ChunkCap); i++ {
		sets := make([]int, n)
		for j := range layout.Offsets {
			sets[j] = int((layout.Offsets[j]+uintptr(i)*strides[j])/64) % L1DataCacheSets
		}
		for j := range sets {
			for k := j + 1; k < len(sets); k++ {
				if sets[j] == sets[k] {
					collisions = append(collisions, collision{i, j, k, sets[j]})
					pairCount[j][k]++
				}
			}
		}
	}

	if len(collisions) == 0 {
		t.Log("\nNo stride-based cache set collisions during full iteration.")
		return
	}

	t.Logf("\nStride-based collision events: %d across %d entities", len(collisions), layout.ChunkCap)

	type pair struct{ a, b, count int }
	var pairs []pair
	for j := range pairCount {
		for k := j + 1; k < len(pairCount[j]); k++ {
			if pairCount[j][k] > 0 {
				pairs = append(pairs, pair{j, k, pairCount[j][k]})
			}
		}
	}
	for i := 1; i < len(pairs); i++ {
		for j := i; j > 0 && pairs[j].count > pairs[j-1].count; j-- {
			pairs[j], pairs[j-1] = pairs[j-1], pairs[j]
		}
	}
	t.Log("\nCollisions per column pair (top 10):")
	for _, p := range pairs[:min(10, len(pairs))] {
		t.Logf("  col[%d](%s) vs col[%d](%s): %d", p.a, names[p.a], p.b, names[p.b], p.count)
	}

	t.Log("\nFirst 20 collision events:")
	var sb strings.Builder
	for _, c := range collisions[:min(20, len(collisions))] {
		fmt.Fprintf(&sb, "  entity=%4d  col[%d](%s) vs col[%d](%s)  set=%d\n",
			c.entity, c.colA, names[c.colA], c.colB, names[c.colB], c.cacheSet)
	}
	t.Log(sb.String())
}
