package core

import (
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	POSTGRES = "postgres"
	SQLITE   = "sqlite3"
	MYSQL    = "mysql"
	MSSQL    = "mssql"
	ORACLE   = "oracle"
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

	Binary     = "BINARY"
	VarBinary  = "VARBINARY"
	TinyBlob   = "TINYBLOB"
	Blob       = "BLOB"
	MediumBlob = "MEDIUMBLOB"
	LongBlob   = "LONGBLOB"
	Bytea      = "BYTEA"

	Bool = "BOOL"

	Serial    = "SERIAL"
	BigSerial = "BIGSERIAL"

	SqlTypes = map[string]bool{
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

		Date:       true,
		DateTime:   true,
		Time:       true,
		TimeStamp:  true,
		TimeStampz: true,

		Decimal: true,
		Numeric: true,

		Binary:     true,
		VarBinary:  true,
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

// !nashtsai! treat following var as interal const values, these are used for reflect.TypeOf comparision
var (
	c_EMPTY_STRING       string
	c_BOOL_DEFAULT       bool
	c_BYTE_DEFAULT       byte
	c_COMPLEX64_DEFAULT  complex64
	c_COMPLEX128_DEFAULT complex128
	c_FLOAT32_DEFAULT    float32
	c_FLOAT64_DEFAULT    float64
	c_INT64_DEFAULT      int64
	c_UINT64_DEFAULT     uint64
	c_INT32_DEFAULT      int32
	c_UINT32_DEFAULT     uint32
	c_INT16_DEFAULT      int16
	c_UINT16_DEFAULT     uint16
	c_INT8_DEFAULT       int8
	c_UINT8_DEFAULT      uint8
	c_INT_DEFAULT        int
	c_UINT_DEFAULT       uint
	c_TIME_DEFAULT       time.Time
)

var (
	IntType   = reflect.TypeOf(c_INT_DEFAULT)
	Int8Type  = reflect.TypeOf(c_INT8_DEFAULT)
	Int16Type = reflect.TypeOf(c_INT16_DEFAULT)
	Int32Type = reflect.TypeOf(c_INT32_DEFAULT)
	Int64Type = reflect.TypeOf(c_INT64_DEFAULT)

	UintType   = reflect.TypeOf(c_UINT_DEFAULT)
	Uint8Type  = reflect.TypeOf(c_UINT8_DEFAULT)
	Uint16Type = reflect.TypeOf(c_UINT16_DEFAULT)
	Uint32Type = reflect.TypeOf(c_UINT32_DEFAULT)
	Uint64Type = reflect.TypeOf(c_UINT64_DEFAULT)

	Float32Type = reflect.TypeOf(c_FLOAT32_DEFAULT)
	Float64Type = reflect.TypeOf(c_FLOAT64_DEFAULT)

	Complex64Type  = reflect.TypeOf(c_COMPLEX64_DEFAULT)
	Complex128Type = reflect.TypeOf(c_COMPLEX128_DEFAULT)

	StringType = reflect.TypeOf(c_EMPTY_STRING)
	BoolType   = reflect.TypeOf(c_BOOL_DEFAULT)
	ByteType   = reflect.TypeOf(c_BYTE_DEFAULT)

	TimeType = reflect.TypeOf(c_TIME_DEFAULT)
)

var (
	PtrIntType   = reflect.PtrTo(IntType)
	PtrInt8Type  = reflect.PtrTo(Int8Type)
	PtrInt16Type = reflect.PtrTo(Int16Type)
	PtrInt32Type = reflect.PtrTo(Int32Type)
	PtrInt64Type = reflect.PtrTo(Int64Type)

	PtrUintType   = reflect.PtrTo(UintType)
	PtrUint8Type  = reflect.PtrTo(Uint8Type)
	PtrUint16Type = reflect.PtrTo(Uint16Type)
	PtrUint32Type = reflect.PtrTo(Uint32Type)
	PtrUint64Type = reflect.PtrTo(Uint64Type)

	PtrFloat32Type = reflect.PtrTo(Float32Type)
	PtrFloat64Type = reflect.PtrTo(Float64Type)

	PtrComplex64Type  = reflect.PtrTo(Complex64Type)
	PtrComplex128Type = reflect.PtrTo(Complex128Type)

	PtrStringType = reflect.PtrTo(StringType)
	PtrBoolType   = reflect.PtrTo(BoolType)
	PtrByteType   = reflect.PtrTo(ByteType)

	PtrTimeType = reflect.PtrTo(TimeType)
)

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
		if t.Elem() == reflect.TypeOf(c_BYTE_DEFAULT) {
			st = SQLType{Blob, 0, 0}
		} else {
			st = SQLType{Text, 0, 0}
		}
	case reflect.Bool:
		st = SQLType{Bool, 0, 0}
	case reflect.String:
		st = SQLType{Varchar, 255, 0}
	case reflect.Struct:
		if t == reflect.TypeOf(c_TIME_DEFAULT) {
			st = SQLType{DateTime, 0, 0}
		} else {
			// TODO need to handle association struct
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
	has = true

	switch t {
	case reflect.TypeOf(&c_EMPTY_STRING):
		st = SQLType{Varchar, 255, 0}
		return
	case reflect.TypeOf(&c_BOOL_DEFAULT):
		st = SQLType{Bool, 0, 0}
	case reflect.TypeOf(&c_COMPLEX64_DEFAULT), reflect.TypeOf(&c_COMPLEX128_DEFAULT):
		st = SQLType{Varchar, 64, 0}
	case reflect.TypeOf(&c_FLOAT32_DEFAULT):
		st = SQLType{Float, 0, 0}
	case reflect.TypeOf(&c_FLOAT64_DEFAULT):
		st = SQLType{Double, 0, 0}
	case reflect.TypeOf(&c_INT64_DEFAULT), reflect.TypeOf(&c_UINT64_DEFAULT):
		st = SQLType{BigInt, 0, 0}
	case reflect.TypeOf(&c_TIME_DEFAULT):
		st = SQLType{DateTime, 0, 0}
	case reflect.TypeOf(&c_INT_DEFAULT), reflect.TypeOf(&c_INT32_DEFAULT), reflect.TypeOf(&c_INT8_DEFAULT), reflect.TypeOf(&c_INT16_DEFAULT), reflect.TypeOf(&c_UINT_DEFAULT), reflect.TypeOf(&c_UINT32_DEFAULT), reflect.TypeOf(&c_UINT8_DEFAULT), reflect.TypeOf(&c_UINT16_DEFAULT):
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
		return reflect.TypeOf(c_TIME_DEFAULT)
	case Decimal, Numeric:
		return reflect.TypeOf("")
	default:
		return reflect.TypeOf("")
	}
}
