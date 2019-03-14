// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type NullType struct {
	Id           int `xorm:"pk autoincr"`
	Name         sql.NullString
	Age          sql.NullInt64
	Height       sql.NullFloat64
	IsMan        sql.NullBool `xorm:"null"`
	CustomStruct CustomStruct `xorm:"valchar(64) null"`
}

type CustomStruct struct {
	Year  int
	Month int
	Day   int
}

func (CustomStruct) String() string {
	return "CustomStruct"
}

func (m *CustomStruct) Scan(value interface{}) error {
	if value == nil {
		m.Year, m.Month, m.Day = 0, 0, 0
		return nil
	}

	if s, ok := value.([]byte); ok {
		seps := strings.Split(string(s), "/")
		m.Year, _ = strconv.Atoi(seps[0])
		m.Month, _ = strconv.Atoi(seps[1])
		m.Day, _ = strconv.Atoi(seps[2])
		return nil
	}

	return errors.New("scan data not fit []byte")
}

func (m CustomStruct) Value() (driver.Value, error) {
	return fmt.Sprintf("%d/%d/%d", m.Year, m.Month, m.Day), nil
}

func TestCreateNullStructTable(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.CreateTables(new(NullType))
	assert.NoError(t, err)
}

func TestDropNullStructTable(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(new(NullType))
	assert.NoError(t, err)
}

func TestNullStructInsert(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(NullType))

	if true {
		item := new(NullType)
		_, err := testEngine.Insert(item)
		assert.NoError(t, err)
		fmt.Println(item)
		assert.EqualValues(t, 1, item.Id)
	}

	if true {
		item := NullType{
			Name:   sql.NullString{"haolei", true},
			Age:    sql.NullInt64{34, true},
			Height: sql.NullFloat64{1.72, true},
			IsMan:  sql.NullBool{true, true},
		}
		_, err := testEngine.Insert(&item)
		assert.NoError(t, err)
		fmt.Println(item)
		assert.EqualValues(t, 2, item.Id)
	}

	if true {
		items := []NullType{}

		for i := 0; i < 5; i++ {
			item := NullType{
				Name:         sql.NullString{"haolei_" + fmt.Sprint(i+1), true},
				Age:          sql.NullInt64{30 + int64(i), true},
				Height:       sql.NullFloat64{1.5 + 1.1*float64(i), true},
				IsMan:        sql.NullBool{true, true},
				CustomStruct: CustomStruct{i, i + 1, i + 2},
			}

			items = append(items, item)
		}

		_, err := testEngine.Insert(&items)
		assert.NoError(t, err)
		fmt.Println(items)
	}
}

func TestNullStructUpdate(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(NullType))

	_, err := testEngine.Insert([]NullType{
		{
			Name: sql.NullString{
				String: "name1",
				Valid:  true,
			},
		},
		{
			Name: sql.NullString{
				String: "name2",
				Valid:  true,
			},
		},
		{
			Name: sql.NullString{
				String: "name3",
				Valid:  true,
			},
		},
		{
			Name: sql.NullString{
				String: "name4",
				Valid:  true,
			},
		},
	})
	assert.NoError(t, err)

	idName := "`" + mapper.Obj2Table("Id") + "`"
	ageName := "`" + mapper.Obj2Table("Age") + "`"
	heightName := "`" + mapper.Obj2Table("Height") + "`"
	isManName := "`" + mapper.Obj2Table("IsMan") + "`"

	if true { // 测试可插入NULL
		item := new(NullType)
		item.Age = sql.NullInt64{23, true}
		item.Height = sql.NullFloat64{0, false} // update to NULL

		affected, err := testEngine.ID(2).Cols(ageName, heightName, isManName).Update(item)
		assert.NoError(t, err)
		assert.EqualValues(t, 1, affected)
	}

	if true { // 测试In update
		item := new(NullType)
		item.Age = sql.NullInt64{23, true}
		affected, err := testEngine.In(idName, 3, 4).Cols(ageName, heightName, isManName).Update(item)
		assert.NoError(t, err)
		assert.EqualValues(t, 2, affected)
	}

	if true { // 测试where
		item := new(NullType)
		item.Name = sql.NullString{"nullname", true}
		item.IsMan = sql.NullBool{true, true}
		item.Age = sql.NullInt64{34, true}

		_, err := testEngine.Where(ageName+" > ?", 34).Update(item)
		assert.NoError(t, err)
	}

	if true { // 修改全部时，插入空值
		item := &NullType{
			Name:   sql.NullString{"winxxp", true},
			Age:    sql.NullInt64{30, true},
			Height: sql.NullFloat64{1.72, true},
			// IsMan:  sql.NullBool{true, true},
		}

		_, err := testEngine.AllCols().ID(6).Update(item)
		assert.NoError(t, err)
		fmt.Println(item)
	}
}

func TestNullStructFind(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(NullType))

	_, err := testEngine.Insert([]NullType{
		{
			Name: sql.NullString{
				String: "name1",
				Valid:  false,
			},
		},
		{
			Name: sql.NullString{
				String: "name2",
				Valid:  true,
			},
		},
		{
			Name: sql.NullString{
				String: "name3",
				Valid:  true,
			},
		},
		{
			Name: sql.NullString{
				String: "name4",
				Valid:  true,
			},
		},
	})
	assert.NoError(t, err)

	if true {
		item := new(NullType)
		has, err := testEngine.ID(1).Get(item)
		assert.NoError(t, err)
		assert.True(t, has)

		fmt.Println(item)
		if item.Id != 1 || item.Name.Valid || item.Age.Valid || item.Height.Valid ||
			item.IsMan.Valid {
			err = errors.New("insert error")
			t.Error(err)
			panic(err)
		}
	}

	if true {
		item := new(NullType)
		item.Id = 2

		has, err := testEngine.Get(item)
		assert.NoError(t, err)
		assert.True(t, has)
		fmt.Println(item)
	}

	if true {
		item := make([]NullType, 0)

		err := testEngine.ID(2).Find(&item)
		assert.NoError(t, err)
		fmt.Println(item)
	}

	if true {
		item := make([]NullType, 0)

		ageName := "`" + mapper.Obj2Table("Age") + "`"
		err := testEngine.Asc(ageName).Find(&item)
		assert.NoError(t, err)
		for k, v := range item {
			fmt.Println(k, v)
		}
	}
}

func TestNullStructIterate(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(NullType))

	if true {
		ageName := "`" + mapper.Obj2Table("Age") + "`"
		err := testEngine.Where(ageName+" IS NOT NULL").OrderBy(ageName).Iterate(new(NullType),
			func(i int, bean interface{}) error {
				nultype := bean.(*NullType)
				fmt.Println(i, nultype)
				return nil
			})
		assert.NoError(t, err)
	}
}

func TestNullStructCount(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(NullType))

	if true {
		item := new(NullType)
		ageName := "`" + mapper.Obj2Table("Age") + "`"
		total, err := testEngine.Where(ageName + " IS NOT NULL").Count(item)
		assert.NoError(t, err)
		fmt.Println(total)
	}
}

func TestNullStructRows(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(NullType))

	item := new(NullType)
	idName := "`" + mapper.Obj2Table("Id") + "`"
	rows, err := testEngine.Where(idName+" > ?", 1).Rows(item)
	assert.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(item)
		assert.NoError(t, err)
		fmt.Println(item)
	}
}

func TestNullStructDelete(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(NullType))

	item := new(NullType)

	_, err := testEngine.ID(1).Delete(item)
	assert.NoError(t, err)

	idName := "`" + mapper.Obj2Table("Id") + "`"
	_, err = testEngine.Where(idName+" > ?", 1).Delete(item)
	assert.NoError(t, err)
}
