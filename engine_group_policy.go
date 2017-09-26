// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"math/rand"
	"sync"
	"time"
)

type GroupPolicy interface {
	Slave(*EngineGroup) *Engine
}

type RandomPolicy struct {
	r *rand.Rand
}

func NewRandomPolicy() *RandomPolicy {
	return &RandomPolicy{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (policy *RandomPolicy) Slave(g *EngineGroup) *Engine {
	return g.Slaves()[policy.r.Intn(len(g.Slaves()))]
}

type WeightRandomPolicy struct {
	weights []int
	rands   []int
	r       *rand.Rand
}

func NewWeightRandomPolicy(weights []int) *WeightRandomPolicy {
	var rands = make([]int, 0, len(weights))
	for i := 0; i < len(weights); i++ {
		for n := 0; n < weights[i]; n++ {
			rands = append(rands, i)
		}
	}

	return &WeightRandomPolicy{
		weights: weights,
		rands:   rands,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (policy *WeightRandomPolicy) Slave(g *EngineGroup) *Engine {
	var slaves = g.Slaves()
	idx := policy.rands[policy.r.Intn(len(policy.rands))]
	if idx >= len(slaves) {
		idx = len(slaves) - 1
	}
	return slaves[idx]
}

type RoundRobinPolicy struct {
	pos  int
	lock sync.Mutex
}

func NewRoundRobinPolicy() *RoundRobinPolicy {
	return &RoundRobinPolicy{pos: -1}
}

func (policy *RoundRobinPolicy) Slave(g *EngineGroup) *Engine {
	var slaves = g.Slaves()
	var pos int
	policy.lock.Lock()
	policy.pos++
	if policy.pos >= len(slaves) {
		policy.pos = 0
	}
	pos = policy.pos
	policy.lock.Unlock()

	return slaves[pos]
}

type WeightRoundRobinPolicy struct {
	weights []int
	rands   []int
	r       *rand.Rand
	lock    sync.Mutex
	pos     int
}

func NewWeightRoundRobinPolicy(weights []int) *WeightRoundRobinPolicy {
	var rands = make([]int, 0, len(weights))
	for i := 0; i < len(weights); i++ {
		for n := 0; n < weights[i]; n++ {
			rands = append(rands, i)
		}
	}

	return &WeightRoundRobinPolicy{
		weights: weights,
		rands:   rands,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		pos:     -1,
	}
}

func (policy *WeightRoundRobinPolicy) Slave(g *EngineGroup) *Engine {
	var slaves = g.Slaves()
	var pos int
	policy.lock.Lock()
	policy.pos++
	if policy.pos >= len(policy.rands) {
		policy.pos = 0
	}
	pos = policy.pos
	policy.lock.Unlock()

	idx := policy.rands[pos]
	if idx >= len(slaves) {
		idx = len(slaves) - 1
	}
	return slaves[idx]
}

type LeastConnPolicy struct {
}

func NewLeastConnPolicy() *LeastConnPolicy {
	return &LeastConnPolicy{}
}

func (policy *LeastConnPolicy) Slave(g *EngineGroup) *Engine {
	var slaves = g.Slaves()
	connections := 0
	idx := 0
	for i, _ := range slaves {
		open_connections := slaves[i].DB().Stats().OpenConnections
		if i == 0 {
			connections = open_connections
			idx = i
		} else if open_connections <= connections {
			connections = open_connections
			idx = i
		}
	}
	return slaves[idx]
}
