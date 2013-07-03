// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

package xorm

import (
	"fmt"
	"reflect"
	//"strconv"
	"strings"
	"time"
)

type Statement struct {
	RefTable     *Table
	Engine       *Engine
	Start        int
	LimitN       int
	WhereStr     string
	Params       []interface{}
	OrderStr     string
	JoinStr      string
	GroupByStr   string
	HavingStr    string
	ColumnStr    string
	AltTableName string
	RawSQL       string
	RawParams    []interface{}
	UseCascade   bool
	BeanArgs     []interface{}
}

func MakeArray(elem string, count int) []string {
	res := make([]string, count)
	for i := 0; i < count; i++ {
		res[i] = elem
	}
	return res
}

func (statement *Statement) Init() {
	statement.RefTable = nil
	statement.Start = 0
	statement.LimitN = 0
	statement.WhereStr = ""
	statement.Params = make([]interface{}, 0)
	statement.OrderStr = ""
	statement.UseCascade = true
	statement.JoinStr = ""
	statement.GroupByStr = ""
	statement.HavingStr = ""
	statement.ColumnStr = ""
	statement.AltTableName = ""
	statement.RawSQL = ""
	statement.RawParams = make([]interface{}, 0)
	statement.BeanArgs = make([]interface{}, 0)
}

func (statement *Statement) Sql(querystring string, args ...interface{}) {
	statement.RawSQL = querystring
	statement.RawParams = args
}

func (statement *Statement) Where(querystring string, args ...interface{}) {
	statement.WhereStr = querystring
	statement.Params = args
}

func (statement *Statement) Table(tableName string) {
	statement.AltTableName = tableName
}

func BuildConditions(engine *Engine, table *Table, bean interface{}) ([]string, []interface{}) {
	colNames := make([]string, 0)
	var args = make([]interface{}, 0)
	for _, col := range table.Columns {
		fieldValue := reflect.Indirect(reflect.ValueOf(bean)).FieldByName(col.FieldName)
		fieldType := reflect.TypeOf(fieldValue.Interface())
		val := fieldValue.Interface()
		switch fieldType.Kind() {
		case reflect.String:
			if fieldValue.String() == "" {
				continue
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			if fieldValue.Int() == 0 {
				continue
			}
		case reflect.Struct:
			if fieldType == reflect.TypeOf(time.Now()) {
				t := fieldValue.Interface().(time.Time)
				if t.IsZero() {
					continue
				}
			}
		default:
			continue
		}
		if table, ok := engine.Tables[fieldValue.Type()]; ok {
			pkField := reflect.Indirect(fieldValue).FieldByName(table.PKColumn().FieldName)
			fmt.Println(pkField.Interface())
			if pkField.Int() != 0 {
				args = append(args, pkField.Interface())
			} else {
				continue
			}
		} else {
			args = append(args, val)
		}
		colNames = append(colNames, fmt.Sprintf("%v%v%v = ?", engine.QuoteIdentifier(),
			col.Name, engine.QuoteIdentifier()))
	}

	return colNames, args
}

func (statement *Statement) TableName() string {
	if statement.AltTableName != "" {
		return statement.AltTableName
	}

	if statement.RefTable != nil {
		return statement.RefTable.Name
	}
	return ""
}

func (statement *Statement) Id(id int64) {
	if statement.WhereStr == "" {
		statement.WhereStr = "(id)=?"
		statement.Params = []interface{}{id}
	} else {
		statement.WhereStr = statement.WhereStr + " and (id)=?"
		statement.Params = append(statement.Params, id)
	}
}

func (statement *Statement) In(column string, args ...interface{}) {
	inStr := fmt.Sprintf("%v in (%v)", column, strings.Join(MakeArray("?", len(args)), ","))
	if statement.WhereStr == "" {
		statement.WhereStr = inStr
		statement.Params = args
	} else {
		statement.WhereStr = statement.WhereStr + " and " + inStr
		statement.Params = append(statement.Params, args...)
	}
}

func (statement *Statement) Limit(limit int, start ...int) {
	statement.LimitN = limit
	if len(start) > 0 {
		statement.Start = start[0]
	}
}

func (statement *Statement) OrderBy(order string) {
	statement.OrderStr = order
}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (statement *Statement) Join(join_operator, tablename, condition string) {
	if statement.JoinStr != "" {
		statement.JoinStr = statement.JoinStr + fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	} else {
		statement.JoinStr = fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	}
}

func (statement *Statement) GroupBy(keys string) {
	statement.GroupByStr = fmt.Sprintf("GROUP BY %v", keys)
}

func (statement *Statement) Having(conditions string) {
	statement.HavingStr = fmt.Sprintf("HAVING %v", conditions)
}

func (statement *Statement) genColumnStr(col *Column) string {
	sql := "`" + col.Name + "` "

	sql += statement.Engine.Dialect.SqlType(col) + " "

	if col.IsPrimaryKey {
		sql += "PRIMARY KEY "
	}

	if col.IsAutoIncrement {
		sql += statement.Engine.AutoIncrIdentifier() + " "
	}

	if col.Nullable {
		sql += "NULL "
	} else {
		sql += "NOT NULL "
	}

	if col.IsUnique {
		sql += "Unique "
	}

	if col.Default != "" {
		sql += "DEFAULT " + col.Default + " "
	}
	return sql
}

func (statement *Statement) selectColumnStr() string {
	table := statement.RefTable
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		if col.MapType != ONLYTODB {
			colNames = append(colNames, statement.TableName()+"."+col.Name)
		}
	}
	return strings.Join(colNames, ", ")
}

func (statement *Statement) genCreateSQL() string {
	sql := "CREATE TABLE IF NOT EXISTS `" + statement.TableName() + "` ("
	for _, col := range statement.RefTable.Columns {
		sql += statement.genColumnStr(&col)
		sql = strings.TrimSpace(sql)
		sql += ", "
	}
	sql = sql[:len(sql)-2] + ");"
	return sql
}

func (statement *Statement) genDropSQL() string {
	sql := "DROP TABLE IF EXISTS `" + statement.TableName() + "`;"
	return sql
}

func (statement Statement) generateSql() string {
	columnStr := statement.selectColumnStr()
	return statement.genSelectSql(columnStr)
}

func (statement Statement) genGetSql(bean interface{}) (string, []interface{}) {
	table := statement.Engine.AutoMap(bean)
	statement.RefTable = table

	colNames, args := BuildConditions(statement.Engine, table, bean)
	statement.ColumnStr = strings.Join(colNames, " and ")
	statement.BeanArgs = args

	return statement.generateSql(), append(statement.Params, statement.BeanArgs...)
}

func (statement Statement) genCountSql(bean interface{}) (string, []interface{}) {
	table := statement.Engine.AutoMap(bean)
	statement.RefTable = table

	colNames, args := BuildConditions(statement.Engine, table, bean)
	statement.ColumnStr = strings.Join(colNames, " and ")
	statement.BeanArgs = args
	return statement.genSelectSql("count(*) as total"), append(statement.Params, statement.BeanArgs...)
}

func (statement Statement) genSelectSql(columnStr string) (a string) {
	if statement.Engine.DriverName == MSSQL {
		if statement.Start > 0 {
			a = fmt.Sprintf("select ROW_NUMBER() OVER(order by %v )as rownum,%v from %v",
				statement.RefTable.PKColumn().Name,
				columnStr,
				statement.TableName())
			if statement.WhereStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
				if statement.ColumnStr != "" {
					a = fmt.Sprintf("%v and %v", a, statement.ColumnStr)
				}
			} else if statement.ColumnStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.ColumnStr)
			}
			a = fmt.Sprintf("select %v from (%v) "+
				"as a where rownum between %v and %v",
				columnStr,
				a,
				statement.Start,
				statement.LimitN)
		} else if statement.LimitN > 0 {
			a = fmt.Sprintf("SELECT top %v %v FROM %v", statement.LimitN, columnStr, statement.TableName())
			if statement.WhereStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
				if statement.ColumnStr != "" {
					a = fmt.Sprintf("%v and %v", a, statement.ColumnStr)
				}
			} else if statement.ColumnStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.ColumnStr)
			}
			if statement.GroupByStr != "" {
				a = fmt.Sprintf("%v %v", a, statement.GroupByStr)
			}
			if statement.HavingStr != "" {
				a = fmt.Sprintf("%v %v", a, statement.HavingStr)
			}
			if statement.OrderStr != "" {
				a = fmt.Sprintf("%v ORDER BY %v", a, statement.OrderStr)
			}
		} else {
			a = fmt.Sprintf("SELECT %v FROM %v", columnStr, statement.TableName())
			if statement.WhereStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
				if statement.ColumnStr != "" {
					a = fmt.Sprintf("%v and %v", a, statement.ColumnStr)
				}
			} else if statement.ColumnStr != "" {
				a = fmt.Sprintf("%v WHERE %v", a, statement.ColumnStr)
			}
			if statement.GroupByStr != "" {
				a = fmt.Sprintf("%v %v", a, statement.GroupByStr)
			}
			if statement.HavingStr != "" {
				a = fmt.Sprintf("%v %v", a, statement.HavingStr)
			}
			if statement.OrderStr != "" {
				a = fmt.Sprintf("%v ORDER BY %v", a, statement.OrderStr)
			}
		}
	} else {
		a = fmt.Sprintf("SELECT %v FROM %v", columnStr, statement.TableName())
		if statement.JoinStr != "" {
			a = fmt.Sprintf("%v %v", a, statement.JoinStr)
		}
		if statement.WhereStr != "" {
			a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
			if statement.ColumnStr != "" {
				a = fmt.Sprintf("%v and %v", a, statement.ColumnStr)
			}
		} else if statement.ColumnStr != "" {
			a = fmt.Sprintf("%v WHERE %v", a, statement.ColumnStr)
		}
		if statement.GroupByStr != "" {
			a = fmt.Sprintf("%v %v", a, statement.GroupByStr)
		}
		if statement.HavingStr != "" {
			a = fmt.Sprintf("%v %v", a, statement.HavingStr)
		}
		if statement.OrderStr != "" {
			a = fmt.Sprintf("%v ORDER BY %v", a, statement.OrderStr)
		}
		if statement.Start > 0 {
			a = fmt.Sprintf("%v LIMIT %v, %v", a, statement.Start, statement.LimitN)
		} else if statement.LimitN > 0 {
			a = fmt.Sprintf("%v LIMIT %v", a, statement.LimitN)
		}
	}
	return
}
