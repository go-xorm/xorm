package xorm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPluck(t *testing.T) {
	var err error
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))
	_, err = testEngine.Insert([]Userinfo{
		{
			Username: "Lenny",
		},
		{
			Username: "Jimmy",
		},
		{
			Username: "John Wick",
		},
	})
	assert.NoError(t, err)
	var ids []int64
	err = testEngine.Table("userinfo").Alias("u").Pluck("u.id", &ids)
	assert.NoError(t, err)
	t.Logf("%+v", ids)

	err = testEngine.Table("userinfo").Alias("u").Pluck("`u`.`id`", &ids)
	assert.NoError(t, err)
	t.Logf("%+v", ids)

	var names []string
	err = testEngine.Table("userinfo").Pluck("username", &names)
	assert.NoError(t, err)
	t.Logf("%+v", names)
}

func TestPluck2(t *testing.T) {
	var err error
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))
	_, err = testEngine.Insert([]Userinfo{
		{
			Username: "Lenny",
		},
		{
			Username: "Jimmy",
		},
		{
			Username: "John Wick",
		},
	})
	assert.NoError(t, err)
	var names []string
	err = testEngine.Pluck("username", &names, new(Userinfo))
	assert.NoError(t, err)
	t.Logf("%+v", names)
}

func TestPluck3(t *testing.T) {
	var err error
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))
	_, err = testEngine.Insert([]Userinfo{
		{
			Username: "Lenny",
		},
		{
			Username: "Jimmy",
		},
		{
			Username: "John Wick",
		},
	})
	assert.NoError(t, err)
	var names []string
	err = testEngine.Table("userinfo").Pluck("username", &names, new(Team))
	assert.NoError(t, err)
	t.Logf("%+v", names)
}

func TestPluck4(t *testing.T) {
	var err error
	assert.NoError(t, prepareEngine())
	assertSync(t, new(Userinfo))
	_, err = testEngine.Insert([]Userinfo{
		{
			Username: "Lenny",
		},
		{
			Username: "Jimmy",
		},
		{
			Username: "John Wick",
		},
	})
	assert.NoError(t, err)
	var names []string
	err = testEngine.Table("userinfo").Select("id, username as name").Pluck("name", &names, new(Team))
	assert.NoError(t, err)
	t.Logf("%+v", names)
}
