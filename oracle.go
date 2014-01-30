package xorm

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type oracle struct {
	base
}

type oracleParser struct {
}

//dataSourceName=user/password@ipv4:port/dbname
//dataSourceName=user/password@[ipv6]:port/dbname
func (p *oracleParser) parse(driverName, dataSourceName string) (*uri, error) {
	db := &uri{dbType: ORACLE_OCI}
	dsnPattern := regexp.MustCompile(
		`^(?P<user>.*)\/(?P<password>.*)@` + // user:password@
			`(?P<net>.*)` + // ip:port
			`\/(?P<dbname>.*)`) // dbname
	matches := dsnPattern.FindStringSubmatch(dataSourceName)
	names := dsnPattern.SubexpNames()
	for i, match := range matches {
		switch names[i] {
		case "dbname":
			db.dbName = match
		}
	}
	if db.dbName == "" {
		return nil, errors.New("dbname is empty")
	}
	return db, nil
}

func (db *oracle) Init(drivername, uri string) error {
	return db.base.init(&oracleParser{}, drivername, uri)
}

func (db *oracle) SqlType(c *Column) string {
	var res string
	switch t := c.SQLType.Name; t {
	case Bit, TinyInt, SmallInt, MediumInt, Int, Integer, BigInt, Bool, Serial, BigSerial:
		return "NUMBER"
	case Binary, VarBinary, Blob, TinyBlob, MediumBlob, LongBlob, Bytea:
		return Blob
	case Time, DateTime, TimeStamp:
		res = TimeStamp
	case TimeStampz:
		res = "TIMESTAMP WITH TIME ZONE"
	case Float, Double, Numeric, Decimal:
		res = "NUMBER"
	case Text, MediumText, LongText:
		res = "CLOB"
	case Char, Varchar, TinyText:
		return "VARCHAR2"
	default:
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

func (db *oracle) SupportInsertMany() bool {
	return true
}

func (db *oracle) QuoteStr() string {
	return "\""
}

func (db *oracle) AutoIncrStr() string {
	return ""
}

func (db *oracle) SupportEngine() bool {
	return false
}

func (db *oracle) SupportCharset() bool {
	return false
}

func (db *oracle) IndexOnTable() bool {
	return false
}

func (db *oracle) IndexCheckSql(tableName, idxName string) (string, []interface{}) {
	args := []interface{}{strings.ToUpper(tableName), strings.ToUpper(idxName)}
	return `SELECT INDEX_NAME FROM USER_INDEXES ` +
		`WHERE TABLE_NAME = ? AND INDEX_NAME = ?`, args
}

func (db *oracle) TableCheckSql(tableName string) (string, []interface{}) {
	args := []interface{}{strings.ToUpper(tableName)}
	return `SELECT table_name FROM user_tables WHERE table_name = ?`, args
}

func (db *oracle) ColumnCheckSql(tableName, colName string) (string, []interface{}) {
	args := []interface{}{strings.ToUpper(tableName), strings.ToUpper(colName)}
	return "SELECT column_name FROM USER_TAB_COLUMNS WHERE table_name = ?" +
		" AND column_name = ?", args
}

func (db *oracle) GetColumns(tableName string) ([]string, map[string]*Column, error) {
	args := []interface{}{strings.ToUpper(tableName)}
	s := "SELECT column_name,data_default,data_type,data_length,data_precision,data_scale," +
		"nullable FROM USER_TAB_COLUMNS WHERE table_name = :1"

	cnn, err := sql.Open(db.driverName, db.dataSourceName)
	if err != nil {
		return nil, nil, err
	}
	defer cnn.Close()
	res, err := query(cnn, s, args...)
	if err != nil {
		return nil, nil, err
	}
	cols := make(map[string]*Column)
	colSeq := make([]string, 0)
	for _, record := range res {
		col := new(Column)
		col.Indexes = make(map[string]bool)
		for name, content := range record {
			switch name {
			case "column_name":
				col.Name = strings.Trim(string(content), `" `)
			case "data_default":
				col.Default = string(content)
				if col.Default == "" {
					col.DefaultIsEmpty = true
				}
			case "nullable":
				if string(content) == "Y" {
					col.Nullable = true
				} else {
					col.Nullable = false
				}
			case "data_type":
				ct := string(content)
				switch ct {
				case "VARCHAR2":
					col.SQLType = SQLType{Varchar, 0, 0}
				case "TIMESTAMP WITH TIME ZONE":
					col.SQLType = SQLType{TimeStamp, 0, 0}
				default:
					col.SQLType = SQLType{strings.ToUpper(ct), 0, 0}
				}
				if _, ok := sqlTypes[col.SQLType.Name]; !ok {
					return nil, nil, errors.New(fmt.Sprintf("unkonw colType %v", ct))
				}
			case "data_length":
				i, err := strconv.Atoi(string(content))
				if err != nil {
					return nil, nil, errors.New("retrieve length error")
				}
				col.Length = i
			case "data_precision":
			case "data_scale":
			}
		}
		if col.SQLType.IsText() {
			if col.Default != "" {
				col.Default = "'" + col.Default + "'"
			}else{
				if col.DefaultIsEmpty {
					col.Default = "''"
				}
			}
		}
		cols[col.Name] = col
		colSeq = append(colSeq, col.Name)
	}

	return colSeq, cols, nil
}

func (db *oracle) GetTables() ([]*Table, error) {
	args := []interface{}{}
	s := "SELECT table_name FROM user_tables"
	cnn, err := sql.Open(db.driverName, db.dataSourceName)
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
			case "table_name":
				table.Name = string(content)
			}
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *oracle) GetIndexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := "SELECT t.column_name,i.table_name,i.uniqueness,i.index_name FROM user_ind_columns t,user_indexes i " +
		"WHERE t.index_name = i.index_name and t.table_name = i.table_name and t.table_name =:1"

	cnn, err := sql.Open(db.driverName, db.dataSourceName)
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
		var colName string

		for name, content := range record {
			switch name {
			case "index_name":
				indexName = strings.Trim(string(content), `" `)
			case "uniqueness":
				c := string(content)
				if c == "UNIQUE" {
					indexType = UniqueType
				} else {
					indexType = IndexType
				}
			case "column_name":
				colName = string(content)
			}
		}

		var index *Index
		var ok bool
		if index, ok = indexes[indexName]; !ok {
			index = new(Index)
			index.Type = indexType
			index.Name = indexName
			indexes[indexName] = index
		}
		index.AddColumn(colName)
	}
	return indexes, nil
}
