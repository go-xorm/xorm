package xorm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-xorm/core"
)

// func init() {
// 	RegisterDialect("oracle", &oracle{})
// }

type oracle struct {
	core.Base
}

func (db *oracle) Init(uri *core.Uri, drivername, dataSourceName string) error {
	return db.Base.Init(db, uri, drivername, dataSourceName)
}

func (db *oracle) SqlType(c *core.Column) string {
	var res string
	switch t := c.SQLType.Name; t {
	case core.Bit, core.TinyInt, core.SmallInt, core.MediumInt, core.Int, core.Integer, core.BigInt, core.Bool, core.Serial, core.BigSerial:
		return "NUMBER"
	case core.Binary, core.VarBinary, core.Blob, core.TinyBlob, core.MediumBlob, core.LongBlob, core.Bytea:
		return core.Blob
	case core.Time, core.DateTime, core.TimeStamp:
		res = core.TimeStamp
	case core.TimeStampz:
		res = "TIMESTAMP WITH TIME ZONE"
	case core.Float, core.Double, core.Numeric, core.Decimal:
		res = "NUMBER"
	case core.Text, core.MediumText, core.LongText:
		res = "CLOB"
	case core.Char, core.Varchar, core.TinyText:
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

func (db *oracle) GetColumns(tableName string) ([]string, map[string]*core.Column, error) {
	args := []interface{}{strings.ToUpper(tableName)}
	s := "SELECT column_name,data_default,data_type,data_length,data_precision,data_scale," +
		"nullable FROM USER_TAB_COLUMNS WHERE table_name = :1"

	cnn, err := core.Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, nil, err
	}
	defer cnn.Close()
	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols := make(map[string]*core.Column)
	colSeq := make([]string, 0)
	for rows.Next() {
		col := new(core.Column)
		col.Indexes = make(map[string]bool)

		var colName, colDefault, nullable, dataType, dataPrecision, dataScale string
		var dataLen int

		err = rows.Scan(&colName, &colDefault, &dataType, &dataLen, &dataPrecision,
			&dataScale, &nullable)
		if err != nil {
			return nil, nil, err
		}

		col.Name = strings.Trim(colName, `" `)
		col.Default = colDefault

		if nullable == "Y" {
			col.Nullable = true
		} else {
			col.Nullable = false
		}

		switch dataType {
		case "VARCHAR2":
			col.SQLType = core.SQLType{core.Varchar, 0, 0}
		case "TIMESTAMP WITH TIME ZONE":
			col.SQLType = core.SQLType{core.TimeStampz, 0, 0}
		default:
			col.SQLType = core.SQLType{strings.ToUpper(dataType), 0, 0}
		}
		if _, ok := core.SqlTypes[col.SQLType.Name]; !ok {
			return nil, nil, errors.New(fmt.Sprintf("unkonw colType %v", dataType))
		}

		col.Length = dataLen

		if col.SQLType.IsText() {
			if col.Default != "" {
				col.Default = "'" + col.Default + "'"
			} else {
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

func (db *oracle) GetTables() ([]*core.Table, error) {
	args := []interface{}{}
	s := "SELECT table_name FROM user_tables"
	cnn, err := core.Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, err
	}

	tables := make([]*core.Table, 0)
	for rows.Next() {
		table := core.NewEmptyTable()
		err = rows.Scan(&table.Name)
		if err != nil {
			return nil, err
		}

		tables = append(tables, table)
	}
	return tables, nil
}

func (db *oracle) GetIndexes(tableName string) (map[string]*core.Index, error) {
	args := []interface{}{tableName}
	s := "SELECT t.column_name,i.uniqueness,i.index_name FROM user_ind_columns t,user_indexes i " +
		"WHERE t.index_name = i.index_name and t.table_name = i.table_name and t.table_name =:1"

	cnn, err := core.Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*core.Index, 0)
	for rows.Next() {
		var indexType int
		var indexName, colName, uniqueness string

		err = rows.Scan(&colName, &uniqueness, &indexName)
		if err != nil {
			return nil, err
		}

		indexName = strings.Trim(indexName, `" `)

		if uniqueness == "UNIQUE" {
			indexType = core.UniqueType
		} else {
			indexType = core.IndexType
		}

		var index *core.Index
		var ok bool
		if index, ok = indexes[indexName]; !ok {
			index = new(core.Index)
			index.Type = indexType
			index.Name = indexName
			indexes[indexName] = index
		}
		index.AddColumn(colName)
	}
	return indexes, nil
}

// PgSeqFilter filter SQL replace ?, ? ... to :1, :2 ...
type OracleSeqFilter struct {
}

func (s *OracleSeqFilter) Do(sql string, dialect core.Dialect, table *core.Table) string {
	counts := strings.Count(sql, "?")
	for i := 1; i <= counts; i++ {
		newstr := ":" + fmt.Sprintf("%v", i)
		sql = strings.Replace(sql, "?", newstr, 1)
	}
	return sql
}

func (db *oracle) Filters() []core.Filter {
	return []core.Filter{&core.QuoteFilter{}, &OracleSeqFilter{}, &core.IdFilter{}}
}
