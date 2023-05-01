package main

import (
	karmem "karmem.org/golang"
	"api/v1"
)

func main() {
	// TODO: Add some initialization here, like logging
}

//go:export ProcessRequest
func ProcessRequest(reqPtr uintptr, reqLen uint32) uint64 {
	reader := karmem.NewReader(ptrToBytes(reqPtr))
	req := new(v1.DataRequest)
	req.ReadAsRoot(reader)

	resp := doProcessRequest(req)

	writer := karmem.NewWriter(20 * 1024)
	if _, err := resp.WriteAsRoot(writer); err != nil {
		panic(err)
	}
	respBytes := writer.Bytes()
	respBytesLen := uint32(len(respBytes))
	ptrResp := Malloc(respBytesLen)
	respBuf := ptrToBytes(ptrResp)
	copy(respBuf, respBytes)
	return joinPtrSize(ptrResp, respBytesLen) // NOTE: That host should free this memory in the end
}

func doProcessRequest(req *v1.DataRequest) *v1.DataResponse {
	// TODO: Add some more realistic processing here
	resp := v1.DataResponse{
		S:    req.S + "bar",
		F64s: []float64{-1, -2, -3, -5},
		I32:  req.I32 + 1,
	}
	return &resp
}
