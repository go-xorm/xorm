// Copyright 2018 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

// AutoTransaction Execute sql wrapped in a transaction(abbr as tx), tx will automatic commit if no errors occurred
func AutoTransaction(f func(*Session) (interface{}, error), engine *Engine) (interface{}, error) {
	session := engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, err
	}

	result, err := f(session)
	if err != nil {
		return nil, err
	}

	if err := session.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
