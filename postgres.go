// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

package xorm

import "strconv"

type postgres struct {
}

func (db *postgres) SqlType(c *Column) string {
	var res string
	switch t := c.SQLType; t {
	case TinyInt:
		res = SmallInt.Name
	case MediumInt, Int, Integer:
		return Integer.Name
	case Serial, BigSerial:
		c.IsAutoIncrement = true
		res = t.Name
	case Binary, VarBinary:
		res = Bytea.Name
	case DateTime:
		res = TimeStamp.Name
	case Float:
		res = Real.Name
	case TinyText, MediumText, LongText:
		res = Text.Name
	case Blob, TinyBlob, MediumBlob, LongBlob:
		res = Bytea.Name
	case Double:
		return "DOUBLE PRECISION"
	default:
		if c.IsAutoIncrement {
			return Serial.Name
		}
		res = t.Name
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
