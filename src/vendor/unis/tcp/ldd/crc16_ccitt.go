package ldd

//static unsigned short crc16_x1(const unsigned char d[], int len)
func ChecksumCCITT(d []byte) uint16 {
	b := 0
	crc := 0xFFFF
	l := len(d)
	for i := 0; i < l; i++ {
		for j := 0; j < 8; j++ {
			b = ((int(d[i]) << uint8(j)) & 0x80) ^ ((crc & 0x8000) >> 8)
			crc <<= 1
			if b != 0 {
				crc ^= 0x1021
			}
		}
	}
	return uint16(crc)
}
