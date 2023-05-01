package main

import (
	"unsafe"
)

// TinyGo adds a pair of functions `malloc` and `free` to wasm module automatically, but stil
// we need custom mem functions because
// 1) TinyGo promises nothing about stability of exported `malloc` and `free`.
// 2) We have to call Malloc for storing the result somehow, it is unclear how to call standard `malloc`.

var allocatedBytes = map[uintptr][]byte{}

//go:export Malloc
func Malloc(size uint32) uintptr {
	buf := make([]byte, size)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	allocatedBytes[unsafePtr] = buf
	return unsafePtr
}

//go:export Free
func Free(ptr uintptr) {
	delete(allocatedBytes, ptr)
}

func getBytes(ptr uintptr) []byte {
	return allocatedBytes[ptr]
}

func packPtrAndSize(ptr uintptr, size uint32) (ptrSize uint64) {
	return uint64(ptr)<<uint64(32) | uint64(size)
}
