// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"reflect"

	"github.com/go-xorm/core"
)

// EagerLoad load bean's belongs to tag field immedicatlly
func (session *Session) EagerLoad(bean interface{}, cols ...string) error {
	if session.isAutoClose {
		defer session.Close()
	}

	v := rValue(bean)
	tb, err := session.engine.autoMapType(v)
	if err != nil {
		return err
	}

	for _, col := range tb.Columns() {
		if col.AssociateTable != nil {
			if col.AssociateType == core.AssociateBelongsTo {
				colV, err := col.ValueOfV(&v)
				if err != nil {
					return err
				}

				pk, err := session.engine.idOfV(*colV)
				if err != nil {
					return err
				}
				var colPtr reflect.Value
				if colV.Kind() == reflect.Ptr {
					colPtr = *colV
				} else {
					colPtr = colV.Addr()
				}

				if !isZero(pk[0]) {
					has, err := session.ID(pk).get(colPtr.Interface())
					if err != nil {
						return err
					}
					if !has {
						return errors.New("load bean does not exist")
					}
				}
			}
		}
	}
	return nil
}
