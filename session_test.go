// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClose(t *testing.T) {
	assert.NoError(t, prepareEngine())

	sess1 := testEngine.NewSession()
	sess1.Close()
	assert.True(t, sess1.IsClosed())

	sess2 := testEngine.Where("a = ?", 1)
	sess2.Close()
	assert.True(t, sess2.IsClosed())
}
