package dialects

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	. "github.com/lunny/xorm/core"
)

func init() {
	RegisterDialect("postgres", &postgres{})
}

type postgres struct {
	Base
}

func (db *postgres) Init(uri *Uri, drivername, dataSourceName string) error {
	return db.Base.Init(db, uri, drivername, dataSourceName)
}

func (db *postgres) SqlType(c *Column) string {
	var res string
	switch t := c.SQLType.Name; t {
	case TinyInt:
		res = SmallInt

	case MediumInt, Int, Integer:
		if c.IsAutoIncrement {
			return Serial
		}
		return Integer
	case Serial, BigSerial:
		c.IsAutoIncrement = true
		c.Nullable = false
		res = t
	case Binary, VarBinary:
		return Bytea
	case DateTime:
		res = TimeStamp
	case TimeStampz:
		return "timestamp with time zone"
	case Float:
		res = Real
	case TinyText, MediumText, LongText:
		res = Text
	case Blob, TinyBlob, MediumBlob, LongBlob:
		return Bytea
	case Double:
		return "DOUBLE PRECISION"
	default:
		if c.IsAutoIncrement {
			return Serial
		}
		res = t
	}

	var hasLen1 bool = (c.Length > 0)
	var hasLen2 bool = (c.Length2 > 0)
	if hasLen1 {
		res += "(" + strconv.Itoa(c.Length) + ")"
	} else if hasLen2 {
		res += "(" + strconv.Itoa(c.Length) + "," + strconv.Itoa(c.Length2) + ")"
	}
	return res
}

func (db *postgres) SupportInsertMany() bool {
	return true
}

func (db *postgres) QuoteStr() string {
	return "\""
}

func (db *postgres) AutoIncrStr() string {
	return ""
}

func (db *postgres) SupportEngine() bool {
	return false
}

func (db *postgres) SupportCharset() bool {
	return false
}

func (db *postgres) IndexOnTable() bool {
	return false
}

func (db *postgres) IndexCheckSql(tableName, idxName string) (string, []interface{}) {
	args := []interface{}{tableName, idxName}
	return `SELECT indexname FROM pg_indexes ` +
		`WHERE tablename = ? AND indexname = ?`, args
}

func (db *postgres) TableCheckSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return `SELECT tablename FROM pg_tables WHERE tablename = ?`, args
}

func (db *postgres) ColumnCheckSql(tableName, colName string) (string, []interface{}) {
	args := []interface{}{tableName, colName}
	return "SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ?" +
		" AND column_name = ?", args
}

func (db *postgres) GetColumns(tableName string) ([]string, map[string]*Column, error) {
	args := []interface{}{tableName}
	s := "SELECT column_name, column_default, is_nullable, data_type, character_maximum_length" +
		", numeric_precision, numeric_precision_radix FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = $1"
	cnn, err := Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, nil, err
	}
	defer cnn.Close()
	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, nil, err
	}
	cols := make(map[string]*Column)
	colSeq := make([]string, 0)

	for rows.Next() {
		col := new(Column)
		col.Indexes = make(map[string]bool)
		var colName, isNullable, dataType string
		var maxLenStr, colDefault, numPrecision, numRadix *string
		err = rows.Scan(&colName, &colDefault, &isNullable, &dataType, &maxLenStr, &numPrecision, &numRadix)
		if err != nil {
			return nil, nil, err
		}

		var maxLen int
		if maxLenStr != nil {
			maxLen, err = strconv.Atoi(*maxLenStr)
			if err != nil {
				return nil, nil, err
			}
		}

		col.Name = strings.Trim(colName, `" `)

		if colDefault != nil {
			if strings.HasPrefix(*colDefault, "nextval") {
				col.IsPrimaryKey = true
			} else {
				col.Default = *colDefault
			}
		}

		if isNullable == "YES" {
			col.Nullable = true
		} else {
			col.Nullable = false
		}

		switch dataType {
		case "character varying", "character":
			col.SQLType = SQLType{Varchar, 0, 0}
		case "timestamp without time zone":
			col.SQLType = SQLType{DateTime, 0, 0}
		case "timestamp with time zone":
			col.SQLType = SQLType{TimeStampz, 0, 0}
		case "double precision":
			col.SQLType = SQLType{Double, 0, 0}
		case "boolean":
			col.SQLType = SQLType{Bool, 0, 0}
		case "time without time zone":
			col.SQLType = SQLType{Time, 0, 0}
		default:
			col.SQLType = SQLType{strings.ToUpper(dataType), 0, 0}
		}
		if _, ok := SqlTypes[col.SQLType.Name]; !ok {
			return nil, nil, errors.New(fmt.Sprintf("unkonw colType %v", dataType))
		}

		col.Length = maxLen

		if col.SQLType.IsText() {
			if col.Default != "" {
				col.Default = "'" + col.Default + "'"
			}
		}
		cols[col.Name] = col
		colSeq = append(colSeq, col.Name)
	}

	return colSeq, cols, nil
}

func (db *postgres) GetTables() ([]*Table, error) {
	args := []interface{}{}
	s := "SELECT tablename FROM pg_tables where schemaname = 'public'"
	cnn, err := Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, err
	}

	tables := make([]*Table, 0)
	for rows.Next() {
		table := NewEmptyTable()
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		table.Name = name
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *postgres) GetIndexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := "SELECT indexname, indexdef FROM pg_indexes WHERE schemaname = 'public' and tablename = $1"

	cnn, err := Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, err
	}

	indexes := make(map[string]*Index, 0)
	for rows.Next() {
		var indexType int
		var indexName, indexdef string
		var colNames []string
		err = rows.Scan(&indexName, &indexdef)
		if err != nil {
			return nil, err
		}
		indexName = strings.Trim(indexName, `" `)

		if strings.HasPrefix(indexdef, "CREATE UNIQUE INDEX") {
			indexType = UniqueType
		} else {
			indexType = IndexType
		}
		cs := strings.Split(indexdef, "(")
		colNames = strings.Split(cs[1][0:len(cs[1])-1], ",")

		if strings.HasSuffix(indexName, "_pkey") {
			continue
		}
		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "UQE_"+tableName) {
			newIdxName := indexName[5+len(tableName) : len(indexName)]
			if newIdxName != "" {
				indexName = newIdxName
			}
		}

		index := &Index{Name: indexName, Type: indexType, Cols: make([]string, 0)}
		for _, colName := range colNames {
			index.Cols = append(index.Cols, strings.Trim(colName, `" `))
		}
		indexes[index.Name] = index
	}
	return indexes, nil
}

// PgSeqFilter filter SQL replace ?, ? ... to $1, $2 ...
type PgSeqFilter struct {
}

func (s *PgSeqFilter) Do(sql string, dialect Dialect, table *Table) string {
	segs := strings.Split(sql, "?")
	size := len(segs)
	res := ""
	for i, c := range segs {
		if i < size-1 {
			res += c + fmt.Sprintf("$%v", i+1)
		}
	}
	res += segs[size-1]
	return res
}

func (db *postgres) Filters() []Filter {
	return []Filter{&IdFilter{}, &QuoteFilter{}, &PgSeqFilter{}}
}
