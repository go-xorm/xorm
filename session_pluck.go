package xorm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Pluck retrieve a list of column values.
// colName is the column name where values from
// slicePtr should be a slice which will filled by the list values
// beans is optional which represents the table, should be a *Struct
func (session *Session) Pluck(colName string, slicePtr interface{}, beans ...interface{}) error {
	if session.isAutoClose {
		defer session.Close()
	}
	return session.pluck(colName, slicePtr, beans...)
}

func (session *Session) pluck(colName string, slicePtr interface{}, beans ...interface{}) error {
	sliceValue := reflect.Indirect(reflect.ValueOf(slicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return errors.New("needs a pointer to a slice")
	}
	sliceElemType := sliceValue.Type().Elem()
	rows := []map[string]interface{}{}
	query := session
	if session.statement.selectStr == "" {
		query.Select(colName)
	}
	if len(beans) != 0 {
		// find table
		bean := beans[0]
		beanValue := reflect.ValueOf(bean)
		if beanValue.Kind() != reflect.Ptr || beanValue.Elem().Kind() != reflect.Struct {
			return errors.New("needs a pointer to a struct")
		}
		session.statement.setRefValue(beanValue)
		tbName := session.statement.TableName()
		query.Table(tbName)
	}
	if err := query.Find(&rows); err != nil {
		return err
	}

	colSegs := strings.Split(colName, ".")
	col := colSegs[len(colSegs)-1] // last one
	col = strings.Trim(col, "`")   // unwrap column name
	nSlice := reflect.New(sliceValue.Type()).Elem()
	rowsLen := len(rows)
	for i := 0; i < rowsLen; i++ {
		if v, ok := rows[i][col]; !ok {
			return fmt.Errorf("cannot find column: %s", col)
		} else {
			nSlice = reflect.Append(nSlice, reflect.ValueOf(v).Convert(sliceElemType))
		}
	}
	sliceValue.Set(nSlice)
	return nil
}
