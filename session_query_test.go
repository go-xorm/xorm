// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"xorm.io/builder"
	"xorm.io/core"

	"github.com/stretchr/testify/assert"
)

func TestQueryString(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type GetVar2 struct {
		Id      int64  `xorm:"autoincr pk"`
		Msg     string `xorm:"varchar(255)"`
		Age     int
		Money   float32
		Created time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(GetVar2)))

	var data = GetVar2{
		Msg:   "hi",
		Age:   28,
		Money: 1.5,
	}
	_, err := testEngine.InsertOne(data)
	assert.NoError(t, err)

	tableName := mapper.Obj2Table("GetVar2")

	records, err := testEngine.QueryString("select * from `" + testEngine.TableName(tableName, true) + "`")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(records))
	assert.Equal(t, 5, len(records[0]))

	idName := mapper.Obj2Table("Id")
	ageName := mapper.Obj2Table("Age")
	msgName := mapper.Obj2Table("Msg")
	moneyName := mapper.Obj2Table("Money")

	assert.Equal(t, "1", records[0][idName])
	assert.Equal(t, "hi", records[0][msgName])
	assert.Equal(t, "28", records[0][ageName])
	assert.Equal(t, "1.5", records[0][moneyName])
}

func TestQueryString2(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type GetVar3 struct {
		Id  int64 `xorm:"autoincr pk"`
		Msg bool  `xorm:"bit"`
	}

	assert.NoError(t, testEngine.Sync2(new(GetVar3)))

	var data = GetVar3{
		Msg: false,
	}
	_, err := testEngine.Insert(data)
	assert.NoError(t, err)

	tableName := "`" + mapper.Obj2Table("GetVar3") + "`"

	idName := mapper.Obj2Table("Id")
	msgName := mapper.Obj2Table("Msg")

	records, err := testEngine.QueryString("select * from " + testEngine.TableName(tableName, true))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(records))
	assert.Equal(t, 2, len(records[0]))
	assert.Equal(t, "1", records[0][idName])
	assert.True(t, "0" == records[0][msgName] || "false" == records[0][msgName])
}

func toString(i interface{}) string {
	switch i.(type) {
	case []byte:
		return string(i.([]byte))
	case string:
		return i.(string)
	}
	return fmt.Sprintf("%v", i)
}

func toInt64(i interface{}) int64 {
	switch i.(type) {
	case []byte:
		n, _ := strconv.ParseInt(string(i.([]byte)), 10, 64)
		return n
	case int:
		return int64(i.(int))
	case int64:
		return i.(int64)
	}
	return 0
}

func toFloat64(i interface{}) float64 {
	switch i.(type) {
	case []byte:
		n, _ := strconv.ParseFloat(string(i.([]byte)), 64)
		return n
	case float64:
		return i.(float64)
	case float32:
		return float64(i.(float32))
	}
	return 0
}

func TestQueryInterface(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type GetVarInterface struct {
		Id      int64  `xorm:"autoincr pk"`
		Msg     string `xorm:"varchar(255)"`
		Age     int
		Money   float32
		Created time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(GetVarInterface)))

	var data = GetVarInterface{
		Msg:   "hi",
		Age:   28,
		Money: 1.5,
	}
	_, err := testEngine.InsertOne(data)
	assert.NoError(t, err)

	tableName := "`" + mapper.Obj2Table("GetVarInterface") + "`"
	idName := mapper.Obj2Table("Id")
	msgName := mapper.Obj2Table("Msg")
	ageName := mapper.Obj2Table("Age")
	moneyName := mapper.Obj2Table("Money")

	records, err := testEngine.QueryInterface("select * from " + testEngine.TableName(tableName, true))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(records))
	assert.Equal(t, 5, len(records[0]))
	assert.EqualValues(t, 1, toInt64(records[0][idName]))
	assert.Equal(t, "hi", toString(records[0][msgName]))
	assert.EqualValues(t, 28, toInt64(records[0][ageName]))
	assert.EqualValues(t, 1.5, toFloat64(records[0][moneyName]))
}

func TestQueryNoParams(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type QueryNoParams struct {
		Id      int64  `xorm:"autoincr pk"`
		Msg     string `xorm:"varchar(255)"`
		Age     int
		Money   float32
		Created time.Time `xorm:"created"`
	}

	testEngine.ShowSQL(true)

	assert.NoError(t, testEngine.Sync2(new(QueryNoParams)))

	var q = QueryNoParams{
		Msg:   "message",
		Age:   20,
		Money: 3000,
	}
	cnt, err := testEngine.Insert(&q)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	idName := mapper.Obj2Table("Id")
	msgName := mapper.Obj2Table("Msg")
	ageName := mapper.Obj2Table("Age")
	moneyName := mapper.Obj2Table("Money")

	assertResult := func(t *testing.T, results []map[string][]byte) {
		assert.EqualValues(t, 1, len(results))
		id, err := strconv.ParseInt(string(results[0][idName]), 10, 64)
		assert.NoError(t, err)
		assert.EqualValues(t, 1, id)
		assert.Equal(t, "message", string(results[0][msgName]))

		age, err := strconv.Atoi(string(results[0][ageName]))
		assert.NoError(t, err)
		assert.EqualValues(t, 20, age)

		money, err := strconv.ParseFloat(string(results[0][moneyName]), 32)
		assert.NoError(t, err)
		assert.EqualValues(t, 3000, money)
	}

	tableName := "`" + mapper.Obj2Table("QueryNoParams") + "`"
	results, err := testEngine.Table(tableName).Limit(10).Query()
	assert.NoError(t, err)
	assertResult(t, results)

	results, err = testEngine.SQL("select * from " + testEngine.TableName(tableName, true)).Query()
	assert.NoError(t, err)
	assertResult(t, results)
}

func TestQueryStringNoParam(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type GetVar4 struct {
		Id  int64 `xorm:"autoincr pk"`
		Msg bool  `xorm:"bit"`
	}

	assert.NoError(t, testEngine.Sync2(new(GetVar4)))

	var data = GetVar4{
		Msg: false,
	}
	_, err := testEngine.Insert(data)
	assert.NoError(t, err)

	tableName := "`" + mapper.Obj2Table("GetVar4") + "`"

	idName := mapper.Obj2Table("Id")
	msgName := mapper.Obj2Table("Msg")

	records, err := testEngine.Table(tableName).Limit(1).QueryString()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(records))
	assert.EqualValues(t, "1", records[0][idName])
	if testEngine.Dialect().DBType() == core.POSTGRES || testEngine.Dialect().DBType() == core.MSSQL {
		assert.EqualValues(t, "false", records[0][msgName])
	} else {
		assert.EqualValues(t, "0", records[0][msgName])
	}

	records, err = testEngine.Table(tableName).Where(builder.Eq{"`" + idName + "`": 1}).QueryString()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(records))
	assert.EqualValues(t, "1", records[0][idName])
	if testEngine.Dialect().DBType() == core.POSTGRES || testEngine.Dialect().DBType() == core.MSSQL {
		assert.EqualValues(t, "false", records[0][msgName])
	} else {
		assert.EqualValues(t, "0", records[0][msgName])
	}
}

func TestQuerySliceStringNoParam(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type GetVar6 struct {
		Id  int64 `xorm:"autoincr pk"`
		Msg bool  `xorm:"bit"`
	}

	assert.NoError(t, testEngine.Sync2(new(GetVar6)))

	var data = GetVar6{
		Msg: false,
	}
	_, err := testEngine.Insert(data)
	assert.NoError(t, err)

	tableName := "`" + mapper.Obj2Table("GetVar6") + "`"

	records, err := testEngine.Table(tableName).Limit(1).QuerySliceString()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(records))
	assert.EqualValues(t, "1", records[0][0])
	if testEngine.Dialect().DBType() == core.POSTGRES || testEngine.Dialect().DBType() == core.MSSQL {
		assert.EqualValues(t, "false", records[0][1])
	} else {
		assert.EqualValues(t, "0", records[0][1])
	}

	idName := mapper.Obj2Table("Id")
	records, err = testEngine.Table(tableName).Where(builder.Eq{"`" + idName + "`": 1}).QuerySliceString()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(records))
	assert.EqualValues(t, "1", records[0][0])
	if testEngine.Dialect().DBType() == core.POSTGRES || testEngine.Dialect().DBType() == core.MSSQL {
		assert.EqualValues(t, "false", records[0][1])
	} else {
		assert.EqualValues(t, "0", records[0][1])
	}
}

func TestQueryInterfaceNoParam(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type GetVar5 struct {
		Id  int64 `xorm:"autoincr pk"`
		Msg bool  `xorm:"bit"`
	}

	assert.NoError(t, testEngine.Sync2(new(GetVar5)))

	var data = GetVar5{
		Msg: false,
	}
	_, err := testEngine.Insert(data)
	assert.NoError(t, err)

	tableName := "`" + mapper.Obj2Table("GetVar5") + "`"
	idName := mapper.Obj2Table("Id")
	msgName := mapper.Obj2Table("Msg")

	records, err := testEngine.Table(tableName).Limit(1).QueryInterface()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(records))
	assert.EqualValues(t, 1, toInt64(records[0][idName]))
	assert.EqualValues(t, 0, toInt64(records[0][msgName]))

	records, err = testEngine.Table(tableName).Where(builder.Eq{"`" + idName + "`": 1}).QueryInterface()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(records))
	assert.EqualValues(t, 1, toInt64(records[0][idName]))
	assert.EqualValues(t, 0, toInt64(records[0][msgName]))
}

func TestQueryWithBuilder(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type QueryWithBuilder struct {
		Id      int64  `xorm:"autoincr pk"`
		Msg     string `xorm:"varchar(255)"`
		Age     int
		Money   float32
		Created time.Time `xorm:"created"`
	}

	testEngine.ShowSQL(true)

	assert.NoError(t, testEngine.Sync2(new(QueryWithBuilder)))

	var q = QueryWithBuilder{
		Msg:   "message",
		Age:   20,
		Money: 3000,
	}
	cnt, err := testEngine.Insert(&q)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	idName := mapper.Obj2Table("Id")
	msgName := mapper.Obj2Table("Msg")
	ageName := mapper.Obj2Table("Age")
	moneyName := mapper.Obj2Table("Money")

	assertResult := func(t *testing.T, results []map[string][]byte) {
		assert.EqualValues(t, 1, len(results))
		id, err := strconv.ParseInt(string(results[0][idName]), 10, 64)
		assert.NoError(t, err)
		assert.EqualValues(t, 1, id)
		assert.Equal(t, "message", string(results[0][msgName]))

		age, err := strconv.Atoi(string(results[0][ageName]))
		assert.NoError(t, err)
		assert.EqualValues(t, 20, age)

		money, err := strconv.ParseFloat(string(results[0][moneyName]), 32)
		assert.NoError(t, err)
		assert.EqualValues(t, 3000, money)
	}

	tableName := mapper.Obj2Table("QueryWithBuilder")
	results, err := testEngine.Query(builder.Select("*").From("`" + testEngine.TableName(tableName, true) + "`"))
	assert.NoError(t, err)
	assertResult(t, results)
}

func TestJoinWithSubQuery(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type JoinWithSubQuery1 struct {
		Id       int64  `xorm:"autoincr pk"`
		Msg      string `xorm:"varchar(255)"`
		DepartId int64
		Money    float32
	}

	type JoinWithSubQueryDepart struct {
		Id   int64 `xorm:"autoincr pk"`
		Name string
	}

	testEngine.ShowSQL(true)

	assert.NoError(t, testEngine.Sync2(new(JoinWithSubQuery1), new(JoinWithSubQueryDepart)))

	var depart = JoinWithSubQueryDepart{
		Name: "depart1",
	}
	cnt, err := testEngine.Insert(&depart)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var q = JoinWithSubQuery1{
		Msg:      "message",
		DepartId: depart.Id,
		Money:    3000,
	}

	cnt, err = testEngine.Insert(&q)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	tableName := "`" + mapper.Obj2Table("JoinWithSubQueryDepart") + "`"
	tableName1 := "`" + mapper.Obj2Table("JoinWithSubQuery1") + "`"

	departID := "`" + mapper.Obj2Table("DepartId") + "`"
	idName := "`" + mapper.Obj2Table("Id") + "`"

	var querys []JoinWithSubQuery1
	err = testEngine.Join("INNER", builder.Select(idName).From(testEngine.Quote(testEngine.TableName(tableName, true))),
		tableName+"."+idName+" = "+tableName1+"."+departID).Find(&querys)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(querys))
	assert.EqualValues(t, q, querys[0])
}
