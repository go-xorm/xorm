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
	err := mysqlDdlImport()
	if err != nil {
		t.Error(err)
		return
	}
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

	sqlResults, _ := engine.Import("tests/mysql_ddl.sql")
	engine.LogDebug("sql results: %v", sqlResults)

	testAll(engine, t)
	testAll2(engine, t)
}

func TestMyMysqlWithCache(t *testing.T) {
	err := mysqlDdlImport()
	if err != nil {
		t.Error(err)
		return
	}
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

	sqlResults, _ := engine.Import("tests/mysql_ddl.sql")
	engine.LogDebug("sql results: %v", sqlResults)

	testAll(engine, t)
	testAll2(engine, t)
}

func newMyMysqlEngine() (*Engine, error) {
	return NewEngine("mymysql", "xorm_test2/root/")
}

func mysqlDdlImport() error {
	engine, err := NewEngine("mymysql", "/root/")
	if err != nil {
		return err
	}
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	sqlResults, _ := engine.Import("tests/mysql_ddl.sql")
	engine.LogDebug("sql results: %v", sqlResults)
	engine.Close()
	return nil
}

func BenchmarkMyMysqlNoCacheInsert(t *testing.B) {
	engine, err := newMyMysqlEngine()
	if err != nil {
		t.Error(err)
		return
	}
	defer engine.Close()

	doBenchInsert(engine, t)
}

func BenchmarkMyMysqlNoCacheFind(t *testing.B) {
	engine, err := newMyMysqlEngine()
	if err != nil {
		t.Error(err)
		return
	}
	defer engine.Close()

	//engine.ShowSQL = true
	doBenchFind(engine, t)
}

func BenchmarkMyMysqlNoCacheFindPtr(t *testing.B) {
	engine, err := newMyMysqlEngine()
	if err != nil {
		t.Error(err)
		return
	}
	defer engine.Close()

	//engine.ShowSQL = true
	doBenchFindPtr(engine, t)
}

func BenchmarkMyMysqlCacheInsert(t *testing.B) {
	engine, err := newMyMysqlEngine()
	if err != nil {
		t.Error(err)
		return
	}

	defer engine.Close()
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

	doBenchInsert(engine, t)
}

func BenchmarkMyMysqlCacheFind(t *testing.B) {
	engine, err := newMyMysqlEngine()
	if err != nil {
		t.Error(err)
		return
	}

	defer engine.Close()
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

	doBenchFind(engine, t)
}

func BenchmarkMyMysqlCacheFindPtr(t *testing.B) {
	engine, err := newMyMysqlEngine()
	if err != nil {
		t.Error(err)
		return
	}

	defer engine.Close()
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

	doBenchFindPtr(engine, t)
}
