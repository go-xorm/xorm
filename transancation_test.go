// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAutoTransaction(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Test struct {
		Id      int64     `xorm:"autoincr pk"`
		Msg     string    `xorm:"varchar(255)"`
		Created time.Time `xorm:"created"`
	}

	assert.NoError(t, testEngine.Sync2(new(Test)))

	engine := testEngine.(*Engine)

	// will success
	AutoTransaction(func(session *Session) (interface{}, error) {
		_, err := session.Insert(Test{Msg: "hi"})
		assert.NoError(t, err)

		return nil, nil
	}, engine)

	has, err := engine.Exist(&Test{Msg: "hi"})
	assert.NoError(t, err)
	assert.EqualValues(t, true, has)

	// will rollback
	_, err = AutoTransaction(func(session *Session) (interface{}, error) {
		_, err := session.Insert(Test{Msg: "hello"})
		assert.NoError(t, err)

		return nil, fmt.Errorf("rollback")
	}, engine)
	assert.Error(t, err)

	has, err = engine.Exist(&Test{Msg: "hello"})
	assert.NoError(t, err)
	assert.EqualValues(t, false, has)
}
