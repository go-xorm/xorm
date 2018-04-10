// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRows(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UserRows struct {
		Id    int64
		IsMan bool
	}

	assert.NoError(t, testEngine.Sync2(new(UserRows)))

	cnt, err := testEngine.Insert(&UserRows{
		IsMan: true,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	rows, err := testEngine.Rows(new(UserRows))
	assert.NoError(t, err)
	defer rows.Close()

	cnt = 0
	user := new(UserRows)
	for rows.Next() {
		err = rows.Scan(user)
		assert.NoError(t, err)
		cnt++
	}
	assert.EqualValues(t, 1, cnt)

	sess := testEngine.NewSession()
	defer sess.Close()

	rows1, err := sess.Prepare().Rows(new(UserRows))
	assert.NoError(t, err)
	defer rows1.Close()

	cnt = 0
	for rows1.Next() {
		err = rows1.Scan(user)
		assert.NoError(t, err)
		cnt++
	}
	assert.EqualValues(t, 1, cnt)

	var tbName = testEngine.Quote(testEngine.TableName(user, true))
	rows2, err := testEngine.SQL("SELECT * FROM " + tbName).Rows(new(UserRows))
	assert.NoError(t, err)
	defer rows2.Close()

	cnt = 0
	for rows2.Next() {
		err = rows2.Scan(user)
		assert.NoError(t, err)
		cnt++
	}
	assert.EqualValues(t, 1, cnt)
}
