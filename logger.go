package xorm

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

// logger interface, log/syslog conform with this interface
type ILogger interface {
	Debug(m string) (err error)
	Err(m string) (err error)
	Info(m string) (err error)
	Warning(m string) (err error)
}

type SimpleLogger struct {
	DEBUG *log.Logger
	ERR   *log.Logger
	INFO  *log.Logger
	WARN  *log.Logger
}

func NewSimpleLogger(out io.Writer) *SimpleLogger {
	return &SimpleLogger{
		DEBUG: log.New(ioutil.Discard, "[xorm] [debug] ", log.Ldate|log.Lmicroseconds),
		ERR:   log.New(ioutil.Discard, "[xorm] [error] ", log.Ldate|log.Lmicroseconds),
		INFO:  log.New(ioutil.Discard, "[xorm] [info]  ", log.Ldate|log.Lmicroseconds),
		WARN:  log.New(ioutil.Discard, "[xorm] [warn]  ", log.Ldate|log.Lmicroseconds),
	}
}

func NewSimpleLogger2(out io.Writer, prefix string, flag int) *SimpleLogger {
	return &SimpleLogger{
		DEBUG: log.New(ioutil.Discard, fmt.Sprintf("%s [debug] ", prefix), log.Ldate|log.Lmicroseconds),
		ERR:   log.New(ioutil.Discard, fmt.Sprintf("%s [error] ", prefix), log.Ldate|log.Lmicroseconds),
		INFO:  log.New(ioutil.Discard, fmt.Sprintf("%s [info]  ", prefix), log.Ldate|log.Lmicroseconds),
		WARN:  log.New(ioutil.Discard, fmt.Sprintf("%s [warn]  ", prefix), log.Ldate|log.Lmicroseconds),
	}
}

func (s *SimpleLogger) Debug(m string) (err error) {
	s.DEBUG.Println(m)
	return
}

func (s *SimpleLogger) Err(m string) (err error) {
	s.ERR.Println(m)
	return
}

func (s *SimpleLogger) Info(m string) (err error) {
	s.INFO.Println(m)
	return
}

func (s *SimpleLogger) Warning(m string) (err error) {
	s.WARN.Println(m)
	return
}
