package qutil

import (
	"bytes"
	"strings"
)

func Rot13(s string) string {
	rot13 := func(r rune) rune {
		const step = 13
		if r >= 'a' && r <= 'z' {
			return ((r - 'a' + step) % 26) + 'a'
		}
		if r >= 'A' && r <= 'Z' {
			return ((r - 'A' + step) % 26) + 'A'
		}
		return r
	}

	return strings.Map(rot13, s)
}

func BytesMerge(b ...[]byte) []byte {
	var t []byte
	buf := bytes.NewBuffer(t)
	for _, arg := range b {
		buf.Write(arg)
	}

	return buf.Bytes()
}

func GetHost(url string) string {
	skip, eidx := 0, len(url)
	if b := strings.Index(url, "://"); b != -1 {
		skip = b + len("://")
		if e := strings.Index(url[skip:], "/"); e != -1 {
			eidx = e + skip
		}
	}

	return strings.ToLower(strings.Trim(
		strings.TrimSpace(url[skip:eidx]), "."))
}
