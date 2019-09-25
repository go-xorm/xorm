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
