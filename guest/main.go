package main

import (
	karmem "karmem.org/golang"
	"api/v1"
)

func main() {
	// Some useful initialization may be added here, e.g. configuration of logging
}

// TODO: reqLen is an unused param, remove it maybe?
//
//go:export ProcessRequest
func ProcessRequest(reqPtr uintptr, reqLen uint32) uint64 {
	reader := karmem.NewReader(getBytes(reqPtr))
	req := new(v1.DataRequest)
	req.ReadAsRoot(reader)

	resp := doProcessRequest(req)

	writer := karmem.NewWriter(4 * 1024)
	if _, err := resp.WriteAsRoot(writer); err != nil {
		panic(err)
	}
	respBytes := writer.Bytes()
	respBytesLen := uint32(len(respBytes))
	ptrResp := Malloc(respBytesLen)
	respBuf := getBytes(ptrResp)
	copy(respBuf, respBytes)
	return packPtrAndSize(ptrResp, respBytesLen) // NOTE: It is the host's responsibility to free this memory!
}

func doProcessRequest(req *v1.DataRequest) *v1.DataResponse {
	result := make([]int32, 0)
	for _, number := range req.Numbers {
		if number > req.K {
			result = append(result, number)
		}
	}
	resp := v1.DataResponse{
		NumbersGreaterK: result,
	}
	return &resp
}
