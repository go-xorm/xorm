// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"strings"

	"github.com/go-xorm/core"
)

type EngineGroup struct {
	*Engine
	slaves []*Engine
	policy GroupPolicy
}

func NewEngineGroup(args1 interface{}, args2 interface{}, policies ...GroupPolicy) (*EngineGroup, error) {
	var eg EngineGroup
	if len(policies) > 0 {
		eg.policy = policies[0]
	} else {
		eg.policy = NewRandomPolicy()
	}

	driverName, ok1 := args1.(string)
	dataSourceNames, ok2 := args2.(string)
	if ok1 && ok2 {
		conns := strings.Split(dataSourceNames, ";")
		engines := make([]*Engine, len(conns))
		for i, conn := range conns {
			engine, err := NewEngine(driverName, conn)
			if err != nil {
				return nil, err
			}
			engine.engineGroup = &eg
			engines[i] = engine
		}

		eg.Engine = engines[0]
		eg.slaves = engines[1:]
		return &eg, nil
	}

	master, ok3 := args1.(*Engine)
	slaves, ok4 := args2.([]*Engine)
	if ok3 && ok4 {
		master.engineGroup = &eg
		for i := 0; i < len(slaves); i++ {
			slaves[i].engineGroup = &eg
		}
		eg.Engine = master
		eg.slaves = slaves
		return &eg, nil
	}
	return nil, ErrParamsType
}

func (eg *EngineGroup) SetPolicy(policy GroupPolicy) *EngineGroup {
	eg.policy = policy
	return eg
}

func (eg *EngineGroup) Master() *Engine {
	return eg.Engine
}

// Slave returns one of the physical databases which is a slave according the policy
func (eg *EngineGroup) Slave() *Engine {
	switch len(eg.slaves) {
	case 0:
		return eg.Engine
	case 1:
		return eg.slaves[0]
	}
	return eg.policy.Slave(eg)
}

func (eg *EngineGroup) Slaves() []*Engine {
	return eg.slaves
}

func (eg *EngineGroup) GetSlave(i int) *Engine {
	return eg.slaves[i]
}

// ShowSQL show SQL statement or not on logger if log level is great than INFO
func (eg *EngineGroup) ShowSQL(show ...bool) {
	eg.Engine.ShowSQL(show...)
	for i := 0; i < len(eg.slaves); i++ {
		eg.slaves[i].ShowSQL(show...)
	}
}

// ShowExecTime show SQL statement and execute time or not on logger if log level is great than INFO
func (eg *EngineGroup) ShowExecTime(show ...bool) {
	eg.Engine.ShowExecTime(show...)
	for i := 0; i < len(eg.slaves); i++ {
		eg.slaves[i].ShowExecTime(show...)
	}
}

// SetMapper set the name mapping rules
func (eg *EngineGroup) SetMapper(mapper core.IMapper) {
	eg.Engine.SetMapper(mapper)
	for i := 0; i < len(eg.slaves); i++ {
		eg.slaves[i].SetMapper(mapper)
	}
}

// SetLogger set the new logger
func (eg *EngineGroup) SetLogger(logger core.ILogger) {
	eg.Engine.SetLogger(logger)
	for i := 0; i < len(eg.slaves); i++ {
		eg.slaves[i].SetLogger(logger)
	}
}

// SetTableMapper set the table name mapping rule
func (eg *EngineGroup) SetTableMapper(mapper core.IMapper) {
	eg.Engine.TableMapper = mapper
	for i := 0; i < len(eg.slaves); i++ {
		eg.slaves[i].TableMapper = mapper
	}
}

// SetColumnMapper set the column name mapping rule
func (eg *EngineGroup) SetColumnMapper(mapper core.IMapper) {
	eg.Engine.ColumnMapper = mapper
	for i := 0; i < len(eg.slaves); i++ {
		eg.slaves[i].ColumnMapper = mapper
	}
}

// SetMaxOpenConns is only available for go 1.2+
func (eg *EngineGroup) SetMaxOpenConns(conns int) {
	eg.Engine.db.SetMaxOpenConns(conns)
	for i := 0; i < len(eg.slaves); i++ {
		eg.slaves[i].db.SetMaxOpenConns(conns)
	}
}

// SetMaxIdleConns set the max idle connections on pool, default is 2
func (eg *EngineGroup) SetMaxIdleConns(conns int) {
	eg.Engine.db.SetMaxIdleConns(conns)
	for i := 0; i < len(eg.slaves); i++ {
		eg.slaves[i].db.SetMaxIdleConns(conns)
	}
}

// Close the engine
func (eg *EngineGroup) Close() error {
	err := eg.Engine.Close()
	if err != nil {
		return err
	}

	for i := 0; i < len(eg.slaves); i++ {
		err := eg.slaves[i].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Ping tests if database is alive
func (eg *EngineGroup) Ping() error {
	if err := eg.Engine.Ping(); err != nil {
		return err
	}

	for _, slave := range eg.slaves {
		if err := slave.Ping(); err != nil {
			return err
		}
	}
	return nil
}
