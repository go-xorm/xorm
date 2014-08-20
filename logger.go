package xorm

import (
	"fmt"
	"github.com/go-xorm/core"
	"io"
	"log"
	"log/syslog"
)

const (
	DEFAULT_LOG_PREFIX = "[xorm]"
	DEFAULT_LOG_FLAG   = log.Ldate | log.Lmicroseconds
	DEFAULT_LOG_LEVEL  = core.LOG_DEBUG
)

type SimpleLogger struct {
	DEBUG *log.Logger
	ERR   *log.Logger
	INFO  *log.Logger
	WARN  *log.Logger
	level core.LogLevel
}

func NewSimpleLogger(out io.Writer) *SimpleLogger {
	return NewSimpleLogger2(out, DEFAULT_LOG_PREFIX, DEFAULT_LOG_FLAG)
}

func NewSimpleLogger2(out io.Writer, prefix string, flag int) *SimpleLogger {
	return NewSimpleLogger3(out, prefix, flag, DEFAULT_LOG_LEVEL)
}

func NewSimpleLogger3(out io.Writer, prefix string, flag int, l core.LogLevel) *SimpleLogger {
	return &SimpleLogger{
		DEBUG: log.New(out, fmt.Sprintf("%s [debug] ", prefix), flag),
		ERR:   log.New(out, fmt.Sprintf("%s [error] ", prefix), flag),
		INFO:  log.New(out, fmt.Sprintf("%s [info]  ", prefix), flag),
		WARN:  log.New(out, fmt.Sprintf("%s [warn]  ", prefix), flag),
		level: l,
	}
}

func (s *SimpleLogger) Err(v ...interface{}) (err error) {
	if s.level > core.LOG_OFF && s.level <= core.LOG_ERR {
		s.ERR.Println(v...)
	}
	return
}

func (s *SimpleLogger) Errf(format string, v ...interface{}) (err error) {
	if s.level > core.LOG_OFF && s.level <= core.LOG_ERR {
		s.ERR.Printf(format, v...)
	}
	return
}

func (s *SimpleLogger) Debug(v ...interface{}) (err error) {
	if s.level > core.LOG_OFF && s.level <= core.LOG_DEBUG {
		s.DEBUG.Println(v...)
	}
	return
}

func (s *SimpleLogger) Debugf(format string, v ...interface{}) (err error) {
	if s.level > core.LOG_OFF && s.level >= core.LOG_DEBUG {
		s.DEBUG.Printf(format, v...)
	}
	return
}

func (s *SimpleLogger) Info(v ...interface{}) (err error) {
	if s.level > core.LOG_OFF && s.level >= core.LOG_INFO {
		s.INFO.Println(v...)
	}
	return
}

func (s *SimpleLogger) Infof(format string, v ...interface{}) (err error) {
	if s.level > core.LOG_OFF && s.level >= core.LOG_INFO {
		s.INFO.Printf(format, v...)
	}
	return
}

func (s *SimpleLogger) Warning(v ...interface{}) (err error) {
	if s.level > core.LOG_OFF && s.level >= core.LOG_WARNING {
		s.WARN.Println(v...)
	}
	return
}

func (s *SimpleLogger) Warningf(format string, v ...interface{}) (err error) {
	if s.level > core.LOG_OFF && s.level >= core.LOG_WARNING {
		s.WARN.Printf(format, v...)
	}
	return
}

func (s *SimpleLogger) Level() core.LogLevel {
	return s.level
}

func (s *SimpleLogger) SetLevel(l core.LogLevel) (err error) {
	s.level = l
	return
}

type SyslogLogger struct {
	w *syslog.Writer
}

func NewSyslogLogger(w *syslog.Writer) *SyslogLogger {
	return &SyslogLogger{w: w}
}

func (s *SyslogLogger) Debug(v ...interface{}) (err error) {
	return s.w.Debug(fmt.Sprint(v...))
}

func (s *SyslogLogger) Debugf(format string, v ...interface{}) (err error) {
	return s.w.Debug(fmt.Sprintf(format, v...))
}

func (s *SyslogLogger) Err(v ...interface{}) (err error) {
	return s.w.Err(fmt.Sprint(v...))
}

func (s *SyslogLogger) Errf(format string, v ...interface{}) (err error) {
	return s.w.Err(fmt.Sprintf(format, v...))
}

func (s *SyslogLogger) Info(v ...interface{}) (err error) {
	return s.w.Info(fmt.Sprint(v...))
}

func (s *SyslogLogger) Infof(format string, v ...interface{}) (err error) {
	return s.w.Info(fmt.Sprintf(format, v...))
}

func (s *SyslogLogger) Warning(v ...interface{}) (err error) {
	return s.w.Warning(fmt.Sprint(v...))
}

func (s *SyslogLogger) Warningf(format string, v ...interface{}) (err error) {
	return s.w.Warning(fmt.Sprintf(format, v...))
}

func (s *SyslogLogger) Level() core.LogLevel {
	return core.LOG_UNKNOWN
}

// SetLevel always return error, as current log/syslog package doesn't allow to set priority level after syslog.Writer created
func (s *SyslogLogger) SetLevel(l core.LogLevel) (err error) {
	return fmt.Errorf("unable to set syslog level")
}
