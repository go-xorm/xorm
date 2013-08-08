// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

package xorm

type sqlite3 struct {
}

func (db *sqlite3) SqlType(c *Column) string {
	switch t := c.SQLType; t {
	case Date, DateTime, TimeStamp, Time:
		return Numeric.Name
	case Char, Varchar, TinyText, Text, MediumText, LongText:
		return Text.Name
	case Bit, TinyInt, SmallInt, MediumInt, Int, Integer, BigInt, Bool:
		return Integer.Name
	case Float, Double, Real:
		return Real.Name
	case Decimal, Numeric:
		return Numeric.Name
	case TinyBlob, Blob, MediumBlob, LongBlob, Bytea, Binary, VarBinary:
		return Blob.Name
	case Serial, BigSerial:
		c.IsAutoIncrement = true
		return Integer.Name
	default:
		return t.Name
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
