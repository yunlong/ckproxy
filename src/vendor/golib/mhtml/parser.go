package mhtml

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/opesun/goquery"
	"golib/iconv"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

var _ = ioutil.ReadAll
var _ = base64.StdEncoding

type MHtml struct {
	ContentLocation string   // The url of this page
	Title           string   // utf8 format
	IframeURLs      []string // The iframe urls
	HtmlUtf8        string   // The html content of UTF8 encoding
	HtmlOriginal    string   // The original html content

	boundary string
	charset  string
}

func New() *MHtml {
	return &MHtml{}
}

func (m *MHtml) DebugPrint() {
	fmt.Printf("url=[%v] title=[%v] charset=[%v] iframe=[%v]\n", m.ContentLocation, m.Title, m.charset, m.IframeURLs)
}

func (m *MHtml) Parse(mht []byte) error {
	if err := m.parseHTML(mht); err != nil {
		return err
	}

	if err := m.parseHTMLElement(); err != nil {
		return err
	}

	if err := m.convert2UTF8(); err != nil {
		return err
	}

	//m.DebugPrint()
	return nil
}

func (m *MHtml) parseHTMLElement() error {
	r := strings.NewReader(m.HtmlOriginal)

	x, err := goquery.Parse(r)
	if err != nil {
		return err
	}

	m.charset = x.Find("head meta").Attr("charset")
	m.Title = x.Find("head title").Text()
	m.IframeURLs = x.Find("iframe").Attrs("src")

	//	c := strings.Count(m.HtmlOriginal, "<iframe")
	//	if c != len(m.IframeURLs) {
	//		fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx <iframe> count =", c, " but we find ", len(m.IframeURLs))
	//	}
	//	fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	//	for _, u := range m.IframeURLs {
	//		fmt.Println(u)
	//	}
	//	fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	return nil
}

func (m *MHtml) parseHTML(mht []byte) error {
	br := bufio.NewReader(bytes.NewReader(mht)) // The buffer reader
	tr := textproto.NewReader(br)
	m.boundary = m.getBoundary(tr)
	if len(m.boundary) == 0 {
		return fmt.Errorf("Cannot get boundary")
	}

	mr := multipart.NewReader(br, m.boundary)
	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}

		//fmt.Println("\n\n================================================================================================================================================================\n\n")
		d := make([]byte, len(mht))
		n, err := part.Read(d)
		if err != nil && err != io.EOF {
			return err
		}
		d = d[:n]
		//fmt.Printf("filename=%v formname=%v n=%v err=%v content=\n", part.FileName(), part.FormName(), n, err)
		//ioutil.WriteFile(
		//	fmt.Sprintf("part-%v.txt", index),
		//	[]byte(fmt.Sprintf("filename=%v formname=%v n=%v err=%v Header=%v content=\n%v", part.FileName(), part.FormName(), n, err, part.Header, string(d))),
		//	0644)
		//index++

		contentType := part.Header["Content-Type"]
		if len(contentType) == 0 {
			continue
		}
		//fmt.Printf("Content-Type=%v\n", contentType[0])
		if contentType[0] == "text/html" {
			m.HtmlOriginal = string(d)
			contentLocation := part.Header["Content-Location"]
			if len(contentLocation) >= 1 {
				m.ContentLocation = contentLocation[0]
			}
			break
		}
	}
	return nil
}

func (m *MHtml) getBoundary(r *textproto.Reader) string {
	mimeHeader, err := r.ReadMIMEHeader()
	if err != nil {
		return ""
	}
	//fmt.Printf("%v %v\n", mimeHeader, err)
	contentType := mimeHeader.Get("Content-Type")
	//fmt.Printf("Content-Type = %v %v\n", contentType)

	_ /*mediatype*/, params, err := mime.ParseMediaType(contentType)
	//fmt.Printf("mediatype=%v,  params=%v %v, err=%v\n", mediatype, len(params), params, err)
	boundary := params["boundary"]
	//fmt.Printf("boundary=%v\n", boundary)
	return boundary
}

func (m *MHtml) convert2UTF8() error {
	if strings.ToLower(m.charset) == "gbk" {
		u, err := iconv.GBK2UTF8(m.Title)
		if err != nil {
			return err
		}
		//fmt.Printf("gbk=%v utf8=%v\n", base64.StdEncoding.EncodeToString([]byte(m.Title)), base64.StdEncoding.EncodeToString([]byte(u)))
		m.Title = u

		u, err = iconv.GBK2UTF8(m.HtmlOriginal)
		if err != nil {
			return err
		}
		m.HtmlUtf8 = u
	}

	m.HtmlUtf8 = m.HtmlOriginal
	return nil
}
