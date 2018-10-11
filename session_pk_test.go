// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"testing"
	"time"

	"github.com/go-xorm/core"
	"github.com/stretchr/testify/assert"
)

type IntID struct {
	ID   int `xorm:"pk autoincr"`
	Name string
}

type Int16ID struct {
	ID   int16 `xorm:"pk autoincr"`
	Name string
}

type Int32ID struct {
	ID   int32 `xorm:"pk autoincr"`
	Name string
}

type UintID struct {
	ID   uint `xorm:"pk autoincr"`
	Name string
}

type Uint16ID struct {
	ID   uint16 `xorm:"pk autoincr"`
	Name string
}

type Uint32ID struct {
	ID   uint32 `xorm:"pk autoincr"`
	Name string
}

type Uint64ID struct {
	ID   uint64 `xorm:"pk autoincr"`
	Name string
}

type StringPK struct {
	ID   string `xorm:"pk notnull"`
	Name string
}

type ID int64
type MyIntPK struct {
	ID   ID `xorm:"pk autoincr"`
	Name string
}

type StrID string
type MyStringPK struct {
	ID   StrID `xorm:"pk notnull"`
	Name string
}

func TestIntID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&IntID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&IntID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&IntID{Name: "test"})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(IntID)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans := make([]IntID, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans2 := make(map[int]IntID)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&IntID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestInt16ID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&Int16ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&Int16ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&Int16ID{Name: "test"})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(Int16ID)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans := make([]Int16ID, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans2 := make(map[int16]Int16ID, 0)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&Int16ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestInt32ID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&Int32ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&Int32ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&Int32ID{Name: "test"})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(Int32ID)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans := make([]Int32ID, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans2 := make(map[int32]Int32ID, 0)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&Int32ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestUintID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&UintID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&UintID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&UintID{Name: "test"})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	var inserts = []UintID{
		{Name: "test1"},
		{Name: "test2"},
	}
	cnt, err = testEngine.Insert(&inserts)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 2 {
		err = errors.New("insert count should be two")
		t.Error(err)
		panic(err)
	}

	bean := new(UintID)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans := make([]UintID, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 3 {
		err = errors.New("get count should be three")
		t.Error(err)
		panic(err)
	}

	beans2 := make(map[uint]UintID, 0)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 3 {
		err = errors.New("get count should be three")
		t.Error(err)
		panic(err)
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&UintID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestUint16ID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&Uint16ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&Uint16ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&Uint16ID{Name: "test"})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(Uint16ID)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans := make([]Uint16ID, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans2 := make(map[uint16]Uint16ID, 0)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&Uint16ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestUint32ID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&Uint32ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&Uint32ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&Uint32ID{Name: "test"})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(Uint32ID)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans := make([]Uint32ID, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans2 := make(map[uint32]Uint32ID, 0)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&Uint32ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestUint64ID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&Uint64ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&Uint64ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	idbean := &Uint64ID{Name: "test"}
	cnt, err := testEngine.Insert(idbean)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(Uint64ID)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if bean.ID != idbean.ID {
		panic(errors.New("should be equal"))
	}

	beans := make([]Uint64ID, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if *bean != beans[0] {
		panic(errors.New("should be equal"))
	}

	beans2 := make(map[uint64]Uint64ID, 0)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if *bean != beans2[bean.ID] {
		panic(errors.New("should be equal"))
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&Uint64ID{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestStringPK(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&StringPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&StringPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&StringPK{ID: "1-1-2", Name: "test"})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(StringPK)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans := make([]StringPK, 0)
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	beans2 := make(map[string]StringPK)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&StringPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

type CompositeKey struct {
	ID1       int64 `xorm:"id1 pk"`
	ID2       int64 `xorm:"id2 pk"`
	UpdateStr string
}

func TestCompositeKey(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&CompositeKey{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&CompositeKey{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&CompositeKey{11, 22, ""})
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("failed to insert CompositeKey{11, 22}"))
	}

	cnt, err = testEngine.Insert(&CompositeKey{11, 22, ""})
	if err == nil || cnt == 1 {
		t.Error(errors.New("inserted CompositeKey{11, 22}"))
	}

	var compositeKeyVal CompositeKey
	has, err := testEngine.ID(core.PK{11, 22}).Get(&compositeKeyVal)
	if err != nil {
		t.Error(err)
	} else if !has {
		t.Error(errors.New("can't get CompositeKey{11, 22}"))
	}

	var compositeKeyVal2 CompositeKey
	// test passing PK ptr, this test seem failed withCache
	has, err = testEngine.ID(&core.PK{11, 22}).Get(&compositeKeyVal2)
	if err != nil {
		t.Error(err)
	} else if !has {
		t.Error(errors.New("can't get CompositeKey{11, 22}"))
	}

	if compositeKeyVal != compositeKeyVal2 {
		t.Error(errors.New("should be equal"))
	}

	var cps = make([]CompositeKey, 0)
	err = testEngine.Find(&cps)
	if err != nil {
		t.Error(err)
	}
	if len(cps) != 1 {
		t.Error(errors.New("should has one record"))
	}
	if cps[0] != compositeKeyVal {
		t.Error(errors.New("should be equal"))
	}

	cnt, err = testEngine.Insert(&CompositeKey{22, 22, ""})
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("failed to insert CompositeKey{22, 22}"))
	}

	cps = make([]CompositeKey, 0)
	err = testEngine.Find(&cps)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(cps), "should has two record")
	assert.EqualValues(t, compositeKeyVal, cps[0], "should be equeal")

	compositeKeyVal = CompositeKey{UpdateStr: "test1"}
	cnt, err = testEngine.ID(core.PK{11, 22}).Update(&compositeKeyVal)
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("can't update CompositeKey{11, 22}"))
	}

	cnt, err = testEngine.ID(core.PK{11, 22}).Delete(&CompositeKey{})
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("can't delete CompositeKey{11, 22}"))
	}
}

func TestCompositeKey2(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type User struct {
		UserID   string `xorm:"varchar(19) not null pk"`
		NickName string `xorm:"varchar(19) not null"`
		GameID   uint32 `xorm:"integer pk"`
		Score    int32  `xorm:"integer"`
	}

	err := testEngine.DropTables(&User{})

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&User{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&User{"11", "nick", 22, 5})
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("failed to insert User{11, 22}"))
	}

	cnt, err = testEngine.Insert(&User{"11", "nick", 22, 6})
	if err == nil || cnt == 1 {
		t.Error(errors.New("inserted User{11, 22}"))
	}

	var user User
	has, err := testEngine.ID(core.PK{"11", 22}).Get(&user)
	if err != nil {
		t.Error(err)
	} else if !has {
		t.Error(errors.New("can't get User{11, 22}"))
	}

	// test passing PK ptr, this test seem failed withCache
	has, err = testEngine.ID(&core.PK{"11", 22}).Get(&user)
	if err != nil {
		t.Error(err)
	} else if !has {
		t.Error(errors.New("can't get User{11, 22}"))
	}

	user = User{NickName: "test1"}
	cnt, err = testEngine.ID(core.PK{"11", 22}).Update(&user)
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("can't update User{11, 22}"))
	}

	cnt, err = testEngine.ID(core.PK{"11", 22}).Delete(&User{})
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("can't delete CompositeKey{11, 22}"))
	}
}

type MyString string
type UserPK2 struct {
	UserID   MyString `xorm:"varchar(19) not null pk"`
	NickName string   `xorm:"varchar(19) not null"`
	GameID   uint32   `xorm:"integer pk"`
	Score    int32    `xorm:"integer"`
}

func TestCompositeKey3(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&UserPK2{})

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&UserPK2{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	cnt, err := testEngine.Insert(&UserPK2{"11", "nick", 22, 5})
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("failed to insert User{11, 22}"))
	}

	cnt, err = testEngine.Insert(&UserPK2{"11", "nick", 22, 6})
	if err == nil || cnt == 1 {
		t.Error(errors.New("inserted User{11, 22}"))
	}

	var user UserPK2
	has, err := testEngine.ID(core.PK{"11", 22}).Get(&user)
	if err != nil {
		t.Error(err)
	} else if !has {
		t.Error(errors.New("can't get User{11, 22}"))
	}

	// test passing PK ptr, this test seem failed withCache
	has, err = testEngine.ID(&core.PK{"11", 22}).Get(&user)
	if err != nil {
		t.Error(err)
	} else if !has {
		t.Error(errors.New("can't get User{11, 22}"))
	}

	user = UserPK2{NickName: "test1"}
	cnt, err = testEngine.ID(core.PK{"11", 22}).Update(&user)
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("can't update User{11, 22}"))
	}

	cnt, err = testEngine.ID(core.PK{"11", 22}).Delete(&UserPK2{})
	if err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Error(errors.New("can't delete CompositeKey{11, 22}"))
	}
}

func TestMyIntID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&MyIntPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&MyIntPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	idbean := &MyIntPK{Name: "test"}
	cnt, err := testEngine.Insert(idbean)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(MyIntPK)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if bean.ID != idbean.ID {
		panic(errors.New("should be equal"))
	}

	var beans []MyIntPK
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if *bean != beans[0] {
		panic(errors.New("should be equal"))
	}

	beans2 := make(map[ID]MyIntPK, 0)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if *bean != beans2[bean.ID] {
		panic(errors.New("should be equal"))
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&MyIntPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestMyStringID(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&MyStringPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&MyStringPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	idbean := &MyStringPK{ID: "1111", Name: "test"}
	cnt, err := testEngine.Insert(idbean)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}

	bean := new(MyStringPK)
	has, err := testEngine.Get(bean)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !has {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if bean.ID != idbean.ID {
		panic(errors.New("should be equal"))
	}

	var beans []MyStringPK
	err = testEngine.Find(&beans)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if *bean != beans[0] {
		panic(errors.New("should be equal"))
	}

	beans2 := make(map[StrID]MyStringPK, 0)
	err = testEngine.Find(&beans2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(beans2) != 1 {
		err = errors.New("get count should be one")
		t.Error(err)
		panic(err)
	}

	if *bean != beans2[bean.ID] {
		panic(errors.New("should be equal"))
	}

	cnt, err = testEngine.ID(bean.ID).Delete(&MyStringPK{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if cnt != 1 {
		err = errors.New("insert count should be one")
		t.Error(err)
		panic(err)
	}
}

func TestSingleAutoIncrColumn(t *testing.T) {
	type Account struct {
		ID int64 `xorm:"pk autoincr"`
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(Account))

	_, err := testEngine.Insert(&Account{})
	assert.NoError(t, err)
}

func TestCompositePK(t *testing.T) {
	type TaskSolution struct {
		UID     string    `xorm:"notnull pk UUID 'uid'"`
		TID     string    `xorm:"notnull pk UUID 'tid'"`
		Created time.Time `xorm:"created"`
		Updated time.Time `xorm:"updated"`
	}

	assert.NoError(t, prepareEngine())

	tables1, err := testEngine.DBMetas()
	assert.NoError(t, err)

	assertSync(t, new(TaskSolution))
	assert.NoError(t, testEngine.Sync2(new(TaskSolution)))

	tables2, err := testEngine.DBMetas()
	assert.NoError(t, err)
	assert.EqualValues(t, 1+len(tables1), len(tables2))

	var table *core.Table
	for _, t := range tables2 {
		if t.Name == testEngine.GetTableMapper().Obj2Table("TaskSolution") {
			table = t
			break
		}
	}

	assert.NotEqual(t, nil, table)

	pkCols := table.PKColumns()
	assert.EqualValues(t, 2, len(pkCols))
	assert.EqualValues(t, "uid", pkCols[0].Name)
	assert.EqualValues(t, "tid", pkCols[1].Name)
}

func TestNoPKIDQueryUpdate(t *testing.T) {
	type NoPKTable struct {
		Username string
	}

	assert.NoError(t, prepareEngine())
	assertSync(t, new(NoPKTable))

	cnt, err := testEngine.Insert(&NoPKTable{
		Username: "test",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var res NoPKTable
	has, err := testEngine.ID("test").Get(&res)
	assert.Error(t, err)
	assert.False(t, has)

	cnt, err = testEngine.ID("test").Update(&NoPKTable{
		Username: "test1",
	})
	assert.Error(t, err)
	assert.EqualValues(t, 0, cnt)

	type UnvalidPKTable struct {
		ID       int `xorm:"id"`
		Username string
	}

	assertSync(t, new(UnvalidPKTable))

	cnt, err = testEngine.Insert(&UnvalidPKTable{
		ID:       1,
		Username: "test",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var res2 UnvalidPKTable
	has, err = testEngine.ID(1).Get(&res2)
	assert.Error(t, err)
	assert.False(t, has)

	cnt, err = testEngine.ID(1).Update(&UnvalidPKTable{
		Username: "test1",
	})
	assert.Error(t, err)
	assert.EqualValues(t, 0, cnt)
}
