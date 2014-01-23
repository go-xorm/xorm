package xorm

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/lunny/xorm"
)

//var connStr string = "dbname=xorm_test user=lunny password=1234 sslmode=disable"

var connStr string = "dbname=xorm_test sslmode=disable"

func newPostgresEngine() (*xorm.Engine, error) {
	orm, err := xorm.NewEngine("postgres", connStr)
	if err != nil {
		return nil, err
	}
	tables, err := orm.DBMetas()
	if err != nil {
		return nil, err
	}
	for _, table := range tables {
		_, err = orm.Exec("drop table \"" + table.Name + "\"")
		if err != nil {
			return nil, err
		}
	}

	return orm, err
}

func newPostgresDriverDB() (*sql.DB, error) {
	return sql.Open("postgres", connStr)
}

func TestPostgres(t *testing.T) {
	engine, err := newPostgresEngine()
	if err != nil {
		t.Error(err)
		return
	}
	defer engine.Close()
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	testAll(engine, t)
	testAllSnakeMapper(engine, t)
	testAll2(engine, t)
	testAll3(engine, t)
}

func TestPostgresWithCache(t *testing.T) {
	engine, err := newPostgresEngine()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))
	defer engine.Close()
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	testAll(engine, t)
	testAllSnakeMapper(engine, t)
	testAll2(engine, t)
}

func TestPostgresSameMapper(t *testing.T) {
	engine, err := newPostgresEngine()
	if err != nil {
		t.Error(err)
		return
	}
	defer engine.Close()
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

func TestPostgresWithCacheSameMapper(t *testing.T) {
	engine, err := newPostgresEngine()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))
	defer engine.Close()
	engine.SetMapper(xorm.SameMapper{})
	engine.ShowSQL = showTestSql
	engine.ShowErr = showTestSql
	engine.ShowWarn = showTestSql
	engine.ShowDebug = showTestSql

	testAll(engine, t)
	testAllSameMapper(engine, t)
	testAll2(engine, t)
}

const (
	createTablePostgres = `CREATE TABLE IF NOT EXISTS "big_struct" ("id" SERIAL PRIMARY KEY  NOT NULL, "name" VARCHAR(255) NULL, "title" VARCHAR(255) NULL, "age" VARCHAR(255) NULL, "alias" VARCHAR(255) NULL, "nick_name" VARCHAR(255) NULL);`
	dropTablePostgres   = `DROP TABLE IF EXISTS "big_struct";`
)

func BenchmarkPostgresDriverInsert(t *testing.B) {
	doBenchDriver(newPostgresDriverDB, createTablePostgres, dropTablePostgres,
		doBenchDriverInsert, t)
}

func BenchmarkPostgresDriverFind(t *testing.B) {
	doBenchDriver(newPostgresDriverDB, createTablePostgres, dropTablePostgres,
		doBenchDriverFind, t)
}

func BenchmarkPostgresNoCacheInsert(t *testing.B) {
	engine, err := newPostgresEngine()

	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchInsert(engine, t)
}

func BenchmarkPostgresNoCacheFind(t *testing.B) {
	engine, err := newPostgresEngine()

	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFind(engine, t)
}

func BenchmarkPostgresNoCacheFindPtr(t *testing.B) {
	engine, err := newPostgresEngine()

	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchFindPtr(engine, t)
}

func BenchmarkPostgresCacheInsert(t *testing.B) {
	engine, err := newPostgresEngine()

	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))

	doBenchInsert(engine, t)
}

func BenchmarkPostgresCacheFind(t *testing.B) {
	engine, err := newPostgresEngine()

	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))

	doBenchFind(engine, t)
}

func BenchmarkPostgresCacheFindPtr(t *testing.B) {
	engine, err := newPostgresEngine()

	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000))

	doBenchFindPtr(engine, t)
}
