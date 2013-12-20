package xorm

//
// +build windows

import (
	"testing"

	_ "github.com/lunny/godbc"
)

/*
CREATE DATABASE IF NOT EXISTS xorm_test CHARACTER SET
utf8 COLLATE utf8_general_ci;
*/

func newMssqlEngine() (*Engine, error) {
	return NewEngine("odbc", "driver={SQL Server};Server=192.168.20.135;Database=xorm_test; uid=sa; pwd=1234;")
}

func TestMssql(t *testing.T) {
	engine, err := newMssqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	testAll(engine, t)
	testAll2(engine, t)
}

func TestMssqlWithCache(t *testing.T) {
	engine, err := newMssqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	testAll(engine, t)
	testAll2(engine, t)
}

func BenchmarkMssqlNoCache(t *testing.B) {
	engine, err := newMssqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFind(engine, t)
}

func BenchmarkMssqlCache(t *testing.B) {
	engine, err := newMssqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	doBenchFind(engine, t)
}
