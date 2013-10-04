package xorm

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

/*
CREATE DATABASE IF NOT EXISTS xorm_test CHARACTER SET
utf8 COLLATE utf8_general_ci;
*/

func TestMysql(t *testing.T) {
	engine, err := NewEngine("mysql", "root:@/xorm_test?charset=utf8")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.ShowSQL = showTestSql

	testAll(engine, t)
	testAll2(engine, t)
}

func TestMysqlWithCache(t *testing.T) {
	engine, err := NewEngine("mysql", "root:@/xorm_test?charset=utf8")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	engine.ShowSQL = showTestSql

	testAll(engine, t)
	testAll2(engine, t)
}

func BenchmarkMysqlNoCache(t *testing.B) {
	engine, err := NewEngine("mysql", "root:@/xorm_test?charset=utf8")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchCacheFind(engine, t)
}

func BenchmarkMysqlCache(t *testing.B) {
	engine, err := NewEngine("mysql", "root:@/xorm_test?charset=utf8")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	doBenchCacheFind(engine, t)
}
