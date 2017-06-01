// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func TestTimeUserTime(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type TimeUser struct {
		Id       string
		OperTime time.Time
	}

	assertSync(t, new(TimeUser))

	var user = TimeUser{
		Id:       "lunny",
		OperTime: time.Now(),
	}

	fmt.Println("user", user.OperTime)

	cnt, err := testEngine.Insert(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var user2 TimeUser
	has, err := testEngine.Get(&user2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, user.OperTime.Unix(), user2.OperTime.Unix())
	assert.EqualValues(t, formatTime(user.OperTime), formatTime(user2.OperTime))
	fmt.Println("user2", user2.OperTime)
}

func TestTimeUserTimeDiffLoc(t *testing.T) {
	assert.NoError(t, prepareEngine())
	loc, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)
	testEngine.TZLocation = loc
	dbLoc, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)
	testEngine.DatabaseTZ = dbLoc

	type TimeUser struct {
		Id       string
		OperTime time.Time
	}

	assertSync(t, new(TimeUser))

	var user = TimeUser{
		Id:       "lunny",
		OperTime: time.Now(),
	}

	fmt.Println("user", user.OperTime)

	cnt, err := testEngine.Insert(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var user2 TimeUser
	has, err := testEngine.Get(&user2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, user.OperTime.Unix(), user2.OperTime.Unix())
	assert.EqualValues(t, formatTime(user.OperTime), formatTime(user2.OperTime))
	fmt.Println("user2", user2.OperTime)
}

func TestTimeUserCreated(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UserCreated struct {
		Id        string
		CreatedAt time.Time `xorm:"created"`
	}

	assertSync(t, new(UserCreated))

	var user = UserCreated{
		Id: "lunny",
	}

	fmt.Println("user", user.CreatedAt)

	cnt, err := testEngine.Insert(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var user2 UserCreated
	has, err := testEngine.Get(&user2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, user.CreatedAt.Unix(), user2.CreatedAt.Unix())
	assert.EqualValues(t, formatTime(user.CreatedAt), formatTime(user2.CreatedAt))
	fmt.Println("user2", user2.CreatedAt)
}

func TestTimeUserCreatedDiffLoc(t *testing.T) {
	assert.NoError(t, prepareEngine())
	loc, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)
	testEngine.TZLocation = loc
	dbLoc, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)
	testEngine.DatabaseTZ = dbLoc

	type UserCreated struct {
		Id        string
		CreatedAt time.Time `xorm:"created"`
	}

	assertSync(t, new(UserCreated))

	var user = UserCreated{
		Id: "lunny",
	}

	fmt.Println("user", user.CreatedAt)

	cnt, err := testEngine.Insert(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var user2 UserCreated
	has, err := testEngine.Get(&user2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, user.CreatedAt.Unix(), user2.CreatedAt.Unix())
	assert.EqualValues(t, formatTime(user.CreatedAt), formatTime(user2.CreatedAt))
	fmt.Println("user2", user2.CreatedAt)
}

func TestTimeUserUpdated(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type UserUpdated struct {
		Id        string
		CreatedAt time.Time `xorm:"created"`
		UpdatedAt time.Time `xorm:"updated"`
	}

	assertSync(t, new(UserUpdated))

	var user = UserUpdated{
		Id: "lunny",
	}

	fmt.Println("user", user.CreatedAt, user.UpdatedAt)

	cnt, err := testEngine.Insert(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var user2 UserUpdated
	has, err := testEngine.Get(&user2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, user.CreatedAt.Unix(), user2.CreatedAt.Unix())
	assert.EqualValues(t, formatTime(user.CreatedAt), formatTime(user2.CreatedAt))
	assert.EqualValues(t, user.UpdatedAt.Unix(), user2.UpdatedAt.Unix())
	assert.EqualValues(t, formatTime(user.UpdatedAt), formatTime(user2.UpdatedAt))
	fmt.Println("user2", user2.CreatedAt, user2.UpdatedAt)

	var user3 = UserUpdated{
		Id: "lunny2",
	}

	cnt, err = testEngine.Update(&user3)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
	assert.True(t, user.UpdatedAt.Unix() <= user3.UpdatedAt.Unix())

	var user4 UserUpdated
	has, err = testEngine.Get(&user4)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, user.CreatedAt.Unix(), user4.CreatedAt.Unix())
	assert.EqualValues(t, formatTime(user.CreatedAt), formatTime(user4.CreatedAt))
	assert.EqualValues(t, user3.UpdatedAt.Unix(), user4.UpdatedAt.Unix())
	assert.EqualValues(t, formatTime(user3.UpdatedAt), formatTime(user4.UpdatedAt))
	fmt.Println("user3", user.CreatedAt, user4.UpdatedAt)
}

func TestTimeUserUpdatedDiffLoc(t *testing.T) {
	assert.NoError(t, prepareEngine())
	loc, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)
	testEngine.TZLocation = loc
	dbLoc, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)
	testEngine.DatabaseTZ = dbLoc

	type UserUpdated struct {
		Id        string
		CreatedAt time.Time `xorm:"created"`
		UpdatedAt time.Time `xorm:"updated"`
	}

	assertSync(t, new(UserUpdated))

	var user = UserUpdated{
		Id: "lunny",
	}

	fmt.Println("user", user.CreatedAt, user.UpdatedAt)

	cnt, err := testEngine.Insert(&user)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var user2 UserUpdated
	has, err := testEngine.Get(&user2)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, user.CreatedAt.Unix(), user2.CreatedAt.Unix())
	assert.EqualValues(t, formatTime(user.CreatedAt), formatTime(user2.CreatedAt))
	assert.EqualValues(t, user.UpdatedAt.Unix(), user2.UpdatedAt.Unix())
	assert.EqualValues(t, formatTime(user.UpdatedAt), formatTime(user2.UpdatedAt))
	fmt.Println("user2", user2.CreatedAt, user2.UpdatedAt)

	var user3 = UserUpdated{
		Id: "lunny2",
	}

	cnt, err = testEngine.Update(&user3)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)
	assert.True(t, user.UpdatedAt.Unix() <= user3.UpdatedAt.Unix())

	var user4 UserUpdated
	has, err = testEngine.Get(&user4)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.EqualValues(t, user.CreatedAt.Unix(), user4.CreatedAt.Unix())
	assert.EqualValues(t, formatTime(user.CreatedAt), formatTime(user4.CreatedAt))
	assert.EqualValues(t, user3.UpdatedAt.Unix(), user4.UpdatedAt.Unix())
	assert.EqualValues(t, formatTime(user3.UpdatedAt), formatTime(user4.UpdatedAt))
	fmt.Println("user3", user.CreatedAt, user4.UpdatedAt)
}
