// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEraseAny(t *testing.T) {
	raw := "SELECT * FROM `table`.[table_name]"
	assert.EqualValues(t, raw, eraseAny(raw))
	assert.EqualValues(t, "SELECT * FROM table.[table_name]", eraseAny(raw, "`"))
	assert.EqualValues(t, "SELECT * FROM table.table_name", eraseAny(raw, "`", "[", "]"))
}
