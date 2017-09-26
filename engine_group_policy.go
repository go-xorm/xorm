// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"math/rand"

	"time"
)

const (
	ENGINE_GROUP_POLICY_RANDOM = iota
	ENGINE_GROUP_POLICY_WEIGHTRANDOM
	ENGINE_GROUP_POLICY_ROUNDROBIN
	ENGINE_GROUP_POLICY_WEIGHTROUNDROBIN
	ENGINE_GROUP_POLICY_LEASTCONNECTIONS
)

type Policy interface {
	Slave() int
	SetEngineGroup(*EngineGroup)
}

type XormEngineGroupPolicy struct {
	pos    int
	slaves []int
	eg     *EngineGroup
	r      *rand.Rand
}

func (xgep *XormEngineGroupPolicy) SetEngineGroup(eg *EngineGroup) {
	xgep.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	xgep.eg = eg
}

func (xgep *XormEngineGroupPolicy) SetWeight() {
	for i, _ := range xgep.eg.weight {
		w := xgep.eg.weight[i]
		for n := 0; n < w; n++ {
			xgep.slaves = append(xgep.slaves, i)
		}
	}
}

func (xgep *XormEngineGroupPolicy) Slave() int {
	switch xgep.eg.p {
	case ENGINE_GROUP_POLICY_RANDOM:
		return xgep.Random()
	case ENGINE_GROUP_POLICY_WEIGHTRANDOM:
		return xgep.WeightRandom()
	case ENGINE_GROUP_POLICY_ROUNDROBIN:
		return xgep.RoundRobin()
	case ENGINE_GROUP_POLICY_WEIGHTROUNDROBIN:
		return xgep.WeightRoundRobin()
	case ENGINE_GROUP_POLICY_LEASTCONNECTIONS:
		return xgep.LeastConnections()
	default:
		return xgep.Random()
	}

}

func (xgep *XormEngineGroupPolicy) Random() int {
	if xgep.eg.s_count <= 1 {
		return 0
	}

	rnd := xgep.r.Intn(xgep.eg.s_count)
	return rnd
}

func (xgep *XormEngineGroupPolicy) WeightRandom() int {
	if xgep.eg.s_count <= 1 {
		return 0
	}

	xgep.SetWeight()
	s := len(xgep.slaves)
	rnd := xgep.r.Intn(s)
	return xgep.slaves[rnd]
}

func (xgep *XormEngineGroupPolicy) RoundRobin() int {
	if xgep.eg.s_count <= 1 {
		return 0
	}

	if xgep.pos >= xgep.eg.s_count {
		xgep.pos = 0
	}
	xgep.pos++

	return xgep.pos - 1
}

func (xgep *XormEngineGroupPolicy) WeightRoundRobin() int {
	if xgep.eg.s_count <= 1 {
		return 0
	}

	xgep.SetWeight()
	count := len(xgep.slaves)
	if xgep.pos >= count {
		xgep.pos = 0
	}
	xgep.pos++

	return xgep.slaves[xgep.pos-1]
}

func (xgep *XormEngineGroupPolicy) LeastConnections() int {
	if xgep.eg.s_count <= 1 {
		return 0
	}
	connections := 0
	slave := 0
	for i, _ := range xgep.eg.slaves {
		open_connections := xgep.eg.slaves[i].Stats()
		if i == 0 {
			connections = open_connections
			slave = i
		} else if open_connections <= connections {
			slave = i
			connections = open_connections
		}
	}
	return slave
}
