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
	store, instance, err := newWasmInstance("../guest/guest.wasm")
	if err != nil {
		panic(err)
	}

	var mem *wasmtime.Memory
	if mem = instance.GetExport(store, "memory").Memory(); mem == nil {
		panic("couln't import memory")
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
		mem.UnsafeData(store)[int32PtrReq:int32PtrReq+reqBytesLen],
		reqBytes,
	)

	respPtrLen, err := processRequest.Call(store, int32PtrReq, reqBytesLen)
	if err != nil {
		panic(err)
	}

	free.Call(store, int32PtrReq)

	respPtr, respLen := splitPtrSize(uint64(respPtrLen.(int64)))

	resp := new(v1.DataResponse)
	resp.ReadAsRoot(karmem.NewReader(mem.UnsafeData(store)[int32(respPtr) : int32(respPtr)+int32(respLen)]))

	fmt.Printf("NumbersGreaterK=%v\n", resp.NumbersGreaterK)

	free.Call(store, respPtr) // This memory was allocated on the guest side, we free it on the host side now

	runtime.KeepAlive(mem)
}

func splitPtrSize(ptrSize uint64) (ptr uintptr, size uint32) {
	ptr = uintptr(ptrSize >> 32)
	size = uint32(ptrSize)
	return
}

func newWasmInstance(wasmBinaryPath string) (*wasmtime.Store, *wasmtime.Instance, error) {
	engine := wasmtime.NewEngine()

	wasmBytes, err := readWasmBytes(engine, wasmBinaryPath)
	if err != nil {
		return nil, nil, err
	}

	module, err := wasmtime.NewModule(engine, wasmBytes)
	if err != nil {
		return nil, nil, err
	}

	linker := wasmtime.NewLinker(engine)
	err = linker.DefineWasi()
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}
	return store, instance, nil
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
