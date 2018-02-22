package symc

/*
#cgo CXXFLAGS:  -I.
#cgo CXXFLAGS:  -g
#cgo CXXFLAGS:  -I/home/s/safe/symc_cgo
#cgo CXXFLAGS:  -I/home/s/safe/include
#cgo CXXFLAGS:  -I/home/s/include
#cgo CXXFLAGS:  -I/usr/local/include
#cgo LDFLAGS:   -I/home/s/safe/symc_cgo/libsymc_cgo.a
#cgo LDFLAGS:   -lqoslib -lsymc
#cgo LDFLAGS:   -L/home/s/safe/lib
#cgo LDFLAGS:   -L/home/s/lib
#cgo LDFLAGS:   -L/usr/local/lib
#cgo LDFLAGS:   -Wl,-rpath=/home/s/safe/lib
#cgo LDFLAGS:   -Wl,-rpath=/home/s/lib

#include "symc.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

type Symc struct {
	s C.symc_t
	p *Pool
}

func NewSymc(vbucketConf, vbucketName string) (s *Symc, err error) {
	s = &Symc{}
	cVbucketConf := C.CString(vbucketConf)
	defer C.free(unsafe.Pointer(cVbucketConf))
	cVbucketName := C.CString(vbucketName)
	defer C.free(unsafe.Pointer(cVbucketName))

	s.s = C.symc_create(cVbucketConf, cVbucketName)
	if s.s == nil {
		return nil, errors.New("create symc object error")
	}
	return s, nil
}

func (s *Symc) Set(kv map[string]string) error {
	if s.s == nil {
		return errors.New("symc nil")
	}
	keys := make([](*_Ctype_char), 0)
	vals := make([](*_Ctype_char), 0)

	for k, v := range kv {
		if len(v) == 0 {
			return errors.New("set kv is error, some value is null")
		}
		ck := C.CString(k)
		defer C.free(unsafe.Pointer(ck))
		cv := C.CString(v)
		defer C.free(unsafe.Pointer(cv))
		keys = append(keys, ck)
		vals = append(vals, cv)
	}
	ret := C.symc_set(s.s, &keys[0], C.int(len(keys)), &vals[0], C.int(len(vals)))
	if ret == false {
		return errors.New("set kv error")
	}
	return nil
}

func (s *Symc) get(kv map[string]string) (map[string]string, error) {
	if s.s == nil {
		return kv, errors.New("symc nil")
	}
	if len(kv) == 0 {
		return kv, errors.New("kv map is null")
	}

	keys := make([](*_Ctype_char), 0)
	vals := make([](*_Ctype_char), len(kv))

	for k, _ := range kv {
		ck := C.CString(k)
		defer C.free(unsafe.Pointer(ck))
		keys = append(keys, ck)
	}
	cValLen := C.int(len(kv))

	ret := C.symc_get(s.s, &keys[0], C.int(len(keys)), &vals[0], &cValLen)
	if ret == false {
		return nil, errors.New("get kv error")
	}

	for i := 0; i < len(vals); i++ {
		k := C.GoString(keys[i])
		kv[k] = C.GoString(vals[i])
		C.free(unsafe.Pointer(vals[i]))
		vals[i] = nil
	}
	return kv, nil
}

func (s *Symc) cget(kv map[string]string) (map[string]string, error) {
	if s.s == nil {
		return kv, errors.New("symc nil")
	}
	if len(kv) == 0 {
		return kv, errors.New("kv map is null")
	}

	keys := make([](*_Ctype_char), 0)

	for k, _ := range kv {
		ck := C.CString(k)
		defer C.free(unsafe.Pointer(ck))
		keys = append(keys, ck)
	}

	ret := C.symc_get_result(s.s, &keys[0], C.int(len(keys)))
	if ret == nil {
		return nil, errors.New("get kv error")
	}

	for C.symc_result_start(ret); !C.symc_result_is_end(ret); C.symc_result_current_next(ret) {
		k := C.GoString(C.symc_result_current_key(ret))
		v := C.GoString(C.symc_result_current_val(ret))
		kv[k] = v
	}
	C.symc_result_destory(ret)
	return kv, nil
}

func (s *Symc) Get(kv map[string]string) (map[string]string, error) {
	return s.cget(kv)
}

func (s *Symc) destory() {
	if s.s != nil {

		C.symc_destory(s.s)
		s.s = nil
	}
}
func (s *Symc) Close() {
	if s.s != nil {
		if s.p == nil {
			s.destory()
		} else {
			s.p.returnSymc(s)
		}
	}
}
