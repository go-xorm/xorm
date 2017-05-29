// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"reflect"

	"github.com/go-xorm/core"
)

// EagerFind load 's belongs to tag field immedicatlly
func (session *Session) EagerFind(slices interface{}, cols ...string) error {
	/*v := reflect.ValueOf(slices)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return errors.New("only slice is supported")
	}

	if v.Len() <= 0 {
		return nil
	}

	vv := v.Index(0)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	tb, err := session.Engine.autoMapType(vv)
	if err != nil {
		return err
	}

	var pks = make(map[string][]core.PK)
	for i := 0; i < v.Len(); i++ {
		ev := v.Index(i)

		for _, col := range tb.Columns() {
			if len(cols) > 0 && !isStringInSlice(col.Name, cols) {
				continue
			}

			if col.AssociateTable != nil {
				if col.AssociateType == core.AssociateBelongsTo {
					colV, err := col.ValueOfV(&ev)
					if err != nil {
						return err
					}

					pk, err := session.Engine.idOfV(*colV)
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
						pks[col.Name] = append(pks[col.Name], pk)
					}
				}
			}
		}
	}

	for colName, pk := range pks {
		slice := reflect.MakeSlice(tp, 0, len(pk))
		err = session.In("", pk).Find(slice.Addr().Interafce())
		if err != nil {
			return err
		}

	}*/
	return nil
}

// EagerGet load bean's belongs to tag field immedicatlly
func (session *Session) EagerGet(bean interface{}, cols ...string) error {
	if session.isAutoClose {
		defer session.Close()
	}

	v := rValue(bean)
	tb, err := session.engine.autoMapType(v)
	if err != nil {
		return err
	}

	for _, col := range tb.Columns() {
		if len(cols) > 0 && !isStringInSlice(col.Name, cols) {
			continue
		}

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
