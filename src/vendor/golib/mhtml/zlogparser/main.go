package main

import (
	"bufio"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/zieckey/goini"
	"golib/mhtml"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

/*
这个程序是用来解析老版本照妖镜日志中mhtml数据，看看解析程序是否有问题。
*/

func main() {
	pattern := "./*"
	if len(os.Args) == 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "-help" {
			fmt.Printf("usage : %v <the file pattern>\n", os.Args[0])
			return
		}
		pattern = os.Args[1]
	}

	if len(os.Args) > 2 {
		fmt.Printf("usage : %v \"the-patern\"\n", os.Args[0])
		return
	}

	//fmt.Printf("argc=%v pattern=%v [%v %v %v]\n", len(os.Args), pattern, os.Args[0], os.Args[1])
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("error happened : %v\n", err.Error())
		return
	}
	var index = 0
	for _, f := range files {
		r, err := OpenFile(f)
		if err != nil {
			fmt.Printf("read file error : %v\n", err.Error())
			continue
		}

		fmt.Printf("Begin to parse file : %v =============================>\n", f)
		for {
			index++
			fmt.Printf("%d\t", index)
			line, err := r.ReadString('\n')
			if err == io.EOF || err != nil {
				break
			}
			m := mhtml.New()
			mime, err := GetMHtmlFromLog(line)
			if err != nil {
				fmt.Printf("GetMHtmlFromLog error: %v\nline=\n[%v]", err.Error(), line)
				continue
			}
			err = m.Parse(mime)
			m.DebugPrint()
			if len(m.IframeURLs) == 0 {
				ioutil.WriteFile(
					fmt.Sprintf("log/%v-line.txt", index),
					[]byte(fmt.Sprintf("filename=%v url=[%v] Title=[%v] HtmlOriginal ==============================>\n%v\n\n\nHtmlUtf8 ==============================>\n%v",
						f, m.ContentLocation, m.Title, m.HtmlOriginal, m.HtmlUtf8)),
					0644)
				ioutil.WriteFile(
					fmt.Sprintf("log/%v-mime.txt", index),
					[]byte(fmt.Sprintf("filename=%v mhtml ==============================>\n%v", f, string(mime))),
					0644)
				ioutil.WriteFile(
					fmt.Sprintf("log/%v-original-log.txt", index),
					[]byte(fmt.Sprintf("%v", string(line))),
					0644)
			}
		}
	}
}

func OpenFile(name string) (*bufio.Reader, error) {
	fr, err := os.Open(name)
	if err != nil {
		fmt.Printf("read file error: %v\n", err.Error())
		return nil, err
	}

	if strings.HasSuffix(name, ".gz") {
		gr, err := gzip.NewReader(fr)
		if err != nil {
			return nil, err
		}
		return bufio.NewReader(gr), nil
	}

	return bufio.NewReader(fr), nil
}

func GetMHtmlFromLog(logxml string) ([]byte, error) {
	ini := goini.New()
	logxml = strings.Trim(logxml, "<>")
	err := ini.Parse([]byte(logxml), "><", ":")
	if err != nil {
		fmt.Printf("parse xml log error : %v\n", err.Error())
		return nil, err
	}

	//	req, ok := ini.Get("req")
	//	if !ok {
	//		fmt.Printf("cannot found 'req'\n")
	//		return nil, errors.New("cannot found 'req'")
	//	}
	//
	//	ini.Reset()
	//	err = ini.Parse([]byte(req), "&", "=")
	//	if err != nil {
	//		fmt.Printf("parse mhtml error : %v\n", err.Error())
	//		return nil, err
	//	}
	mhtmlenc, ok := ini.Get("mhtml")
	if !ok {
		fmt.Printf("cannot found 'mhtml'\n")
		return nil, errors.New("cannot found 'mhtml'")
	}

	mhtml, err := base64.StdEncoding.DecodeString(mhtmlenc)
	if err != nil {
		fmt.Printf("'mhtml' base64 decode error : %v\n", err.Error())
		return nil, err
	}
	ioutil.WriteFile("mime.txt", mhtml, 0644)

	return mhtml, nil
}

//func GetMHtmlFromLog(logxml string) ([]byte, error) {
//	ini := goini.New()
//	logxml = strings.Trim(logxml, "<>")
//	err := ini.Parse([]byte(logxml), "><", ":")
//	if err != nil {
//		fmt.Printf("parse xml log error : %v\n", err.Error())
//		return nil, err
//	}
//
//	req, ok := ini.Get("req")
//	if !ok {
//		fmt.Printf("cannot found 'req'\n")
//		return nil, errors.New("cannot found 'req'")
//	}
//
//	ini.Reset()
//	err = ini.Parse([]byte(req), "&", "=")
//	if err != nil {
//		fmt.Printf("parse mhtml error : %v\n", err.Error())
//		return nil, err
//	}
//	mhtmlenc, ok := ini.Get("mhtml")
//	if !ok {
//		fmt.Printf("cannot found 'mhtml'\n")
//		return nil, errors.New("cannot found 'mhtml'")
//	}
//
//	mhtml, err := base64.StdEncoding.DecodeString(mhtmlenc)
//	if err != nil {
//		fmt.Printf("'mhtml' base64 decode error : %v\n", err.Error())
//		return nil, err
//	}
//	ioutil.WriteFile("mime.txt", mhtml, 0644)
//
//	return mhtml, nil
//}
