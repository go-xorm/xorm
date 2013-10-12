package xorm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type postgres struct {
	base
	dbname string
}

type values map[string]string

func (vs values) Set(k, v string) {
	vs[k] = v
}

func (vs values) Get(k string) (v string) {
	return vs[k]
}

type Error error

func errorf(s string, args ...interface{}) {
	panic(Error(fmt.Errorf("pq: %s", fmt.Sprintf(s, args...))))
}

func parseOpts(name string, o values) {
	if len(name) == 0 {
		return
	}

	name = strings.TrimSpace(name)

	ps := strings.Split(name, " ")
	for _, p := range ps {
		kv := strings.Split(p, "=")
		if len(kv) < 2 {
			errorf("invalid option: %q", p)
		}
		o.Set(kv[0], kv[1])
	}
}

func (db *postgres) Init(drivername, uri string) error {
	db.base.init(drivername, uri)

	o := make(values)
	parseOpts(uri, o)

	db.dbname = o.Get("dbname")
	if db.dbname == "" {
		return errors.New("dbname is empty")
	}
	return nil
}

func (db *postgres) SqlType(c *Column) string {
	var res string
	switch t := c.SQLType.Name; t {
	case TinyInt:
		res = SmallInt
	case MediumInt, Int, Integer:
		return Integer
	case Serial, BigSerial:
		c.IsAutoIncrement = true
		c.Nullable = false
		res = t
	case Binary, VarBinary:
		return Bytea
	case DateTime:
		res = TimeStamp
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

func (db *postgres) GetColumns(tableName string) (map[string]*Column, error) {
	args := []interface{}{tableName}
	s := "SELECT COLUMN_NAME, column_default, is_nullable, data_type, character_maximum_length" +
		" FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ?"

	cnn, err := sql.Open(db.drivername, db.dataSourceName)
	if err != nil {
		return nil, err
	}
	res, err := query(cnn, s, args...)
	if err != nil {
		return nil, err
	}
	cols := make(map[string]*Column)
	for _, record := range res {
		col := new(Column)

		for name, content := range record {
			switch name {
			case "COLUMN_NAME":
				col.Name = string(content)
			case "column_default":
				if strings.HasPrefix(string(content), "") {
					col.IsPrimaryKey
				}
			}
		}
	}

	return nil, ErrNotImplemented
}

func (db *postgres) GetTables() ([]*Table, error) {
	return nil, ErrNotImplemented
}

func (db *postgres) GetIndexes(tableName string) (map[string]*Index, error) {
	return nil, ErrNotImplemented
}
