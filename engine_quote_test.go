// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuoteColumns(t *testing.T) {
	cols := []string{"f1", "f2", "f3"}
	quoteFunc := func(value string) string {
		return "[" + value + "]"
	}

	assert.EqualValues(t, "[f1], [f2], [f3]", quoteJoinFunc(cols, quoteFunc, ","))
}

func TestChangeQuotePolicy(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type ChangeQuotePolicy struct {
		Id   int64
		Name string
	}

	testEngine.SetColumnQuotePolicy(QuotePolicyNone)
	defer func() {
		testEngine.SetColumnQuotePolicy(colQuotePolicy)
	}()

	assertSync(t, new(ChangeQuotePolicy))

	var obj1 = ChangeQuotePolicy{
		Name: "obj1",
	}
	_, err := testEngine.Insert(&obj1)
	assert.NoError(t, err)

	var obj2 ChangeQuotePolicy
	_, err = testEngine.ID(obj1.Id).Get(&obj2)
	assert.NoError(t, err)

	var objs []ChangeQuotePolicy
	err = testEngine.Find(&objs)
	assert.NoError(t, err)

	_, err = testEngine.ID(obj1.Id).Update(&ChangeQuotePolicy{
		Name: "obj2",
	})
	assert.NoError(t, err)

	_, err = testEngine.ID(obj1.Id).Delete(new(ChangeQuotePolicy))
	assert.NoError(t, err)
}

func TestChangeQuotePolicy2(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type ChangeQuotePolicy2 struct {
		Id    int64
		Name  string
		User  string
		Index int
	}

	testEngine.SetColumnQuotePolicy(QuotePolicyReserved)
	defer func() {
		testEngine.SetColumnQuotePolicy(colQuotePolicy)
	}()

	assertSync(t, new(ChangeQuotePolicy2))

	var obj1 = ChangeQuotePolicy2{
		Name: "obj1",
	}
	_, err := testEngine.Insert(&obj1)
	assert.NoError(t, err)

	var obj2 ChangeQuotePolicy2
	_, err = testEngine.ID(obj1.Id).Get(&obj2)
	assert.NoError(t, err)

	var objs []ChangeQuotePolicy2
	err = testEngine.Find(&objs)
	assert.NoError(t, err)

	_, err = testEngine.ID(obj1.Id).Update(&ChangeQuotePolicy2{
		Name: "obj2",
	})
	assert.NoError(t, err)

	_, err = testEngine.ID(obj1.Id).Delete(new(ChangeQuotePolicy2))
	assert.NoError(t, err)
}
