package core

import (
	"unsafe"

	"github.com/kjkrol/uid"
)

type PageIdx uint32 // Index of the page in Memo.Pages slice
// PageSlot is a type alias for the index within a page.
// Using uint32 ensures alignment and supports >255 entities per page.
type PageSlot uint32

//------------------------------------------------------------------------------
//                          Memo (Memory Manager)
//------------------------------------------------------------------------------

type Memo struct {
	Pages    []Page
	Layout   PageLayout
	Len      uint32  // Total number of active entities across all pages (O(1) tracking).
	Reserved PageIdx // Prevents truncation of the currently processed page during chunked iteration/creation.
}

func (b *Memo) Init(compInfos []ComponentInfo) {
	b.Layout = CalculateLayout(compInfos)

	b.Pages = make([]Page, 0, 16)
	b.Len = 0

	b.addPage()
}

func (b *Memo) AllocSlot() (*Page, PageIdx, PageSlot) {
	lastIdx := PageIdx(len(b.Pages) - 1)
	page := &b.Pages[lastIdx]

	if page.Len >= PageSlot(b.Layout.PageCap) {
		b.addPage()
		lastIdx++
		page = &b.Pages[lastIdx]
	}

	slot := page.Len
	page.Len++
	b.Len++

	return page, lastIdx, slot
}

func (b *Memo) GetPage(idx PageIdx) *Page {
	return &b.Pages[idx]
}

func (b *Memo) addPage() {
	data := make([]byte, b.Layout.PageBytes)
	b.Pages = append(b.Pages, Page{
		data: data,
		Ptr:  unsafe.Pointer(&data[0]),
		Len:  0,
	})
}

func (b *Memo) AddPages(n int) {
	pageBytes := b.Layout.PageBytes
	bigBlock := make([]byte, uintptr(n)*pageBytes)
	for i := range n {
		offset := uintptr(i) * pageBytes
		b.Pages = append(b.Pages, Page{
			data: bigBlock[offset : offset+pageBytes : offset+pageBytes],
			Ptr:  unsafe.Pointer(&bigBlock[offset]),
			Len:  0,
		})
	}
}

// ResolveTail returns the index and pointer of the last page that actually contains data.
// It performs a "sanity check" by truncating any empty trailing pages from the Pages slice.
func (b *Memo) ResolveTail() (PageIdx, *Page) {
	lastIdx := len(b.Pages) - 1
	floor := int(b.Reserved)

	for lastIdx > floor && b.Pages[lastIdx].Len == 0 {
		b.Pages = b.Pages[:lastIdx]
		lastIdx--
	}

	tailIdx := lastIdx
	for tailIdx > 0 && b.Pages[tailIdx].Len == 0 {
		tailIdx--
	}

	return PageIdx(tailIdx), &b.Pages[tailIdx]
}

func (b *Memo) Clear() {
	for i := range b.Pages {
		clear(b.Pages[i].data)
		b.Pages[i].Len = 0
	}

	b.Pages = b.Pages[:0]
	b.Len = 0

	b.addPage()
}

//------------------------------------------------------------------------------
//                          page
//------------------------------------------------------------------------------

type Page struct {
	// ┌────────────────────────────────────────────────────────────┐
	// │ data []byte                                                │
	// ├────────────────────────────────────────────────────────────┤
	// │ Entity Column │ CompA Column │ CompB Column │ ...          │
	// └────────────────────────────────────────────────────────────┘
	data []byte
	Ptr  unsafe.Pointer
	Len  PageSlot
}

//------------------------------------------------------------------------------
//                          CalculateLayout
//------------------------------------------------------------------------------

type PageLayout struct {
	PageCap   uint32
	PageBytes uintptr
	Offsets   []uintptr
}

// CalculateLayout computes the optimal memory layout for a page.
func CalculateLayout(compInfos []ComponentInfo) PageLayout {

	totalStride := unsafe.Sizeof(uid.UID64(0))
	for _, info := range compInfos {
		totalStride += info.Size
	}

	capacity := uintptr(L1DataCacheSize) / totalStride
	if capacity == 0 {
		capacity = 1
	}

	for capacity >= 1 {
		offsets := make([]uintptr, len(compInfos)+1)
		currentOffset := uintptr(0)

		entityAlign := unsafe.Alignof(uid.UID64(0))
		currentOffset = alignUp(currentOffset, entityAlign)
		offsets[0] = currentOffset
		currentOffset += unsafe.Sizeof(uid.UID64(0)) * capacity

		for i, info := range compInfos {
			currentOffset = alignUp(currentOffset, info.Align)
			offsets[i+1] = currentOffset
			currentOffset += info.Size * capacity
		}

		if capacity == 1 || currentOffset <= L1DataCacheSize {
			return PageLayout{
				PageCap:   uint32(capacity),
				PageBytes: currentOffset,
				Offsets:   offsets,
			}
		}

		capacity--
	}

	panic("unreachable")
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

func (p *Page) GetPointer(offset uintptr, itemSize uintptr, slot PageSlot) unsafe.Pointer {
	return unsafe.Add(p.Ptr, offset+(uintptr(slot)*itemSize))
}
