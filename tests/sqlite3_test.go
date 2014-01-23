package xorm

import (
	"database/sql"
	"os"
	"testing"

	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
)

func newSqlite3Engine() (*xorm.Engine, error) {
	os.Remove("./test.db")
	return xorm.NewEngine("sqlite3", "./test.db")
}

func newSqlite3DriverDB() (*sql.DB, error) {
	os.Remove("./test.db")
	return sql.Open("sqlite3", "./test.db")
}

func TestSqlite3(t *testing.T) {
	engine, err := newSqlite3Engine()
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
	testAllSnakeMapper(engine, t)
	testAll2(engine, t)
	testAll3(engine, t)
}

func TestSqlite3WithCache(t *testing.T) {
	engine, err := newSqlite3Engine()
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
	testAllSnakeMapper(engine, t)
	testAll2(engine, t)
}

func TestSqlite3SameMapper(t *testing.T) {
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetMapper(xorm.SameMapper{})
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	testAll(engine, t)
	testAllSameMapper(engine, t)
	testAll2(engine, t)
	testAll3(engine, t)
}

func TestSqlite3WithCacheSameMapper(t *testing.T) {
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetMapper(xorm.SameMapper{})
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	testAll(engine, t)
	testAllSameMapper(engine, t)
	testAll2(engine, t)
}

const (
	createTableSqlite3 = "CREATE TABLE IF NOT EXISTS `big_struct` (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `name` TEXT NULL, `title` TEXT NULL, `age` TEXT NULL, `alias` TEXT NULL, `nick_name` TEXT NULL);"
	dropTableSqlite3   = "DROP TABLE IF EXISTS `big_struct`;"
)

func BenchmarkSqlite3DriverInsert(t *testing.B) {
	doBenchDriver(newSqlite3DriverDB, createTableSqlite3, dropTableSqlite3,
		doBenchDriverInsert, t)
}

func BenchmarkSqlite3DriverFind(t *testing.B) {
	doBenchDriver(newSqlite3DriverDB, createTableSqlite3, dropTableSqlite3,
		doBenchDriverFind, t)
}

func BenchmarkSqlite3NoCacheInsert(t *testing.B) {
	t.StopTimer()
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchInsert(engine, t)
}

func BenchmarkSqlite3NoCacheFind(t *testing.B) {
	t.StopTimer()
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFind(engine, t)
}

func BenchmarkSqlite3NoCacheFindPtr(t *testing.B) {
	t.StopTimer()
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFindPtr(engine, t)
}

func BenchmarkSqlite3CacheInsert(t *testing.B) {
	t.StopTimer()
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))
	doBenchInsert(engine, t)
}

func BenchmarkSqlite3CacheFind(t *testing.B) {
	t.StopTimer()
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))
	doBenchFind(engine, t)
}

func BenchmarkSqlite3CacheFindPtr(t *testing.B) {
	t.StopTimer()
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))
	doBenchFindPtr(engine, t)
}
