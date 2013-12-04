package xorm

import (
	_ "github.com/ziutek/mymysql/godrv"
	"testing"
)

/*
CREATE DATABASE IF NOT EXISTS xorm_test CHARACTER SET
utf8 COLLATE utf8_general_ci;
*/

var showTestSql bool = true

func TestMyMysql(t *testing.T) {
	engine, err := NewEngine("mymysql", "xorm_test/root/")
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

func TestMyMysqlWithCache(t *testing.T) {
	engine, err := NewEngine("mymysql", "xorm_test2/root/")
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

func newMyMysqlEngine() (*Engine, error) {
	return NewEngine("mymysql", "xorm_test2/root/")
}

func BenchmarkMyMysqlNoCacheInsert(t *testing.B) {
	engine, err := newMyMysqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	doBenchInsert(engine, t)
}

func BenchmarkMyMysqlNoCacheFind(t *testing.B) {
	engine, err := newMyMysqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFind(engine, t)
}

func BenchmarkMyMysqlNoCacheFindPtr(t *testing.B) {
	engine, err := newMyMysqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFindPtr(engine, t)
}

func BenchmarkMyMysqlCacheInsert(t *testing.B) {
	engine, err := newMyMysqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

	doBenchInsert(engine, t)
}

func BenchmarkMyMysqlCacheFind(t *testing.B) {
	engine, err := newMyMysqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

	doBenchFind(engine, t)
}

func BenchmarkMyMysqlCacheFindPtr(t *testing.B) {
	engine, err := newMyMysqlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

	doBenchFindPtr(engine, t)
}
