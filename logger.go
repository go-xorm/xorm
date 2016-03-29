// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"io"
	"log"

	"github.com/go-xorm/core"
)

const (
	DEFAULT_LOG_PREFIX = "[xorm]"
	DEFAULT_LOG_FLAG   = log.Ldate | log.Lmicroseconds
	DEFAULT_LOG_LEVEL  = core.LOG_DEBUG
)

// SimpleLogger is the default implment of core.ILogger
type SimpleLogger struct {
	DEBUG   *log.Logger
	ERR     *log.Logger
	INFO    *log.Logger
	WARN    *log.Logger
	level   core.LogLevel
	showSQL bool
}

var _ core.ILogger = &SimpleLogger{}

// NewSimpleLogger use a special io.Writer as logger output
func NewSimpleLogger(out io.Writer) *SimpleLogger {
	return NewSimpleLogger2(out, DEFAULT_LOG_PREFIX, DEFAULT_LOG_FLAG)
}

// NewSimpleLogger2 let you customrize your logger prefix and flag
func NewSimpleLogger2(out io.Writer, prefix string, flag int) *SimpleLogger {
	return NewSimpleLogger3(out, prefix, flag, DEFAULT_LOG_LEVEL)
}

// NewSimpleLogger3 let you customrize your logger prefix and flag and logLevel
func NewSimpleLogger3(out io.Writer, prefix string, flag int, l core.LogLevel) *SimpleLogger {
	return &SimpleLogger{
		DEBUG: log.New(out, fmt.Sprintf("%s [debug] ", prefix), flag),
		ERR:   log.New(out, fmt.Sprintf("%s [error] ", prefix), flag),
		INFO:  log.New(out, fmt.Sprintf("%s [info]  ", prefix), flag),
		WARN:  log.New(out, fmt.Sprintf("%s [warn]  ", prefix), flag),
		level: l,
	}
}

// Err implement core.ILogger
func (s *SimpleLogger) Err(v ...interface{}) (err error) {
	if s.level <= core.LOG_ERR {
		s.ERR.Println(v...)
	}
	return
}

// Errf implement core.ILogger
func (s *SimpleLogger) Errf(format string, v ...interface{}) (err error) {
	if s.level <= core.LOG_ERR {
		s.ERR.Printf(format, v...)
	}
	return
}

// Debug implement core.ILogger
func (s *SimpleLogger) Debug(v ...interface{}) (err error) {
	if s.level <= core.LOG_DEBUG {
		s.DEBUG.Println(v...)
	}
	return
}

// Debugf implement core.ILogger
func (s *SimpleLogger) Debugf(format string, v ...interface{}) (err error) {
	if s.level <= core.LOG_DEBUG {
		s.DEBUG.Printf(format, v...)
	}
	return
}

// Info implement core.ILogger
func (s *SimpleLogger) Info(v ...interface{}) (err error) {
	if s.level <= core.LOG_INFO {
		s.INFO.Println(v...)
	}
	return
}

// Infof implement core.ILogger
func (s *SimpleLogger) Infof(format string, v ...interface{}) (err error) {
	if s.level <= core.LOG_INFO {
		s.INFO.Printf(format, v...)
	}
	return
}

// Warning implement core.ILogger
func (s *SimpleLogger) Warning(v ...interface{}) (err error) {
	if s.level <= core.LOG_WARNING {
		s.WARN.Println(v...)
	}
	return
}

// Warningf implement core.ILogger
func (s *SimpleLogger) Warningf(format string, v ...interface{}) (err error) {
	if s.level <= core.LOG_WARNING {
		s.WARN.Printf(format, v...)
	}
	return
}

// Level implement core.ILogger
func (s *SimpleLogger) Level() core.LogLevel {
	return s.level
}

// SetLevel implement core.ILogger
func (s *SimpleLogger) SetLevel(l core.LogLevel) (err error) {
	s.level = l
	return
}

// ShowSQL implement core.ILogger
func (s *SimpleLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		s.showSQL = true
		return
	}
	s.showSQL = show[0]
}

// IsShowSQL implement core.ILogger
func (s *SimpleLogger) IsShowSQL() bool {
	return s.showSQL
}
