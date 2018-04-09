// Copyright 2018 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/go-xorm/core"
)

// TableNameWithSchema will automatically add schema prefix on table name
func (engine *Engine) TableNameWithSchema(v string) string {
	// Add schema name as prefix of table name.
	// Only for postgres database.
	if engine.dialect.DBType() == core.POSTGRES &&
		engine.dialect.URI().Schema != "" &&
		engine.dialect.URI().Schema != postgresPublicSchema &&
		strings.Index(v, ".") == -1 {
		return engine.dialect.URI().Schema + "." + v
	}
	return v
}

func (engine *Engine) tableName(bean interface{}) string {
	return engine.TableNameWithSchema(engine.tbNameNoSchemaString(bean))
}

func (engine *Engine) tbNameForMap(v reflect.Value) string {
	t := v.Type()
	if tb, ok := v.Interface().(TableName); ok {
		return tb.TableName()
	}
	if v.CanAddr() {
		if tb, ok := v.Addr().Interface().(TableName); ok {
			return tb.TableName()
		}
	}
	return engine.TableMapper.Obj2Table(t.Name())
}

func (engine *Engine) tbNameNoSchema(w io.Writer, tablename interface{}) {
	switch tablename.(type) {
	case []string:
		t := tablename.([]string)
		if len(t) > 1 {
			fmt.Fprintf(w, "%v AS %v", engine.Quote(t[0]), engine.Quote(t[1]))
		} else if len(t) == 1 {
			fmt.Fprintf(w, engine.Quote(t[0]))
		}
	case []interface{}:
		t := tablename.([]interface{})
		l := len(t)
		var table string
		if l > 0 {
			f := t[0]
			switch f.(type) {
			case string:
				table = f.(string)
			case TableName:
				table = f.(TableName).TableName()
			default:
				v := rValue(f)
				t := v.Type()
				if t.Kind() == reflect.Struct {
					fmt.Fprintf(w, engine.tbNameForMap(v))
				} else {
					fmt.Fprintf(w, engine.Quote(fmt.Sprintf("%v", f)))
				}
			}
		}
		if l > 1 {
			fmt.Fprintf(w, "%v AS %v", engine.Quote(table),
				engine.Quote(fmt.Sprintf("%v", t[1])))
		} else if l == 1 {
			fmt.Fprintf(w, engine.Quote(table))
		}
	case TableName:
		fmt.Fprintf(w, tablename.(TableName).TableName())
	case string:
		fmt.Fprintf(w, tablename.(string))
	default:
		v := rValue(tablename)
		t := v.Type()
		if t.Kind() == reflect.Struct {
			fmt.Fprintf(w, engine.tbNameForMap(v))
		} else {
			fmt.Fprintf(w, engine.Quote(fmt.Sprintf("%v", tablename)))
		}
	}
}

func (engine *Engine) tbNameNoSchemaString(tablename interface{}) string {
	var buf bytes.Buffer
	engine.tbNameNoSchema(&buf, tablename)
	return buf.String()
}
