package xorm

import (
	"database/sql"
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
	s := "SELECT column_name, column_default, is_nullable, data_type, character_maximum_length" +
		", numeric_precision, numeric_precision_radix FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = $1"

	cnn, err := sql.Open(db.drivername, db.dataSourceName)
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	res, err := query(cnn, s, args...)
	if err != nil {
		return nil, err
	}
	cols := make(map[string]*Column)
	for _, record := range res {
		col := new(Column)
		col.Indexes = make(map[string]bool)
		for name, content := range record {
			switch name {
			case "column_name":
				col.Name = strings.Trim(string(content), `" `)
			case "column_default":
				if strings.HasPrefix(string(content), "nextval") {
					col.IsPrimaryKey = true
				} else {
					col.Default = string(content)
				}
			case "is_nullable":
				if string(content) == "YES" {
					col.Nullable = true
				} else {
					col.Nullable = false
				}
			case "data_type":
				ct := string(content)
				switch ct {
				case "character varying", "character":
					col.SQLType = SQLType{Varchar, 0, 0}
				case "timestamp without time zone":
					col.SQLType = SQLType{DateTime, 0, 0}
				case "double precision":
					col.SQLType = SQLType{Double, 0, 0}
				case "boolean":
					col.SQLType = SQLType{Bool, 0, 0}
				case "time without time zone":
					col.SQLType = SQLType{Time, 0, 0}
				default:
					col.SQLType = SQLType{strings.ToUpper(ct), 0, 0}
				}
				if _, ok := sqlTypes[col.SQLType.Name]; !ok {
					return nil, errors.New(fmt.Sprintf("unkonw colType %v", ct))
				}
			case "character_maximum_length":
				i, err := strconv.Atoi(string(content))
				if err != nil {
					return nil, errors.New("retrieve length error")
				}
				col.Length = i
			case "numeric_precision":
			case "numeric_precision_radix":
			}
		}
		if col.SQLType.IsText() {
			if col.Default != "" {
				col.Default = "'" + col.Default + "'"
			}
		}
		cols[col.Name] = col
	}

	return cols, nil
}

func (db *postgres) GetTables() ([]*Table, error) {
	args := []interface{}{}
	s := "SELECT tablename FROM pg_tables where schemaname = 'public'"
	cnn, err := sql.Open(db.drivername, db.dataSourceName)
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	res, err := query(cnn, s, args...)
	if err != nil {
		return nil, err
	}

	tables := make([]*Table, 0)
	for _, record := range res {
		table := new(Table)
		for name, content := range record {
			switch name {
			case "tablename":
				table.Name = string(content)
			}
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *postgres) GetIndexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := "SELECT tablename, indexname, indexdef FROM pg_indexes WHERE schemaname = 'public' and tablename = $1"

	cnn, err := sql.Open(db.drivername, db.dataSourceName)
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	res, err := query(cnn, s, args...)
	if err != nil {
		return nil, err
	}

	indexes := make(map[string]*Index, 0)
	for _, record := range res {
		var indexType int
		var indexName string
		var colNames []string

		for name, content := range record {
			switch name {
			case "indexname":
				indexName = strings.Trim(string(content), `" `)
			case "indexdef":
				c := string(content)
				if strings.HasPrefix(c, "CREATE UNIQUE INDEX") {
					indexType = UniqueType
				} else {
					indexType = IndexType
				}
				cs := strings.Split(c, "(")
				colNames = strings.Split(cs[1][0:len(cs[1])-1], ",")
			}
		}
		if strings.HasSuffix(indexName, "_pkey") {
			continue
		}
		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "QUE_"+tableName) {
			indexName = indexName[5+len(tableName) : len(indexName)]
		}

		index := &Index{Name: indexName, Type: indexType, Cols: make([]string, 0)}
		for _, colName := range colNames {
			index.Cols = append(index.Cols, strings.Trim(colName, `" `))
		}
		indexes[index.Name] = index
	}
	return indexes, nil
}
