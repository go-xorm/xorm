// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMap(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateTable struct {
		Id   int64
		Name string
		Age  int
	}

	assert.NoError(t, testEngine.Sync2(new(UpdateTable)))
	var tb = UpdateTable{
		Name: "test",
		Age:  35,
	}
	_, err := testEngine.Insert(&tb)
	assert.NoError(t, err)

	cnt, err := testEngine.Table("update_table").Where("id = ?", tb.Id).Update(map[string]interface{}{
		"name": "test2",
		"age":  36,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
}

func TestUpdateLimit(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateTable struct {
		Id   int64
		Name string
		Age  int
	}

	assert.NoError(t, testEngine.Sync2(new(UpdateTable)))
	var tb = UpdateTable{
		Name: "test1",
		Age:  35,
	}
	cnt, err := testEngine.Insert(&tb)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	tb.Name = "test2"
	tb.Id = 0
	cnt, err = testEngine.Insert(&tb)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.OrderBy("name desc").Limit(1).Update(&UpdateTable{
		Age: 30,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var uts []UpdateTable
	err = testEngine.Find(&uts)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(uts))
	assert.EqualValues(t, 35, uts[0].Age)
	assert.EqualValues(t, 30, uts[1].Age)
}

func TestUpdateMapConcurrently(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateTable struct {
		Id   int64
		Name string
		Age  int
	}

	assert.NoError(t, testEngine.Sync2(new(UpdateTable)))
	var tb = UpdateTable{
		Name: "test",
		Age:  35,
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			// We don't check return errors for the sake of simplicity as the main point is data races.
			testEngine.Insert(&tb)
			testEngine.Table("update_table").Where("id = ?", tb.Id).Update(map[string]interface{}{
				"name": "test2",
				"age":  fmt.Sprintf("%d", i),
			})
			wg.Done()
		}()
	}
	wg.Wait()
}
