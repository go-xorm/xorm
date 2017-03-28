// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinLimit(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Salary struct {
		Id  int64
		Lid int64
	}

	type CheckList struct {
		Id  int64
		Eid int64
	}

	type Empsetting struct {
		Id   int64
		Name string
	}

	assert.NoError(t, testEngine.Sync2(new(Salary), new(CheckList), new(Empsetting)))

	var emp Empsetting
	cnt, err := testEngine.Insert(&emp)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var checklist = CheckList{
		Eid: emp.Id,
	}
	cnt, err = testEngine.Insert(&checklist)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var salary = Salary{
		Lid: checklist.Id,
	}
	cnt, err = testEngine.Insert(&salary)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var salaries []Salary
	err = testEngine.Table("salary").
		Join("INNER", "check_list", "check_list.id = salary.lid").
		Join("LEFT", "empsetting", "empsetting.id = check_list.eid").
		Limit(10, 0).
		Find(&salaries)
	assert.NoError(t, err)
}
