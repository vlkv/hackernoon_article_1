package v1

import (
	karmem "karmem.org/golang"
	"unsafe"
)

var _ unsafe.Pointer

var _Null = make([]byte, 24)
var _NullReader = karmem.NewReader(_Null)

type (
	PacketIdentifier uint64
)

const (
	PacketIdentifierDataRequest  = 15680009911858927649
	PacketIdentifierDataResponse = 3179384322344759246
)

type DataRequest struct {
	Numbers []int32
	K       int32
}

func NewDataRequest() DataRequest {
	return DataRequest{}
}

func (x *DataRequest) PacketIdentifier() PacketIdentifier {
	return PacketIdentifierDataRequest
}

func (x *DataRequest) Reset() {
	x.Read((*DataRequestViewer)(unsafe.Pointer(&_Null)), _NullReader)
}

func (x *DataRequest) WriteAsRoot(writer *karmem.Writer) (offset uint, err error) {
	return x.Write(writer, 0)
}

func (x *DataRequest) Write(writer *karmem.Writer, start uint) (offset uint, err error) {
	offset = start
	size := uint(24)
	if offset == 0 {
		offset, err = writer.Alloc(size)
		if err != nil {
			return 0, err
		}
	}
	__NumbersSize := uint(4 * len(x.Numbers))
	__NumbersOffset, err := writer.Alloc(__NumbersSize)
	if err != nil {
		return 0, err
	}
	writer.Write4At(offset+0, uint32(__NumbersOffset))
	writer.Write4At(offset+0+4, uint32(__NumbersSize))
	writer.Write4At(offset+0+4+4, 4)
	__NumbersSlice := *(*[3]uint)(unsafe.Pointer(&x.Numbers))
	__NumbersSlice[1] = __NumbersSize
	__NumbersSlice[2] = __NumbersSize
	writer.WriteAt(__NumbersOffset, *(*[]byte)(unsafe.Pointer(&__NumbersSlice)))
	__KOffset := offset + 12
	writer.Write4At(__KOffset, *(*uint32)(unsafe.Pointer(&x.K)))

	return offset, nil
}

func (x *DataRequest) ReadAsRoot(reader *karmem.Reader) {
	x.Read(NewDataRequestViewer(reader, 0), reader)
}

func (x *DataRequest) Read(viewer *DataRequestViewer, reader *karmem.Reader) {
	__NumbersSlice := viewer.Numbers(reader)
	__NumbersLen := len(__NumbersSlice)
	if __NumbersLen > cap(x.Numbers) {
		x.Numbers = append(x.Numbers, make([]int32, __NumbersLen-len(x.Numbers))...)
	}
	if __NumbersLen > len(x.Numbers) {
		x.Numbers = x.Numbers[:__NumbersLen]
	}
	copy(x.Numbers, __NumbersSlice)
	x.Numbers = x.Numbers[:__NumbersLen]
	x.K = viewer.K()
}

type DataResponse struct {
	NumbersGreaterK []int32
}

func NewDataResponse() DataResponse {
	return DataResponse{}
}

func (x *DataResponse) PacketIdentifier() PacketIdentifier {
	return PacketIdentifierDataResponse
}

func (x *DataResponse) Reset() {
	x.Read((*DataResponseViewer)(unsafe.Pointer(&_Null)), _NullReader)
}

func (x *DataResponse) WriteAsRoot(writer *karmem.Writer) (offset uint, err error) {
	return x.Write(writer, 0)
}

func (x *DataResponse) Write(writer *karmem.Writer, start uint) (offset uint, err error) {
	offset = start
	size := uint(16)
	if offset == 0 {
		offset, err = writer.Alloc(size)
		if err != nil {
			return 0, err
		}
	}
	__NumbersGreaterKSize := uint(4 * len(x.NumbersGreaterK))
	__NumbersGreaterKOffset, err := writer.Alloc(__NumbersGreaterKSize)
	if err != nil {
		return 0, err
	}
	writer.Write4At(offset+0, uint32(__NumbersGreaterKOffset))
	writer.Write4At(offset+0+4, uint32(__NumbersGreaterKSize))
	writer.Write4At(offset+0+4+4, 4)
	__NumbersGreaterKSlice := *(*[3]uint)(unsafe.Pointer(&x.NumbersGreaterK))
	__NumbersGreaterKSlice[1] = __NumbersGreaterKSize
	__NumbersGreaterKSlice[2] = __NumbersGreaterKSize
	writer.WriteAt(__NumbersGreaterKOffset, *(*[]byte)(unsafe.Pointer(&__NumbersGreaterKSlice)))

	return offset, nil
}

func (x *DataResponse) ReadAsRoot(reader *karmem.Reader) {
	x.Read(NewDataResponseViewer(reader, 0), reader)
}

func (x *DataResponse) Read(viewer *DataResponseViewer, reader *karmem.Reader) {
	__NumbersGreaterKSlice := viewer.NumbersGreaterK(reader)
	__NumbersGreaterKLen := len(__NumbersGreaterKSlice)
	if __NumbersGreaterKLen > cap(x.NumbersGreaterK) {
		x.NumbersGreaterK = append(x.NumbersGreaterK, make([]int32, __NumbersGreaterKLen-len(x.NumbersGreaterK))...)
	}
	if __NumbersGreaterKLen > len(x.NumbersGreaterK) {
		x.NumbersGreaterK = x.NumbersGreaterK[:__NumbersGreaterKLen]
	}
	copy(x.NumbersGreaterK, __NumbersGreaterKSlice)
	x.NumbersGreaterK = x.NumbersGreaterK[:__NumbersGreaterKLen]
}

type DataRequestViewer struct {
	_data [24]byte
}

func NewDataRequestViewer(reader *karmem.Reader, offset uint32) (v *DataRequestViewer) {
	if !reader.IsValidOffset(offset, 24) {
		return (*DataRequestViewer)(unsafe.Pointer(&_Null))
	}
	v = (*DataRequestViewer)(unsafe.Add(reader.Pointer, offset))
	return v
}

func (x *DataRequestViewer) size() uint32 {
	return 24
}
func (x *DataRequestViewer) Numbers(reader *karmem.Reader) (v []int32) {
	offset := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 0))
	size := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 0+4))
	if !reader.IsValidOffset(offset, size) {
		return []int32{}
	}
	length := uintptr(size / 4)
	slice := [3]uintptr{
		uintptr(unsafe.Add(reader.Pointer, offset)), length, length,
	}
	return *(*[]int32)(unsafe.Pointer(&slice))
}
func (x *DataRequestViewer) K() (v int32) {
	return *(*int32)(unsafe.Add(unsafe.Pointer(&x._data), 12))
}

type DataResponseViewer struct {
	_data [16]byte
}

func NewDataResponseViewer(reader *karmem.Reader, offset uint32) (v *DataResponseViewer) {
	if !reader.IsValidOffset(offset, 16) {
		return (*DataResponseViewer)(unsafe.Pointer(&_Null))
	}
	v = (*DataResponseViewer)(unsafe.Add(reader.Pointer, offset))
	return v
}

func (x *DataResponseViewer) size() uint32 {
	return 16
}
func (x *DataResponseViewer) NumbersGreaterK(reader *karmem.Reader) (v []int32) {
	offset := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 0))
	size := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 0+4))
	if !reader.IsValidOffset(offset, size) {
		return []int32{}
	}
	length := uintptr(size / 4)
	slice := [3]uintptr{
		uintptr(unsafe.Add(reader.Pointer, offset)), length, length,
	}
	return *(*[]int32)(unsafe.Pointer(&slice))
}
