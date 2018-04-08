package xorm

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/ziutek/mymysql/godrv"
)

var (
	testEngine EngineInterface
	dbType     string
	connString string

	db         = flag.String("db", "sqlite3", "the tested database")
	showSQL    = flag.Bool("show_sql", true, "show generated SQLs")
	ptrConnStr = flag.String("conn_str", "./test.db?cache=shared&mode=rwc", "test database connection string")
	mapType    = flag.String("map_type", "snake", "indicate the name mapping")
	cache      = flag.Bool("cache", false, "if enable cache")
	cluster    = flag.Bool("cluster", false, "if this is a cluster")
	splitter   = flag.String("splitter", ";", "the splitter on connstr for cluster")
	schema     = flag.String("schema", "", "specify the schema")
)

func createEngine(dbType, connStr string) error {
	if testEngine == nil {
		var err error

		if !*cluster {
			testEngine, err = NewEngine(dbType, connStr)
		} else {
			testEngine, err = NewEngineGroup(dbType, strings.Split(connStr, *splitter))
		}
		if err != nil {
			return err
		}

		if *schema != "" {
			testEngine.SetSchema(*schema)
		}
		testEngine.ShowSQL(*showSQL)
		testEngine.SetLogLevel(core.LOG_DEBUG)
		if *cache {
			cacher := NewLRUCacher(NewMemoryStore(), 100000)
			testEngine.SetDefaultCacher(cacher)
		}

		if len(*mapType) > 0 {
			switch *mapType {
			case "snake":
				testEngine.SetMapper(core.SnakeMapper{})
			case "same":
				testEngine.SetMapper(core.SameMapper{})
			case "gonic":
				testEngine.SetMapper(core.LintGonicMapper)
			}
		}
	}

	tables, err := testEngine.DBMetas()
	if err != nil {
		return err
	}
	var tableNames = make([]interface{}, 0, len(tables))
	for _, table := range tables {
		tableNames = append(tableNames, table.Name)
	}
	if err = testEngine.DropTables(tableNames...); err != nil {
		return err
	}
	return nil
}

func prepareEngine() error {
	return createEngine(dbType, connString)
}

func TestMain(m *testing.M) {
	flag.Parse()

	dbType = *db
	if *db == "sqlite3" {
		if ptrConnStr == nil {
			connString = "./test.db?cache=shared&mode=rwc"
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

	dbs := strings.Split(*db, "::")
	conns := strings.Split(connString, "::")

	var res int
	for i := 0; i < len(dbs); i++ {
		dbType = dbs[i]
		connString = conns[i]
		testEngine = nil
		fmt.Println("testing", dbType, connString)

		if err := prepareEngine(); err != nil {
			fmt.Println(err)
			return
		}

		code := m.Run()
		if code > 0 {
			res = code
		}
	}

	os.Exit(res)
}

func TestPing(t *testing.T) {
	if err := testEngine.Ping(); err != nil {
		t.Fatal(err)
	}
}
