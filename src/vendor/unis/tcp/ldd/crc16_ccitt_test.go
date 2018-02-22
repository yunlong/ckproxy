package ldd

import (
	"testing"
)

func TestXYZ(t *testing.T) {
	s := []byte("abc")
	println(ChecksumCCITT(s))
}
