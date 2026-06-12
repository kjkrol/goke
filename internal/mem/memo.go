package mem

import (
	"unsafe"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/core"
)

type PageIdx uint32
type PageSlot uint32

type Memo struct {
	Pages    []Page
	Layout   PageLayout
	Len      uint32
	Reserved PageIdx
}

func (b *Memo) Init(compInfos []core.ComponentInfo) {
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

type Page struct {
	data []byte
	Ptr  unsafe.Pointer
	Len  PageSlot
}

func (p *Page) GetPointer(offset uintptr, itemSize uintptr, slot PageSlot) unsafe.Pointer {
	return unsafe.Add(p.Ptr, offset+(uintptr(slot)*itemSize))
}

type PageLayout struct {
	PageCap   uint32
	PageBytes uintptr
	Offsets   []uintptr
}

func CalculateLayout(compInfos []core.ComponentInfo) PageLayout {
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

func CopyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func ZeroMemory(ptr unsafe.Pointer, size uintptr) {
	clear(unsafe.Slice((*byte)(ptr), size))
}
