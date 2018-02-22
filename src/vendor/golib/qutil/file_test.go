package qutil

import (
	"runtime"
	"testing"
)

func TestIsExist(t *testing.T) {
	path := "/etc/passwd"
	if runtime.GOOS == "windows" {
		path = "c:/Windows/system.ini"
	}

	if !IsExist(path) {
		t.Errorf("path [%v] MUST exist", path)
	}

	path = "Not exist path"

	if IsExist(path) {
		t.Errorf("path [%v] MUST NOT exist", path)
	}
}
