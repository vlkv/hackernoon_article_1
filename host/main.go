package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"api/v1"

	"github.com/bytecodealliance/wasmtime-go"
	karmem "karmem.org/golang"
)

func main() {
	instance, store, mem, err := newWasmInstance("../guest/guest.wasm")
	if err != nil {
		panic(err)
	}

	main := instance.GetFunc(store, "_start")
	malloc := instance.GetFunc(store, "Malloc")
	free := instance.GetFunc(store, "Free")
	processRequest := instance.GetFunc(store, "ProcessRequest")

	_, err = main.Call(store)
	if err != nil {
		panic(err)
	}

	req := v1.DataRequest{
		Numbers: []int32{10, 43, 13, 24, 56, 16},
		K: 42,
	}
	fmt.Printf("Numbers=%v, K=%v\n", req.Numbers, req.K)
	writer := karmem.NewWriter(20 * 1024)
	if _, err := req.WriteAsRoot(writer); err != nil {
		panic(err)
	}
	reqBytes := writer.Bytes()
	reqBytesLen := int32(len(reqBytes))

	ptrReq, err := malloc.Call(store, reqBytesLen)
	if err != nil {
		panic(err)
	}

	int32PtrReq := ptrReq.(int32)
	copy(
		mem.UnsafeData(store)[int32PtrReq : int32PtrReq+reqBytesLen],
		reqBytes,
	)

	respPtrLen, err := processRequest.Call(store, int32PtrReq, reqBytesLen)
	if err != nil {
		panic(err)
	}

	free.Call(store, int32PtrReq)

	respPtr, respLen := unpackPtrAndSize(uint64(respPtrLen.(int64)))

	resp := new(v1.DataResponse)
	respBytes := mem.UnsafeData(store)[int32(respPtr) : int32(respPtr)+int32(respLen)]
	resp.ReadAsRoot(karmem.NewReader(respBytes))

	fmt.Printf("NumbersGreaterK=%v\n", resp.NumbersGreaterK)

	free.Call(store, respPtr) // This memory was allocated on the guest side, we free it on the host side here

	runtime.KeepAlive(mem)
}

func unpackPtrAndSize(ptrSize uint64) (ptr uintptr, size uint32) {
	ptr = uintptr(ptrSize >> 32)
	size = uint32(ptrSize)
	return
}

func newWasmInstance(wasmBinaryPath string) (*wasmtime.Instance, *wasmtime.Store, *wasmtime.Memory, error) {
	engine := wasmtime.NewEngine()

	wasmBytes, err := readWasmBytes(engine, wasmBinaryPath)
	if err != nil {
		return nil, nil, nil, err
	}

	module, err := wasmtime.NewModule(engine, wasmBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	linker := wasmtime.NewLinker(engine)
	err = linker.DefineWasi()
	if err != nil {
		return nil, nil, nil, err
	}

	store := wasmtime.NewStore(engine)
	wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.SetArgv([]string{wasmBinaryPath})
	wasiConfig.InheritStdin()
	wasiConfig.InheritStderr()
	wasiConfig.InheritStdout()
	store.SetWasi(wasiConfig)

	instance, err := linker.Instantiate(store, module)
	if err != nil {
		return nil, nil, nil, err
	}

	var mem *wasmtime.Memory
	if mem = instance.GetExport(store, "memory").Memory(); mem == nil {
		panic("couldn't import memory")
	}

	return instance, store, mem, nil
}

func readWasmBytes(engine *wasmtime.Engine, path string) ([]byte, error) {
	wasmFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	wasm, err := io.ReadAll(wasmFile)
	if err != nil {
		return nil, err
	}

	if err := wasmtime.ModuleValidate(engine, wasm); err != nil {
		return nil, err
	}

	return wasm, nil
}
