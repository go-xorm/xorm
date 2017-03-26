// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/go-xorm/builder"
	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	assert.NoError(t, prepareEngine())

	const (
		OpEqual int = iota
		OpGreatThan
		OpLessThan
	)

	type Condition struct {
		Id        int64
		TableName string
		ColName   string
		Op        int
		Value     string
	}

	err := testEngine.CreateTables(&Condition{})
	assert.NoError(t, err)

	_, err = testEngine.Insert(&Condition{TableName: "table1", ColName: "col1", Op: OpEqual, Value: "1"})
	assert.NoError(t, err)

	var cond Condition
	has, err := testEngine.Where(builder.Eq{"col_name": "col1"}).Get(&cond)
	assert.NoError(t, err)
	assert.Equal(t, true, has, "records should exist")

	has, err = testEngine.Where(builder.Eq{"col_name": "col1"}.
		And(builder.Eq{"op": OpEqual})).
		NoAutoCondition().
		Get(&cond)
	assert.NoError(t, err)
	assert.Equal(t, true, has, "records should exist")

	has, err = testEngine.Where(builder.Eq{"col_name": "col1", "op": OpEqual, "value": "1"}).
		NoAutoCondition().
		Get(&cond)
	assert.NoError(t, err)
	assert.Equal(t, true, has, "records should exist")

	has, err = testEngine.Where(builder.Eq{"col_name": "col1"}.
		And(builder.Neq{"op": OpEqual})).
		NoAutoCondition().
		Get(&cond)
	assert.NoError(t, err)
	assert.Equal(t, false, has, "records should not exist")

	var conds []Condition
	err = testEngine.Where(builder.Eq{"col_name": "col1"}.
		And(builder.Eq{"op": OpEqual})).
		Find(&conds)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(conds), "records should exist")

	conds = make([]Condition, 0)
	err = testEngine.Where(builder.Like{"col_name", "col"}).Find(&conds)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(conds), "records should exist")

	conds = make([]Condition, 0)
	err = testEngine.Where(builder.Expr("col_name = ?", "col1")).Find(&conds)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(conds), "records should exist")

	conds = make([]Condition, 0)
	err = testEngine.Where(builder.In("col_name", "col1", "col2")).Find(&conds)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(conds), "records should exist")

	// complex condtions
	var where = builder.NewCond()
	if true {
		where = where.And(builder.Eq{"col_name": "col1"})
		where = where.Or(builder.And(builder.In("col_name", "col1", "col2"), builder.Expr("col_name = ?", "col1")))
	}

	conds = make([]Condition, 0)
	err = testEngine.Where(where).Find(&conds)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(conds), "records should exist")
}
