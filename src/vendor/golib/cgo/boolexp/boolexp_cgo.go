package boolexp

/*
#cgo CXXFLAGS:  -I. -I.. -I../ -I ../../include
#cgo CXXFLAGS:  -I/home/s/safe/include/protobuf3
#cgo CXXFLAGS:  -I/home/s/safe/include
#cgo CXXFLAGS:  -I/home/s/include
#cgo CXXFLAGS:  -I/usr/local/include
#cgo LDFLAGS:   /home/s/safe/lib/libboolexp.a
#cgo LDFLAGS:   -lpcre -lprotobuf
#cgo LDFLAGS:   -L /home/s/safe/protobuf3/lib/ -L/usr/local/lib
#cgo LDFLAGS:   -L/home/s/safe/lib
#cgo LDFLAGS:   -L/home/s/lib
#cgo LDFLAGS:   -Wl,-rpath=/home/s/safe/protobuf3/lib/
#cgo LDFLAGS:   -Wl,-rpath=/home/s/safe/lib
#cgo LDFLAGS:   -Wl,-rpath=/home/s/lib

#include "boolexp_cgo.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"github.com/golang/protobuf/proto"
	"unsafe"
)

type BoolExp struct {
	be C.boolexp_t
}

func NewBoolExp(input string) (*BoolExp, error) {
	cInput := C.CString(input)
	b := &BoolExp{}

	b.be = C.boolexp_create(cInput)
	C.free(unsafe.Pointer(cInput))

	if b.be == nil {
		return nil, errors.New("c boolexp create failed.")
	}
	return b, nil
}

func (b *BoolExp) Process(req *Request) (*Response, error) {
	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("data is nill")
	}
	var result = unsafe.Pointer(nil)
	resultLen := C.int(0)
	ret := C.boolexp_process(b.be, unsafe.Pointer(&data[0]), C.int(len(data)), &result, &resultLen)
	if bool(ret) == false {
		return nil, errors.New("c boolexp process failed.")
	}
	defer C.free(unsafe.Pointer(result))

	r := C.GoBytes(result, resultLen)
	resp := &Response{}
	err = proto.Unmarshal(r, resp)
	return resp, err
}

func (b *BoolExp) Close() {
	if b.be != nil {
		C.boolexp_destory(b.be)
		b.be = nil
	}
}
