// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

// Warning: All contents in this file will be removed from xorm some times after
package xorm

// @deprecated : please use NewSession instead
func (engine *Engine) MakeSession() (Session, error) {
	s := engine.NewSession()
	return *s, nil
}

// @deprecated : please use NewEngine instead
func Create(driverName string, dataSourceName string) Engine {
	engine, _ := NewEngine(driverName, dataSourceName)
	return *engine
}
