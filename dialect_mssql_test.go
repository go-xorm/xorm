// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"reflect"
	"testing"

	"github.com/go-xorm/core"
)

func TestParseMSSQL(t *testing.T) {
	tests := []struct {
		in       string
		expected string
		valid    bool
	}{
		{"sqlserver://sa:yourStrong(!)Password@localhost:1433?database=db&connection+timeout=30", "db", true},
		{"server=localhost;user id=sa;password=yourStrong(!)Password;database=db", "db", true},
	}

	driver := core.QueryDriver("mssql")

	for _, test := range tests {
		uri, err := driver.Parse("mssql", test.in)

		if err != nil && test.valid {
			t.Errorf("%q got unexpected error: %s", test.in, err)
		} else if err == nil && !reflect.DeepEqual(test.expected, uri.DbName) {
			t.Errorf("%q got: %#v want: %#v", test.in, uri.DbName, test.expected)
		}
	}
}
