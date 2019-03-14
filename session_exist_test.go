// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistStruct(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type RecordExist struct {
		Id   int64
		Name string
	}

	assertSync(t, new(RecordExist))

	has, err := testEngine.Exist(new(RecordExist))
	assert.NoError(t, err)
	assert.False(t, has)

	cnt, err := testEngine.Insert(&RecordExist{
		Name: "test1",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	has, err = testEngine.Exist(new(RecordExist))
	assert.NoError(t, err)
	assert.True(t, has)

	has, err = testEngine.Exist(&RecordExist{
		Name: "test1",
	})
	assert.NoError(t, err)
	assert.True(t, has)

	has, err = testEngine.Exist(&RecordExist{
		Name: "test2",
	})
	assert.NoError(t, err)
	assert.False(t, has)

	nameName := "`" + mapper.Obj2Table("Name") + "`"

	has, err = testEngine.Where(nameName+" = ?", "test1").Exist(&RecordExist{})
	assert.NoError(t, err)
	assert.True(t, has)

	has, err = testEngine.Where(nameName+" = ?", "test2").Exist(&RecordExist{})
	assert.NoError(t, err)
	assert.False(t, has)

	tableName := mapper.Obj2Table("RecordExist")

	has, err = testEngine.SQL("select * from "+testEngine.TableName(tableName, true)+" where "+nameName+" = ?", "test1").Exist()
	assert.NoError(t, err)
	assert.True(t, has)

	has, err = testEngine.SQL("select * from "+testEngine.TableName(tableName, true)+" where "+nameName+" = ?", "test2").Exist()
	assert.NoError(t, err)
	assert.False(t, has)

	has, err = testEngine.Table(tableName).Exist()
	assert.NoError(t, err)
	assert.True(t, has)

	has, err = testEngine.Table(tableName).Where(nameName+" = ?", "test1").Exist()
	assert.NoError(t, err)
	assert.True(t, has)

	has, err = testEngine.Table(tableName).Where(nameName+" = ?", "test2").Exist()
	assert.NoError(t, err)
	assert.False(t, has)
}
