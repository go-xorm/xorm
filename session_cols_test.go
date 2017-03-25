// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import "testing"

func TestSetExpr(t *testing.T) {
	type User struct {
		Id   int64
		Show bool
	}

	testEngine.SetExpr("show", "NOT show").Id(1).Update(new(User))
}
