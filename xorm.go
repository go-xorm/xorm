package xorm

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync"
)

const (
	version string = "0.1.9"
)

func close(engine *Engine) {
	engine.Close()
}

// new a db manager according to the parameter. Currently support four
// drivers
func NewEngine(driverName string, dataSourceName string) (*Engine, error) {
	engine := &Engine{DriverName: driverName, Mapper: SnakeMapper{},
		DataSourceName: dataSourceName, Filters: make([]Filter, 0)}

	if driverName == SQLITE {
		engine.Dialect = &sqlite3{}
	} else if driverName == MYSQL {
		engine.Dialect = &mysql{}
	} else if driverName == POSTGRES {
		engine.Dialect = &postgres{}
		engine.Filters = append(engine.Filters, &PgSeqFilter{})
		engine.Filters = append(engine.Filters, &QuoteFilter{})
	} else if driverName == MYMYSQL {
		engine.Dialect = &mymysql{}
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupported driver name: %v", driverName))
	}
	err := engine.Dialect.Init(dataSourceName)
	if err != nil {
		return nil, err
	}

	engine.Tables = make(map[reflect.Type]*Table)
	engine.mutex = &sync.Mutex{}
	engine.TagIdentifier = "xorm"

	engine.Filters = append(engine.Filters, &IdFilter{})
	engine.Logger = os.Stdout

	//engine.Pool = NewSimpleConnectPool()
	//engine.Pool = NewNoneConnectPool()
	//engine.Cacher = NewLRUCacher()
	err = engine.SetPool(NewSysConnectPool())
	runtime.SetFinalizer(engine, close)
	return engine, err
}
