package iconv

import (
	"encoding/base64"
	"testing"
)

func TestGBK2UTF8(t *testing.T) {
	gbk, _ := base64.StdEncoding.DecodeString("IC03ocEyNNChyrHIq8zs19S2r7eiu/XGvcyo")
	utf8, _ := base64.StdEncoding.DecodeString("IC03w5cyNOWwj+aXtuWFqOWkqeiHquWKqOWPkei0p+W5s+WPsA==")
	u, err := GBK2UTF8(string(gbk))
	if err != nil || u != string(utf8) {
		t.Errorf("GBK2UTF8 test failed")
	}
}

func d64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func TestUTF8ToGBK(t *testing.T) {
	gbk, _ := d64("IC03ocEyNNChyrHIq8zs19S2r7eiu/XGvcyo")
	utf8, _ := d64("IC03w5cyNOWwj+aXtuWFqOWkqeiHquWKqOWPkei0p+W5s+WPsA==")

	if g, err := UTF8ToGBK(string(utf8)); err != nil || g != string(gbk) {
		t.Fatal("UTF8ToGBK test failed")
	}
}

func TestUTF16ToUTF8(t *testing.T) {
	utf8, _ := d64("IC03w5cyNOWwj+aXtuWFqOWkqeiHquWKqOWPkei0p+W5s+WPsA==")
	utf16, _ := d64("/v8AIAAtADcA1wAyADRcD2X2UWhZKYHqUqhT0Y0nXnNT8A==")

	if u8, err := UTF16ToUTF8(string(utf16)); err != nil || u8 != string(utf8) {
		t.Fatal("UTF16ToUTF8 test failed")
	}
}
