package xorm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-xorm/core"
)

// func init() {
// 	RegisterDialect("mssql", &mssql{})
// }

type mssql struct {
	core.Base
}

func (db *mssql) Init(d *core.DB, uri *core.Uri, drivername, dataSourceName string) error {
	return db.Base.Init(d, db, uri, drivername, dataSourceName)
}

func (db *mssql) SqlType(c *core.Column) string {
	var res string
	switch t := c.SQLType.Name; t {
	case core.Bool:
		res = core.TinyInt
	case core.Serial:
		c.IsAutoIncrement = true
		c.IsPrimaryKey = true
		c.Nullable = false
		res = core.Int
	case core.BigSerial:
		c.IsAutoIncrement = true
		c.IsPrimaryKey = true
		c.Nullable = false
		res = core.BigInt
	case core.Bytea, core.Blob, core.Binary, core.TinyBlob, core.MediumBlob, core.LongBlob:
		res = core.VarBinary
		if c.Length == 0 {
			c.Length = 50
		}
	case core.TimeStamp:
		res = core.DateTime
	case core.TimeStampz:
		res = "DATETIMEOFFSET"
		c.Length = 7
	case core.MediumInt:
		res = core.Int
	case core.MediumText, core.TinyText, core.LongText:
		res = core.Text
	case core.Double:
		res = core.Real
	default:
		res = t
	}

	if res == core.Int {
		return core.Int
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

func (db *mssql) SupportInsertMany() bool {
	return true
}

func (db *mssql) QuoteStr() string {
	return "\""
}

func (db *mssql) SupportEngine() bool {
	return false
}

func (db *mssql) AutoIncrStr() string {
	return "IDENTITY"
}

func (db *mssql) DropTableSql(tableName string) string {
	return fmt.Sprintf("IF EXISTS (SELECT * FROM sysobjects WHERE id = "+
		"object_id(N'%s') and OBJECTPROPERTY(id, N'IsUserTable') = 1) "+
		"DROP TABLE \"%s\"", tableName, tableName)
}

func (db *mssql) SupportCharset() bool {
	return false
}

func (db *mssql) IndexOnTable() bool {
	return true
}

func (db *mssql) IndexCheckSql(tableName, idxName string) (string, []interface{}) {
	args := []interface{}{idxName}
	sql := "select name from sysindexes where id=object_id('" + tableName + "') and name=?"
	return sql, args
}

/*func (db *mssql) ColumnCheckSql(tableName, colName string) (string, []interface{}) {
	args := []interface{}{tableName, colName}
	sql := `SELECT "COLUMN_NAME" FROM "INFORMATION_SCHEMA"."COLUMNS" WHERE "TABLE_NAME" = ? AND "COLUMN_NAME" = ?`
	return sql, args
}*/

func (db *mssql) IsColumnExist(tableName string, col *core.Column) (bool, error) {
	query := `SELECT "COLUMN_NAME" FROM "INFORMATION_SCHEMA"."COLUMNS" WHERE "TABLE_NAME" = ? AND "COLUMN_NAME" = ?`

	return db.HasRecords(query, tableName, col.Name)
}

func (db *mssql) TableCheckSql(tableName string) (string, []interface{}) {
	args := []interface{}{}
	sql := "select * from sysobjects where id = object_id(N'" + tableName + "') and OBJECTPROPERTY(id, N'IsUserTable') = 1"
	return sql, args
}

func (db *mssql) GetColumns(tableName string) ([]string, map[string]*core.Column, error) {
	args := []interface{}{}
	s := `select a.name as name, b.name as ctype,a.max_length,a.precision,a.scale
from sys.columns a left join sys.types b on a.user_type_id=b.user_type_id
where a.object_id=object_id('` + tableName + `')`

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols := make(map[string]*core.Column)
	colSeq := make([]string, 0)
	for rows.Next() {
		var name, ctype, precision, scale string
		var maxLen int
		err = rows.Scan(&name, &ctype, &maxLen, &precision, &scale)
		if err != nil {
			return nil, nil, err
		}

		col := new(core.Column)
		col.Indexes = make(map[string]bool)
		col.Length = maxLen
		col.Name = strings.Trim(name, "` ")

		ct := strings.ToUpper(ctype)
		switch ct {
		case "DATETIMEOFFSET":
			col.SQLType = core.SQLType{core.TimeStampz, 0, 0}
		case "NVARCHAR":
			col.SQLType = core.SQLType{core.Varchar, 0, 0}
		case "IMAGE":
			col.SQLType = core.SQLType{core.VarBinary, 0, 0}
		default:
			if _, ok := core.SqlTypes[ct]; ok {
				col.SQLType = core.SQLType{ct, 0, 0}
			} else {
				return nil, nil, errors.New(fmt.Sprintf("unknow colType %v for %v - %v",
					ct, tableName, col.Name))
			}
		}

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

func (db *mssql) GetTables() ([]*core.Table, error) {
	args := []interface{}{}
	s := `select name from sysobjects where xtype ='U'`

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*core.Table, 0)
	for rows.Next() {
		table := core.NewEmptyTable()
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		table.Name = strings.Trim(name, "` ")
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mssql) GetIndexes(tableName string) (map[string]*core.Index, error) {
	args := []interface{}{tableName}
	s := `SELECT
IXS.NAME                    AS  [INDEX_NAME],
C.NAME                      AS  [COLUMN_NAME],
IXS.is_unique AS [IS_UNIQUE]
FROM SYS.INDEXES IXS
INNER JOIN SYS.INDEX_COLUMNS   IXCS
ON IXS.OBJECT_ID=IXCS.OBJECT_ID  AND IXS.INDEX_ID = IXCS.INDEX_ID
INNER   JOIN SYS.COLUMNS C  ON IXS.OBJECT_ID=C.OBJECT_ID
AND IXCS.COLUMN_ID=C.COLUMN_ID
WHERE IXS.TYPE_DESC='NONCLUSTERED' and OBJECT_NAME(IXS.OBJECT_ID) =?
`

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*core.Index, 0)
	for rows.Next() {
		var indexType int
		var indexName, colName, isUnique string

		err = rows.Scan(&indexName, &colName, &isUnique)
		if err != nil {
			return nil, err
		}

		i, err := strconv.ParseBool(isUnique)
		if err != nil {
			return nil, err
		}

		if i {
			indexType = core.UniqueType
		} else {
			indexType = core.IndexType
		}

		colName = strings.Trim(colName, "` ")

		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "UQE_"+tableName) {
			indexName = indexName[5+len(tableName) : len(indexName)]
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

func (db *mssql) CreateTableSql(table *core.Table, tableName, storeEngine, charset string) string {
	var sql string
	if tableName == "" {
		tableName = table.Name
	}

	sql = "IF NOT EXISTS (SELECT [name] FROM sys.tables WHERE [name] = '" + tableName + "' ) CREATE TABLE "

	sql += db.QuoteStr() + tableName + db.QuoteStr() + " ("

	pkList := table.PrimaryKeys

	for _, colName := range table.ColumnsSeq() {
		col := table.GetColumn(colName)
		if col.IsPrimaryKey && len(pkList) == 1 {
			sql += col.String(db)
		} else {
			sql += col.StringNoPk(db)
		}
		sql = strings.TrimSpace(sql)
		sql += ", "
	}

	if len(pkList) > 1 {
		sql += "PRIMARY KEY ( "
		sql += strings.Join(pkList, ",")
		sql += " ), "
	}

	sql = sql[:len(sql)-2] + ")"
	sql += ";"
	return sql
}

func (db *mssql) Filters() []core.Filter {
	return []core.Filter{&core.IdFilter{}, &core.QuoteFilter{}}
}
