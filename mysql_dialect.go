package xorm

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-xorm/core"
)

// func init() {
// 	RegisterDialect("mysql", &mysql{})
// }

var (
	reservedWords = map[string]bool{
		"ADD":true,
		"ALL":true,
		"ALTER":true,
"ANALYZE":true,
"AND":true,
"AS": true,
"ASC":true,
"ASENSITIVE":true,
"BEFORE":true,
"BETWEEN":true,
"BIGINT":true,
"BINARY":true,
"BLOB":true,
"BOTH":true,
"BY":true,
"CALL":true,
"CASCADE":true,
"CASE":true,
"CHANGE":true,
"CHAR":true,
"CHARACTER":true,
"CHECK":true,
"COLLATE":true,
"COLUMN":true,
"CONDITION":true,
"CONNECTION":true,
"CONSTRAINT":true,
"CONTINUE":true,
"CONVERT":true,
"CREATE":true,
"CROSS":true,
"CURRENT_DATE":true,
"CURRENT_TIME":true,
"CURRENT_TIMESTAMP":true,
"CURRENT_USER":true,
"CURSOR":true,
"DATABASE":true,
"DATABASES":true,
"DAY_HOUR":true,
"DAY_MICROSECOND":true,
"DAY_MINUTE":true,
"DAY_SECOND":true,
DEC	DECIMAL	DECLARE
DEFAULT	DELAYED	DELETE
DESC	DESCRIBE	DETERMINISTIC
DISTINCT	DISTINCTROW	DIV
DOUBLE	DROP	DUAL
EACH	ELSE	ELSEIF
ENCLOSED	ESCAPED	EXISTS
EXIT	EXPLAIN	FALSE
FETCH	FLOAT	FLOAT4
FLOAT8	FOR	FORCE
FOREIGN	FROM	FULLTEXT
GOTO	GRANT	GROUP
HAVING	HIGH_PRIORITY	HOUR_MICROSECOND
HOUR_MINUTE	HOUR_SECOND	IF
IGNORE	IN	INDEX
INFILE	INNER	INOUT
INSENSITIVE	INSERT	INT
INT1	INT2	INT3
INT4	INT8	INTEGER
INTERVAL	INTO	IS
ITERATE	JOIN	KEY
KEYS	KILL	LABEL
LEADING	LEAVE	LEFT
LIKE	LIMIT	LINEAR
LINES	LOAD	LOCALTIME
LOCALTIMESTAMP	LOCK	LONG
LONGBLOB	LONGTEXT	LOOP
LOW_PRIORITY	MATCH	MEDIUMBLOB
MEDIUMINT	MEDIUMTEXT	MIDDLEINT
MINUTE_MICROSECOND	MINUTE_SECOND	MOD
MODIFIES	NATURAL	NOT
NO_WRITE_TO_BINLOG	NULL	NUMERIC
ON	OPTIMIZE	OPTION
OPTIONALLY	OR	ORDER
OUT	OUTER	OUTFILE
PRECISION	PRIMARY	PROCEDURE
PURGE	RAID0	RANGE
READ	READS	REAL
REFERENCES	REGEXP	RELEASE
RENAME	REPEAT	REPLACE
REQUIRE	RESTRICT	RETURN
REVOKE	RIGHT	RLIKE
SCHEMA	SCHEMAS	SECOND_MICROSECOND
SELECT	SENSITIVE	SEPARATOR
SET	SHOW	SMALLINT
SPATIAL	SPECIFIC	SQL
SQLEXCEPTION	SQLSTATE	SQLWARNING
SQL_BIG_RESULT	SQL_CALC_FOUND_ROWS	SQL_SMALL_RESULT
SSL	STARTING	STRAIGHT_JOIN
TABLE	TERMINATED	THEN
TINYBLOB	TINYINT	TINYTEXT
TO	TRAILING	TRIGGER
TRUE	UNDO	UNION
UNIQUE	UNLOCK	UNSIGNED
UPDATE	USAGE	USE
USING	UTC_DATE	UTC_TIME
UTC_TIMESTAMP	VALUES	VARBINARY
"VARCHAR":true,
"VARCHARACTER":true,
"VARYING":true,
"WHEN":true,
"WHERE":true,
"WHILE":true,
"WITH":true,
"WRITE":true,
"X509":true,
"XOR":true,
"YEAR_MONTH":true,
"ZEROFILL":true,
	}
)

type mysql struct {
	core.Base
	net               string
	addr              string
	params            map[string]string
	loc               *time.Location
	timeout           time.Duration
	tls               *tls.Config
	allowAllFiles     bool
	allowOldPasswords bool
	clientFoundRows   bool
}

func (db *mysql) Init(d *core.DB, uri *core.Uri, drivername, dataSourceName string) error {
	return db.Base.Init(d, db, uri, drivername, dataSourceName)
}

func (db *mysql) SqlType(c *core.Column) string {
	var res string
	switch t := c.SQLType.Name; t {
	case core.Bool:
		res = core.TinyInt
		c.Length = 1
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
	case core.Bytea:
		res = core.Blob
	case core.TimeStampz:
		res = core.Char
		c.Length = 64
	case core.Enum: //mysql enum
		res = core.Enum
		res += "("
		opts := ""
		for v, _ := range c.EnumOptions {
			opts += fmt.Sprintf(",'%v'", v)
		}
		res += strings.TrimLeft(opts, ",")
		res += ")"
	case core.Set: //mysql set
		res = core.Set
		res += "("
		opts := ""
		for v, _ := range c.SetOptions {
			opts += fmt.Sprintf(",'%v'", v)
		}
		res += strings.TrimLeft(opts, ",")
		res += ")"
	default:
		res = t
	}

	var hasLen1 bool = (c.Length > 0)
	var hasLen2 bool = (c.Length2 > 0)

	if res == core.BigInt && !hasLen1 && !hasLen2 {
		c.Length = 20
		hasLen1 = true
	}

	if hasLen2 {
		res += "(" + strconv.Itoa(c.Length) + "," + strconv.Itoa(c.Length2) + ")"
	} else if hasLen1 {
		res += "(" + strconv.Itoa(c.Length) + ")"
	}
	return res
}

func (db *mysql) SupportInsertMany() bool {
	return true
}

func (db *mysql) QuoteStr() string {
	return "`"
}

func (db *mysql) SupportEngine() bool {
	return true
}

func (db *mysql) AutoIncrStr() string {
	return "AUTO_INCREMENT"
}

func (db *mysql) SupportCharset() bool {
	return true
}

func (db *mysql) IndexOnTable() bool {
	return true
}

func (db *mysql) IndexCheckSql(tableName, idxName string) (string, []interface{}) {
	args := []interface{}{db.DbName, tableName, idxName}
	sql := "SELECT `INDEX_NAME` FROM `INFORMATION_SCHEMA`.`STATISTICS`"
	sql += " WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ? AND `INDEX_NAME`=?"
	return sql, args
}

/*func (db *mysql) ColumnCheckSql(tableName, colName string) (string, []interface{}) {
	args := []interface{}{db.DbName, tableName, colName}
	sql := "SELECT `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ? AND `COLUMN_NAME` = ?"
	return sql, args
}*/

func (db *mysql) TableCheckSql(tableName string) (string, []interface{}) {
	args := []interface{}{db.DbName, tableName}
	sql := "SELECT `TABLE_NAME` from `INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA`=? and `TABLE_NAME`=?"
	return sql, args
}

func (db *mysql) GetColumns(tableName string) ([]string, map[string]*core.Column, error) {
	args := []interface{}{db.DbName, tableName}
	s := "SELECT `COLUMN_NAME`, `IS_NULLABLE`, `COLUMN_DEFAULT`, `COLUMN_TYPE`," +
		" `COLUMN_KEY`, `EXTRA` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols := make(map[string]*core.Column)
	colSeq := make([]string, 0)
	for rows.Next() {
		col := new(core.Column)
		col.Indexes = make(map[string]bool)

		var columnName, isNullable, colType, colKey, extra string
		var colDefault *string
		err = rows.Scan(&columnName, &isNullable, &colDefault, &colType, &colKey, &extra)
		if err != nil {
			return nil, nil, err
		}
		//fmt.Println(columnName, isNullable, colType, colKey, extra, colDefault)
		col.Name = strings.Trim(columnName, "` ")
		if "YES" == isNullable {
			col.Nullable = true
		}

		if colDefault != nil {
			col.Default = *colDefault
			if col.Default == "" {
				col.DefaultIsEmpty = true
			}
		}

		cts := strings.Split(colType, "(")
		colName := cts[0]
		colType = strings.ToUpper(colName)
		var len1, len2 int
		if len(cts) == 2 {
			idx := strings.Index(cts[1], ")")
			if colType == core.Enum && cts[1][0] == '\'' { //enum
				options := strings.Split(cts[1][0:idx], ",")
				col.EnumOptions = make(map[string]int)
				for k, v := range options {
					v = strings.TrimSpace(v)
					v = strings.Trim(v, "'")
					col.EnumOptions[v] = k
				}
			} else if colType == core.Set && cts[1][0] == '\'' {
				options := strings.Split(cts[1][0:idx], ",")
				col.SetOptions = make(map[string]int)
				for k, v := range options {
					v = strings.TrimSpace(v)
					v = strings.Trim(v, "'")
					col.SetOptions[v] = k
				}
			} else {
				lens := strings.Split(cts[1][0:idx], ",")
				len1, err = strconv.Atoi(strings.TrimSpace(lens[0]))
				if err != nil {
					return nil, nil, err
				}
				if len(lens) == 2 {
					len2, err = strconv.Atoi(lens[1])
					if err != nil {
						return nil, nil, err
					}
				}
			}
		}
		if colType == "FLOAT UNSIGNED" {
			colType = "FLOAT"
		}
		col.Length = len1
		col.Length2 = len2
		if _, ok := core.SqlTypes[colType]; ok {
			col.SQLType = core.SQLType{colType, len1, len2}
		} else {
			return nil, nil, errors.New(fmt.Sprintf("unkonw colType %v", colType))
		}

		if colKey == "PRI" {
			col.IsPrimaryKey = true
		}
		if colKey == "UNI" {
			//col.is
		}

		if extra == "auto_increment" {
			col.IsAutoIncrement = true
		}

		if col.SQLType.IsText() || col.SQLType.IsTime() {
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

func (db *mysql) GetTables() ([]*core.Table, error) {
	args := []interface{}{db.DbName}
	s := "SELECT `TABLE_NAME`, `ENGINE`, `TABLE_ROWS`, `AUTO_INCREMENT` from `INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA`=?"

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*core.Table, 0)
	for rows.Next() {
		table := core.NewEmptyTable()
		var name, engine, tableRows string
		var autoIncr *string
		err = rows.Scan(&name, &engine, &tableRows, &autoIncr)
		if err != nil {
			return nil, err
		}

		table.Name = name
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mysql) GetIndexes(tableName string) (map[string]*core.Index, error) {
	args := []interface{}{db.DbName, tableName}
	s := "SELECT `INDEX_NAME`, `NON_UNIQUE`, `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`STATISTICS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*core.Index, 0)
	for rows.Next() {
		var indexType int
		var indexName, colName, nonUnique string
		err = rows.Scan(&indexName, &nonUnique, &colName)
		if err != nil {
			return nil, err
		}

		if indexName == "PRIMARY" {
			continue
		}

		if "YES" == nonUnique || nonUnique == "1" {
			indexType = core.IndexType
		} else {
			indexType = core.UniqueType
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

func (db *mysql) Filters() []core.Filter {
	return []core.Filter{&core.IdFilter{}}
}
