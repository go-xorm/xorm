// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

package xorm

import (
	"reflect"
	//"strconv"
	//"strings"
	"time"
)

type SQLType struct {
	Name           string
	DefaultLength  int
	DefaultLength2 int
}

var (
	Bit       = SQLType{"BIT", 0, 0}
	TinyInt   = SQLType{"TINYINT", 0, 0}
	SmallInt  = SQLType{"SMALLINT", 0, 0}
	MediumInt = SQLType{"MEDIUMINT", 0, 0}
	Int       = SQLType{"INT", 0, 0}
	Integer   = SQLType{"INTEGER", 0, 0}
	BigInt    = SQLType{"BIGINT", 0, 0}

	Char       = SQLType{"CHAR", 0, 0}
	Varchar    = SQLType{"VARCHAR", 64, 0}
	TinyText   = SQLType{"TINYTEXT", 0, 0}
	Text       = SQLType{"TEXT", 0, 0}
	MediumText = SQLType{"MEDIUMTEXT", 0, 0}
	LongText   = SQLType{"LONGTEXT", 0, 0}
	Binary     = SQLType{"BINARY", 0, 0}
	VarBinary  = SQLType{"VARBINARY", 0, 0}

	Date      = SQLType{"DATE", 0, 0}
	DateTime  = SQLType{"DATETIME", 0, 0}
	Time      = SQLType{"TIME", 0, 0}
	TimeStamp = SQLType{"TIMESTAMP", 0, 0}

	Decimal = SQLType{"DECIMAL", 26, 2}
	Numeric = SQLType{"NUMERIC", 0, 0}

	Real   = SQLType{"REAL", 0, 0}
	Float  = SQLType{"FLOAT", 0, 0}
	Double = SQLType{"DOUBLE", 0, 0}
	//Money  = SQLType{"MONEY", 0, 0}

	TinyBlob   = SQLType{"TINYBLOB", 0, 0}
	Blob       = SQLType{"BLOB", 0, 0}
	MediumBlob = SQLType{"MEDIUMBLOB", 0, 0}
	LongBlob   = SQLType{"LONGBLOB", 0, 0}
	Bytea      = SQLType{"BYTEA", 0, 0}

	Bool = SQLType{"BOOL", 0, 0}

	Serial    = SQLType{"SERIAL", 0, 0}
	BigSerial = SQLType{"BIGSERIAL", 0, 0}
)

var b byte
var tm time.Time

func Type2SQLType(t reflect.Type) (st SQLType) {
	switch k := t.Kind(); k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		st = Int
	case reflect.Int64, reflect.Uint64:
		st = BigInt
	case reflect.Float32:
		st = Float
	case reflect.Float64:
		st = Double
	case reflect.Complex64, reflect.Complex128:
		st = Varchar
	case reflect.Array, reflect.Slice:
		if t.Elem() == reflect.TypeOf(b) {
			st = Blob
		}
	case reflect.Bool:
		st = TinyInt
	case reflect.String:
		st = Varchar
	case reflect.Struct:
		if t == reflect.TypeOf(tm) {
			st = DateTime
		}
	default:
		st = Varchar
	}
	return
}

const (
	TWOSIDES = iota + 1
	ONLYTODB
	ONLYFROMDB
)

const (
	NONEINDEX = iota
	SINGLEINDEX
	UNIONINDEX
)

const (
	NONEUNIQUE = iota
	SINGLEUNIQUE
	UNIONUNIQUE
)

type Column struct {
	Name            string
	FieldName       string
	SQLType         SQLType
	Length          int
	Length2         int
	Nullable        bool
	Default         string
	UniqueType      int
	UniqueName      string
	IndexType       int
	IndexName       string
	IsPrimaryKey    bool
	IsAutoIncrement bool
	MapType         int
}

func (col *Column) String(engine *Engine) string {
	sql := engine.Quote(col.Name) + " "

	sql += engine.SqlType(col) + " "

	if col.IsPrimaryKey {
		sql += "PRIMARY KEY "
	}

	if col.IsAutoIncrement {
		sql += engine.AutoIncrStr() + " "
	}

	if col.Nullable {
		sql += "NULL "
	} else {
		sql += "NOT NULL "
	}

	if col.Default != "" {
		sql += "DEFAULT " + col.Default + " "
	}
	return sql
}

type Table struct {
	Name       string
	Type       reflect.Type
	Columns    map[string]Column
	Indexes    map[string][]string
	Uniques    map[string][]string
	PrimaryKey string
}

func (table *Table) PKColumn() Column {
	return table.Columns[table.PrimaryKey]
}

type Conversion interface {
	FromDB([]byte) error
	ToDB() ([]byte, error)
}
