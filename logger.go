package xorm

import (
	"fmt"
	"io"
	"log"
)

type LogLevel int

const (
	LOG_ERR LogLevel = iota + 3
	LOG_WARNING
	LOG_INFO = iota + 6
	LOG_DEBUG
)

const (
	DEFAULT_LOG_PREFIX = "[xorm]"
	DEFAULT_LOG_FLAG   = log.Ldate | log.Lmicroseconds
	DEFAULT_LOG_LEVEL  = LOG_DEBUG
)

type SimpleLogger struct {
	DEBUG    *log.Logger
	ERR      *log.Logger
	INFO     *log.Logger
	WARN     *log.Logger
	LogLevel LogLevel
}

func NewSimpleLogger(out io.Writer) *SimpleLogger {
	return NewSimpleLogger2(out, DEFAULT_LOG_PREFIX, DEFAULT_LOG_FLAG)
}

func NewSimpleLogger2(out io.Writer, prefix string, flag int) *SimpleLogger {
	return NewSimpleLogger3(out, prefix, flag, DEFAULT_LOG_LEVEL)
}

func NewSimpleLogger3(out io.Writer, prefix string, flag int, logLevel LogLevel) *SimpleLogger {
	return &SimpleLogger{
		DEBUG:    log.New(out, fmt.Sprintf("%s [debug] ", prefix), flag),
		ERR:      log.New(out, fmt.Sprintf("%s [error] ", prefix), flag),
		INFO:     log.New(out, fmt.Sprintf("%s [info]  ", prefix), flag),
		WARN:     log.New(out, fmt.Sprintf("%s [warn]  ", prefix), flag),
		LogLevel: logLevel,
	}
}

func (s *SimpleLogger) Debug(m string) (err error) {
	if s.LogLevel >= LOG_DEBUG {
		s.DEBUG.Println(m)
	}
	return
}

func (s *SimpleLogger) Err(m string) (err error) {
	s.ERR.Println(m)
	return
}

func (s *SimpleLogger) Info(m string) (err error) {
	if s.LogLevel >= LOG_INFO {
		s.INFO.Println(m)
	}
	return
}

func (s *SimpleLogger) Warning(m string) (err error) {
	if s.LogLevel >= LOG_WARNING {
		s.WARN.Println(m)
	}
	return
}
