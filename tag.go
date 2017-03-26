// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-xorm/core"
)

type tagContext struct {
	tagName         string
	params          []string
	preTag, nextTag string
	table           *core.Table
	col             *core.Column
	fieldValue      reflect.Value
	isIndex         bool
	isUnique        bool
	indexNames      map[string]int
	engine          *Engine
	hasCacheTag     bool
	hasNoCacheTag   bool
}

// tagHandler describes tag handler for XORM
type tagHandler func(ctx *tagContext) error

var (
	// defaultTagHandlers enumerates all the default tag handler
	defaultTagHandlers = map[string]tagHandler{
		"<-":       DirectTagHandler,
		"->":       DirectTagHandler,
		"PK":       PKTagHandler,
		"NULL":     NULLTagHandler,
		"NOT":      IgnoreTagHandler,
		"AUTOINCR": AutoIncrTagHandler,
		"DEFAULT":  DefaultTagHandler,
		"CREATED":  CreatedTagHandler,
		"UPDATED":  UpdatedTagHandler,
		"DELETED":  DeletedTagHandler,
		"VERSION":  VersionTagHandler,
		"UTC":      UTCTagHandler,
		"LOCAL":    LocalTagHandler,
		"NOTNULL":  NotNullTagHandler,
		"INDEX":    IndexTagHandler,
		"UNIQUE":   UniqueTagHandler,
		"CACHE":    CacheTagHandler,
		"NOCACHE":  CacheTagHandler,
	}
)

func init() {
	for k, _ := range core.SqlTypes {
		defaultTagHandlers[k] = SQLTypeTagHandler
	}
}

func IgnoreTagHandler(ctx *tagContext) error {
	return nil
}

// DirectTagHandler describes handle mapping type handler
func DirectTagHandler(ctx *tagContext) error {
	if ctx.tagName == "<-" {
		ctx.col.MapType = core.ONLYFROMDB
	} else if ctx.tagName == "->" {
		ctx.col.MapType = core.ONLYTODB
	}
	return nil
}

// PKTagHandler decribes handle pk
func PKTagHandler(ctx *tagContext) error {
	ctx.col.IsPrimaryKey = true
	ctx.col.Nullable = false
	return nil
}

// NULLTagHandler
func NULLTagHandler(ctx *tagContext) error {
	if len(ctx.preTag) == 0 {
		ctx.col.Nullable = true
	} else {
		ctx.col.Nullable = (strings.ToUpper(ctx.preTag) != "NOT")
	}
	return nil
}

func NotNullTagHandler(ctx *tagContext) error {
	ctx.col.Nullable = false
	return nil
}

func AutoIncrTagHandler(ctx *tagContext) error {
	ctx.col.IsAutoIncrement = true
	//col.AutoIncrStart = 1

	// TODO: for postgres how add autoincr?
	/*case strings.HasPrefix(k, "AUTOINCR(") && strings.HasSuffix(k, ")"):
	col.IsAutoIncrement = true

	autoStart := k[len("AUTOINCR")+1 : len(k)-1]
	autoStartInt, err := strconv.Atoi(autoStart)
	if err != nil {
		engine.LogError(err)
	}
	col.AutoIncrStart = autoStartInt*/
	return nil
}

func DefaultTagHandler(ctx *tagContext) error {
	ctx.col.Default = ctx.nextTag
	return nil
}

func CreatedTagHandler(ctx *tagContext) error {
	ctx.col.IsCreated = true
	return nil
}

func VersionTagHandler(ctx *tagContext) error {
	ctx.col.IsVersion = true
	ctx.col.Default = "1"
	return nil
}

func UTCTagHandler(ctx *tagContext) error {
	ctx.col.TimeZone = time.UTC
	return nil
}

func LocalTagHandler(ctx *tagContext) error {
	if len(ctx.params) == 0 {
		ctx.col.TimeZone = time.Local
	} else {
		var err error
		ctx.col.TimeZone, err = time.LoadLocation(ctx.params[0])
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdatedTagHandler(ctx *tagContext) error {
	ctx.col.IsUpdated = true
	return nil
}

func DeletedTagHandler(ctx *tagContext) error {
	ctx.col.IsDeleted = true
	return nil
}

func IndexTagHandler(ctx *tagContext) error {
	if len(ctx.params) > 0 {
		ctx.indexNames[ctx.params[0]] = core.IndexType
	} else {
		ctx.isIndex = true
	}
	return nil
}

func UniqueTagHandler(ctx *tagContext) error {
	if len(ctx.params) > 0 {
		ctx.indexNames[ctx.params[0]] = core.UniqueType
	} else {
		ctx.isUnique = true
	}
	return nil
}

func SQLTypeTagHandler(ctx *tagContext) error {
	ctx.col.SQLType = core.SQLType{Name: ctx.tagName}
	if len(ctx.params) > 0 {
		if ctx.tagName == core.Enum {
			ctx.col.EnumOptions = make(map[string]int)
			for k, v := range ctx.params {
				v = strings.TrimSpace(v)
				v = strings.Trim(v, "'")
				ctx.col.EnumOptions[v] = k
			}
		} else if ctx.tagName == core.Set {
			ctx.col.SetOptions = make(map[string]int)
			for k, v := range ctx.params {
				v = strings.TrimSpace(v)
				v = strings.Trim(v, "'")
				ctx.col.SetOptions[v] = k
			}
		} else {
			var err error
			if len(ctx.params) == 2 {
				ctx.col.Length, err = strconv.Atoi(ctx.params[0])
				if err != nil {
					return err
				}
				ctx.col.Length2, err = strconv.Atoi(ctx.params[1])
				if err != nil {
					return err
				}
			} else if len(ctx.params) == 1 {
				ctx.col.Length, err = strconv.Atoi(ctx.params[0])
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func ExtendsTagHandler(ctx *tagContext) error {
	var fieldValue = ctx.fieldValue
	switch fieldValue.Kind() {
	case reflect.Ptr:
		f := fieldValue.Type().Elem()
		if f.Kind() == reflect.Struct {
			fieldPtr := fieldValue
			fieldValue = fieldValue.Elem()
			if !fieldValue.IsValid() || fieldPtr.IsNil() {
				fieldValue = reflect.New(f).Elem()
			}
		}
		fallthrough
	case reflect.Struct:
		parentTable, err := ctx.engine.mapType(fieldValue)
		if err != nil {
			return err
		}
		for _, col := range parentTable.Columns() {
			col.FieldName = fmt.Sprintf("%v.%v", ctx.col.FieldName, col.FieldName)
			ctx.table.AddColumn(col)
			for indexName, indexType := range col.Indexes {
				addIndex(indexName, ctx.table, col, indexType)
			}
		}
	default:
		//TODO: warning
	}
	return nil
}

func CacheTagHandler(ctx *tagContext) error {
	if ctx.tagName == "CACHE" {
		if !ctx.hasCacheTag {
			ctx.hasCacheTag = true
		}
	} else if ctx.tagName == "NOCACHE" {
		if !ctx.hasNoCacheTag {
			ctx.hasNoCacheTag = true
		}
	}
	return nil
}
