package xorm

import (
	"io"
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
	logger *log.Logger
}

func NewSimpleLogger(out io.Writer) *SimpleLogger {
	return &SimpleLogger{
		logger: log.New(out, "[xorm] ", log.Ldate|log.Lmicroseconds)}
}

func NewSimpleLogger2(out io.Writer, prefix string, flag int) *SimpleLogger {
	return &SimpleLogger{
		logger: log.New(out, prefix, flag)}
}

func (s *SimpleLogger) Debug(m string) (err error) {
	s.logger.Println("[debug]", m)
	return
}

func (s *SimpleLogger) Err(m string) (err error) {
	s.logger.Println("[error]", m)
	return
}

func (s *SimpleLogger) Info(m string) (err error) {
	s.logger.Println("[info]", m)
	return
}

func (s *SimpleLogger) Warning(m string) (err error) {
	s.logger.Println("[warning]", m)
	return
}
