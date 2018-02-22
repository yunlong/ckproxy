// +build !windows

package qlog

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

var (
	// file name
	outputFile = "info.log"
	configFile = "qlog.conf"

	// test string
	v = "abcdefadfa#X2%sx"
	s = "xxxafqwer23r12/xxs\\/sdfdsfsdfsf"
)

func setup() error {
	var configData string = `
#####################################################################
#   CloudSafeLine.Token LOG CONFIG
#####################################################################
qlog.additivity.CloudSafeLine.Token=FALSE
qlog.logger.CloudSafeLine.Token=ALL, Token

qlog.appender.Token=FileAppender
qlog.appender.Token.Schedule=MINUTELY
qlog.appender.Token.ImmediateFlush=true
qlog.appender.Token.File=info.log

qlog.appender.Token.layout=PatternLayout
qlog.appender.Token.layout.ConversionPattern=%D [PID=%P] [MODULE=%c] [HOST=%x] [%-5p] %m%n`

	return ioutil.WriteFile(configFile, []byte(configData), 0644)
}

func clean() {
	os.Remove(configFile)
	os.Remove(outputFile)
}

func TestQLog(t *testing.T) {
	l := New(configFile)

	l.Trace("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_TRACE, "trace")
	l.Debug("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_DEBUG, "debug")
	l.Info("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_INFO, "info")
	l.Warn("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_WARN, "warn")
	l.Error("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_ERROR, "error")

	_, file, line, _ := callerFuncName(0, true)
	l.LogAll("CloudSafeLine.Token", LOG_OFF, int(1), file, line, "%v : %s - %d > %s", v, s, LOG_OFF, "xxxxx")

	var err error
	var outputData []byte
	if outputData, err = ioutil.ReadFile(outputFile); err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(outputData, []byte("[MODULE=CloudSafeLine.Token] [HOST=] [TRACE] abcdefadfa#X2%sx : xxxafqwer23r12/xxs\\/sdfdsfsdfsf - 0, trace")) {
		t.Fatal("Trace method error")
	}

	if !bytes.Contains(outputData, []byte("[MODULE=CloudSafeLine.Token] [HOST=] [DEBUG] abcdefadfa#X2%sx : xxxafqwer23r12/xxs\\/sdfdsfsdfsf - 10000, debug")) {
		t.Fatal("Debug method error")
	}

	if !bytes.Contains(outputData, []byte("[MODULE=CloudSafeLine.Token] [HOST=] [INFO ] abcdefadfa#X2%sx : xxxafqwer23r12/xxs\\/sdfdsfsdfsf - 20000, info")) {
		t.Fatal("Info method error")
	}

	if !bytes.Contains(outputData, []byte("[MODULE=CloudSafeLine.Token] [HOST=] [WARN ] abcdefadfa#X2%sx : xxxafqwer23r12/xxs\\/sdfdsfsdfsf - 30000, warn")) {
		t.Fatal("Warn method error")
	}

	if !bytes.Contains(outputData, []byte("[MODULE=CloudSafeLine.Token] [HOST=] [ERROR] abcdefadfa#X2%sx : xxxafqwer23r12/xxs\\/sdfdsfsdfsf - 40000, error")) {
		t.Fatal("Error method error")
	}

	// XXX  "-- (1)Operation not permitted" 这个信息为传入的errno(1)的信息描述, 由qlog库自动生成
	if !bytes.Contains(outputData, []byte("[MODULE=CloudSafeLine.Token] [HOST=] [OFF  ] abcdefadfa#X2%sx : xxxafqwer23r12/xxs\\/sdfdsfsdfsf - 60000 > xxxxx -- (1)Operation not permitted")) {
		t.Fatal("LogAll method error")
	}

	l.Close()
	defer func() {
		if err := recover(); err != ErrInvalidInstance {
			t.Fatal("invalid instance call error")
		}
	}()
	l.Info("CloudSafeLine.Token", "%s", "invalid qlog instance")
}

func BenchmarkQLog(b *testing.B) {
	l := New(configFile)
	_, file, line, _ := callerFuncName(0, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Trace("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_TRACE, "trace")
		l.Debug("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_DEBUG, "debug")
		l.Info("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_INFO, "info")
		l.Warn("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_WARN, "warn")
		l.Error("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_ERROR, "error")
		l.LogAll("CloudSafeLine.Token", LOG_OFF, int(1), file, line, "%v : %s - %d > %s", v, s, LOG_OFF, "xxxxx")
	}
}

func TestMain(m *testing.M) {
	if setup() != nil {
		os.Exit(-1)
	}

	ret := m.Run()

	clean()
	os.Exit(ret)
}
