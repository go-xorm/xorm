// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

/*
Package xorm is a simple and powerful ORM for Go. It makes
database operation simple.

First, we should new a engine for a database

	engine, err = xorm.NewEngine(driverName, dataSourceName)

Method NewEngine's parameters is the same as sql.Open. It depends
drivers' implementation. Generally, one engine is enough.

engine.Get(...)
engine.Insert(...)
engine.Find(...)
engine.Iterate(...)
engine.Delete(...)
*/
package xorm
