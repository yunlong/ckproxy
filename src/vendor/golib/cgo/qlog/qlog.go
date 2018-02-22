// +build !windows

package qlog

// #cgo CFLAGS:  -I/home/s/include -I.
// #cgo LDFLAGS: -L/home/s/lib -lqlog
// #cgo LDFLAGS: -Wl,-rpath=/home/s/lib
//
// #include "qlog/c.h"
// #include <stdlib.h>
//
// void gLogAll(const char* name, int level, int eNo,
//         const char* file, int line, const char* content) {
//     cLogAll(name, level, eNo, file, line, content);
// }
//
// void gLogConfig(const char* configfile) {
//     cLogConfig(configfile);
// }
//
// void gLogCleanConfig() {
//     cLogCleanConfig();
// }
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"unsafe"
)

const (
	LOG_ALL   int = 0
	LOG_TRACE int = 0
	LOG_DEBUG int = 10000
	LOG_INFO  int = 20000
	LOG_WARN  int = 30000
	LOG_ERROR int = 40000
	LOG_FATAL int = 50000
	LOG_OFF   int = 60000
)

var ErrInvalidInstance = errors.New("invalid qlog instance")
var ErrInvalidLogLevel = errors.New("invalid qlog level")

type QLog struct {
	config *C.char
	level  int
	inited bool
}

func New(confFilePath string) *QLog {
	conf := C.CString(confFilePath)
	C.gLogConfig(conf)

	return &QLog{
		config: conf,
		level:  LOG_ALL,
		inited: true,
	}
}

func (log *QLog) SetLevel(level int) error {
	switch level {
	case LOG_ALL, LOG_DEBUG, LOG_INFO, LOG_WARN, LOG_ERROR, LOG_FATAL, LOG_OFF:
		log.level = level
		return nil
	}

	return ErrInvalidLogLevel
}

func (log *QLog) Trace(name, format string, args ...interface{}) {
	log.callLogAll(name, LOG_TRACE, format, args...)
}

func (log *QLog) Debug(name, format string, args ...interface{}) {
	log.callLogAll(name, LOG_DEBUG, format, args...)
}

func (log *QLog) Info(name, format string, args ...interface{}) {
	log.callLogAll(name, LOG_INFO, format, args...)
}

func (log *QLog) Warn(name, format string, args ...interface{}) {
	log.callLogAll(name, LOG_WARN, format, args...)
}

func (log *QLog) Error(name, format string, args ...interface{}) {
	log.callLogAll(name, LOG_ERROR, format, args...)
}

func (log *QLog) LogAll(name string, level int, eno int,
	file string, line int, format string, args ...interface{}) {
	if !log.inited {
		panic(ErrInvalidInstance)
	}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cfile := C.CString(file)
	defer C.free(unsafe.Pointer(cfile))

	content := C.CString(strings.Replace(fmt.Sprintf(format, args...), "%", "%%", -1))
	defer C.free(unsafe.Pointer(content))

	C.gLogAll(cname, C.int(level), C.int(eno), cfile, C.int(line), content)
}

func (log *QLog) Close() {
	if !log.inited {
		return
	}

	C.gLogCleanConfig()
	C.free(unsafe.Pointer(log.config))
	log.inited = false
}

func (log *QLog) callLogAll(name string, level int, format string, args ...interface{}) {
	if level < log.level {
		return
	}

	if _, file, line, ok := callerFuncName(2, false); ok {
		log.LogAll(name, level, 0, file, line, format, args...)
		return
	}

	log.LogAll(name, level, 0, "", 0, format, args...)
}

func callerFuncName(skip int, useFuncName bool) (name, file string, line int, ok bool) {
	var pc uintptr
	if pc, file, line, ok = runtime.Caller(skip + 1); !ok {
		return
	}

	if useFuncName {
		name = runtime.FuncForPC(pc).Name()
	}

	return
}
