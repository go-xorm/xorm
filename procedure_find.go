package xorm

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

func find(this *Procedure, sliceValue reflect.Value) (err error) {
	//拼接SQL语句
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

	elem := sliceValue.Type().Elem() //切片中结构体
	//fmt.Println("elem:", elem)
	numField := elem.NumField() //结构体中字段个数
	//fmt.Println("numField:", numField)
	elemKind := elem.Kind() //切片的类型
	switch elemKind {
	case reflect.Struct:
		var sqlMap map[string]string
		var elemStruct reflect.Value

		values := make([]reflect.Value, 0) //定义一个 reflect.Value 的切片

		//查询到的数组数据，遍历赋值
		for i := 0; i < lens; i++ {
			sqlMap = results[i]

			elemStruct = reflect.New(elem).Elem() //new一个elem类型的结构体

			//将每一个查询到的条目，赋值到结构体
			for i := 0; i < numField; i++ {
				field := elem.Field(i)
				fieldName := field.Name
				fieldType := field.Type
				//fmt.Printf("name:%v    type:%v \n", fieldName, fieldType)
				tag := field.Tag.Get("xorm") //获取jorm的tag
				var column string
				if tag != "" {
					column = tag
				} else {
					column = convertColumn(fieldName)
				}

				if sqlMap[column] != "" {
					value, err := convertValue(fieldType, sqlMap[column]) //搜索出来的值转换成 reflect.Value 值
					if err != nil {
						fmt.Println("err:", err)
					}
					//fmt.Println("value:", value)
					elemStruct.Field(i).Set(value) //给elem类型结构体的每一个属性阻断赋值
				}
			}

			//fmt.Println("elemStruct:", elemStruct.Elem())
			values = append(values, elemStruct) //将每一个elem类型的结构体的值，添加到 reflect.Value 的切片
		}

		reflectValues := reflect.Append(sliceValue, values...) //将 reflect.Value 的切片 添加到sliceValue,得到一个reflectValue

		sliceValue.Set(reflectValues) //sliceValue 赋值值
	case reflect.Ptr:
		return errors.New("slice does not support pointer")
	}
	return nil
}
