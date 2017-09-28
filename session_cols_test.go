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

	type UserExpr struct {
		Id   int64
		Show bool
	}

	assert.NoError(t, testEngine.Sync2(new(UserExpr)))

	cnt, err := testEngine.Insert(&UserExpr{
		Show: true,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var not = "NOT"
	if testEngine.Dialect().DBType() == core.MSSQL {
		not = "~"
	}
	cnt, err = testEngine.SetExpr("show", not+" `show`").ID(1).Update(new(UserExpr))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
}

func TestCols(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type ColsTable struct {
		Id   int64
		Col1 string
		Col2 string
	}

	assertSync(t, new(ColsTable))

	_, err := testEngine.Insert(&ColsTable{
		Col1: "1",
		Col2: "2",
	})
	assert.NoError(t, err)

	sess := testEngine.ID(1)
	_, err = sess.Cols("col1").Cols("col2").Update(&ColsTable{
		Col1: "",
		Col2: "",
	})
	assert.NoError(t, err)

	var tb ColsTable
	has, err := testEngine.ID(1).Get(&tb)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, "", tb.Col1)
	assert.EqualValues(t, "", tb.Col2)
}
