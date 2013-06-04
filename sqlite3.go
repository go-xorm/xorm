package xorm

type sqlite3 struct {
}

func (db sqlite3) SqlType(c *Column) string {
	switch t := c.SQLType; t {
	case Date, DateTime, TimeStamp:
		return "NUMERIC"
	case Char, Varchar, Text:
		return "TEXT"
	case TinyInt, SmallInt, MediumInt, Int, BigInt:
		return "INTEGER"
	case Float, Double:
		return "REAL"
	case Decimal:
		return "NUMERIC"
	case Blob:
		return "BLOB"
	default:
		return t.Name
	}
}
