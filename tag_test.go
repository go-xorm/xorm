// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-xorm/core"
	"github.com/stretchr/testify/assert"
)

type UserCU struct {
	Id      int64
	Name    string
	Created time.Time `xorm:"created"`
	Updated time.Time `xorm:"updated"`
}

func TestCreatedAndUpdated(t *testing.T) {
	assert.NoError(t, prepareEngine())

	u := new(UserCU)
	err := testEngine.DropTables(u)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(u)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	u.Name = "sss"
	cnt, err := testEngine.Insert(u)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert not returned 1")
		t.Error(err)
		panic(err)
		return
	}

	u.Name = "xxx"
	cnt, err = testEngine.ID(u.Id).Update(u)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("update not returned 1")
		t.Error(err)
		panic(err)
		return
	}

	u.Id = 0
	u.Created = time.Now().Add(-time.Hour * 24 * 365)
	u.Updated = u.Created
	fmt.Println(u)
	cnt, err = testEngine.NoAutoTime().Insert(u)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert not returned 1")
		t.Error(err)
		panic(err)
		return
	}
}

type StrangeName struct {
	Id_t int64 `xorm:"pk autoincr"`
	Name string
}

func TestStrangeName(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(new(StrangeName))
	if err != nil {
		t.Error(err)
	}

	err = testEngine.CreateTables(new(StrangeName))
	if err != nil {
		t.Error(err)
	}

	_, err = testEngine.Insert(&StrangeName{Name: "sfsfdsfds"})
	if err != nil {
		t.Error(err)
	}

	beans := make([]StrangeName, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
	}
}

func TestCreatedUpdated(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type CreatedUpdated struct {
		Id       int64
		Name     string
		Value    float64   `xorm:"numeric"`
		Created  time.Time `xorm:"created"`
		Created2 time.Time `xorm:"created"`
		Updated  time.Time `xorm:"updated"`
	}

	err := testEngine.Sync2(&CreatedUpdated{})
	assert.NoError(t, err)

	c := &CreatedUpdated{Name: "test"}
	_, err = testEngine.Insert(c)
	assert.NoError(t, err)

	c2 := new(CreatedUpdated)
	has, err := testEngine.ID(c.Id).Get(c2)
	assert.NoError(t, err)

	assert.True(t, has)

	c2.Value -= 1
	_, err = testEngine.ID(c2.Id).Update(c2)
	assert.NoError(t, err)
}

func TestCreatedUpdatedInt64(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type CreatedUpdatedInt64 struct {
		Id       int64
		Name     string
		Value    float64 `xorm:"numeric"`
		Created  int64   `xorm:"created"`
		Created2 int64   `xorm:"created"`
		Updated  int64   `xorm:"updated"`
	}

	assertSync(t, &CreatedUpdatedInt64{})

	c := &CreatedUpdatedInt64{Name: "test"}
	_, err := testEngine.Insert(c)
	assert.NoError(t, err)

	c2 := new(CreatedUpdatedInt64)
	has, err := testEngine.ID(c.Id).Get(c2)
	assert.NoError(t, err)
	assert.True(t, has)

	c2.Value -= 1
	_, err = testEngine.ID(c2.Id).Update(c2)
	assert.NoError(t, err)
}

type Lowercase struct {
	Id    int64
	Name  string
	ended int64 `xorm:"-"`
}

func TestLowerCase(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.Sync2(&Lowercase{})
	_, err = testEngine.Where("(id) > 0").Delete(&Lowercase{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	_, err = testEngine.Insert(&Lowercase{ended: 1})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	ls := make([]Lowercase, 0)
	err = testEngine.Find(&ls)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if len(ls) != 1 {
		err = errors.New("should be 1")
		t.Error(err)
		panic(err)
	}
}

func TestAutoIncrTag(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type TestAutoIncr1 struct {
		Id int64
	}

	tb := testEngine.TableInfo(new(TestAutoIncr1))
	cols := tb.Columns()
	assert.EqualValues(t, 1, len(cols))
	assert.True(t, cols[0].IsAutoIncrement)
	assert.True(t, cols[0].IsPrimaryKey)
	assert.Equal(t, "id", cols[0].Name)

	type TestAutoIncr2 struct {
		Id int64 `xorm:"id"`
	}

	tb = testEngine.TableInfo(new(TestAutoIncr2))
	cols = tb.Columns()
	assert.EqualValues(t, 1, len(cols))
	assert.False(t, cols[0].IsAutoIncrement)
	assert.False(t, cols[0].IsPrimaryKey)
	assert.Equal(t, "id", cols[0].Name)

	type TestAutoIncr3 struct {
		Id int64 `xorm:"'ID'"`
	}

	tb = testEngine.TableInfo(new(TestAutoIncr3))
	cols = tb.Columns()
	assert.EqualValues(t, 1, len(cols))
	assert.False(t, cols[0].IsAutoIncrement)
	assert.False(t, cols[0].IsPrimaryKey)
	assert.Equal(t, "ID", cols[0].Name)

	type TestAutoIncr4 struct {
		Id int64 `xorm:"pk"`
	}

	tb = testEngine.TableInfo(new(TestAutoIncr4))
	cols = tb.Columns()
	assert.EqualValues(t, 1, len(cols))
	assert.False(t, cols[0].IsAutoIncrement)
	assert.True(t, cols[0].IsPrimaryKey)
	assert.Equal(t, "id", cols[0].Name)
}

func TestTagComment(t *testing.T) {
	assert.NoError(t, prepareEngine())
	// FIXME: only support mysql
	if testEngine.Dialect().DriverName() != core.MYSQL {
		return
	}

	type TestComment1 struct {
		Id int64 `xorm:"comment(主键)"`
	}

	assert.NoError(t, testEngine.Sync2(new(TestComment1)))

	tables, err := testEngine.DBMetas()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(tables))
	assert.EqualValues(t, 1, len(tables[0].Columns()))
	assert.EqualValues(t, "主键", tables[0].Columns()[0].Comment)

	assert.NoError(t, testEngine.DropTables(new(TestComment1)))

	type TestComment2 struct {
		Id int64 `xorm:"comment('主键')"`
	}

	assert.NoError(t, testEngine.Sync2(new(TestComment2)))

	tables, err = testEngine.DBMetas()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(tables))
	assert.EqualValues(t, 1, len(tables[0].Columns()))
	assert.EqualValues(t, "主键", tables[0].Columns()[0].Comment)
}

func TestTagDefault(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type DefaultStruct struct {
		Id   int64
		Name string
		Age  int `xorm:"default(10)"`
	}

	assertSync(t, new(DefaultStruct))

	cnt, err := testEngine.Omit("age").Insert(&DefaultStruct{
		Name: "test",
		Age:  20,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var s DefaultStruct
	has, err := testEngine.ID(1).Get(&s)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, 10, s.Age)
	assert.EqualValues(t, "test", s.Name)
}

func TestTagsDirection(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type OnlyFromDBStruct struct {
		Id   int64
		Name string
		Uuid string `xorm:"<- default '1'"`
	}

	assertSync(t, new(OnlyFromDBStruct))

	cnt, err := testEngine.Insert(&OnlyFromDBStruct{
		Name: "test",
		Uuid: "2",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var s OnlyFromDBStruct
	has, err := testEngine.ID(1).Get(&s)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, "1", s.Uuid)
	assert.EqualValues(t, "test", s.Name)

	cnt, err = testEngine.ID(1).Update(&OnlyFromDBStruct{
		Uuid: "3",
		Name: "test1",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var s3 OnlyFromDBStruct
	has, err = testEngine.ID(1).Get(&s3)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, "1", s3.Uuid)
	assert.EqualValues(t, "test1", s3.Name)

	type OnlyToDBStruct struct {
		Id   int64
		Name string
		Uuid string `xorm:"->"`
	}

	assertSync(t, new(OnlyToDBStruct))

	cnt, err = testEngine.Insert(&OnlyToDBStruct{
		Name: "test",
		Uuid: "2",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var s2 OnlyToDBStruct
	has, err = testEngine.ID(1).Get(&s2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, "", s2.Uuid)
	assert.EqualValues(t, "test", s2.Name)
}

func TestTagTime(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type TagUTCStruct struct {
		Id      int64
		Name    string
		Created time.Time `xorm:"created utc"`
	}

	assertSync(t, new(TagUTCStruct))

	assert.EqualValues(t, time.Local.String(), testEngine.GetTZLocation().String())

	s := TagUTCStruct{
		Name: "utc",
	}
	cnt, err := testEngine.Insert(&s)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var u TagUTCStruct
	has, err := testEngine.ID(1).Get(&u)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, s.Created.Format("2006-01-02 15:04:05"), u.Created.Format("2006-01-02 15:04:05"))

	var tm string
	has, err = testEngine.Table("tag_u_t_c_struct").Cols("created").Get(&tm)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, s.Created.UTC().Format("2006-01-02 15:04:05"),
		strings.Replace(strings.Replace(tm, "T", " ", -1), "Z", "", -1))
}
