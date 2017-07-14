// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-xorm/builder"
	"github.com/go-xorm/core"
)

// Exist returns true if the record exist otherwise return false
func (session *Session) Exist(bean ...interface{}) (bool, error) {
	defer session.resetStatement()
	if session.IsAutoClose {
		defer session.Close()
	}

	var sqlStr string
	var args []interface{}
	var err error

	if session.Statement.RawSQL == "" {
		if len(bean) == 0 {
			tableName := session.Statement.TableName()
			if len(tableName) <= 0 {
				return false, ErrTableNotFound
			}

			if session.Statement.cond.IsValid() {
				condSQL, condArgs, err := builder.ToSQL(session.Statement.cond)
				if err != nil {
					return false, err
				}

				sqlStr = fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1", tableName, condSQL)
				args = condArgs
			} else {
				sqlStr = fmt.Sprintf("SELECT * FROM %s LIMIT 1", tableName)
				args = []interface{}{}
			}
		} else {
			beanValue := reflect.ValueOf(bean[0])
			if beanValue.Kind() != reflect.Ptr {
				return false, errors.New("needs a pointer")
			}

			if beanValue.Elem().Kind() == reflect.Struct {
				if err := session.Statement.setRefValue(beanValue.Elem()); err != nil {
					return false, err
				}
			}

			if len(session.Statement.TableName()) <= 0 {
				return false, ErrTableNotFound
			}
			session.Statement.Limit(1)
			sqlStr, args, err = session.Statement.genGetSQL(bean[0])
			if err != nil {
				return false, err
			}
		}
	} else {
		sqlStr = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	session.queryPreprocess(&sqlStr, args...)

	var rawRows *core.Rows
	if session.IsAutoCommit {
		_, rawRows, err = session.innerQuery(sqlStr, args...)
	} else {
		rawRows, err = session.Tx.Query(sqlStr, args...)
	}
	if err != nil {
		return false, err
	}

	defer rawRows.Close()

	return rawRows.Next(), nil
}
