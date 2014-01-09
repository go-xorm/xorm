package dialects

import (
	"strings"

	. "github.com/lunny/xorm/core"
)

func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

type sqlite3 struct {
	Base
}

func (db *sqlite3) Init(uri *Uri, drivername, dataSourceName string) error {
	return db.Base.Init(db, uri, drivername, dataSourceName)
}

func (db *sqlite3) SqlType(c *Column) string {
	switch t := c.SQLType.Name; t {
	case Date, DateTime, TimeStamp, Time:
		return Numeric
	case TimeStampz:
		return Text
	case Char, Varchar, TinyText, Text, MediumText, LongText:
		return Text
	case Bit, TinyInt, SmallInt, MediumInt, Int, Integer, BigInt, Bool:
		return Integer
	case Float, Double, Real:
		return Real
	case Decimal, Numeric:
		return Numeric
	case TinyBlob, Blob, MediumBlob, LongBlob, Bytea, Binary, VarBinary:
		return Blob
	case Serial, BigSerial:
		c.IsPrimaryKey = true
		c.IsAutoIncrement = true
		c.Nullable = false
		return Integer
	default:
		return t
	}
}

func (db *sqlite3) SupportInsertMany() bool {
	return true
}

func (db *sqlite3) QuoteStr() string {
	return "`"
}

func (db *sqlite3) AutoIncrStr() string {
	return "AUTOINCREMENT"
}

func (db *sqlite3) SupportEngine() bool {
	return false
}

func (db *sqlite3) SupportCharset() bool {
	return false
}

func (db *sqlite3) IndexOnTable() bool {
	return false
}

func (db *sqlite3) IndexCheckSql(tableName, idxName string) (string, []interface{}) {
	args := []interface{}{idxName}
	return "SELECT name FROM sqlite_master WHERE type='index' and name = ?", args
}

func (db *sqlite3) TableCheckSql(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ?", args
}

func (db *sqlite3) ColumnCheckSql(tableName, colName string) (string, []interface{}) {
	args := []interface{}{tableName}
	sql := "SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`" + colName + "`%') or (sql like '%[" + colName + "]%'))"
	return sql, args
}

func (db *sqlite3) GetColumns(tableName string) ([]string, map[string]*Column, error) {
	args := []interface{}{tableName}
	s := "SELECT sql FROM sqlite_master WHERE type='table' and name = ?"
	cnn, err := Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, nil, err
	}
	defer cnn.Close()

	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var name string
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return nil, nil, err
		}
	}

	nStart := strings.Index(name, "(")
	nEnd := strings.Index(name, ")")
	colCreates := strings.Split(name[nStart+1:nEnd], ",")
	cols := make(map[string]*Column)
	colSeq := make([]string, 0)
	for _, colStr := range colCreates {
		fields := strings.Fields(strings.TrimSpace(colStr))
		col := new(Column)
		col.Indexes = make(map[string]bool)
		col.Nullable = true
		for idx, field := range fields {
			if idx == 0 {
				col.Name = strings.Trim(field, "`[] ")
				continue
			} else if idx == 1 {
				col.SQLType = SQLType{field, 0, 0}
			}
			switch field {
			case "PRIMARY":
				col.IsPrimaryKey = true
			case "AUTOINCREMENT":
				col.IsAutoIncrement = true
			case "NULL":
				if fields[idx-1] == "NOT" {
					col.Nullable = false
				} else {
					col.Nullable = true
				}
			}
		}
		cols[col.Name] = col
		colSeq = append(colSeq, col.Name)
	}
	return colSeq, cols, nil
}

func (db *sqlite3) GetTables() ([]*Table, error) {
	args := []interface{}{}
	s := "SELECT name FROM sqlite_master WHERE type='table'"

	cnn, err := Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*Table, 0)
	for rows.Next() {
		table := NewEmptyTable()
		err = rows.Scan(&table.Name)
		if err != nil {
			return nil, err
		}
		if table.Name == "sqlite_sequence" {
			continue
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *sqlite3) GetIndexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := "SELECT sql FROM sqlite_master WHERE type='index' and tbl_name = ?"
	cnn, err := Open(db.DriverName(), db.DataSourceName())
	if err != nil {
		return nil, err
	}
	defer cnn.Close()
	rows, err := cnn.Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*Index, 0)
	for rows.Next() {
		var sql string
		err = rows.Scan(&sql)
		if err != nil {
			return nil, err
		}

		if sql == "" {
			continue
		}

		index := new(Index)
		nNStart := strings.Index(sql, "INDEX")
		nNEnd := strings.Index(sql, "ON")
		if nNStart == -1 || nNEnd == -1 {
			continue
		}

		indexName := strings.Trim(sql[nNStart+6:nNEnd], "` []")
		//fmt.Println(indexName)
		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "UQE_"+tableName) {
			index.Name = indexName[5+len(tableName) : len(indexName)]
		} else {
			index.Name = indexName
		}

		if strings.HasPrefix(sql, "CREATE UNIQUE INDEX") {
			index.Type = UniqueType
		} else {
			index.Type = IndexType
		}

		nStart := strings.Index(sql, "(")
		nEnd := strings.Index(sql, ")")
		colIndexes := strings.Split(sql[nStart+1:nEnd], ",")

		index.Cols = make([]string, 0)
		for _, col := range colIndexes {
			index.Cols = append(index.Cols, strings.Trim(col, "` []"))
		}
		indexes[index.Name] = index
	}

	return indexes, nil
}

func (db *sqlite3) Filters() []Filter {
	return []Filter{&IdFilter{}}
}
