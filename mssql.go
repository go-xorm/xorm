package xorm

import (
	//"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	//"regexp"
	"strconv"
	"strings"
	//"time"
)

type mssql struct {
	base
	quoteFilter Filter
}

type mssqlParser struct {
}

func (p *mssqlParser) parse(driverName, dataSourceName string) (*uri, error) {
	return &uri{dbName: "xorm_test", dbType: MSSQL}, nil
}

func (db *mssql) Init(drivername, uri string) error {
	db.quoteFilter = &QuoteFilter{}
	return db.base.init(&mssqlParser{}, drivername, uri)
}

func (db *mssql) SqlType(c *Column) string {
	var res string
	switch t := c.SQLType.Name; t {
	case Bool:
		res = TinyInt
	case Serial:
		c.IsAutoIncrement = true
		c.IsPrimaryKey = true
		c.Nullable = false
		res = Int
	case BigSerial:
		c.IsAutoIncrement = true
		c.IsPrimaryKey = true
		c.Nullable = false
		res = BigInt
	case Bytea, Blob, Binary, TinyBlob, MediumBlob, LongBlob:
		res = VarBinary
		if c.Length == 0 {
			c.Length = 50
		}
	case TimeStamp:
		res = DateTime
	case TimeStampz:
		res = "DATETIMEOFFSET"
		c.Length = 7
	case MediumInt:
		res = Int
	case MediumText, TinyText, LongText:
		res = Text
	case Double:
		res = Real
	default:
		res = t
	}

	if res == Int {
		return Int
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

func (db *mssql) ColumnCheckSql(tableName, colName string) (string, []interface{}) {
	args := []interface{}{tableName, colName}
	sql := `SELECT "COLUMN_NAME" FROM "INFORMATION_SCHEMA"."COLUMNS" WHERE "TABLE_NAME" = ? AND "COLUMN_NAME" = ?`
	return sql, args
}

func (db *mssql) TableCheckSql(tableName string) (string, []interface{}) {
	args := []interface{}{}
	sql := "select * from sysobjects where id = object_id(N'" + tableName + "') and OBJECTPROPERTY(id, N'IsUserTable') = 1"
	return sql, args
}

func (db *mssql) GetColumns(tableName string) ([]string, map[string]*Column, error) {
	args := []interface{}{}
	s := `select a.name as name, b.name as ctype,a.max_length,a.precision,a.scale 
from sys.columns a left join sys.types b on a.user_type_id=b.user_type_id 
where a.object_id=object_id('` + tableName + `')`
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
			case "name":
				col.Name = strings.Trim(string(content), "` ")
			case "ctype":
				ct := strings.ToUpper(string(content))
				switch ct {
				case "DATETIMEOFFSET":
					col.SQLType = SQLType{TimeStampz, 0, 0}
				default:
					if _, ok := sqlTypes[ct]; ok {
						col.SQLType = SQLType{ct, 0, 0}
					} else {
						return nil, nil, errors.New(fmt.Sprintf("unknow colType %v", ct))
					}
				}

			case "max_length":
				len1, err := strconv.Atoi(strings.TrimSpace(string(content)))
				if err != nil {
					return nil, nil, err
				}
				col.Length = len1
			}
		}
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

func (db *mssql) GetTables() ([]*Table, error) {
	args := []interface{}{}
	s := `select name from sysobjects where xtype ='U'`
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
			case "name":
				table.Name = strings.Trim(string(content), "` ")
			}
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mssql) GetIndexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := `SELECT  
IXS.NAME                    AS  [INDEX_NAME],  
C.NAME                      AS  [COLUMN_NAME], 
IXS.is_unique AS [IS_UNIQUE], 
CASE    IXCS.IS_INCLUDED_COLUMN   
WHEN    0   THEN    'NONE' 
ELSE    'INCLUDED'  END     AS  [IS_INCLUDED_COLUMN]  
FROM SYS.INDEXES IXS  
INNER JOIN SYS.INDEX_COLUMNS   IXCS  
ON IXS.OBJECT_ID=IXCS.OBJECT_ID  AND IXS.INDEX_ID = IXCS.INDEX_ID  
INNER   JOIN SYS.COLUMNS C  ON IXS.OBJECT_ID=C.OBJECT_ID  
AND IXCS.COLUMN_ID=C.COLUMN_ID  
WHERE IXS.TYPE_DESC='NONCLUSTERED' and OBJECT_NAME(IXS.OBJECT_ID) =?
`
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
		fmt.Println("-----", record, "-----")
		var indexType int
		var indexName, colName string
		for name, content := range record {
			switch name {
			case "IS_UNIQUE":
				i, err := strconv.ParseBool(string(content))
				if err != nil {
					return nil, err
				}

				fmt.Println(name, string(content), i)

				if i {
					indexType = UniqueType
				} else {
					indexType = IndexType
				}
			case "INDEX_NAME":
				indexName = string(content)
			case "COLUMN_NAME":
				colName = strings.Trim(string(content), "` ")
			}
		}

		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "UQE_"+tableName) {
			indexName = indexName[5+len(tableName) : len(indexName)]
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
		fmt.Print("------end------")
	}
	return indexes, nil
}
