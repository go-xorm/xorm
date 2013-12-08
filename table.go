package xorm

import (
	"reflect"
	"sort"
	"strings"
	"time"
)

// xorm SQL types
type SQLType struct {
	Name           string
	DefaultLength  int
	DefaultLength2 int
}

func (s *SQLType) IsText() bool {
	return s.Name == Char || s.Name == Varchar || s.Name == TinyText ||
		s.Name == Text || s.Name == MediumText || s.Name == LongText
}

func (s *SQLType) IsBlob() bool {
	return (s.Name == TinyBlob) || (s.Name == Blob) ||
		s.Name == MediumBlob || s.Name == LongBlob ||
		s.Name == Binary || s.Name == VarBinary || s.Name == Bytea
}

const ()

var (
	Bit       = "BIT"
	TinyInt   = "TINYINT"
	SmallInt  = "SMALLINT"
	MediumInt = "MEDIUMINT"
	Int       = "INT"
	Integer   = "INTEGER"
	BigInt    = "BIGINT"

	Char       = "CHAR"
	Varchar    = "VARCHAR"
	TinyText   = "TINYTEXT"
	Text       = "TEXT"
	MediumText = "MEDIUMTEXT"
	LongText   = "LONGTEXT"
	Binary     = "BINARY"
	VarBinary  = "VARBINARY"

	Date       = "DATE"
	DateTime   = "DATETIME"
	Time       = "TIME"
	TimeStamp  = "TIMESTAMP"
	TimeStampz = "TIMESTAMPZ"

	Decimal = "DECIMAL"
	Numeric = "NUMERIC"

	Real   = "REAL"
	Float  = "FLOAT"
	Double = "DOUBLE"

	TinyBlob   = "TINYBLOB"
	Blob       = "BLOB"
	MediumBlob = "MEDIUMBLOB"
	LongBlob   = "LONGBLOB"
	Bytea      = "BYTEA"

	Bool = "BOOL"

	Serial    = "SERIAL"
	BigSerial = "BIGSERIAL"

	sqlTypes = map[string]bool{
		Bit:       true,
		TinyInt:   true,
		SmallInt:  true,
		MediumInt: true,
		Int:       true,
		Integer:   true,
		BigInt:    true,

		Char:       true,
		Varchar:    true,
		TinyText:   true,
		Text:       true,
		MediumText: true,
		LongText:   true,
		Binary:     true,
		VarBinary:  true,

		Date:       true,
		DateTime:   true,
		Time:       true,
		TimeStamp:  true,
		TimeStampz: true,

		Decimal: true,
		Numeric: true,

		Real:       true,
		Float:      true,
		Double:     true,
		TinyBlob:   true,
		Blob:       true,
		MediumBlob: true,
		LongBlob:   true,
		Bytea:      true,

		Bool: true,

		Serial:    true,
		BigSerial: true,
	}

	intTypes  = sort.StringSlice{"*int", "*int16", "*int32", "*int8"}
	uintTypes = sort.StringSlice{"*uint", "*uint16", "*uint32", "*uint8"}
)

var b byte
var tm time.Time

func Type2SQLType(t reflect.Type) (st SQLType) {
	switch k := t.Kind(); k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		st = SQLType{Int, 0, 0}
	case reflect.Int64, reflect.Uint64:
		st = SQLType{BigInt, 0, 0}
	case reflect.Float32:
		st = SQLType{Float, 0, 0}
	case reflect.Float64:
		st = SQLType{Double, 0, 0}
	case reflect.Complex64, reflect.Complex128:
		st = SQLType{Varchar, 64, 0}
	case reflect.Array, reflect.Slice, reflect.Map:
		if t.Elem() == reflect.TypeOf(b) {
			st = SQLType{Blob, 0, 0}
		} else {
			st = SQLType{Text, 0, 0}
		}
	case reflect.Bool:
		st = SQLType{Bool, 0, 0}
	case reflect.String:
		st = SQLType{Varchar, 255, 0}
	case reflect.Struct:
		if t == reflect.TypeOf(tm) {
			st = SQLType{DateTime, 0, 0}
		} else {
			st = SQLType{Text, 0, 0}
		}
	case reflect.Ptr:
		st, _ = ptrType2SQLType(t)
	default:
		st = SQLType{Text, 0, 0}
	}
	return
}

func ptrType2SQLType(t reflect.Type) (st SQLType, has bool) {
	typeStr := t.String()
	has = true

	switch typeStr {
	case "*string":
		st = SQLType{Varchar, 255, 0}
	case "*bool":
		st = SQLType{Bool, 0, 0}
	case "*complex64", "*complex128":
		st = SQLType{Varchar, 64, 0}
	case "*float32":
		st = SQLType{Float, 0, 0}
	case "*float64":
		st = SQLType{Double, 0, 0}
	case "*int64", "*uint64":
		st = SQLType{BigInt, 0, 0}
	case "*time.Time":
		st = SQLType{DateTime, 0, 0}
	case "*int", "*int16", "*int32", "*int8", "*uint", "*uint16", "*uint32", "*uint8":
		st = SQLType{Int, 0, 0}
	default:
		has = false
	}
	return
}

// default sql type change to go types
func SQLType2Type(st SQLType) reflect.Type {
	name := strings.ToUpper(st.Name)
	switch name {
	case Bit, TinyInt, SmallInt, MediumInt, Int, Integer, Serial:
		return reflect.TypeOf(1)
	case BigInt, BigSerial:
		return reflect.TypeOf(int64(1))
	case Float, Real:
		return reflect.TypeOf(float32(1))
	case Double:
		return reflect.TypeOf(float64(1))
	case Char, Varchar, TinyText, Text, MediumText, LongText:
		return reflect.TypeOf("")
	case TinyBlob, Blob, LongBlob, Bytea, Binary, MediumBlob, VarBinary:
		return reflect.TypeOf([]byte{})
	case Bool:
		return reflect.TypeOf(true)
	case DateTime, Date, Time, TimeStamp, TimeStampz:
		return reflect.TypeOf(tm)
	case Decimal, Numeric:
		return reflect.TypeOf("")
	default:
		return reflect.TypeOf("")
	}
}

const (
	IndexType = iota + 1
	UniqueType
)

// database index
type Index struct {
	Name string
	Type int
	Cols []string
}

// add columns which will be composite index
func (index *Index) AddColumn(cols ...string) {
	for _, col := range cols {
		index.Cols = append(index.Cols, col)
	}
}

// new an index
func NewIndex(name string, indexType int) *Index {
	return &Index{name, indexType, make([]string, 0)}
}

const (
	TWOSIDES = iota + 1
	ONLYTODB
	ONLYFROMDB
)

// database column
type Column struct {
	Name            string
	FieldName       string
	SQLType         SQLType
	Length          int
	Length2         int
	Nullable        bool
	Default         string
	Indexes         map[string]bool
	IsPrimaryKey    bool
	IsAutoIncrement bool
	MapType         int
	IsCreated       bool
	IsUpdated       bool
	IsCascade       bool
	IsVersion       bool
}

// generate column description string according dialect
func (col *Column) String(d dialect) string {
	sql := d.QuoteStr() + col.Name + d.QuoteStr() + " "

	sql += d.SqlType(col) + " "

	if col.IsPrimaryKey {
		sql += "PRIMARY KEY "
	}

	if col.IsAutoIncrement {
		sql += d.AutoIncrStr() + " "
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

// return col's filed of struct's value
func (col *Column) ValueOf(bean interface{}) reflect.Value {
	var fieldValue reflect.Value
	if strings.Contains(col.FieldName, ".") {
		fields := strings.Split(col.FieldName, ".")
		if len(fields) > 2 {
			return reflect.ValueOf(nil)
		}

		fieldValue = reflect.Indirect(reflect.ValueOf(bean)).FieldByName(fields[0])
		fieldValue = fieldValue.FieldByName(fields[1])
	} else {
		fieldValue = reflect.Indirect(reflect.ValueOf(bean)).FieldByName(col.FieldName)
	}
	return fieldValue
}

// database table
type Table struct {
	Name       string
	Type       reflect.Type
	ColumnsSeq []string
	Columns    map[string]*Column
	Indexes    map[string]*Index
	PrimaryKey string
	Created    map[string]bool
	Updated    string
	Version    string
	Cacher     Cacher
}

/*
func NewTable(name string, t reflect.Type) *Table {
	return &Table{Name: name, Type: t,
		ColumnsSeq: make([]string, 0),
		Columns:    make(map[string]*Column),
		Indexes:    make(map[string]*Index),
		Created:    make(map[string]bool),
	}
}*/

// if has primary key, return column
func (table *Table) PKColumn() *Column {
	return table.Columns[table.PrimaryKey]
}

func (table *Table) VersionColumn() *Column {
	return table.Columns[table.Version]
}

// add a column to table
func (table *Table) AddColumn(col *Column) {
	table.ColumnsSeq = append(table.ColumnsSeq, col.Name)
	table.Columns[col.Name] = col
	if col.IsPrimaryKey {
		table.PrimaryKey = col.Name
	}
	if col.IsCreated {
		table.Created[col.Name] = true
	}
	if col.IsUpdated {
		table.Updated = col.Name
	}
	if col.IsVersion {
		table.Version = col.Name
	}
}

// add an index or an unique to table
func (table *Table) AddIndex(index *Index) {
	table.Indexes[index.Name] = index
}

func (table *Table) genCols(session *Session, bean interface{}, useCol bool, includeQuote bool) ([]string, []interface{}, error) {
	colNames := make([]string, 0)
	args := make([]interface{}, 0)

	for _, col := range table.Columns {
		if useCol && !col.IsVersion && !col.IsCreated && !col.IsUpdated {
			if _, ok := session.Statement.columnMap[col.Name]; !ok {
				continue
			}
		}
		if col.MapType == ONLYFROMDB {
			continue
		}

		fieldValue := col.ValueOf(bean)
		if col.IsAutoIncrement && fieldValue.Int() == 0 {
			continue
		}

		if session.Statement.ColumnStr != "" {
			if _, ok := session.Statement.columnMap[col.Name]; !ok {
				continue
			}
		}
		if session.Statement.OmitStr != "" {
			if _, ok := session.Statement.columnMap[col.Name]; ok {
				continue
			}
		}

		if (col.IsCreated || col.IsUpdated) && session.Statement.UseAutoTime {
			args = append(args, time.Now())
		} else if col.IsVersion && session.Statement.checkVersion {
			args = append(args, 1)
		} else {
			arg, err := session.value2Interface(col, fieldValue)
			if err != nil {
				return colNames, args, err
			}
			args = append(args, arg)
		}

		if includeQuote {
			colNames = append(colNames, session.Engine.Quote(col.Name)+" = ?")
		} else {
			colNames = append(colNames, col.Name)
		}
	}
	return colNames, args, nil
}

// Conversion is an interface. A type implements Conversion will according
// the custom method to fill into database and retrieve from database.
type Conversion interface {
	FromDB([]byte) error
	ToDB() ([]byte, error)
}
