// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

func find(this *Procedure, sliceValue reflect.Value) (err error) {

	buffer := new(bytes.Buffer)
	buffer.WriteString("call ")
	buffer.WriteString(this.funcName)
	buffer.WriteString(this.sql)
	sqlSlice := []interface{}{buffer.String()}
	sqlSlice = append(sqlSlice, this.inParams...)
	//fmt.Println("sql:", sqlSlice)

	//do SQL
	results, err := this.engine.QueryString(sqlSlice...)
	if err != nil {
		return err
	}
	lens := len(results)
	if lens <= 0 {
		sliceValue.Set(reflect.Zero(sliceValue.Type()))
		return nil
	}

	elem := sliceValue.Type().Elem() //
	//fmt.Println("elem:", elem)
	numField := elem.NumField() //
	//fmt.Println("numField:", numField)
	elemKind := elem.Kind() //
	switch elemKind {
	case reflect.Struct:
		var sqlMap map[string]string
		var elemStruct reflect.Value

		values := make([]reflect.Value, 0) //init reflect.Value's slice

		//results
		for i := 0; i < lens; i++ {
			sqlMap = results[i]

			elemStruct = reflect.New(elem).Elem() //new elem struct

			//item to struct
			for i := 0; i < numField; i++ {
				field := elem.Field(i)
				fieldName := field.Name
				fieldType := field.Type
				//fmt.Printf("name:%v    type:%v \n", fieldName, fieldType)
				tag := field.Tag.Get("xorm")
				var column string
				if tag != "" {
					column = tag
				} else {
					column = convertColumn(fieldName)
				}

				if sqlMap[column] != "" {
					value, err := convertValue(fieldType, sqlMap[column])
					if err != nil {
						fmt.Println("err:", err)
					}
					//fmt.Println("value:", value)
					elemStruct.Field(i).Set(value)
				}
			}

			//fmt.Println("elemStruct:", elemStruct.Elem())
			values = append(values, elemStruct)
		}

		reflectValues := reflect.Append(sliceValue, values...)

		sliceValue.Set(reflectValues) //sliceValue
	case reflect.Ptr:
		return errors.New("slice does not support pointer")
	}
	return nil
}
