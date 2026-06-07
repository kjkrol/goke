package core

import (
	"sort"
	"unsafe"
)

// sortLinkKey encodes (ArchId, PageIdx, PageRow) into a single uint64 so the
// monotonic detector can compare entries with a single CPU instruction instead
// of three branched comparisons. Field widths:
//   - ArchId  : 16 bits (high)   → ≤ 65 535 archetypes
//   - PageIdx : 16 bits (middle) → ≤ 65 535 pages per archetype
//   - PageRow : 32 bits (low)    → ≤ 4.29 billion rows per page
// If any field overflows its width, the function still returns a deterministic
// key but the monotonic optimization may produce false negatives (fallback to
// full sort) — never false positives, so correctness is preserved.
func sortLinkKey(l EntityArchLink) uint64 {
	return uint64(l.ArchId)<<48 | uint64(l.PageIdx)<<32 | uint64(l.PageRow)
}

// insertionSortByLink performs an in-place SoA insertion sort on the first n
// entries of cache. Used as a non-allocating, branch-friendly alternative to
// sort.Sort for small n where interface dispatch dominates wall time
// (empirically: faster than sort.Sort for n ≲ 24 on Apple M1 Max).
func insertionSortByLink(cache *FilterCache, n int) {
	for i := 1; i < n; i++ {
		// Snapshot the row at position i — we shift later entries to its right.
		keyEnt := cache.Entity[i]
		keyLink := cache.Link[i]
		keyMatched := cache.Matched[i]
		keyOrig := cache.OriginalIndex[i]
		keyK := sortLinkKey(keyLink)

		j := i - 1
		for j >= 0 && sortLinkKey(cache.Link[j]) > keyK {
			cache.Entity[j+1] = cache.Entity[j]
			cache.Link[j+1] = cache.Link[j]
			cache.Matched[j+1] = cache.Matched[j]
			cache.OriginalIndex[j+1] = cache.OriginalIndex[j]
			j--
		}
		cache.Entity[j+1] = keyEnt
		cache.Link[j+1] = keyLink
		cache.Matched[j+1] = keyMatched
		cache.OriginalIndex[j+1] = keyOrig
	}
}

// insertionSortThreshold marks the point above which sort.Sort beats inline
// insertion sort on the SoA cache. Below it the avoided interface dispatch
// and improved branch predictability pay off.
const insertionSortThreshold = 24

// FilterCache is a reusable scratchpad buffer for View.Filter operations.
// Holding it in a system field (or sync.Pool) eliminates allocations.
// All four primary slices are parallel — the i-th element of each describes
// the same entity (SoA layout, improving cache coherence during Link-based
// sort).
//
// SingleIdx is a fixed-size buffer used by the per-entity Filter path
// generated for thin views (View1, View2) — it lets those adapters yield
// 1-element Indices slices without heap allocations and without disturbing
// the OriginalIndex slice used by the chunked path.
type FilterCache struct {
	Entity        []Entity
	Link          []EntityArchLink
	Matched       []*MatchedArch
	OriginalIndex []int    // position in the input `selected` slice
	SingleIdx     [1]int   // scratch for the per-entity inline Filter (thin views)
}

// Reset clears the length of all buffers while preserving their capacity.
// Called automatically by the chunked Filter path on entry.
func (c *FilterCache) Reset() {
	c.Entity = c.Entity[:0]
	c.Link = c.Link[:0]
	c.Matched = c.Matched[:0]
	c.OriginalIndex = c.OriginalIndex[:0]
}

// Grow optionally pre-allocates capacity (e.g. at system startup) so the
// first few frames don't pay for repeated reallocation/copy as the cache
// grows to its working set size.
func (c *FilterCache) Grow(n int) {
	if cap(c.Entity) < n {
		c.Entity = make([]Entity, 0, n)
		c.Link = make([]EntityArchLink, 0, n)
		c.Matched = make([]*MatchedArch, 0, n)
		c.OriginalIndex = make([]int, 0, n)
	}
}

// filterCacheSort is a type alias used to sort the cache in SoA layout via
// sort.Interface. The conversion (*filterCacheSort)(cache) is a zero-cost
// pointer cast — both types share the same memory layout because
// filterCacheSort's extra method set does not change the struct fields.
type filterCacheSort FilterCache

func (s *filterCacheSort) Len() int { return len(s.Link) }

func (s *filterCacheSort) Less(i, j int) bool {
	return sortLinkKey(s.Link[i]) < sortLinkKey(s.Link[j])
}

func (s *filterCacheSort) Swap(i, j int) {
	s.Entity[i], s.Entity[j] = s.Entity[j], s.Entity[i]
	s.Link[i], s.Link[j] = s.Link[j], s.Link[i]
	s.Matched[i], s.Matched[j] = s.Matched[j], s.Matched[i]
	s.OriginalIndex[i], s.OriginalIndex[j] = s.OriginalIndex[j], s.OriginalIndex[i]
}

// FilterPage represents a contiguous run of entities sharing the same
// archetype page (not necessarily the entire page — see StartRow/Count).
// Passed to the visitor by WalkFilteredPages.
//
// Type-safe View*.Filter adapters wrap BasePtr+offsets with unsafe.Slice
// to give user code typed Go slices over native memory — therefore
// downstream user code sees these values as ordinary "pages", consistent
// with View.All().
type FilterPage struct {
	Matched  *MatchedArch
	BasePtr  unsafe.Pointer
	StartRow uintptr // first row of this page span within the physical page
	Count    int     // number of contiguous rows in the span
	Indices  []int   // sub-slice of cache.OriginalIndex aligned with the rows
}

// FilterPageVisitor receives a single FilterPage; return false to abort
// iteration (mirrors the iter.Seq yield semantics).
type FilterPageVisitor func(page FilterPage) bool

// WalkFilteredPages resolves `selected` entities against `view`, sorts them
// by physical location (ArchId, PageIdx, PageRow) and invokes `visit` once
// per maximal contiguous run within the same archetype page.
//
// Two fast paths shortcut the most common access patterns:
//   - Monotonic input  : if the resolve phase observes keys arriving in
//                        non-decreasing order, sort is skipped entirely
//                        (common when `selected` comes from a View.All scan
//                        or a spatial index in scan order).
//   - Small n          : for n ≤ insertionSortThreshold the cache is sorted
//                        with an inline insertion sort that avoids
//                        sort.Interface dispatch overhead.
//
// All transient state lives in `cache`, so repeated calls with a reused cache
// allocate nothing in the hot path. The visitor receives raw pointers; typed
// View*.Filter adapters wrap them with unsafe.Slice to recover Go slices over
// component columns.
//
// Note: thin views (1–2 stateful components) bypass this function entirely.
// Their generated adapters in View1.Filter / View2.Filter inline a naive
// per-entity loop because that avoids both the resolve materialization and
// the closure-call overhead of one yield per entity through this visitor.
// From three components upward the chunked algorithm wins on ordered inputs
// and the visitor cost is amortized over a whole page.
func WalkFilteredPages(view *View, selected []Entity, cache *FilterCache, visit FilterPageVisitor) {
	// STEP 1: Resolve — unpack links and matched archetypes into SoA cache.
	// While doing so, track whether keys arrive in non-decreasing order so we
	// can skip the sort entirely when callers pass already-ordered subsets.
	cache.Reset()
	store := &view.Reg.ArchetypeRegistry.EntityLinkStore

	monotonic := true
	var prevKey uint64

	for i, e := range selected {
		link, ok := store.Get(e)
		if !ok {
			continue
		}
		ma := view.GetMatchedArch(link.ArchId)
		if ma == nil {
			continue // entity does not match this view — drop early
		}
		cache.Entity = append(cache.Entity, e)
		cache.Link = append(cache.Link, link)
		cache.Matched = append(cache.Matched, ma)
		cache.OriginalIndex = append(cache.OriginalIndex, i)

		// Cheap monotonic check via packed key. Single comparison per entry.
		k := sortLinkKey(link)
		if monotonic && len(cache.Link) > 1 && k < prevKey {
			monotonic = false
		}
		prevKey = k
	}

	// STEP 2: Sort SoA — but only if we actually need to.
	n := len(cache.Link)
	if !monotonic {
		if n <= insertionSortThreshold {
			insertionSortByLink(cache, n)
		} else {
			sort.Sort((*filterCacheSort)(cache))
		}
	}

	// STEP 3: Iterate in contiguous runs (same Arch + Page, consecutive Row).
	for i := 0; i < n; {
		startLink := cache.Link[i]
		startMatched := cache.Matched[i]
		count := 1

		for j := i + 1; j < n; j++ {
			next := cache.Link[j]
			if next.ArchId == startLink.ArchId &&
				next.PageIdx == startLink.PageIdx &&
				next.PageRow == startLink.PageRow+PageRow(count) {
				count++
				continue
			}
			break
		}

		physPage := startMatched.Arch.Memory.Pages[startLink.PageIdx]
		if !visit(FilterPage{
			Matched:  startMatched,
			BasePtr:  physPage.Ptr,
			StartRow: uintptr(startLink.PageRow),
			Count:    count,
			Indices:  cache.OriginalIndex[i : i+count],
		}) {
			return
		}
		i += count
	}
}
