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

	type Face struct {
		Id   int64
		Name string
	}

	type Nose struct {
		Id   int64
		Face Face `xorm:"belongs_to"`
	}

	err := testEngine.Sync2(new(Nose), new(Face))
	assert.NoError(t, err)

	var face = Face{
		Name: "face1",
	}
	_, err = testEngine.Insert(&face)
	assert.NoError(t, err)

	var cfgFace Face
	has, err := testEngine.Get(&cfgFace)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, face, cfgFace)

	var nose = Nose{Face: face}
	_, err = testEngine.Insert(&nose)
	assert.NoError(t, err)

	var cfgNose Nose
	has, err = testEngine.Get(&cfgNose)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, nose.Id, cfgNose.Id)
	assert.Equal(t, nose.Face.Id, cfgNose.Face.Id)
	assert.Equal(t, "", cfgNose.Face.Name)

	err = testEngine.EagerGet(&cfgNose)
	assert.NoError(t, err)
	assert.Equal(t, nose.Id, cfgNose.Id)
	assert.Equal(t, nose.Face.Id, cfgNose.Face.Id)
	assert.Equal(t, "face1", cfgNose.Face.Name)

	var cfgNose2 Nose
	has, err = testEngine.Cascade().Get(&cfgNose2)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, nose.Id, cfgNose2.Id)
	assert.Equal(t, nose.Face.Id, cfgNose2.Face.Id)
	assert.Equal(t, "face1", cfgNose2.Face.Name)
}

func TestBelongsTo_GetPtr(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Face struct {
		Id   int64
		Name string
	}

	type Nose struct {
		Id   int64
		Face *Face `xorm:"belongs_to"`
	}

	err := testEngine.Sync2(new(Nose), new(Face))
	assert.NoError(t, err)

	var face = Face{
		Name: "face1",
	}
	_, err = testEngine.Insert(&face)
	assert.NoError(t, err)

	var cfgFace Face
	has, err := testEngine.Get(&cfgFace)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, face, cfgFace)

	var nose = Nose{Face: &face}
	_, err = testEngine.Insert(&nose)
	assert.NoError(t, err)

	var cfgNose Nose
	has, err = testEngine.Get(&cfgNose)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, nose.Id, cfgNose.Id)
	assert.Equal(t, nose.Face.Id, cfgNose.Face.Id)

	err = testEngine.EagerGet(&cfgNose)
	assert.NoError(t, err)
	assert.Equal(t, nose.Id, cfgNose.Id)
	assert.Equal(t, nose.Face.Id, cfgNose.Face.Id)
	assert.Equal(t, "face1", cfgNose.Face.Name)

	var cfgNose2 Nose
	has, err = testEngine.Cascade().Get(&cfgNose2)
	assert.NoError(t, err)
	assert.Equal(t, true, has)
	assert.Equal(t, nose.Id, cfgNose2.Id)
	assert.Equal(t, nose.Face.Id, cfgNose2.Face.Id)
	assert.Equal(t, "face1", cfgNose2.Face.Name)
}

func TestBelongsTo_Find(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Face struct {
		Id   int64
		Name string
	}

	type Nose struct {
		Id   int64
		Face Face `xorm:"belongs_to"`
	}

	err := testEngine.Sync2(new(Nose), new(Face))
	assert.NoError(t, err)

	var face1 = Face{
		Name: "face1",
	}
	var face2 = Face{
		Name: "face2",
	}
	_, err = testEngine.Insert(&face1, &face2)
	assert.NoError(t, err)

	var noses = []Nose{
		{Face: face1},
		{Face: face2},
	}
	_, err = testEngine.Insert(&noses)
	assert.NoError(t, err)

	var noses1 []Nose
	err = testEngine.Find(&noses1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(noses1))
	assert.Equal(t, face1.Id, noses1[0].Face.Id)
	assert.Equal(t, face2.Id, noses1[1].Face.Id)
	assert.Equal(t, "", noses1[0].Face.Name)
	assert.Equal(t, "", noses1[1].Face.Name)

	var noses2 []Nose
	err = testEngine.Cascade().Find(&noses2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(noses2))
	assert.Equal(t, face1.Id, noses2[0].Face.Id)
	assert.Equal(t, face2.Id, noses2[1].Face.Id)
	assert.Equal(t, "face1", noses2[0].Face.Name)
	assert.Equal(t, "face2", noses2[1].Face.Name)
}

func TestBelongsTo_FindPtr(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type Face struct {
		Id   int64
		Name string
	}

	type Nose struct {
		Id   int64
		Face *Face `xorm:"belongs_to"`
	}

	err := testEngine.Sync2(new(Nose), new(Face))
	assert.NoError(t, err)

	var face1 = Face{
		Name: "face1",
	}
	var face2 = Face{
		Name: "face2",
	}
	_, err = testEngine.Insert(&face1, &face2)
	assert.NoError(t, err)

	var noses = []Nose{
		{Face: &face1},
		{Face: &face2},
	}
	_, err = testEngine.Insert(&noses)
	assert.NoError(t, err)

	var noses1 []Nose
	err = testEngine.Find(&noses1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(noses1))
	assert.Equal(t, face1.Id, noses1[0].Face.Id)
	assert.Equal(t, face2.Id, noses1[1].Face.Id)
	assert.Equal(t, "", noses1[0].Face.Name)
	assert.Equal(t, "", noses1[1].Face.Name)

	var noses2 []Nose
	err = testEngine.Cascade().Find(&noses2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(noses2))
	assert.NotNil(t, noses2[0].Face)
	assert.NotNil(t, noses2[1].Face)
	assert.Equal(t, face1.Id, noses2[0].Face.Id)
	assert.Equal(t, face2.Id, noses2[1].Face.Id)
	assert.Equal(t, "face1", noses2[0].Face.Name)
	assert.Equal(t, "face2", noses2[1].Face.Name)
}
