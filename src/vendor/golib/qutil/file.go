package qutil

import (
	"os"
)

func IsExist(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
