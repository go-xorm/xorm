package xorm

import "strconv"

type mysql struct {
}

func (db mysql) SqlType(c *Column) string {
	switch t := c.SQLType; t {
	case Date, DateTime, TimeStamp:
		return "DATETIME"
	case Varchar:
		return t.Name + "(" + strconv.Itoa(c.Length) + ")"
	case Decimal:
		return t.Name + "(" + strconv.Itoa(c.Length) + "," + strconv.Itoa(c.Length2) + ")"
	default:
		return t.Name
	}
}
