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
	err := mymysqlDdlImport()
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
	err := mymysqlDdlImport()
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

func mymysqlDdlImport() error {
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

func BenchmarkMyMysqlNoCache(t *testing.B) {
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
	//engine.ShowSQL = true
	doBenchCacheFind(engine, t)
}

func BenchmarkMyMysqlCache(t *testing.B) {
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
	doBenchCacheFind(engine, t)
}
