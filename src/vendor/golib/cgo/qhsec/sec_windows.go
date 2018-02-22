// qhsec
package qhsec

import (
	"errors"
	"hash/crc32"
	"encoding/binary"
	"bytes"
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
	ErrInitNppConfig          = errors.New("initial NppConfig failed")
	ErrInitMessagePacker      = errors.New("initial MessagePacker handle")
	ErrInitMessageUnpacker    = errors.New("initial MessageUnpacker failed")
	ErrInvalidMessagePacker   = errors.New("invalid MessagePacker handle")
	ErrInvalidMessageUnpacker = errors.New("invalid MessageUnpacker handle")
	ErrInvalidOption          = errors.New("invalid option name")
	ErrPackData               = errors.New("pack data failed")
	ErrUnpackData             = errors.New("unpack data failed")
)

type NppConfig struct {
}

func NewNppConfig(symmetric_key_file, business, asymmetric_key_file string) (*NppConfig, error) {
	return &NppConfig{}, nil
}

func (npp *NppConfig) Close() {
}


type ServerUnpacker struct {
}

func NewServerUnpacker(npp *NppConfig) (*ServerUnpacker, error) {
	return &ServerUnpacker{}, nil
}

func (s *ServerUnpacker) Pack(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	
	// 0x0C, 0x0B, 0x00, 0x00, 0x00, 0x00, crc32, 0x00, 0x00
	buf.Write([]byte{0x0C, 0x0B, 0x00, 0x00, 0x00, 0x00})
	binary.Write(buf, binary.BigEndian, crc32.ChecksumIEEE(data))
	buf.Write([]byte{0x00, 0x00})
	
	buf.Write(data)
	return buf.Bytes(), nil
}

func (s *ServerUnpacker) Unpack(data []byte) ([]byte, error) {
	return data[12:], nil
}

func (s *ServerUnpacker) Close() {
}

type ClientPacker struct {
	ServerUnpacker
}

func NewClientPacker(npp *NppConfig, protoVersion int) (*ClientPacker, error) {
	return &ClientPacker{}, nil
}

func (c *ClientPacker) SetOption(option string, value int) error {
	return nil
}

func GetNetMethod(data []byte) int {
	return 0
}
