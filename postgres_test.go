package xorm

import (
	"fmt"
	_ "github.com/bylevel/pq"
	"testing"
)

func TestPostgres(t *testing.T) {
	engine, err := NewEngine("postgres", "dbname=xorm_test sslmode=disable")
	if err != nil {
		t.Error(err)
		return
	}
	defer engine.Close()
	engine.ShowSQL = showTestSql

	testAll(engine, t)
	testAll2(engine, t)
}

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
	fmt.Println("-------------- transaction --------------")
	transaction(engine, t)
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
}

func BenchmarkPostgresNoCache(t *testing.B) {
	engine, err := NewEngine("postgres", "dbname=xorm_test sslmode=disable")

	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	//engine.ShowSQL = true
	doBenchCacheFind(engine, t)
}

func BenchmarkPostgresCache(t *testing.B) {
	engine, err := NewEngine("postgres", "dbname=xorm_test sslmode=disable")

	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.SetDefaultCacher(NewLRUCacher(NewMemoryStore(), 1000))
	doBenchCacheFind(engine, t)
}
