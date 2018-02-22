// +build !windows

package qhsec

/*
#cgo CFLAGS: -I/home/s/safe/include
#cgo LDFLAGS: -L/home/s/safe/lib -L/home/s/lib
#cgo LDFLAGS: /home/s/safe/lib/libnetproto.a /home/s/safe/lib/libcryptopp.a /home/s/safe/lib/libnppnacl.a /home/s/safe/lib/libqoslib.a -lcrypto -lz -lm -lstdc++
#cgo LDFLAGS: -Wl,-rpath=/home/s/safe/lib
#cgo LDFLAGS: -Wl,-rpath=/home/s/lib

#include <stdlib.h>
#include <netproto/include/qhsec.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Version
const (
	V6  = 6
	V10 = 10
	V11 = 11
	V14 = 14
)

// Client Option
const (
	OptCompressMethod           = "compress_method"
	OptSymmetricMethod          = "symmetric_method"
	OptSymmetricKeyNo           = "symmetric_key_no"
	OptAsymmetricMethod         = "asymmetric_method"
	OptAsymmetricKeyNo          = "asymmetric_key_no"
	OptResponseSymmetricKeyType = "response_symmetric_key_type"
	OptNetMethod                = "net_method"
)

// Symmetric Method
const (
	PLAIN = iota
	XOR
	IDEAECB
	AES128ECB
	DESECB
	AES128CBC
)

// Asymmetric Method
const (
	OpenSSLRSA  = 0
	SimpleRSA   = 1
	NaclEC      = 4
	CryptoppRSA = 5
	OpenSSLEDCH = 6
)

var (
	ErrInitNppConfig             = errors.New("initial NppConfig failed")
	ErrInitMessagePacker         = errors.New("initial MessagePacker handle")
	ErrInitMessageUnpacker       = errors.New("initial MessageUnpacker failed")
	ErrInvalidNppConfig          = errors.New("invalid NppConfig handle")
	ErrInvalidMessagePacker      = errors.New("invalid MessagePacker handle")
	ErrInvalidMessagePackerStage = errors.New("invalid MessagePacker handle stage")
	ErrInvalidMessageUnpacker    = errors.New("invalid MessageUnpacker handle")
	ErrInvalidOption             = errors.New("invalid option name")
	ErrPackData                  = errors.New("pack data failed")
	ErrUnpackData                = errors.New("unpack data failed")
)

type NppConfig struct {
	handler unsafe.Pointer
}

func NewNppConfig(symmetric_key_file, business, asymmetric_key_file string) (*NppConfig, error) {
	cSymmetric := C.CString(symmetric_key_file)
	cBusiness := C.CString(business)
	cAsymmetric := C.CString(asymmetric_key_file)
	defer C.free(unsafe.Pointer(cSymmetric))
	defer C.free(unsafe.Pointer(cBusiness))
	defer C.free(unsafe.Pointer(cAsymmetric))

	if h := C.qhsec_create_handler(cSymmetric, cBusiness, cAsymmetric); h != nil {
		return &NppConfig{handler: h}, nil
	}

	return &NppConfig{handler: nil}, ErrInitNppConfig
}

func (npp *NppConfig) AddSymmetricKey(symmetricMethod, keyNo int, key []byte) bool {
	if npp.handler == nil {
		return false
	}

	klen := len(key)
	if klen == 0 {
		return false
	}

	ret := C.qhsec_add_symmetric_key(npp.handler, C.int(symmetricMethod), C.int(keyNo),
		unsafe.Pointer(&key[0]), C.size_t(klen))

	return bool(ret)
}

func (npp *NppConfig) Close() {
	if npp.handler == nil {
		return
	}

	C.qhsec_destroy_handler(npp.handler)
	npp.handler = nil
}

type ServerUnpacker struct {
	nppConfig *NppConfig
	unpacker  unsafe.Pointer
}

func NewServerUnpacker(npp *NppConfig) (*ServerUnpacker, error) {
	if npp == nil || npp.handler == nil {
		return nil, ErrInvalidNppConfig
	}

	unpacker := C.qhsec_create_s_unpacker(npp.handler)
	if unpacker == nil {
		return nil, ErrInitMessageUnpacker
	}

	return &ServerUnpacker{nppConfig: npp, unpacker: unpacker}, nil
}

func (s *ServerUnpacker) Unpack(data []byte) ([]byte, error) {
	if s.unpacker == nil {
		return nil, ErrInvalidMessageUnpacker
	}

	dlen := len(data)
	if dlen == 0 {
		return make([]byte, dlen), nil
	}

	size := C.int(dlen)
	var l *C.int = &size
	str := C.qhsec_s_unpack(s.unpacker, unsafe.Pointer(&data[0]), l)
	if str == nil {
		return nil, ErrUnpackData
	}

	return C.GoBytes(str, size), nil
}

func (s *ServerUnpacker) GetSymmetricEncryptKey() ([]byte, error) {
	if s.unpacker == nil {
		return nil, ErrInvalidMessageUnpacker
	}

	size := C.int(0)
	var l *C.int = &size
	str := C.qhsec_s_get_symmetric_key(s.unpacker, l)
	if str == nil {
		return nil, ErrUnpackData
	}

	defer C.free(str)
	return C.GoBytes(str, size), nil
}

func (s *ServerUnpacker) Pack(data []byte) ([]byte, error) {
	if s.unpacker == nil {
		return nil, ErrInvalidMessageUnpacker
	}

	dlen := len(data)
	if dlen == 0 {
		return make([]byte, dlen), nil
	}

	size := C.int(dlen)
	var l *C.int = &size
	str := C.qhsec_s_pack(s.unpacker, unsafe.Pointer(&data[0]), l)
	if str == nil {
		return nil, ErrPackData
	}

	defer C.free(str)
	return C.GoBytes(str, size), nil
}

func (s *ServerUnpacker) NppConfig() (*NppConfig, bool) {
	if s.nppConfig == nil {
		return nil, false
	}

	return s.nppConfig, true
}

func (s *ServerUnpacker) Close() {
	if s.unpacker == nil {
		return
	}

	C.qhsec_destroy_s_unpacker(s.unpacker)

	s.unpacker = nil
	s.nppConfig = nil
}

type ClientPacker struct {
	nppConfig *NppConfig
	packer    unsafe.Pointer

	unpacked bool
}

func NewClientPacker(npp *NppConfig, protoVersion int) (*ClientPacker, error) {
	if npp == nil || npp.handler == nil {
		return nil, ErrInvalidNppConfig
	}

	packer := C.qhsec_create_c_packer(npp.handler, C.int(protoVersion))
	if packer == nil {
		return nil, ErrInitMessagePacker
	}

	return &ClientPacker{nppConfig: npp, packer: packer, unpacked: false}, nil
}

func (c *ClientPacker) SetOption(option string, value int) error {
	if c.packer == nil {
		return ErrInvalidMessagePacker
	}

	cOption := C.CString(option)
	defer C.free(unsafe.Pointer(cOption))

	status := C.qhsec_c_packer_set_option(c.packer, cOption, C.int(value))
	if status == 1 {
		return ErrInvalidOption
	}

	if status == 2 {
		return ErrInvalidMessagePacker
	}

	return nil
}

func (c *ClientPacker) Pack(data []byte) ([]byte, error) {
	if c.packer == nil {
		return nil, ErrInvalidMessagePacker
	}

	dlen := len(data)
	if dlen == 0 {
		return make([]byte, dlen), nil
	}

	size := C.int(dlen)
	var l *C.int = &size
	cStr := C.qhsec_c_pack(c.packer, unsafe.Pointer(&data[0]), l)
	if cStr == nil {
		return nil, ErrPackData
	}

	return C.GoBytes(cStr, size), nil
}

func (c *ClientPacker) Unpack(data []byte) ([]byte, error) {
	if c.packer == nil {
		return nil, ErrInvalidMessagePacker
	}

	dlen := len(data)
	if dlen == 0 {
		return make([]byte, dlen), nil
	}

	size := C.int(dlen)
	var l *C.int = &size
	str := C.qhsec_c_unpack(c.packer, unsafe.Pointer(&data[0]), l)
	if str == nil {
		return nil, ErrUnpackData
	}

	c.unpacked = true

	defer C.free(str)
	return C.GoBytes(str, size), nil
}

type SymmetricKey struct {
	ExpiredTimestamp int64
	Key              []byte
	Method           int
	KeyNo            int
}

func (c *ClientPacker) SessionSymmetricKey() (*SymmetricKey, error) {
	if c.nppConfig == nil {
		return nil, ErrInvalidNppConfig
	}

	if !c.unpacked {
		return nil, ErrInvalidMessagePackerStage
	}

	sk := C.qhsec_get_session_key(c.nppConfig.handler)
	return &SymmetricKey{
		ExpiredTimestamp: int64(sk.expired_timestamp),
		Key: C.GoBytes(unsafe.Pointer(sk.symm_key),
			C.int(sk.symm_key_len)),
		Method: int(sk.symm_key_type),
		KeyNo:  int(sk.symm_key_id),
	}, nil
}

func (c *ClientPacker) NppConfig() (*NppConfig, bool) {
	if c.nppConfig == nil {
		return nil, false
	}

	return c.nppConfig, true
}

func (c *ClientPacker) Close() {
	if c.packer == nil {
		return
	}

	C.qhsec_destroy_c_packer(c.packer)
	c.packer = nil
	c.nppConfig = nil
}

func GetNetMethod(data []byte) int {
	dlen := len(data)
	if dlen == 0 {
		return -1
	}

	size := C.int(len(data))
	method := C.qhsec_get_net_method(unsafe.Pointer(&data[0]), size)
	return int(method)
}
