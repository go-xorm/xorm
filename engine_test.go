// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoIncrTag(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type TestAutoIncr1 struct {
		Id int64
	}

	tb := testEngine.TableInfo(new(TestAutoIncr1))
	cols := tb.Columns()
	assert.EqualValues(t, 1, len(cols))
	assert.True(t, cols[0].IsAutoIncrement)
	assert.True(t, cols[0].IsPrimaryKey)
	assert.Equal(t, "id", cols[0].Name)

	type TestAutoIncr2 struct {
		Id int64 `xorm:"id"`
	}

	tb = testEngine.TableInfo(new(TestAutoIncr2))
	cols = tb.Columns()
	assert.EqualValues(t, 1, len(cols))
	assert.False(t, cols[0].IsAutoIncrement)
	assert.False(t, cols[0].IsPrimaryKey)
	assert.Equal(t, "id", cols[0].Name)

	type TestAutoIncr3 struct {
		Id int64 `xorm:"'ID'"`
	}

	tb = testEngine.TableInfo(new(TestAutoIncr3))
	cols = tb.Columns()
	assert.EqualValues(t, 1, len(cols))
	assert.False(t, cols[0].IsAutoIncrement)
	assert.False(t, cols[0].IsPrimaryKey)
	assert.Equal(t, "ID", cols[0].Name)

	type TestAutoIncr4 struct {
		Id int64 `xorm:"pk"`
	}

	tb = testEngine.TableInfo(new(TestAutoIncr4))
	cols = tb.Columns()
	assert.EqualValues(t, 1, len(cols))
	assert.False(t, cols[0].IsAutoIncrement)
	assert.True(t, cols[0].IsPrimaryKey)
	assert.Equal(t, "id", cols[0].Name)
}
