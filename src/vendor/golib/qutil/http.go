package qutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

func HTTPPost(url string, data []byte, timeout time.Duration) ([]byte, error) {
	r, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/octet-stream")
	r.Header.Set("User-Agent", "unis fetcher")
	r.Header.Set("Connection", "close")

	resp, err := timeoutClient(timeout).Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid HTTP Code: [%v]", resp.StatusCode)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func timeoutClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(timeout)
				c, err := net.DialTimeout(netw, addr, timeout)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
}

// RemoteAddr 从HTTP头信息中取得客户端IP地址
// XXX 新照妖镜前端nginx代理使用了X-Real-IP, 所以优先
func RemoteAddr(r *http.Request) string {
	if r == nil {
		return "0.0.0.0"
	}

	addrHeaderKeys := []string{
		"X-Real-IP",
		"HTTP_CLIENT_IP",
		"HTTP_X_FORWARDED_FOR",
		"HTTP_X_FORWARDED",
		"HTTP_X_CLUSTER_CLIENT_IP",
		"HTTP_FORWARDED_FOR",
		"REMOTE_ADDR",
	}
	for _, header := range addrHeaderKeys {
		if addr := r.Header.Get(header); addr != "" {
			return addr
		}
	}

	// The HTTP server in this package
	// sets RemoteAddr to an "IP:port" address before invoking a
	// handler.
	return strings.SplitN(r.RemoteAddr, ":", 2)[0]
}
