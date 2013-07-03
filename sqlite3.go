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

func (db *sqlite3) SupportInsertMany() bool {
	return true
}

func (db *sqlite3) QuoteIdentifier() string {
	return "`"
}

func (db *sqlite3) AutoIncrIdentifier() string {
	return "AUTOINCREMENT"
}
