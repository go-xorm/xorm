package xorm

import (
	//"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

func newSqlite3Engine() (*Engine, error) {
	os.Remove("./test.db")
	return NewEngine("sqlite3", "./test.db")
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
	testAll2(engine, t)
}

func TestSqlite3WithCache(t *testing.T) {
	engine, err := newSqlite3Engine()
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

/*func BenchmarkSqlite3DriverInsert(t *testing.B) {
	t.StopTimer()
	engine, err := newSqlite3Engine()
	if err != nil {
		t.Error(err)
		return
	}

	err = engine.CreateTables(&BigStruct{})
	if err != nil {
		t.Error(err)
		return
	}
	engine.Close()

	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		t.Error(err)
		return
	}

	doBenchDriverInsertS(db, t)

	db.Close()

	engine, err = newSqlite3Engine()
	if err != nil {
		t.Error(err)
		return
	}

	err = engine.DropTables(&BigStruct{})
	if err != nil {
		t.Error(err)
		return
	}
	defer engine.Close()
}

func BenchmarkSqlite3DriverFind(t *testing.B) {
	t.StopTimer()
	engine, err := newSqlite3Engine()
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}

	err = engine.CreateTables(&BigStruct{})
	if err != nil {
		t.Error(err)
		return
	}
	engine.Close()

	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	doBenchDriverFindS(db, t)

	db.Close()

	engine, err = newSqlite3Engine()
	if err != nil {
		t.Error(err)
		return
	}

	err = engine.DropTables(&BigStruct{})
	if err != nil {
		t.Error(err)
		return
	}
	defer engine.Close()
}*/

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
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
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
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
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
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	doBenchFindPtr(engine, t)
}
