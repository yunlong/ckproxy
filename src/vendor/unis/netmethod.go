package unis

import (
	"bytes"
	"encoding/binary"
)

func NetVersion(buf []byte) int {
	return byte2int(buf[1])
}

func byte2int(b byte) int {
	buf := []byte{b}
	bBuf := bytes.NewBuffer(buf)
	var i int8
	binary.Read(bBuf, binary.BigEndian, &i)
	return int(i)
}

func GetNetMethod(buf []byte) int {
	version := NetVersion(buf)
	netmethod := -1
	len := len(buf)
	if version == 6 {
		if len < 14 {
			return -1
		}
		netmethod = byte2int(buf[12])
	} else if version == 11 || version == 12 {
		if len < 12 {
			return -1
		}
		netmethod = byte2int(buf[10])
	} else if version == 1 || version == 4 || version == 3 {
		if len < 10 {
			return -1
		}
		netmethod = byte2int(buf[8])
	}

	return netmethod
}
