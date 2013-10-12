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
	version string = "0.2.0"
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
		engine.dialect = &sqlite3{}
	} else if driverName == MYSQL {
		engine.dialect = &mysql{}
	} else if driverName == POSTGRES {
		engine.dialect = &postgres{}
		engine.Filters = append(engine.Filters, &PgSeqFilter{})
		engine.Filters = append(engine.Filters, &QuoteFilter{})
	} else if driverName == MYMYSQL {
		engine.dialect = &mymysql{}
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupported driver name: %v", driverName))
	}
	err := engine.dialect.Init(driverName, dataSourceName)
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
