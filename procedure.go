// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"unicode"
)

type Procedure struct {
	engine   *Engine
	funcName string
	inLen    int
	outLen   int
	sql      string
	inParams []interface{}
}

//Input Parameter
//    engine：engine *Engine
//    funcName：procedure function name
//    inLen：input parameter length
//    outLen：output parameter length
func callProcedure(engine *Engine, funcName string, inLen, outLen int) (p *Procedure) {
	p = new(Procedure)
	p.engine = engine
	p.funcName = funcName
	p.inLen = inLen
	p.outLen = outLen

	buffer := new(bytes.Buffer)
	buffer.WriteString("(")
	if inLen == 0 {
		if outLen == 0 {
			buffer.WriteString(")")
			p.sql = buffer.String()
			return
		} else {
			for j := 0; j < outLen-1; j++ {
				buffer.WriteString("@out,")
			}

			buffer.WriteString("@out)")
			p.sql = buffer.String()
			return
		}
	} else {
		for i := 0; i < inLen-1; i++ {
			buffer.WriteString("?,")
		}

		if outLen == 0 {
			buffer.WriteString("?)")
			p.sql = buffer.String()
			return
		} else {
			buffer.WriteString("?,")
			for j := 0; j < outLen-1; j++ {
				buffer.WriteString("@out,")
			}
			buffer.WriteString("@out)")
			p.sql = buffer.String()
		}
	}
	return
}

//Input Parameter
//    inParams：Procedure input parameter
func (this *Procedure) InParams(inParams ...interface{}) (p *Procedure) {
	this.inParams = inParams
	return this
}

// Query runs a raw sql and return records as []map[string]string
func (this *Procedure) Query() (result []map[string]string, err error) {
	if len(this.inParams) != this.inLen {
		return nil, errors.New("inLen should be the same as the input parameter length")
	}
	buffer := new(bytes.Buffer)
	buffer.WriteString("call ")
	buffer.WriteString(this.funcName)
	buffer.WriteString(this.sql)

	sqlSlice := []interface{}{buffer.String()}
	sqlSlice = append(sqlSlice, this.inParams...)

	strings, err := this.engine.QueryString(sqlSlice...)
	if err != nil {
		return nil, err
	}
	return strings, nil
}

// Get retrieve one record from database, beanPtr's non-empty fields
// will be as conditions
func (this *Procedure) Get(beanPtr interface{}) (has bool, err error) {
	if len(this.inParams) != this.inLen {
		return false, errors.New("inLen should be the same as the input parameter length")
	}
	beanValue := reflect.ValueOf(beanPtr)
	if beanValue.Kind() != reflect.Ptr {
		return false, errors.New("needs a pointer to a value")
	} else if beanValue.Elem().Kind() != reflect.Struct {
		return false, errors.New("needs a struct to a pointer")
	}
	return get(this, beanValue.Elem())
}

// Find retrieve records from table, beanSlicePtr's non-empty fields
// are conditions. beans could be []Struct, []*Struct
func (this *Procedure) Find(beanSlicePtr interface{}) (err error) {
	if len(this.inParams) != this.inLen {
		return errors.New("inLen should be the same as the input parameter length")
	}
	beanSliceValue := reflect.ValueOf(beanSlicePtr)

	if beanSliceValue.Kind() != reflect.Ptr {
		return errors.New("needs a pointer to a value")
	}
	sliceValue := beanSliceValue.Elem()

	if sliceValue.Kind() != reflect.Slice {
		return errors.New("intut interface{} must be Slice")
	}
	elemKind := sliceValue.Type().Elem().Kind()

	if elemKind != reflect.Struct && elemKind != reflect.Ptr {
		return errors.New("slice must be struct or pointer struct")
	}
	return find(this, sliceValue)
}

//convert column name
func convertColumn(name string) (key string) {
	s := []rune(name)
	sLen := len(s)
	buffer := new(bytes.Buffer)
	var isFirst = true

	for i := 0; i < sLen; i++ {
		if unicode.IsUpper(s[i]) {
			if isFirst {
				lower := unicode.ToLower(s[i])
				buffer.WriteString(string(lower))
				isFirst = false
			} else {
				buffer.WriteString("_")
				lower := unicode.ToLower(s[i])
				buffer.WriteString(string(lower))
			}
		} else {
			buffer.WriteString(string(s[i]))
		}
	}
	key = buffer.String()
	return
}

//string to reflect.Value
func convertValue(fieldType reflect.Type, strValue string) (value reflect.Value, err error) {
	var result interface{}
	var defReturn = reflect.Zero(fieldType)
	switch fieldType.Kind() {
	case reflect.Int:
		result, err = strconv.Atoi(strValue)
		if err != nil {
			return defReturn, fmt.Errorf("convert %s to int error: %s", strValue, err.Error())
		}
	case reflect.Int32:
		i, err := strconv.Atoi(strValue)
		if err != nil {
			return defReturn, fmt.Errorf("convert %s to int32 error: %s", strValue, err.Error())
		}
		result = int32(i)
	case reflect.Int64:
		i, err := strconv.Atoi(strValue)
		if err != nil {
			return defReturn, fmt.Errorf("convert %s to int64 error: %s", strValue, err.Error())
		}
		result = int64(i)
	case reflect.Float32:
		f, err := strconv.ParseFloat(strValue, 32)
		if err != nil {
			return defReturn, fmt.Errorf("convert %s to float32 error: %s", strValue, err.Error())
		}
		result = float32(f)
	case reflect.Float64:
		f, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return defReturn, fmt.Errorf("convert %s to float64 error: %s", strValue, err.Error())
		}
		result = f
	case reflect.String:
		result = strValue
	default:
		return defReturn, errors.New("parameter type can not convert.")
	}
	return reflect.ValueOf(result).Convert(fieldType), nil
}
