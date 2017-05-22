// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"reflect"
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

func TestInsertMulti(t *testing.T) {

	assert.NoError(t, prepareEngine())
	type TestMulti struct {
		Id   int64  `xorm:"int(11) pk"`
		Name string `xorm:"varchar(255)"`
	}

	assert.NoError(t, testEngine.Sync2(new(TestMulti)))

	num, err := insertMultiDatas(1,
		append([]TestMulti{}, TestMulti{1, "test1"}, TestMulti{2, "test2"}, TestMulti{3, "test3"}))
	assert.NoError(t, err)
	assert.EqualValues(t, 3, num)
}

func insertMultiDatas(step int, datas interface{}) (num int64, err error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(datas))
	var iLen int64
	if sliceValue.Kind() != reflect.Slice {
		return 0, fmt.Errorf("not silce")
	}
	iLen = int64(sliceValue.Len())
	if iLen == 0 {
		return
	}

	session := testEngine.NewSession()
	defer session.Close()

	if err = callbackLooper(datas, step,
		func(innerDatas interface{}) error {
			n, e := session.InsertMulti(innerDatas)
			if e != nil {
				return e
			}
			num += n
			return nil
		}); err != nil {
		return 0, err
	} else if num != iLen {
		return 0, fmt.Errorf("num error: %d - %d", num, iLen)
	}
	return
}

func callbackLooper(datas interface{}, step int, actionFunc func(interface{}) error) (err error) {

	sliceValue := reflect.Indirect(reflect.ValueOf(datas))
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("not slice")
	}
	if sliceValue.Len() <= 0 {
		return
	}

	tempLen := 0
	processedLen := sliceValue.Len()
	for i := 0; i < sliceValue.Len(); i += step {
		if processedLen > step {
			tempLen = i + step
		} else {
			tempLen = sliceValue.Len()
		}
		var tempInterface []interface{}
		for j := i; j < tempLen; j++ {
			tempInterface = append(tempInterface, sliceValue.Index(j).Interface())
		}
		if err = actionFunc(tempInterface); err != nil {
			return
		}
		processedLen = processedLen - step
	}
	return
}

func TestInsertOneIfPkIsPoint(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type TestPoint struct {
		Id      *int64     `xorm:"autoincr pk notnull 'id'"`
		Msg     *string    `xorm:"varchar(255)"`
		Created *time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(TestPoint)))
	msg := "hi"
	data := TestPoint{Msg: &msg}
	_, err := testEngine.InsertOne(&data)
	assert.NoError(t, err)
}

func TestInsertOneIfPkIsPointRename(t *testing.T) {
	assert.NoError(t, prepareEngine())
	type ID *int64
	type TestPoint struct {
		Id      ID         `xorm:"autoincr pk notnull 'id'"`
		Msg     *string    `xorm:"varchar(255)"`
		Created *time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(TestPoint)))
	msg := "hi"
	data := TestPoint{Msg: &msg}
	_, err := testEngine.InsertOne(&data)
	assert.NoError(t, err)
}
