package core

import "unsafe"

const PageSize = L1DataCacheSize

type PageIdx uint32 // Index of the page in Memo.Pages slice
// PageRow is a type alias for the index within a page.
// Using uint32 ensures alignment and supports >255 entities per page.
type PageRow uint32

//------------------------------------------------------------------------------
//                          Memo (Memory Manager)
//------------------------------------------------------------------------------

type Memo struct {
	// Pages holds pointers to all allocated pages.
	// Using a slice allows O(1) access by PageIdx, which is crucial for EntityLinkStore.
	Pages  []*page
	Layout PageLayout
	Len    uint32 // Global entity count (optional, but useful)
}

// Computes how components are packed into fixed size (by default 16KB) pages
// and initialize memory pages (Pages)
func (b *Memo) Init(compInfos []ComponentInfo) {
	b.Layout = CalculateLayout(compInfos)

	// Pre-allocate slice capacity to avoid frequent resizing at start
	b.Pages = make([]*page, 0, 16)
	b.Len = 0

	b.addPage()
}

// AllocSlot allocates space for a new entity.
// It returns:
// 1. *page  -> Pointer for immediate data writing (fastest access)
// 2. uint32  -> PageIdx (to store in EntityLinkStore)
// 3. PageRow -> Row index within the page (to store in EntityLinkStore)
func (b *Memo) AllocSlot() (*page, PageIdx, PageRow) {
	lastIdx := PageIdx(len(b.Pages) - 1)
	c := b.Pages[lastIdx]

	// If the current page is full, create a new one
	if c.Len >= PageRow(b.Layout.PageCap) {
		b.addPage()
		lastIdx++
		c = b.Pages[lastIdx]
	}

	row := c.Len
	c.Len++
	b.Len++

	return c, lastIdx, row
}

// AllocBatch reserves up to 'count' slots in a SINGLE contiguous page.
// It returns:
// 1. *page   -> Pointer for data writing
// 2. PageIdx -> Index of the page
// 3. PageRow -> Starting row in the page
// 4. int      -> How many slots were actually allocated (could be less than count)
func (b *Memo) AllocBatch(count int) (*page, PageIdx, PageRow, int) {
	lastIdx := PageIdx(len(b.Pages) - 1)
	c := b.Pages[lastIdx]

	available := int(b.Layout.PageCap) - int(c.Len)

	if available == 0 {
		b.addPage()
		lastIdx++
		c = b.Pages[lastIdx]
		available = int(b.Layout.PageCap)
	}

	allocated := min(count, available)

	startRow := c.Len

	c.Len += PageRow(allocated)
	b.Len += uint32(allocated)

	return c, lastIdx, startRow, allocated
}

// GetPage provides O(1) access to a page by its index.
// Used when moving/removing entities based on EntityLinkStore data.
func (b *Memo) GetPage(idx PageIdx) *page {
	// In production, you might skip bounds check if you trust LinkStore
	return b.Pages[idx]
}

func (b *Memo) addPage() {
	data := make([]byte, PageSize)

	newPage := &page{
		data: data,
		Ptr:  unsafe.Pointer(&data[0]),
		Len:  0,
	}

	b.Pages = append(b.Pages, newPage)
}

// ResolveTail returns the index and pointer of the last page that actually contains data.
// It performs a "sanity check" by truncating any empty trailing pages from the Pages slice.
func (b *Memo) ResolveTail() (PageIdx, *page) {
	lastIdx := len(b.Pages) - 1

	// Loop backwards to remove empty trailing pages.
	// We keep at least one page (index 0) even if it's empty.
	for lastIdx > 0 && b.Pages[lastIdx].Len == 0 {
		// Optional: you could clear(b.Pages[lastIdx].data) here if
		// you want to be ultra-aggressive with GC, but Clear() should handle it.
		b.Pages = b.Pages[:lastIdx]
		lastIdx--
	}

	return PageIdx(lastIdx), b.Pages[lastIdx]
}

func (b *Memo) Clear() {
	// 1. Zero out memory for GC safety
	for _, c := range b.Pages {
		clear(c.data)
		c.Len = 0
	}

	// 2. Reset the slice
	// We can keep the underlying array capacity to avoid re-allocations on restart
	b.Pages = b.Pages[:0]
	b.Len = 0

	// 3. Immediately add the first fresh page
	b.addPage()
}

//------------------------------------------------------------------------------
//                          page
//------------------------------------------------------------------------------

type page struct {
	data []byte
	Ptr  unsafe.Pointer
	Len  PageRow
}

//------------------------------------------------------------------------------
//                          CalculateLayout
//------------------------------------------------------------------------------

type PageLayout struct {
	PageCap uint32
	Offsets []uintptr
}

// CalculateLayout computes the optimal memory layout for a page.
func CalculateLayout(compInfos []ComponentInfo) PageLayout {

	totalStride := unsafe.Sizeof(Entity(0))
	for _, info := range compInfos {
		totalStride += info.Size
	}

	capacity := uintptr(PageSize) / totalStride
	if capacity == 0 {
		panic("Entity layout too large for a single memory page (16KB)")
	}

	for capacity > 0 {
		offsets := make([]uintptr, len(compInfos)+1)
		currentOffset := uintptr(0)
		fits := true

		// --- STEP A: Entity ID ---
		entityAlign := unsafe.Alignof(Entity(0))
		currentOffset = alignUp(currentOffset, entityAlign)
		offsets[0] = currentOffset
		currentOffset += unsafe.Sizeof(Entity(0)) * capacity

		// --- STEP B: Components ---
		for i, info := range compInfos {
			currentOffset = alignUp(currentOffset, info.Align)
			offsets[i+1] = currentOffset
			currentOffset += info.Size * capacity

			if currentOffset > PageSize {
				fits = false
				break
			}
		}

		if fits {
			return PageLayout{
				PageCap: uint32(capacity),
				Offsets: offsets,
			}
		}

		capacity--
	}

	panic("Components too large for PageSize")
}

func alignUp(ptr, align uintptr) uintptr {
	return (ptr + align - 1) & ^(align - 1)
}

// -----------------------------------------------------------------------------
// Low-Level Helpers
// -----------------------------------------------------------------------------

func copyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func zeroMemory(ptr unsafe.Pointer, size uintptr) {
	clear(unsafe.Slice((*byte)(ptr), size))
}
