// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterate(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UserIterate struct {
		Id    int64
		IsMan bool
	}

	assert.NoError(t, testEngine.Sync2(new(UserIterate)))

	cnt, err := testEngine.Insert(&UserIterate{
		IsMan: true,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt = 0
	err = testEngine.Iterate(new(UserIterate), func(i int, bean interface{}) error {
		user := bean.(*UserIterate)
		assert.EqualValues(t, 1, user.Id)
		assert.EqualValues(t, true, user.IsMan)
		cnt++
		return nil
	})
	assert.EqualValues(t, 1, cnt)
}
