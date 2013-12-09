package xorm

import (
    _ "github.com/go-sql-driver/mysql"
    "testing"
)

/*
CREATE DATABASE IF NOT EXISTS xorm_test CHARACTER SET
utf8 COLLATE utf8_general_ci;
*/

var mysqlShowTestSql bool = true

func TestMysql(t *testing.T) {
    err := mysqlDdlImport()
    if err != nil {
        t.Error(err)
        return
    }

    engine, err := NewEngine("mysql", "root:@/xorm_test?charset=utf8")
    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    engine.ShowSQL = mysqlShowTestSql
    engine.ShowErr = mysqlShowTestSql
    engine.ShowWarn = mysqlShowTestSql
    engine.ShowDebug = mysqlShowTestSql

    testAll(engine, t)
    testAll2(engine, t)
    testAll3(engine, t)
}

func TestMysqlWithCache(t *testing.T) {
    err := mysqlDdlImport()
    if err != nil {
        t.Error(err)
        return
    }

    engine, err := NewEngine("mysql", "root:@/xorm_test?charset=utf8")
    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
    engine.ShowSQL = mysqlShowTestSql
    engine.ShowErr = mysqlShowTestSql
    engine.ShowWarn = mysqlShowTestSql
    engine.ShowDebug = mysqlShowTestSql

    testAll(engine, t)
    testAll2(engine, t)
}

func newMysqlEngine() (*Engine, error) {
    return NewEngine("mysql", "root:@/xorm_test?charset=utf8")
}

func mysqlDdlImport() error {
    engine, err := NewEngine("mysql", "root:@/?charset=utf8")
    if err != nil {
        return err
    }
    engine.ShowSQL = mysqlShowTestSql
    engine.ShowErr = mysqlShowTestSql
    engine.ShowWarn = mysqlShowTestSql
    engine.ShowDebug = mysqlShowTestSql

    sqlResults, _ := engine.Import("tests/mysql_ddl.sql")
    engine.LogDebug("sql results: %v", sqlResults)
    engine.Close()
    return nil
}

func BenchmarkMysqlNoCacheInsert(t *testing.B) {
    engine, err := newMysqlEngine()
    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    //engine.ShowSQL = true
    doBenchInsert(engine, t)
}

func BenchmarkMysqlNoCacheFind(t *testing.B) {
    engine, err := newMysqlEngine()
    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    //engine.ShowSQL = true
    doBenchFind(engine, t)
}

func BenchmarkMysqlNoCacheFindPtr(t *testing.B) {
    engine, err := newMysqlEngine()
    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    //engine.ShowSQL = true
    doBenchFindPtr(engine, t)
}

func BenchmarkMysqlCacheInsert(t *testing.B) {
    engine, err := newMysqlEngine()
    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

    doBenchInsert(engine, t)
}

func BenchmarkMysqlCacheFind(t *testing.B) {
    engine, err := newMysqlEngine()
    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

    doBenchFind(engine, t)
}

func BenchmarkMysqlCacheFindPtr(t *testing.B) {
    engine, err := newMysqlEngine()
    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

    doBenchFindPtr(engine, t)
}
