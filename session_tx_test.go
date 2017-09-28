// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-xorm/core"
	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	counter := func() {
		total, err := testEngine.Count(&Userinfo{})
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("----now total %v records\n", total)
	}

	counter()
	//defer counter()

	session := testEngine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		t.Error(err)
		panic(err)
		return
	}

	user1 := Userinfo{Username: "xiaoxiao", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
		return
	}

	user2 := Userinfo{Username: "yyy"}
	_, err = session.Where("(id) = ?", 0).Update(&user2)
	if err != nil {
		session.Rollback()
		fmt.Println(err)
		//t.Error(err)
		return
	}

	_, err = session.Delete(&user2)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
		return
	}

	err = session.Commit()
	if err != nil {
		t.Error(err)
		panic(err)
		return
	}
	// panic(err) !nashtsai! should remove this
}

func TestCombineTransaction(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))

	counter := func() {
		total, err := testEngine.Count(&Userinfo{})
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("----now total %v records\n", total)
	}

	counter()
	//defer counter()
	session := testEngine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		t.Error(err)
		panic(err)
	}

	user1 := Userinfo{Username: "xiaoxiao2", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
	}
	user2 := Userinfo{Username: "zzz"}
	_, err = session.Where("id = ?", 0).Update(&user2)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
	}

	_, err = session.Exec("delete from userinfo where username = ?", user2.Username)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
	}

	err = session.Commit()
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func TestCombineTransactionSameMapper(t *testing.T) {
	assert.NoError(t, prepareEngine())

	oldMapper := testEngine.GetColumnMapper()
	testEngine.UnMapType(rValue(new(Userinfo)).Type())
	testEngine.SetMapper(core.SameMapper{})
	defer func() {
		testEngine.UnMapType(rValue(new(Userinfo)).Type())
		testEngine.SetMapper(oldMapper)
	}()

	assertSync(t, new(Userinfo))

	counter := func() {
		total, err := testEngine.Count(&Userinfo{})
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("----now total %v records\n", total)
	}

	counter()
	defer counter()
	session := testEngine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		t.Error(err)
		panic(err)
		return
	}

	user1 := Userinfo{Username: "xiaoxiao2", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
		return
	}

	user2 := Userinfo{Username: "zzz"}
	_, err = session.Where("(id) = ?", 0).Update(&user2)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
		return
	}

	_, err = session.Exec("delete from `Userinfo` where `Username` = ?", user2.Username)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
		return
	}

	err = session.Commit()
	if err != nil {
		t.Error(err)
		panic(err)
	}
}
