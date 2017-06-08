// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"fmt"
	"testing"
	"time"

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
	cnt, err = testEngine.Id(u.Id).Update(u)
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

type CreatedUpdated struct {
	Id       int64
	Name     string
	Value    float64   `xorm:"numeric"`
	Created  time.Time `xorm:"created"`
	Created2 time.Time `xorm:"created"`
	Updated  time.Time `xorm:"updated"`
}

func TestCreatedUpdated(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.Sync(&CreatedUpdated{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	c := &CreatedUpdated{Name: "test"}
	_, err = testEngine.Insert(c)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	c2 := new(CreatedUpdated)
	has, err := testEngine.Id(c.Id).Get(c2)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !has {
		panic(errors.New("no id"))
	}

	c2.Value -= 1
	_, err = testEngine.Id(c2.Id).Update(c2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

type Lowercase struct {
	Id    int64
	Name  string
	ended int64 `xorm:"-"`
}

func TestLowerCase(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.Sync(&Lowercase{})
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