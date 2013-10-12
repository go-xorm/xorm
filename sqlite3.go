package xorm

type sqlite3 struct {
	base
}

func (db *sqlite3) Init(drivername, dataSourceName string) error {
	db.base.init(drivername, dataSourceName)
	return nil
}

func (db *sqlite3) SqlType(c *Column) string {
	switch t := c.SQLType.Name; t {
	case Date, DateTime, TimeStamp, Time:
		return Numeric
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
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ? and sql like '%`" + colName + "`%'", args
}

func (db *sqlite3) GetColumns(tableName string) (map[string]*Column, error) {
	/*args := []interface{}{db.dbname, tableName}

	SELECT sql FROM sqlite_master WHERE type='table' and name = 'category';
	sql := "SELECT `COLUMN_NAME`, `IS_NULLABLE`, `COLUMN_DEFAULT`, `COLUMN_TYPE`," +
		" `COLUMN_KEY`, `EXTRA` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"

	return sql, args*/
	return nil, ErrNotImplemented
}

func (db *sqlite3) GetTables() ([]*Table, error) {
	return nil, ErrNotImplemented
}

func (db *sqlite3) GetIndexes(tableName string) (map[string]*Index, error) {
	return nil, ErrNotImplemented
}
