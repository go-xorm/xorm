package xorm

import (
	"flag"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	testEngine *Engine
	connString string

	db         = flag.String("db", "sqlite3", "the tested database")
	showSQL    = flag.Bool("show_sql", true, "show generated SQLs")
	ptrConnStr = flag.String("conn_str", "", "test database connection string")
	mapType    = flag.String("map_type", "snake", "indicate the name mapping")
	cache      = flag.Bool("cache", false, "if enable cache")
)

func createEngine(dbType, connStr string) error {
	if testEngine == nil {
		var err error
		testEngine, err = NewEngine(dbType, connStr)
		if err != nil {
			return err
		}

		testEngine.ShowSQL(*showSQL)
	}

	tables, err := testEngine.DBMetas()
	if err != nil {
		return err
	}
	var tableNames = make([]interface{}, 0, len(tables))
	for _, table := range tables {
		tableNames = append(tableNames, table.Name)
	}
	return testEngine.DropTables(tableNames...)
}

func prepareEngine() error {
	return createEngine(*db, connString)
}

func TestMain(m *testing.M) {
	flag.Parse()

	if *db == "sqlite3" {
		if ptrConnStr == nil {
			connString = "./test.db"
		} else {
			connString = *ptrConnStr
		}
	} else {
		if ptrConnStr == nil {
			fmt.Println("you should indicate conn string")
			return
		}
		connString = *ptrConnStr
	}

	if err := prepareEngine(); err != nil {
		fmt.Println(err)
		return
	}
	os.Exit(m.Run())
}

func TestPing(t *testing.T) {
	if err := testEngine.Ping(); err != nil {
		t.Fatal(err)
	}
}
