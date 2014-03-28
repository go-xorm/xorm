package xorm

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/ql-driver"
)

func newQlEngine() (*Engine, error) {
	os.Remove("./ql.db")
	return NewEngine("ql", "./ql.db")
}

func newQlDriverDB() (*sql.DB, error) {
	os.Remove("./ql.db")
	return sql.Open("ql", "./ql.db")
}

func TestQl(t *testing.T) {
	engine, err := newQlEngine()
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

func TestQlWithCache(t *testing.T) {
	engine, err := newQlEngine()
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

const (
	createTableQl = "CREATE TABLE IF NOT EXISTS `big_struct` (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `name` TEXT NULL, `title` TEXT NULL, `age` TEXT NULL, `alias` TEXT NULL, `nick_name` TEXT NULL);"
	dropTableQl   = "DROP TABLE IF EXISTS `big_struct`;"
)

func BenchmarkQlDriverInsert(t *testing.B) {
	doBenchDriver(newQlDriverDB, createTableQl, dropTableQl,
		doBenchDriverInsert, t)
}

func BenchmarkQlDriverFind(t *testing.B) {
	doBenchDriver(newQlDriverDB, createTableQl, dropTableQl,
		doBenchDriverFind, t)
}

func BenchmarkQlNoCacheInsert(t *testing.B) {
	t.StopTimer()
	engine, err := newQlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchInsert(engine, t)
}

func BenchmarkQlNoCacheFind(t *testing.B) {
	t.StopTimer()
	engine, err := newQlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFind(engine, t)
}

func BenchmarkQlNoCacheFindPtr(t *testing.B) {
	t.StopTimer()
	engine, err := newQlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFindPtr(engine, t)
}

func BenchmarkQlCacheInsert(t *testing.B) {
	t.StopTimer()
	engine, err := newQlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	doBenchInsert(engine, t)
}

func BenchmarkQlCacheFind(t *testing.B) {
	t.StopTimer()
	engine, err := newQlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	doBenchFind(engine, t)
}

func BenchmarkQlCacheFindPtr(t *testing.B) {
	t.StopTimer()
	engine, err := newQlEngine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	doBenchFindPtr(engine, t)
}
