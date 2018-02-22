package qlog

import (
	"testing"
)

func test2(l *QLog) {
	v := "abcdef"
	s := "xxxafq"
	l.Trace("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_TRACE, "trace")
	l.Debug("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_DEBUG, "debug")
	l.Info("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_INFO, "info")
	l.Warn("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_WARN, "warn")
	l.Error("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_ERROR, "error")
}

func TestQLogWindows(t *testing.T) {
	l := New("")
	defer l.Close()

	v := "abcdef"
	s := "xxxafq"

	l.Trace("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_TRACE, "trace")
	l.Debug("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_DEBUG, "debug")
	l.Info("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_INFO, "info")
	l.Warn("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_WARN, "warn")
	l.Error("CloudSafeLine.Token", "%v : %s - %d, %s", v, s, LOG_ERROR, "error")

	test2(l)
}
