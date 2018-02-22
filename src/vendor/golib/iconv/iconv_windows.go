package iconv

import (
	//	"golang.org/x/text/encoding"
	//	"golang.org/x/text/encoding/charmap"
	//	"golang.org/x/text/encoding/japanese"
	//	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	//	"golang.org/x/text/encoding/traditionalchinese"
	//	"golang.org/x/text/encoding/unicode"
	//	"golang.org/x/text/transform"
)

func GBK2UTF8(gbkstr string) (utf8str string, err error) {
	t := simplifiedchinese.GBK.NewDecoder()
	gbkBuf := []byte(gbkstr)
	dst := make([]byte, len(gbkBuf)*3)
	nDst, _, err := t.Transform(dst, gbkBuf, true)
	if err != nil {
		return "", err
	}
	return string(dst[:nDst]), nil
}
