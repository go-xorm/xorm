// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.8

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingContext(t *testing.T) {
	assert.NoError(t, prepareEngine())

	// TODO: Since EngineInterface should be compitable with old Go version, PingContext is not supported.
	/*
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := testEngine.PingContext(ctx)
		assert.NoError(t, err)
	*/
}
