package xorm

import (
	"errors"
)

var (
	ParamsTypeError      error = errors.New("params type error")
	TableNotFoundError   error = errors.New("not found table")
	UnSupportedTypeError error = errors.New("unsupported type error")
)
