// +build !windows

package iconv

import (
	cv "github.com/zieckey/iconv"
)

func GBK2UTF8(gbk string) (utf8str string, err error) {
	return Convert(gbk, "utf-8", "gbk")
}

func UTF8ToGBK(utf8 string) (string, error) {
	return Convert(utf8, "gbk", "utf-8")
}

func UTF16ToUTF8(utf16 string) (string, error) {
	return Convert(utf16, "utf-8", "utf-16")
}

func Convert(s, tocode, fromcode string) (string, error) {
	cd, err := cv.Open(tocode, fromcode)
	if err != nil {
		return "", err
	}
	defer cd.Close()

	return cd.ConvString(s), nil
}
