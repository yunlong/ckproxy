package qlog

import (
	"fmt"
	"log"
	"os"
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

type QLog struct {
	logger *log.Logger
}

func New(confFilePath string) *QLog {
	return &QLog{log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)}
}

func (logger *QLog) Trace(name, format string, args ...interface{}) {
	logger.callLogAll(name, LOG_TRACE, format, args...)
}

func (logger *QLog) Debug(name, format string, args ...interface{}) {
	logger.callLogAll(name, LOG_DEBUG, format, args...)
}

func (logger *QLog) Info(name, format string, args ...interface{}) {
	logger.callLogAll(name, LOG_INFO, format, args...)
}

func (logger *QLog) Warn(name, format string, args ...interface{}) {
	logger.callLogAll(name, LOG_WARN, format, args...)
}

func (logger *QLog) Error(name, format string, args ...interface{}) {
	logger.callLogAll(name, LOG_ERROR, format, args...)
}

func (logger *QLog) LogAll(name string, level int, eno int,
	file string, line int, format string, args ...interface{}) {
	logger.logger.Output(3, fmt.Sprintf(format, args...))
}

func (logger *QLog) Close() {
}

func (logger *QLog) callLogAll(name string, level int, format string, args ...interface{}) {
	logger.logger.Output(3, fmt.Sprintf(format, args...))
}
