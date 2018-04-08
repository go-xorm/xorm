// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBelongsTo_Get(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Face1 struct {
		Id   int64
		Name string
	}

	type Nose1 struct {
		Id   int64
		Face Face1 `xorm:"belongs_to"`
	}

	err := testEngine.Sync2(new(Nose1), new(Face1))
	assert.NoError(t, err)

	var face = Face1{
		Name: "face1",
	}
	_, err = testEngine.Insert(&face)
	assert.NoError(t, err)

	var cfgFace Face1
	has, err := testEngine.Get(&cfgFace)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, face, cfgFace)

	var nose = Nose1{Face: face}
	_, err = testEngine.Insert(&nose)
	assert.NoError(t, err)

	var cfgNose Nose1
	has, err = testEngine.Get(&cfgNose)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, nose.Id, cfgNose.Id)
	assert.Equal(t, nose.Face.Id, cfgNose.Face.Id)
	assert.Equal(t, "", cfgNose.Face.Name)

	err = testEngine.Load(&cfgNose)
	assert.NoError(t, err)
	assert.Equal(t, nose.Id, cfgNose.Id)
	assert.Equal(t, nose.Face.Id, cfgNose.Face.Id)
	assert.Equal(t, "face1", cfgNose.Face.Name)

	var cfgNose2 Nose1
	has, err = testEngine.Cascade().Get(&cfgNose2)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, nose.Id, cfgNose2.Id)
	assert.Equal(t, nose.Face.Id, cfgNose2.Face.Id)
	assert.Equal(t, "face1", cfgNose2.Face.Name)
}

func TestBelongsTo_GetPtr(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Face2 struct {
		Id   int64
		Name string
	}

	type Nose2 struct {
		Id   int64
		Face *Face2 `xorm:"belongs_to"`
	}

	err := testEngine.Sync2(new(Nose2), new(Face2))
	assert.NoError(t, err)

	var face = Face2{
		Name: "face1",
	}
	_, err = testEngine.Insert(&face)
	assert.NoError(t, err)

	var cfgFace Face2
	has, err := testEngine.Get(&cfgFace)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, face, cfgFace)

	var nose = Nose2{Face: &face}
	_, err = testEngine.Insert(&nose)
	assert.NoError(t, err)

	var cfgNose Nose2
	has, err = testEngine.Get(&cfgNose)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, nose.Id, cfgNose.Id)
	assert.Equal(t, nose.Face.Id, cfgNose.Face.Id)

	err = testEngine.Load(&cfgNose)
	assert.NoError(t, err)
	assert.Equal(t, nose.Id, cfgNose.Id)
	assert.Equal(t, nose.Face.Id, cfgNose.Face.Id)
	assert.Equal(t, "face1", cfgNose.Face.Name)

	var cfgNose2 Nose2
	has, err = testEngine.Cascade().Get(&cfgNose2)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, nose.Id, cfgNose2.Id)
	assert.Equal(t, nose.Face.Id, cfgNose2.Face.Id)
	assert.Equal(t, "face1", cfgNose2.Face.Name)
}

func TestBelongsTo_Find(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Face3 struct {
		Id   int64
		Name string
	}

	type Nose3 struct {
		Id   int64
		Face Face3 `xorm:"belongs_to"`
	}

	err := testEngine.Sync2(new(Nose3), new(Face3))
	assert.NoError(t, err)

	var face1 = Face3{
		Name: "face1",
	}
	var face2 = Face3{
		Name: "face2",
	}
	_, err = testEngine.Insert(&face1, &face2)
	assert.NoError(t, err)

	var noses = []Nose3{
		{Face: face1},
		{Face: face2},
	}
	_, err = testEngine.Insert(&noses)
	assert.NoError(t, err)

	var noses1 []Nose3
	err = testEngine.Find(&noses1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(noses1))
	assert.Equal(t, face1.Id, noses1[0].Face.Id)
	assert.Equal(t, face2.Id, noses1[1].Face.Id)
	assert.Equal(t, "", noses1[0].Face.Name)
	assert.Equal(t, "", noses1[1].Face.Name)

	var noses2 []Nose3
	err = testEngine.Cascade().Find(&noses2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(noses2))
	assert.Equal(t, face1.Id, noses2[0].Face.Id)
	assert.Equal(t, face2.Id, noses2[1].Face.Id)
	assert.Equal(t, "face1", noses2[0].Face.Name)
	assert.Equal(t, "face2", noses2[1].Face.Name)

	err = testEngine.Load(noses1, "face")
	assert.NoError(t, err)
	assert.Equal(t, "face1", noses1[0].Face.Name)
	assert.Equal(t, "face2", noses1[1].Face.Name)
}

func TestBelongsTo_FindPtr(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Face4 struct {
		Id   int64
		Name string
	}

	type Nose4 struct {
		Id   int64
		Face *Face4 `xorm:"belongs_to"`
	}

	err := testEngine.Sync2(new(Nose4), new(Face4))
	assert.NoError(t, err)

	var face1 = Face4{
		Name: "face1",
	}
	var face2 = Face4{
		Name: "face2",
	}
	_, err = testEngine.Insert(&face1, &face2)
	assert.NoError(t, err)

	var noses = []Nose4{
		{Face: &face1},
		{Face: &face2},
	}
	_, err = testEngine.Insert(&noses)
	assert.NoError(t, err)

	var noses1 []Nose4
	err = testEngine.Find(&noses1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(noses1))
	assert.Equal(t, face1.Id, noses1[0].Face.Id)
	assert.Equal(t, face2.Id, noses1[1].Face.Id)
	assert.Equal(t, "", noses1[0].Face.Name)
	assert.Equal(t, "", noses1[1].Face.Name)

	var noses2 []Nose4
	err = testEngine.Cascade().Find(&noses2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(noses2))
	assert.NotNil(t, noses2[0].Face)
	assert.NotNil(t, noses2[1].Face)
	assert.Equal(t, face1.Id, noses2[0].Face.Id)
	assert.Equal(t, face2.Id, noses2[1].Face.Id)
	assert.Equal(t, "face1", noses2[0].Face.Name)
	assert.Equal(t, "face2", noses2[1].Face.Name)

	err = testEngine.Load(noses2, "face")
	assert.NoError(t, err)
}
