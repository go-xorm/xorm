// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInsertOne(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Test struct {
		Id      int64     `xorm:"autoincr pk"`
		Msg     string    `xorm:"varchar(255)"`
		Created time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(Test)))

	data := Test{Msg: "hi"}
	_, err := testEngine.InsertOne(data)
	assert.NoError(t, err)
}


func TestInsertOneIfPkIsPoint(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type TestPoint struct {
		Id	 *int64     `xorm:"autoincr pk notnull 'id'"`
		Msg      *string    `xorm:"varchar(255)"`
		Created  *time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(TestPoint)))
	msg := "hi"
	data := TestPoint{Msg: &msg}
	_, err := testEngine.InsertOne(&data)
	assert.NoError(t, err)
}

func TestInsertOneIfPkIsPointRename (t *testing.T) {
	assert.NoError(t, prepareEngine())
	type ID *int64
	type TestPoint struct {
		Id	 ID         `xorm:"autoincr pk notnull 'id'"`
		Msg      *string    `xorm:"varchar(255)"`
		Created  *time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(TestPoint)))
	msg := "hi"
	data := TestPoint{Msg: &msg}
	_, err := testEngine.InsertOne(&data)
	assert.NoError(t, err)
}