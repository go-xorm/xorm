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
	engine, err := NewEngine("mymysql", "xorm_test2/root/")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.ShowSQL = showTestSql

	testAll(engine, t)
	testAll2(engine, t)
}

func BenchmarkMyMysqlNoCache(t *testing.B) {
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
	engine, err := NewEngine("mymysql", "xorm_test2/root/")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	doBenchCacheFind(engine, t)
}
