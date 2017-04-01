// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueryString(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type GetVar struct {
		Id      int64  `xorm:"autoincr pk"`
		Msg     string `xorm:"varchar(255)"`
		Age     int
		Money   float32
		Created time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(GetVar)))

	var data = GetVar{
		Msg:   "hi",
		Age:   28,
		Money: 1.5,
	}
	_, err := testEngine.InsertOne(data)
	assert.NoError(t, err)

	records, err := testEngine.QueryString("select * from get_var")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(records))
	assert.Equal(t, 5, len(records[0]))
	assert.Equal(t, "1", records[0]["id"])
	assert.Equal(t, "hi", records[0]["msg"])
	assert.Equal(t, "28", records[0]["age"])
	assert.Equal(t, "1.5", records[0]["money"])
}
