package unis

import (
	//	"strconv"
	"bufio"
	"fmt"
	"io"
	_ "net/http/pprof"
	"os"
	"strings"
//	"golib/cgo/qhsec"
)

const (
	StatusOK  = "OK"
	QPollerOK = "ok"
	MAINTAIN  = "MAINTAIN"
	FAILED    = "FAILED"
)

type MonitorModule struct {
}

func (m *MonitorModule) Initialize() error {
	HandleFunc("/status.html", DefaultPlain, m.Status, m)
	HandleFunc("/qpoller/status.html", DefaultPlain, m.Status, m)
	return nil
}

/**
func (m *MonitorModule) NppConfig() *qhsec.NppConfig {
    return nil
}
**/

func (m *MonitorModule) Status(recv []byte, ctx *Context) (send []byte, err error) {
	fmt.Printf("url=%v post data=[%v]\n", ctx.Method, string(recv))
	file, err := os.Open(DefaultFramework.statusFilePath)
	if err != nil {
		fmt.Printf("ERROR open file <%v> failed : %v\n", DefaultFramework.statusFilePath, err.Error())
		return []byte(FAILED), nil
	}
	defer file.Close()

	r := bufio.NewReader(file)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Printf("ERROR read the first line from file <%v> failed : %v\n", DefaultFramework.statusFilePath, err.Error())
		return []byte(FAILED), nil
	}

	line = strings.TrimSpace(line)
	if strings.ToLower(line) == "ok" {
		if ctx.Method == "/status.html" {
			return []byte(StatusOK), nil
		} else if ctx.Method == "/qpoller/status.html" {
			return []byte(QPollerOK), nil
		} else {
			return []byte(StatusOK), nil
		}
	}

	if strings.ToUpper(line) == MAINTAIN {
		return []byte(MAINTAIN), nil
	}

	fmt.Printf("ERROR the first line from file <%v> format wrong: [%v]\n", DefaultFramework.statusFilePath, line)

	//TODO FIXME AWS : HTTP Code 403
	return []byte(FAILED), nil
}
