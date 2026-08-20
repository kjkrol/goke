package chunk

import "unsafe"

func CopyMemory(dst, src unsafe.Pointer, size uintptr) {
	copy(unsafe.Slice((*byte)(dst), size), unsafe.Slice((*byte)(src), size))
}

func ZeroMemory(ptr unsafe.Pointer, size uintptr) {
	clear(unsafe.Slice((*byte)(ptr), size))
}

func ScannableBytes(n uintptr) []byte {
	wordSize := unsafe.Sizeof(unsafe.Pointer(nil))
	words := (n + wordSize - 1) / wordSize
	arr := make([]unsafe.Pointer, words)
	return unsafe.Slice((*byte)(unsafe.Pointer(&arr[0])), n)
}
