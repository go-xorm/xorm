package xorm

import (
	"bytes"
	"reflect"
)

//获取结果赋值到结构体
func get(this *Procedure, beanElem reflect.Value) (has bool, err error) {
	//拼接SQL语句
	buffer := new(bytes.Buffer)
	buffer.WriteString("call ")
	buffer.WriteString(this.funcName)
	buffer.WriteString(this.sql)
	sqlSlice := []interface{}{buffer.String()}
	sqlSlice = append(sqlSlice, this.inParams...)
	//fmt.Println("sql:", sqlSlice)
	//执行SQL请求
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

	numField := beanElem.NumField() //结构体中字段个数
	elemType := beanElem.Type()     //结构体类型
	var column string
	for i := 0; i < numField; i++ {
		field := elemType.Field(i) //遍历每一个字段
		fieldName := field.Name
		fieldType := field.Type
		tag := field.Tag.Get("xorm") //获取jorm的tag
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
