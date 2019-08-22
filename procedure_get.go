// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"bytes"
	"reflect"
)

func get(this *Procedure, beanElem reflect.Value) (has bool, err error) {

	buffer := new(bytes.Buffer)
	buffer.WriteString("call ")
	buffer.WriteString(this.funcName)
	buffer.WriteString(this.sql)
	sqlSlice := []interface{}{buffer.String()}
	sqlSlice = append(sqlSlice, this.inParams...)
	//fmt.Println("sql:", sqlSlice)
	//do Sql
	results, err := this.engine.QueryString(sqlSlice...)
	if err != nil {
		return false, err
	}
	if len(results) <= 0 {
		beanElem.Set(reflect.Zero(beanElem.Type()))
		return false, nil
	}
	result := results[0]

	elemStruct := reflect.New(beanElem.Type()).Elem()

	numField := beanElem.NumField()
	elemType := beanElem.Type()
	var column string
	for i := 0; i < numField; i++ {
		field := elemType.Field(i)
		fieldName := field.Name
		fieldType := field.Type
		tag := field.Tag.Get("xorm")
		if tag != "" {
			column = tag
		} else {
			column = convertColumn(fieldName)
		}
		if result[column] != "" {
			value, err := convertValue(fieldType, result[column])
			if err != nil {
				return false, err
			}
			elemStruct.Field(i).Set(value)
		}
	}

	beanElem.Set(elemStruct)
	return true, nil
}
