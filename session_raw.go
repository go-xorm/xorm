// Copyright 2016 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/go-xorm/core"
)

func rows2maps(rows *core.Rows) (resultsSlice []map[string][]byte, err error) {
	fields, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		result, err := row2map(rows, fields)
		if err != nil {
			return nil, err
		}
		resultsSlice = append(resultsSlice, result)
	}

	return resultsSlice, nil
}

func value2Bytes(rawValue *reflect.Value) (data []byte, err error) {
	var str string
	str, err = reflect2value(rawValue)
	if err != nil {
		return
	}
	data = []byte(str)
	return
}

func row2map(rows *core.Rows, fields []string) (resultsMap map[string][]byte, err error) {
	result := make(map[string][]byte)
	scanResultContainers := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		var scanResultContainer interface{}
		scanResultContainers[i] = &scanResultContainer
	}
	if err := rows.Scan(scanResultContainers...); err != nil {
		return nil, err
	}

	for ii, key := range fields {
		rawValue := reflect.Indirect(reflect.ValueOf(scanResultContainers[ii]))
		//if row is null then ignore
		if rawValue.Interface() == nil {
			//fmt.Println("ignore ...", key, rawValue)
			continue
		}

		if data, err := value2Bytes(&rawValue); err == nil {
			result[key] = data
		} else {
			return nil, err // !nashtsai! REVIEW, should return err or just error log?
		}
	}
	return result, nil
}

func rows2Strings(rows *core.Rows) (resultsSlice []map[string]string, err error) {
	fields, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		result, err := row2mapStr(rows, fields)
		if err != nil {
			return nil, err
		}
		resultsSlice = append(resultsSlice, result)
	}

	return resultsSlice, nil
}

func reflect2value(rawValue *reflect.Value) (str string, err error) {
	aa := reflect.TypeOf((*rawValue).Interface())
	vv := reflect.ValueOf((*rawValue).Interface())
	switch aa.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(vv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(vv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(vv.Float(), 'f', -1, 64)
	case reflect.String:
		str = vv.String()
	case reflect.Array, reflect.Slice:
		switch aa.Elem().Kind() {
		case reflect.Uint8:
			data := rawValue.Interface().([]byte)
			str = string(data)
		default:
			err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
		}
	// time type
	case reflect.Struct:
		if aa.ConvertibleTo(core.TimeType) {
			str = vv.Convert(core.TimeType).Interface().(time.Time).Format(time.RFC3339Nano)
		} else {
			err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
		}
	case reflect.Bool:
		str = strconv.FormatBool(vv.Bool())
	case reflect.Complex128, reflect.Complex64:
		str = fmt.Sprintf("%v", vv.Complex())
	/* TODO: unsupported types below
	   case reflect.Map:
	   case reflect.Ptr:
	   case reflect.Uintptr:
	   case reflect.UnsafePointer:
	   case reflect.Chan, reflect.Func, reflect.Interface:
	*/
	default:
		err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
	}
	return
}

func value2String(rawValue *reflect.Value) (data string, err error) {
	data, err = reflect2value(rawValue)
	if err != nil {
		return
	}
	return
}

func row2mapStr(rows *core.Rows, fields []string) (resultsMap map[string]string, err error) {
	result := make(map[string]string)
	scanResultContainers := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		var scanResultContainer interface{}
		scanResultContainers[i] = &scanResultContainer
	}
	if err := rows.Scan(scanResultContainers...); err != nil {
		return nil, err
	}

	for ii, key := range fields {
		rawValue := reflect.Indirect(reflect.ValueOf(scanResultContainers[ii]))
		// if row is null then as empty string
		if rawValue.Interface() == nil {
			result[key] = ""
			continue
		}

		if data, err := value2String(&rawValue); err == nil {
			result[key] = data
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (session *Session) queryPreprocess(sqlStr *string, paramStr ...interface{}) {
	for _, filter := range session.engine.dialect.Filters() {
		*sqlStr = filter.Do(*sqlStr, session.engine.dialect, session.statement.RefTable)
	}

	session.lastSQL = *sqlStr
	session.lastSQLArgs = paramStr
}

func (session *Session) queryRows(sqlStr string, args ...interface{}) (*core.Rows, error) {
	defer session.resetStatement()

	session.queryPreprocess(&sqlStr, args...)

	if session.engine.showSQL {
		if session.engine.showExecTime {
			b4ExecTime := time.Now()
			defer func() {
				execDuration := time.Since(b4ExecTime)
				if len(args) > 0 {
					session.engine.logger.Infof("[SQL] %s %#v - took: %v", sqlStr, args, execDuration)
				} else {
					session.engine.logger.Infof("[SQL] %s - took: %v", sqlStr, execDuration)
				}
			}()
		} else {
			if len(args) > 0 {
				session.engine.logger.Infof("[SQL] %v %#v", sqlStr, args)
			} else {
				session.engine.logger.Infof("[SQL] %v", sqlStr)
			}
		}
	}

	if session.isAutoCommit {
		if session.prepareStmt {
			// don't clear stmt since session will cache them
			stmt, err := session.doPrepare(sqlStr)
			if err != nil {
				return nil, err
			}

			rows, err := stmt.Query(args...)
			if err != nil {
				return nil, err
			}
			return rows, nil
		}

		rows, err := session.DB().Query(sqlStr, args...)
		if err != nil {
			return nil, err
		}
		return rows, nil
	}

	rows, err := session.tx.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (session *Session) queryRow(sqlStr string, args ...interface{}) *core.Row {
	return core.NewRow(session.queryRows(sqlStr, args...))
}

func (session *Session) queryBytes(sqlStr string, args ...interface{}) ([]map[string][]byte, error) {
	rows, err := session.queryRows(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows2maps(rows)
}

// Query runs a raw sql and return records as []map[string][]byte
func (session *Session) Query(sqlStr string, args ...interface{}) ([]map[string][]byte, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	return session.queryBytes(sqlStr, args...)
}

// QueryString runs a raw sql and return records as []map[string]string
func (session *Session) QueryString(sqlStr string, args ...interface{}) ([]map[string]string, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	rows, err := session.queryRows(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows2Strings(rows)
}

func (session *Session) exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	defer session.resetStatement()

	session.queryPreprocess(&sqlStr, args...)

	if session.engine.showSQL {
		if session.engine.showExecTime {
			b4ExecTime := time.Now()
			defer func() {
				execDuration := time.Since(b4ExecTime)
				if len(args) > 0 {
					session.engine.logger.Infof("[SQL] %s %#v - took: %v", sqlStr, args, execDuration)
				} else {
					session.engine.logger.Infof("[SQL] %s - took: %v", sqlStr, execDuration)
				}
			}()
		} else {
			if len(args) > 0 {
				session.engine.logger.Infof("[SQL] %v %#v", sqlStr, args)
			} else {
				session.engine.logger.Infof("[SQL] %v", sqlStr)
			}
		}
	}

	if !session.isAutoCommit {
		return session.tx.Exec(sqlStr, args...)
	}

	if session.prepareStmt {
		stmt, err := session.doPrepare(sqlStr)
		if err != nil {
			return nil, err
		}

		res, err := stmt.Exec(args...)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	return session.DB().Exec(sqlStr, args...)
}

// Exec raw sql
func (session *Session) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	return session.exec(sqlStr, args...)
}
