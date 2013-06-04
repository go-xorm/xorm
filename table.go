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
	TinyInt   = SQLType{"TINYINT", 0, 0}
	SmallInt  = SQLType{"SMALLINT", 0, 0}
	MediumInt = SQLType{"MEDIUMINT", 0, 0}
	Int       = SQLType{"INT", 11, 0}
	BigInt    = SQLType{"BIGINT", 0, 0}
	Char      = SQLType{"CHAR", 1, 0}
	Varchar   = SQLType{"VARCHAR", 64, 0}
	Text      = SQLType{"TEXT", 16, 0}
	Date      = SQLType{"DATE", 24, 0}
	DateTime  = SQLType{"DATETIME", 0, 0}
	Decimal   = SQLType{"DECIMAL", 26, 2}
	Float     = SQLType{"FLOAT", 31, 0}
	Double    = SQLType{"DOUBLE", 31, 0}
	Blob      = SQLType{"BLOB", 0, 0}
	TimeStamp = SQLType{"TIMESTAMP", 0, 0}
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

type Column struct {
	Name            string
	FieldName       string
	SQLType         SQLType
	Length          int
	Length2         int
	Nullable        bool
	Default         string
	IsUnique        bool
	IsPrimaryKey    bool
	IsAutoIncrement bool
}

type Table struct {
	Name       string
	Type       reflect.Type
	Columns    map[string]Column
	PrimaryKey string
}

func (table *Table) PKColumn() Column {
	return table.Columns[table.PrimaryKey]
}
