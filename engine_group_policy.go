// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"math/rand"
	"sync"
	"time"
)

// GroupPolicy is be used by chosing the current slave from slaves
type GroupPolicy interface {
	Slave(*EngineGroup) *Engine
}

// GroupPolicyHandler should be used when a function is a GroupPolicy
type GroupPolicyHandler func(*EngineGroup) *Engine

// Slave implements the chosen of slaves
func (h GroupPolicyHandler) Slave(eg *EngineGroup) *Engine {
	return h(eg)
}

// RandomPolicy implmentes randomly chose the slave of slaves
type RandomPolicy struct {
	r *rand.Rand
}

// NewRandomPolicy creates a RandomPolicy
func NewRandomPolicy() *RandomPolicy {
	return &RandomPolicy{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Slave randomly choses the slave of slaves
func (policy *RandomPolicy) Slave(g *EngineGroup) *Engine {
	return g.Slaves()[policy.r.Intn(len(g.Slaves()))]
}

// WeightRandomPolicy implmentes randomly chose the slave of slaves
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

// LeastConnPolicy implements GroupPolicy, every time will get the least connections slave
func LeastConnPolicy() GroupPolicyHandler {
	return func(g *EngineGroup) *Engine {
		var slaves = g.Slaves()
		connections := 0
		idx := 0
		for i := 0; i < len(slaves); i++ {
			openConnections := slaves[i].DB().Stats().OpenConnections
			if i == 0 {
				connections = openConnections
				idx = i
			} else if openConnections <= connections {
				connections = openConnections
				idx = i
			}
		}
		return slaves[idx]
	}
}
