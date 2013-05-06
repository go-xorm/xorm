package xorm

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

type SQLType struct {
	Name           string
	DefaultLength  int
	DefaultLength2 int
}

var (
	Int     = SQLType{"int", 11, 0}
	Char    = SQLType{"char", 1, 0}
	Bool    = SQLType{"int", 1, 0}
	Varchar = SQLType{"varchar", 50, 0}
	Date    = SQLType{"date", 24, 0}
	Decimal = SQLType{"decimal", 26, 2}
	Float   = SQLType{"float", 31, 0}
	Double  = SQLType{"double", 31, 0}
)

func (sqlType SQLType) genSQL(length int) string {
	if sqlType == Date {
		return " datetime "
	}
	return sqlType.Name + "(" + strconv.Itoa(length) + ")"
}

func Type2SQLType(t reflect.Type) (st SQLType) {
	switch k := t.Kind(); k {
	case reflect.Int, reflect.Int32, reflect.Int64:
		st = Int
	case reflect.Bool:
		st = Bool
	case reflect.String:
		st = Varchar
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			st = Date
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

func (table *Table) ColumnStr() string {
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		colNames = append(colNames, col.Name)
	}
	return strings.Join(colNames, ", ")
}

/*func (table *Table) PlaceHolders() string {
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		colNames = append(colNames, "?")
	}
	return strings.Join(colNames, ", ")
}*/

func (table *Table) PKColumn() Column {
	return table.Columns[table.PrimaryKey]
}
