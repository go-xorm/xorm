// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayField(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type ArrayStruct struct {
		Id   int64
		Name [20]byte `xorm:"char(80)"`
	}

	assert.NoError(t, testEngine.Sync2(new(ArrayStruct)))

	var as = ArrayStruct{
		Name: [20]byte{
			96, 96, 96, 96, 96,
			96, 96, 96, 96, 96,
			96, 96, 96, 96, 96,
			96, 96, 96, 96, 96,
		},
	}
	cnt, err := testEngine.Insert(&as)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var arr ArrayStruct
	has, err := testEngine.Id(1).Get(&arr)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, as.Name, arr.Name)

	var arrs []ArrayStruct
	err = testEngine.Find(&arrs)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(arrs))
	assert.Equal(t, as.Name, arrs[0].Name)

	var newName = [20]byte{
		90, 96, 96, 96, 96,
		96, 96, 96, 96, 96,
		96, 96, 96, 96, 96,
		96, 96, 96, 96, 96,
	}

	cnt, err = testEngine.ID(1).Update(&ArrayStruct{
		Name: newName,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var newArr ArrayStruct
	has, err = testEngine.ID(1).Get(&newArr)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, newName, newArr.Name)

	cnt, err = testEngine.ID(1).Delete(new(ArrayStruct))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var cfgArr ArrayStruct
	has, err = testEngine.ID(1).Get(&cfgArr)
	assert.NoError(t, err)
	assert.Equal(t, false, has)
}

func TestGetBytes(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Varbinary struct {
		Data []byte `xorm:"VARBINARY"`
	}

	err := testEngine.Sync2(new(Varbinary))
	assert.NoError(t, err)

	cnt, err := testEngine.Insert(&Varbinary{
		Data: []byte("test"),
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var b Varbinary
	has, err := testEngine.Get(&b)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, "test", string(b.Data))
}
