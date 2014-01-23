package xorm

import (
	"database/sql"
	"testing"

	"github.com/lunny/xorm"
	_ "github.com/ziutek/mymysql/godrv"
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
	engine, err := xorm.NewEngine("mymysql", "xorm_test/root/")
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
	testAll3(engine, t)
}

func TestMyMysqlWithCache(t *testing.T) {
	err := mymysqlDdlImport()
	if err != nil {
		t.Error(err)
		return
	}
	engine, err := xorm.NewEngine("mymysql", "xorm_test2/root/")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	testAll(engine, t)
	testAll2(engine, t)
}

func newMyMysqlEngine() (*xorm.Engine, error) {
	return xorm.NewEngine("mymysql", "xorm_test2/root/")
}

func newMyMysqlDriverDB() (*sql.DB, error) {
	return sql.Open("mymysql", "xorm_test2/root/")
}

func BenchmarkMyMysqlDriverInsert(t *testing.B) {
	doBenchDriver(newMyMysqlDriverDB, createTableMySql, dropTableMySql,
		doBenchDriverInsert, t)
}

func BenchmarkMyMysqlDriverFind(t *testing.B) {
	doBenchDriver(newMyMysqlDriverDB, createTableMySql, dropTableMySql,
		doBenchDriverFind, t)
}

func mymysqlDdlImport() error {
	engine, err := xorm.NewEngine("mymysql", "/root/")
	if err != nil {
		return err
	}
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	sqlResults, _ := engine.Import("testdata/mysql_ddl.sql")
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
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))

	doBenchInsert(engine, t)
}

func BenchmarkMyMysqlCacheFind(t *testing.B) {
	engine, err := newMyMysqlEngine()
	if err != nil {
		t.Error(err)
		return
	}

	defer engine.Close()
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))

	doBenchFind(engine, t)
}

func BenchmarkMyMysqlCacheFindPtr(t *testing.B) {
	engine, err := newMyMysqlEngine()
	if err != nil {
		t.Error(err)
		return
	}

	defer engine.Close()
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))

	doBenchFindPtr(engine, t)
}
