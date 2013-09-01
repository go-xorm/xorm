package xorm

import "strconv"

type postgres struct {
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
