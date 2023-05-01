package main

import (
	"unsafe"
)

// Custom mem functions are needed because tinygo
// promises nothing about stability of exported `malloc`` and `free`.
// More about mem management: https://wazero.io/languages/tinygo/#memory

var allocatedMemory = map[uintptr][]byte{}

func ptrToBytes(ptr uintptr) []byte {
	return allocatedMemory[ptr]
}

//go:export Malloc
func Malloc(size uint32) uintptr {
	buf := make([]byte, size)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	allocatedMemory[unsafePtr] = buf
	return unsafePtr
}

//go:export Free
func Free(ptr uintptr) {
	delete(allocatedMemory, ptr)
}

func joinPtrSize(ptr uintptr, size uint32) (ptrSize uint64) {
	return uint64(ptr)<<uint64(32) | uint64(size)
}
