package xorm

import (
    "database/sql"
    "testing"

    _ "github.com/lib/pq"
)

func newPostgresEngine() (*Engine, error) {
    return NewEngine("postgres", "dbname=xorm_test sslmode=disable")
}

func newPostgresDriverDB() (*sql.DB, error) {
    return sql.Open("postgres", "dbname=xorm_test sslmode=disable")
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
    testAll2(engine, t)
    testAll3(engine, t)
}

func TestPostgresWithCache(t *testing.T) {
    engine, err := newPostgresEngine()
    if err != nil {
        t.Error(err)
        return
    }
    engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
    defer engine.Close()
    engine.ShowSQL = showTestSql
    engine.ShowErr = showTestSql
    engine.ShowWarn = showTestSql
    engine.ShowDebug = showTestSql

    testAll(engine, t)
    testAll2(engine, t)
}

/*
func TestPostgres2(t *testing.T) {
    engine, err := NewEngine("postgres", "dbname=xorm_test sslmode=disable")
    if err != nil {
        t.Error(err)
        return
    }
    defer engine.Close()
    engine.ShowSQL = showTestSql
    engine.Mapper = SameMapper{}

    fmt.Println("-------------- directCreateTable --------------")
    directCreateTable(engine, t)
    fmt.Println("-------------- mapper --------------")
    mapper(engine, t)
    fmt.Println("-------------- insert --------------")
    insert(engine, t)
    fmt.Println("-------------- querySameMapper --------------")
    querySameMapper(engine, t)
    fmt.Println("-------------- execSameMapper --------------")
    execSameMapper(engine, t)
    fmt.Println("-------------- insertAutoIncr --------------")
    insertAutoIncr(engine, t)
    fmt.Println("-------------- insertMulti --------------")
    insertMulti(engine, t)
    fmt.Println("-------------- insertTwoTable --------------")
    insertTwoTable(engine, t)
    fmt.Println("-------------- updateSameMapper --------------")
    updateSameMapper(engine, t)
    fmt.Println("-------------- testdelete --------------")
    testdelete(engine, t)
    fmt.Println("-------------- get --------------")
    get(engine, t)
    fmt.Println("-------------- cascadeGet --------------")
    cascadeGet(engine, t)
    fmt.Println("-------------- find --------------")
    find(engine, t)
    fmt.Println("-------------- find2 --------------")
    find2(engine, t)
    fmt.Println("-------------- findMap --------------")
    findMap(engine, t)
    fmt.Println("-------------- findMap2 --------------")
    findMap2(engine, t)
    fmt.Println("-------------- count --------------")
    count(engine, t)
    fmt.Println("-------------- where --------------")
    where(engine, t)
    fmt.Println("-------------- in --------------")
    in(engine, t)
    fmt.Println("-------------- limit --------------")
    limit(engine, t)
    fmt.Println("-------------- orderSameMapper --------------")
    orderSameMapper(engine, t)
    fmt.Println("-------------- joinSameMapper --------------")
    joinSameMapper(engine, t)
    fmt.Println("-------------- havingSameMapper --------------")
    havingSameMapper(engine, t)
    fmt.Println("-------------- combineTransactionSameMapper --------------")
    combineTransactionSameMapper(engine, t)
    fmt.Println("-------------- table --------------")
    table(engine, t)
    fmt.Println("-------------- createMultiTables --------------")
    createMultiTables(engine, t)
    fmt.Println("-------------- tableOp --------------")
    tableOp(engine, t)
    fmt.Println("-------------- testColsSameMapper --------------")
    testColsSameMapper(engine, t)
    fmt.Println("-------------- testCharst --------------")
    testCharst(engine, t)
    fmt.Println("-------------- testStoreEngine --------------")
    testStoreEngine(engine, t)
    fmt.Println("-------------- testExtends --------------")
    testExtends(engine, t)
    fmt.Println("-------------- testColTypes --------------")
    testColTypes(engine, t)
    fmt.Println("-------------- testCustomType --------------")
    testCustomType(engine, t)
    fmt.Println("-------------- testCreatedAndUpdated --------------")
    testCreatedAndUpdated(engine, t)
    fmt.Println("-------------- testIndexAndUnique --------------")
    testIndexAndUnique(engine, t)
    fmt.Println("-------------- testMetaInfo --------------")
    testMetaInfo(engine, t)
    fmt.Println("-------------- testIterate --------------")
    testIterate(engine, t)
    fmt.Println("-------------- testStrangeName --------------")
    testStrangeName(engine, t)
    fmt.Println("-------------- testVersion --------------")
    testVersion(engine, t)
    fmt.Println("-------------- testDistinct --------------")
    testDistinct(engine, t)
    fmt.Println("-------------- testUseBool --------------")
    testUseBool(engine, t)
    fmt.Println("-------------- transaction --------------")
    transaction(engine, t)
}*/

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
    engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

    doBenchInsert(engine, t)
}

func BenchmarkPostgresCacheFind(t *testing.B) {
    engine, err := newPostgresEngine()

    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

    doBenchFind(engine, t)
}

func BenchmarkPostgresCacheFindPtr(t *testing.B) {
    engine, err := newPostgresEngine()

    defer engine.Close()
    if err != nil {
        t.Error(err)
        return
    }
    engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))

    doBenchFindPtr(engine, t)
}
