// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBefore_Get(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type BeforeTable struct {
		Id   int64
		Name string
		Val  string `xorm:"-"`
	}

	assert.NoError(t, testEngine.Sync2(new(BeforeTable)))

	cnt, err := testEngine.Insert(&BeforeTable{
		Name: "test",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var be BeforeTable
	has, err := testEngine.Before(func(bean interface{}) {
		bean.(*BeforeTable).Val = "val"
	}).Get(&be)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, "val", be.Val)
	assert.Equal(t, "test", be.Name)
}

func TestBefore_Find(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type BeforeTable struct {
		Id   int64
		Name string
		Val  string `xorm:"-"`
	}

	assert.NoError(t, testEngine.Sync2(new(BeforeTable)))

	cnt, err := testEngine.Insert([]BeforeTable{
		{Name: "test1"},
		{Name: "test2"},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 2, cnt)

	var be []BeforeTable
	err = testEngine.Before(func(bean interface{}) {
		bean.(*BeforeTable).Val = "val"
	}).Find(&be)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(be))
	assert.Equal(t, "val", be[0].Val)
	assert.Equal(t, "test1", be[0].Name)
	assert.Equal(t, "val", be[1].Val)
	assert.Equal(t, "test2", be[1].Name)
}
