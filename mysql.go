package xorm

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type uri struct {
	dbType  string
	proto   string
	host    string
	port    string
	dbName  string
	user    string
	passwd  string
	charset string
	laddr   string
	raddr   string
	timeout time.Duration
}

type parser interface {
	parse(driverName, dataSourceName string) (*uri, error)
}

type mysqlParser struct {
}

func (p *mysqlParser) parse(driverName, dataSourceName string) (*uri, error) {
	//cfg.params = make(map[string]string)
	dsnPattern := regexp.MustCompile(
		`^(?:(?P<user>.*?)(?::(?P<passwd>.*))?@)?` + // [user[:password]@]
			`(?:(?P<net>[^\(]*)(?:\((?P<addr>[^\)]*)\))?)?` + // [net[(addr)]]
			`\/(?P<dbname>.*?)` + // /dbname
			`(?:\?(?P<params>[^\?]*))?$`) // [?param1=value1&paramN=valueN]
	matches := dsnPattern.FindStringSubmatch(dataSourceName)
	//tlsConfigRegister := make(map[string]*tls.Config)
	names := dsnPattern.SubexpNames()

	uri := &uri{dbType: MYSQL}

	for i, match := range matches {
		switch names[i] {
		case "dbname":
			uri.dbName = match
		}
	}
	return uri, nil
}

type base struct {
	parser         parser
	driverName     string
	dataSourceName string
	*uri
}

func (b *base) init(parser parser, drivername, dataSourceName string) (err error) {
	b.parser = parser
	b.driverName, b.dataSourceName = drivername, dataSourceName
	b.uri, err = b.parser.parse(b.driverName, b.dataSourceName)
	return
}

func (b *base) DBType() string {
	return b.uri.dbType
}

type mysql struct {
	base
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

func (db *mysql) Init(drivername, uri string) error {
	return db.base.init(&mysqlParser{}, drivername, uri)
}

func (db *mysql) SqlType(c *Column) string {
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
	case Bytea:
		res = Blob
	case TimeStampz:
		res = Char
		c.Length = 64
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
	args := []interface{}{db.dbName, tableName, idxName}
	sql := "SELECT `INDEX_NAME` FROM `INFORMATION_SCHEMA`.`STATISTICS`"
	sql += " WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ? AND `INDEX_NAME`=?"
	return sql, args
}

func (db *mysql) ColumnCheckSql(tableName, colName string) (string, []interface{}) {
	args := []interface{}{db.dbName, tableName, colName}
	sql := "SELECT `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ? AND `COLUMN_NAME` = ?"
	return sql, args
}

func (db *mysql) TableCheckSql(tableName string) (string, []interface{}) {
	args := []interface{}{db.dbName, tableName}
	sql := "SELECT `TABLE_NAME` from `INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA`=? and `TABLE_NAME`=?"
	return sql, args
}

func (db *mysql) GetColumns(tableName string) ([]string, map[string]*Column, error) {
	args := []interface{}{db.dbName, tableName}
	s := "SELECT `COLUMN_NAME`, `IS_NULLABLE`, `COLUMN_DEFAULT`, `COLUMN_TYPE`," +
		" `COLUMN_KEY`, `EXTRA` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"
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
			case "COLUMN_NAME":
				col.Name = strings.Trim(string(content), "` ")
			case "IS_NULLABLE":
				if "YES" == string(content) {
					col.Nullable = true
				}
			case "COLUMN_DEFAULT":
				// add ''
				col.Default = string(content)
			case "COLUMN_TYPE":
				cts := strings.Split(string(content), "(")
				var len1, len2 int
				if len(cts) == 2 {
					idx := strings.Index(cts[1], ")")
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
				colName := cts[0]
				colType := strings.ToUpper(colName)
				col.Length = len1
				col.Length2 = len2
				if _, ok := sqlTypes[colType]; ok {
					col.SQLType = SQLType{colType, len1, len2}
				} else {
					return nil, nil, errors.New(fmt.Sprintf("unkonw colType %v", colType))
				}
			case "COLUMN_KEY":
				key := string(content)
				if key == "PRI" {
					col.IsPrimaryKey = true
				}
				if key == "UNI" {
					//col.is
				}
			case "EXTRA":
				extra := string(content)
				if extra == "auto_increment" {
					col.IsAutoIncrement = true
				}
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

func (db *mysql) GetTables() ([]*Table, error) {
	args := []interface{}{db.dbName}
	s := "SELECT `TABLE_NAME`, `ENGINE`, `TABLE_ROWS`, `AUTO_INCREMENT` from `INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA`=?"
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
			case "TABLE_NAME":
				table.Name = strings.Trim(string(content), "` ")
			case "ENGINE":
			}
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mysql) GetIndexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{db.dbName, tableName}
	s := "SELECT `INDEX_NAME`, `NON_UNIQUE`, `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`STATISTICS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"
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
		var indexName, colName string
		for name, content := range record {
			switch name {
			case "NON_UNIQUE":
				if "YES" == string(content) || string(content) == "1" {
					indexType = IndexType
				} else {
					indexType = UniqueType
				}
			case "INDEX_NAME":
				indexName = string(content)
			case "COLUMN_NAME":
				colName = strings.Trim(string(content), "` ")
			}
		}
		if indexName == "PRIMARY" {
			continue
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
	}
	return indexes, nil
}
