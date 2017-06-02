// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/go-xorm/core"
	"github.com/stretchr/testify/assert"
)

func TestSetExpr(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type User struct {
		Id   int64
		Show bool
	}

	assert.NoError(t, testEngine.Sync2(new(User)))

	cnt, err := testEngine.Insert(&User{
		Show: true,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var not = "NOT"
	if testEngine.dialect.DBType() == core.MSSQL {
		not = "~"
	}
	cnt, err = testEngine.SetExpr("show", not+" `show`").Id(1).Update(new(User))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
}
