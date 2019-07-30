// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitTag(t *testing.T) {
	var cases = []struct {
		tag  string
		tags []string
	}{
		{"not null default '2000-01-01 00:00:00' TIMESTAMP", []string{"not", "null", "default", "'2000-01-01 00:00:00'", "TIMESTAMP"}},
		{"TEXT", []string{"TEXT"}},
		{"default('2000-01-01 00:00:00')", []string{"default('2000-01-01 00:00:00')"}},
		{"json  binary", []string{"json", "binary"}},
	}

	for _, kase := range cases {
		tags := splitTag(kase.tag)
		if !sliceEq(tags, kase.tags) {
			t.Fatalf("[%d]%v is not equal [%d]%v", len(tags), tags, len(kase.tags), kase.tags)
		}
	}
}

func TestEraseAny(t *testing.T) {
	raw := "SELECT * FROM `table`.[table_name]"
	assert.EqualValues(t, raw, eraseAny(raw))
	assert.EqualValues(t, "SELECT * FROM table.[table_name]", eraseAny(raw, "`"))
	assert.EqualValues(t, "SELECT * FROM table.table_name", eraseAny(raw, "`", "[", "]"))
}

func TestQuoteColumns(t *testing.T) {
	cols := []string{"f1", "f2", "f3"}
	quoteFunc := func(value string) string {
		return "[" + value + "]"
	}

	assert.EqualValues(t, "[f1], [f2], [f3]", quoteColumns(cols, quoteFunc, ","))
}
