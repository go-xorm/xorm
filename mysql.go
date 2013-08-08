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
	var res string
	switch t := c.SQLType; t {
	case Bool:
		res = TinyInt.Name
	case Serial:
		c.IsAutoIncrement = true
		res = Int.Name
	case BigSerial:
		c.IsAutoIncrement = true
		res = Integer.Name
	case Bytea:
		res = Blob.Name
	default:
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

func (db *mysql) SupportInsertMany() bool {
	return true
}

func (db *mysql) QuoteStr() string {
	return "`"
}

func (db *mysql) AutoIncrStr() string {
	return "AUTO_INCREMENT"
}

func (db *mysql) SupportEngine() bool {
	return true
}

func (db *mysql) SupportCharset() bool {
	return true
}
