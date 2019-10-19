package xorm

import (
	"reflect"

	"xorm.io/builder"
	"xorm.io/core"
)

type SoftDelete interface {
	getDeleteValue() interface{}
	getSelectFilter(deleteField string) builder.Cond
	setBeanConumenAttr(bean interface{}, col *core.Column, val interface{})
}

type DefaultSoftDeleteHandler struct {
}

func (h *DefaultSoftDeleteHandler) setBeanConumenAttr(bean interface{}, col *core.Column, val interface{}) {
	t := val.(int64)
	v, err := col.ValueOf(bean)
	if err != nil {
		return
	}
	if v.CanSet() {
		switch v.Type().Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32:
			v.SetInt(t)
		case reflect.Uint, reflect.Uint64, reflect.Uint32:
			v.SetUint(uint64(t))
		}
	}
}
func (h *DefaultSoftDeleteHandler) getDeleteValue() interface{} {
	return int64(1)
}
func (h *DefaultSoftDeleteHandler) getSelectFilter(deleteField string) builder.Cond {
	return builder.Eq{deleteField: int64(0)}
}
