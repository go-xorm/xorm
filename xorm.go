// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

package xorm

import (
	//"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"
	//"time"
)

const (
	version string = "0.1.6"
)

func NewEngine(driverName string, dataSourceName string) (*Engine, error) {
	engine := &Engine{ShowSQL: false, DriverName: driverName, Mapper: SnakeMapper{},
		DataSourceName: dataSourceName}

	engine.Tables = make(map[reflect.Type]*Table)
	engine.mutex = &sync.Mutex{}
	//engine.InsertMany = true
	engine.TagIdentifier = "xorm"
	//engine.QuoteIdentifier = "`"
	if driverName == SQLITE {
		engine.Dialect = &sqlite3{}
		//engine.AutoIncrement = "AUTOINCREMENT"
		//engine.Pool = NoneConnectPool{}
	} else if driverName == MYSQL {
		engine.Dialect = &mysql{}
		//engine.AutoIncrement = "AUTO_INCREMENT"
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupported driver name: %v", driverName))
	}

	//engine.Pool = NewSimpleConnectPool()
	//engine.Pool = NewNoneConnectPool()
	engine.pool = NewSysConnectPool()
	err := engine.pool.Init(engine)

	return engine, err
}
