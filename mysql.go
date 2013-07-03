// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

package xorm

import "strconv"

type mysql struct {
}

func (db *mysql) SqlType(c *Column) string {
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

func (db *mysql) SupportInsertMany() bool {
	return true
}

func (db *mysql) QuoteIdentifier() string {
	return "`"
}

func (db *mysql) AutoIncrIdentifier() string {
	return "AUTO_INCREMENT"
}
