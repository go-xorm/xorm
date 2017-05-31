// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMap(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateTable struct {
		Id   int64
		Name string
		Age  int
	}

	assert.NoError(t, testEngine.Sync2(new(UpdateTable)))
	var tb = UpdateTable{
		Name: "test",
		Age:  35,
	}
	_, err := testEngine.Insert(&tb)
	assert.NoError(t, err)

	cnt, err := testEngine.Table("update_table").Where("id = ?", tb.Id).Update(map[string]interface{}{
		"name": "test2",
		"age":  36,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
}

func TestUpdateLimit(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UpdateTable struct {
		Id   int64
		Name string
		Age  int
	}

	assert.NoError(t, testEngine.Sync2(new(UpdateTable)))
	var tb = UpdateTable{
		Name: "test1",
		Age:  35,
	}
	cnt, err := testEngine.Insert(&tb)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	tb.Name = "test2"
	tb.Id = 0
	cnt, err = testEngine.Insert(&tb)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	cnt, err = testEngine.OrderBy("name desc").Limit(1).Update(&UpdateTable{
		Age: 30,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var uts []UpdateTable
	err = testEngine.Find(&uts)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(uts))
	assert.EqualValues(t, 35, uts[0].Age)
	assert.EqualValues(t, 30, uts[1].Age)
}

type ForUpdate struct {
	Id   int64 `xorm:"pk"`
	Name string
}

func setupForUpdate(engine *Engine) error {
	v := new(ForUpdate)
	err := engine.DropTables(v)
	if err != nil {
		return err
	}
	err = engine.CreateTables(v)
	if err != nil {
		return err
	}

	list := []ForUpdate{
		{1, "data1"},
		{2, "data2"},
		{3, "data3"},
	}

	for _, f := range list {
		_, err = engine.Insert(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestForUpdate(t *testing.T) {
	if testEngine.DriverName() != "mysql" && testEngine.DriverName() != "mymysql" {
		return
	}

	err := setupForUpdate(testEngine)
	if err != nil {
		t.Error(err)
		return
	}

	session1 := testEngine.NewSession()
	session2 := testEngine.NewSession()
	session3 := testEngine.NewSession()
	defer session1.Close()
	defer session2.Close()
	defer session3.Close()

	// start transaction
	err = session1.Begin()
	if err != nil {
		t.Error(err)
		return
	}

	// use lock
	fList := make([]ForUpdate, 0)
	session1.ForUpdate()
	session1.Where("(id) = ?", 1)
	err = session1.Find(&fList)
	switch {
	case err != nil:
		t.Error(err)
		return
	case len(fList) != 1:
		t.Errorf("find not returned single row")
		return
	case fList[0].Name != "data1":
		t.Errorf("for_update.name must be `data1`")
		return
	}

	// wait for lock
	wg := &sync.WaitGroup{}

	// lock is used
	wg.Add(1)
	go func() {
		f2 := new(ForUpdate)
		session2.Where("(id) = ?", 1).ForUpdate()
		has, err := session2.Get(f2) // wait release lock
		switch {
		case err != nil:
			t.Error(err)
		case !has:
			t.Errorf("cannot find target row. for_update.id = 1")
		case f2.Name != "updated by session1":
			t.Errorf("read lock failed")
		}
		wg.Done()
	}()

	// lock is NOT used
	wg.Add(1)
	go func() {
		f3 := new(ForUpdate)
		session3.Where("(id) = ?", 1)
		has, err := session3.Get(f3) // wait release lock
		switch {
		case err != nil:
			t.Error(err)
		case !has:
			t.Errorf("cannot find target row. for_update.id = 1")
		case f3.Name != "data1":
			t.Errorf("read lock failed")
		}
		wg.Done()
	}()

	// wait for go rountines
	time.Sleep(50 * time.Millisecond)

	f := new(ForUpdate)
	f.Name = "updated by session1"
	session1.Where("(id) = ?", 1)
	session1.Update(f)

	// release lock
	err = session1.Commit()
	if err != nil {
		t.Error(err)
		return
	}

	wg.Wait()
}

func TestWithIn(t *testing.T) {
	type temp3 struct {
		Id   int64
		Name string `xorm:"Name"`
		Test bool   `xorm:"Test"`
	}

	assert.NoError(t, prepareEngine())
	assert.NoError(t, testEngine.Sync(new(temp3)))

	testEngine.Insert(&[]temp3{
		{
			Name: "user1",
		},
		{
			Name: "user1",
		},
		{
			Name: "user1",
		},
	})

	cnt, err := testEngine.In("Id", 1, 2, 3, 4).Update(&temp3{Name: "aa"}, &temp3{Name: "user1"})
	assert.NoError(t, err)
	assert.EqualValues(t, 3, cnt)
}
