package v1

import (
	karmem "karmem.org/golang"
	"unsafe"
)

var _ unsafe.Pointer

var _Null = make([]byte, 32)
var _NullReader = karmem.NewReader(_Null)

type (
	PacketIdentifier uint64
)

const (
	PacketIdentifierDataRequest  = 15680009911858927649
	PacketIdentifierDataResponse = 3179384322344759246
)

type DataRequest struct {
	S    string
	F64s []float64
	I32  int32
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
	size := uint(32)
	if offset == 0 {
		offset, err = writer.Alloc(size)
		if err != nil {
			return 0, err
		}
	}
	__SSize := uint(1 * len(x.S))
	__SOffset, err := writer.Alloc(__SSize)
	if err != nil {
		return 0, err
	}
	writer.Write4At(offset+0, uint32(__SOffset))
	writer.Write4At(offset+0+4, uint32(__SSize))
	writer.Write4At(offset+0+4+4, 1)
	__SSlice := [3]uint{*(*uint)(unsafe.Pointer(&x.S)), __SSize, __SSize}
	writer.WriteAt(__SOffset, *(*[]byte)(unsafe.Pointer(&__SSlice)))
	__F64sSize := uint(8 * len(x.F64s))
	__F64sOffset, err := writer.Alloc(__F64sSize)
	if err != nil {
		return 0, err
	}
	writer.Write4At(offset+12, uint32(__F64sOffset))
	writer.Write4At(offset+12+4, uint32(__F64sSize))
	writer.Write4At(offset+12+4+4, 8)
	__F64sSlice := *(*[3]uint)(unsafe.Pointer(&x.F64s))
	__F64sSlice[1] = __F64sSize
	__F64sSlice[2] = __F64sSize
	writer.WriteAt(__F64sOffset, *(*[]byte)(unsafe.Pointer(&__F64sSlice)))
	__I32Offset := offset + 24
	writer.Write4At(__I32Offset, *(*uint32)(unsafe.Pointer(&x.I32)))

	return offset, nil
}

func (x *DataRequest) ReadAsRoot(reader *karmem.Reader) {
	x.Read(NewDataRequestViewer(reader, 0), reader)
}

func (x *DataRequest) Read(viewer *DataRequestViewer, reader *karmem.Reader) {
	__SString := viewer.S(reader)
	if x.S != __SString {
		__SStringCopy := make([]byte, len(__SString))
		copy(__SStringCopy, __SString)
		x.S = *(*string)(unsafe.Pointer(&__SStringCopy))
	}
	__F64sSlice := viewer.F64s(reader)
	__F64sLen := len(__F64sSlice)
	if __F64sLen > cap(x.F64s) {
		x.F64s = append(x.F64s, make([]float64, __F64sLen-len(x.F64s))...)
	}
	if __F64sLen > len(x.F64s) {
		x.F64s = x.F64s[:__F64sLen]
	}
	copy(x.F64s, __F64sSlice)
	x.F64s = x.F64s[:__F64sLen]
	x.I32 = viewer.I32()
}

type DataResponse struct {
	S    string
	F64s []float64
	I32  int32
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
	size := uint(32)
	if offset == 0 {
		offset, err = writer.Alloc(size)
		if err != nil {
			return 0, err
		}
	}
	__SSize := uint(1 * len(x.S))
	__SOffset, err := writer.Alloc(__SSize)
	if err != nil {
		return 0, err
	}
	writer.Write4At(offset+0, uint32(__SOffset))
	writer.Write4At(offset+0+4, uint32(__SSize))
	writer.Write4At(offset+0+4+4, 1)
	__SSlice := [3]uint{*(*uint)(unsafe.Pointer(&x.S)), __SSize, __SSize}
	writer.WriteAt(__SOffset, *(*[]byte)(unsafe.Pointer(&__SSlice)))
	__F64sSize := uint(8 * len(x.F64s))
	__F64sOffset, err := writer.Alloc(__F64sSize)
	if err != nil {
		return 0, err
	}
	writer.Write4At(offset+12, uint32(__F64sOffset))
	writer.Write4At(offset+12+4, uint32(__F64sSize))
	writer.Write4At(offset+12+4+4, 8)
	__F64sSlice := *(*[3]uint)(unsafe.Pointer(&x.F64s))
	__F64sSlice[1] = __F64sSize
	__F64sSlice[2] = __F64sSize
	writer.WriteAt(__F64sOffset, *(*[]byte)(unsafe.Pointer(&__F64sSlice)))
	__I32Offset := offset + 24
	writer.Write4At(__I32Offset, *(*uint32)(unsafe.Pointer(&x.I32)))

	return offset, nil
}

func (x *DataResponse) ReadAsRoot(reader *karmem.Reader) {
	x.Read(NewDataResponseViewer(reader, 0), reader)
}

func (x *DataResponse) Read(viewer *DataResponseViewer, reader *karmem.Reader) {
	__SString := viewer.S(reader)
	if x.S != __SString {
		__SStringCopy := make([]byte, len(__SString))
		copy(__SStringCopy, __SString)
		x.S = *(*string)(unsafe.Pointer(&__SStringCopy))
	}
	__F64sSlice := viewer.F64s(reader)
	__F64sLen := len(__F64sSlice)
	if __F64sLen > cap(x.F64s) {
		x.F64s = append(x.F64s, make([]float64, __F64sLen-len(x.F64s))...)
	}
	if __F64sLen > len(x.F64s) {
		x.F64s = x.F64s[:__F64sLen]
	}
	copy(x.F64s, __F64sSlice)
	x.F64s = x.F64s[:__F64sLen]
	x.I32 = viewer.I32()
}

type DataRequestViewer struct {
	_data [32]byte
}

func NewDataRequestViewer(reader *karmem.Reader, offset uint32) (v *DataRequestViewer) {
	if !reader.IsValidOffset(offset, 32) {
		return (*DataRequestViewer)(unsafe.Pointer(&_Null))
	}
	v = (*DataRequestViewer)(unsafe.Add(reader.Pointer, offset))
	return v
}

func (x *DataRequestViewer) size() uint32 {
	return 32
}
func (x *DataRequestViewer) S(reader *karmem.Reader) (v string) {
	offset := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 0))
	size := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 0+4))
	if !reader.IsValidOffset(offset, size) {
		return ""
	}
	length := uintptr(size / 1)
	slice := [3]uintptr{
		uintptr(unsafe.Add(reader.Pointer, offset)), length, length,
	}
	return *(*string)(unsafe.Pointer(&slice))
}
func (x *DataRequestViewer) F64s(reader *karmem.Reader) (v []float64) {
	offset := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 12))
	size := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 12+4))
	if !reader.IsValidOffset(offset, size) {
		return []float64{}
	}
	length := uintptr(size / 8)
	slice := [3]uintptr{
		uintptr(unsafe.Add(reader.Pointer, offset)), length, length,
	}
	return *(*[]float64)(unsafe.Pointer(&slice))
}
func (x *DataRequestViewer) I32() (v int32) {
	return *(*int32)(unsafe.Add(unsafe.Pointer(&x._data), 24))
}

type DataResponseViewer struct {
	_data [32]byte
}

func NewDataResponseViewer(reader *karmem.Reader, offset uint32) (v *DataResponseViewer) {
	if !reader.IsValidOffset(offset, 32) {
		return (*DataResponseViewer)(unsafe.Pointer(&_Null))
	}
	v = (*DataResponseViewer)(unsafe.Add(reader.Pointer, offset))
	return v
}

func (x *DataResponseViewer) size() uint32 {
	return 32
}
func (x *DataResponseViewer) S(reader *karmem.Reader) (v string) {
	offset := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 0))
	size := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 0+4))
	if !reader.IsValidOffset(offset, size) {
		return ""
	}
	length := uintptr(size / 1)
	slice := [3]uintptr{
		uintptr(unsafe.Add(reader.Pointer, offset)), length, length,
	}
	return *(*string)(unsafe.Pointer(&slice))
}
func (x *DataResponseViewer) F64s(reader *karmem.Reader) (v []float64) {
	offset := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 12))
	size := *(*uint32)(unsafe.Add(unsafe.Pointer(&x._data), 12+4))
	if !reader.IsValidOffset(offset, size) {
		return []float64{}
	}
	length := uintptr(size / 8)
	slice := [3]uintptr{
		uintptr(unsafe.Add(reader.Pointer, offset)), length, length,
	}
	return *(*[]float64)(unsafe.Pointer(&slice))
}
func (x *DataResponseViewer) I32() (v int32) {
	return *(*int32)(unsafe.Add(unsafe.Pointer(&x._data), 24))
}
