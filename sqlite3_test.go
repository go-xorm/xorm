package xorm

import (
	_ "github.com/mattn/go-sqlite3"
	//"os"
	"testing"
)

func TestSqlite3(t *testing.T) {
	//os.Remove("./test.db")
	engine, err := NewEngine("sqlite3", "./test.db")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.ShowSQL = true

	testAll(engine, t)
	testAll2(engine, t)
}
