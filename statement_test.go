package xorm

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"strings"

	"github.com/go-xorm/core"
)

var colStrTests = []struct {
	omitColumn        string
	onlyToDBColumnNdx int
	expected          string
}{
	{"", -1, "`ID`, `IsDeleted`, `Caption`, `Code1`, `Code2`, `Code3`, `ParentID`, `Latitude`, `Longitude`"},
	{"Code2", -1, "`ID`, `IsDeleted`, `Caption`, `Code1`, `Code3`, `ParentID`, `Latitude`, `Longitude`"},
	{"", 1, "`ID`, `Caption`, `Code1`, `Code2`, `Code3`, `ParentID`, `Latitude`, `Longitude`"},
	{"Code3", 1, "`ID`, `Caption`, `Code1`, `Code2`, `ParentID`, `Latitude`, `Longitude`"},
	{"Longitude", 1, "`ID`, `Caption`, `Code1`, `Code2`, `Code3`, `ParentID`, `Latitude`"},
	{"", 8, "`ID`, `IsDeleted`, `Caption`, `Code1`, `Code2`, `Code3`, `ParentID`, `Latitude`"},
}

// !nemec784! Only for Statement object creation
const driverName = "mysql"
const dataSourceName = "Server=TestServer;Database=TestDB;Uid=testUser;Pwd=testPassword;"

func init() {
	core.RegisterDriver(driverName, &mysqlDriver{})
}

func TestColumnsStringGeneration(t *testing.T) {

	var statement *Statement

	for ndx, testCase := range colStrTests {

		statement = createTestStatement()

		if testCase.omitColumn != "" {
			statement.Omit(testCase.omitColumn) // !nemec784! Column must be skipped
		}

		if testCase.onlyToDBColumnNdx >= 0 {
			columns := statement.RefTable.Columns()
			columns[testCase.onlyToDBColumnNdx].MapType = core.ONLYTODB // !nemec784! Column must be skipped
		}

		actual := statement.genColumnStr()

		if actual != testCase.expected {
			t.Errorf("[test #%d] Unexpected columns string:\nwant:\t%s\nhave:\t%s", ndx, testCase.expected, actual)
		}
	}
}

func BenchmarkColumnsStringGeneration(b *testing.B) {

	b.StopTimer()

	statement := createTestStatement()

	testCase := colStrTests[0]

	if testCase.omitColumn != "" {
		statement.Omit(testCase.omitColumn) // !nemec784! Column must be skipped
	}

	if testCase.onlyToDBColumnNdx >= 0 {
		columns := statement.RefTable.Columns()
		columns[testCase.onlyToDBColumnNdx].MapType = core.ONLYTODB // !nemec784! Column must be skipped
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		actual := statement.genColumnStr()

		if actual != testCase.expected {
			b.Errorf("Unexpected columns string:\nwant:\t%s\nhave:\t%s", testCase.expected, actual)
		}
	}
}

func BenchmarkGetFlagForColumnWithICKey_ContainsKey(b *testing.B) {

	b.StopTimer()

	mapCols := make(map[string]bool)
	cols := []*core.Column{
		&core.Column{Name: `ID`},
		&core.Column{Name: `IsDeleted`},
		&core.Column{Name: `Caption`},
		&core.Column{Name: `Code1`},
		&core.Column{Name: `Code2`},
		&core.Column{Name: `Code3`},
		&core.Column{Name: `ParentID`},
		&core.Column{Name: `Latitude`},
		&core.Column{Name: `Longitude`},
	}

	for _, col := range cols {
		mapCols[strings.ToLower(col.Name)] = true
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {

		for _, col := range cols {

			if _, ok := getFlagForColumn(mapCols, col); !ok {
				b.Fatal("Unexpected result")
			}
		}
	}
}

func BenchmarkGetFlagForColumnWithICKey_EmptyMap(b *testing.B) {

	b.StopTimer()

	mapCols := make(map[string]bool)
	cols := []*core.Column{
		&core.Column{Name: `ID`},
		&core.Column{Name: `IsDeleted`},
		&core.Column{Name: `Caption`},
		&core.Column{Name: `Code1`},
		&core.Column{Name: `Code2`},
		&core.Column{Name: `Code3`},
		&core.Column{Name: `ParentID`},
		&core.Column{Name: `Latitude`},
		&core.Column{Name: `Longitude`},
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {

		for _, col := range cols {

			if _, ok := getFlagForColumn(mapCols, col); ok {
				b.Fatal("Unexpected result")
			}
		}
	}
}

type TestType struct {
	ID        int64   `xorm:"ID PK"`
	IsDeleted bool    `xorm:"IsDeleted"`
	Caption   string  `xorm:"Caption"`
	Code1     string  `xorm:"Code1"`
	Code2     string  `xorm:"Code2"`
	Code3     string  `xorm:"Code3"`
	ParentID  int64   `xorm:"ParentID"`
	Latitude  float64 `xorm:"Latitude"`
	Longitude float64 `xorm:"Longitude"`
}

func (TestType) TableName() string {
	return "TestTable"
}

func createTestStatement() *Statement {

	engine := createTestEngine()

	statement := &Statement{}
	statement.Init()
	statement.Engine = engine
	statement.setRefValue(reflect.ValueOf(TestType{}))

	return statement
}

func createTestEngine() *Engine {
	driver := core.QueryDriver(driverName)
	uri, err := driver.Parse(driverName, dataSourceName)

	if err != nil {
		panic(err)
	}

	dialect := &mysql{}
	err = dialect.Init(nil, uri, driverName, dataSourceName)

	if err != nil {
		panic(err)
	}

	engine := &Engine{
		dialect:       dialect,
		Tables:        make(map[reflect.Type]*core.Table),
		mutex:         &sync.RWMutex{},
		TagIdentifier: "xorm",
		TZLocation:    time.Local,
	}
	engine.SetMapper(core.NewCacheMapper(new(core.SnakeMapper)))

	return engine
}
