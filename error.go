package xorm

import (
	"errors"
)

var (
	ErrParamsType      error = errors.New("params type error")
	ErrTableNotFound   error = errors.New("not found table")
	ErrUnSupportedType error = errors.New("unsupported type error")
	ErrNotExist        error = errors.New("not exist error")
	ErrCacheFailed     error = errors.New("cache failed")
)
