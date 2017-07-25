// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UserinfoDelete struct {
		Uid   int64 `xorm:"id pk not null autoincr"`
		IsMan bool
	}

	assert.NoError(t, testEngine.Sync2(new(UserinfoDelete)))

	user := UserinfoDelete{Uid: 1}
	cnt, err := testEngine.Insert(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.Delete(&UserinfoDelete{Uid: user.Uid})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	user.Uid = 0
	user.IsMan = true
	has, err := testEngine.Id(1).Get(&user)
	assert.NoError(t, err)
	assert.False(t, has)

	cnt, err = testEngine.Insert(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.Where("id=?", user.Uid).Delete(&UserinfoDelete{})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	user.Uid = 0
	user.IsMan = true
	has, err = testEngine.Id(2).Get(&user)
	assert.NoError(t, err)
	assert.False(t, has)
}

func TestDeleted(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Deleted struct {
		Id        int64 `xorm:"pk"`
		Name      string
		DeletedAt time.Time `xorm:"deleted"`
	}

	err := testEngine.DropTables(&Deleted{})
	assert.NoError(t, err)

	err = testEngine.CreateTables(&Deleted{})
	assert.NoError(t, err)

	_, err = testEngine.InsertOne(&Deleted{Id: 1, Name: "11111"})
	assert.NoError(t, err)

	_, err = testEngine.InsertOne(&Deleted{Id: 2, Name: "22222"})
	assert.NoError(t, err)

	_, err = testEngine.InsertOne(&Deleted{Id: 3, Name: "33333"})
	assert.NoError(t, err)

	// Test normal Find()
	var records1 []Deleted
	err = testEngine.Where("`"+testEngine.ColumnMapper.Obj2Table("Id")+"` > 0").Find(&records1, &Deleted{})
	assert.EqualValues(t, 3, len(records1))

	// Test normal Get()
	record1 := &Deleted{}
	has, err := testEngine.Id(1).Get(record1)
	assert.NoError(t, err)
	assert.True(t, has)

	// Test Delete() with deleted
	affected, err := testEngine.Id(1).Delete(&Deleted{})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, affected)

	has, err = testEngine.Id(1).Get(&Deleted{})
	assert.NoError(t, err)
	assert.False(t, has)

	var records2 []Deleted
	err = testEngine.Where("`" + testEngine.ColumnMapper.Obj2Table("Id") + "` > 0").Find(&records2)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(records2))

	// Test no rows affected after Delete() again.
	affected, err = testEngine.Id(1).Delete(&Deleted{})
	assert.NoError(t, err)
	assert.EqualValues(t, 0, affected)

	// Deleted.DeletedAt must not be updated.
	affected, err = testEngine.Id(2).Update(&Deleted{Name: "2", DeletedAt: time.Now()})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, affected)

	record2 := &Deleted{}
	has, err = testEngine.Id(2).Get(record2)
	assert.NoError(t, err)
	assert.True(t, record2.DeletedAt.IsZero())

	// Test find all records whatever `deleted`.
	var unscopedRecords1 []Deleted
	err = testEngine.Unscoped().Where("`"+testEngine.ColumnMapper.Obj2Table("Id")+"` > 0").Find(&unscopedRecords1, &Deleted{})
	assert.NoError(t, err)
	assert.EqualValues(t, 3, len(unscopedRecords1))

	// Delete() must really delete a record with Unscoped()
	affected, err = testEngine.Unscoped().Id(1).Delete(&Deleted{})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, affected)

	var unscopedRecords2 []Deleted
	err = testEngine.Unscoped().Where("`"+testEngine.ColumnMapper.Obj2Table("Id")+"` > 0").Find(&unscopedRecords2, &Deleted{})
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(unscopedRecords2))

	var records3 []Deleted
	err = testEngine.Where("`"+testEngine.ColumnMapper.Obj2Table("Id")+"` > 0").And("`"+testEngine.ColumnMapper.Obj2Table("Id")+"`> 1").
		Or("`"+testEngine.ColumnMapper.Obj2Table("Id")+"` = ?", 3).Find(&records3)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(records3))
}
