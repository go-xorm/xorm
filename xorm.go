package xorm

import (
	//"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"
	//"time"
)

func NewEngine(driverName string, dataSourceName string) (*Engine, error) {
	engine := &Engine{ShowSQL: false, DriverName: driverName, Mapper: SnakeMapper{},
		DataSourceName: dataSourceName}

	engine.Tables = make(map[reflect.Type]*Table)
	engine.mutex = &sync.Mutex{}
	engine.InsertMany = true
	engine.TagIdentifier = "xorm"
	engine.QuoteIdentifier = "`"
	if driverName == SQLITE {
		engine.Dialect = sqlite3{}
		engine.AutoIncrement = "AUTOINCREMENT"
		//engine.Pool = NoneConnectPool{}
	} else if driverName == MYSQL {
		engine.Dialect = mysql{}
		engine.AutoIncrement = "AUTO_INCREMENT"
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupported driver name: %v", driverName))
	}

	/*engine.Pool = &SimpleConnectPool{
		releasedSessions: make([]*sql.DB, 30),
		usingSessions:    map[*sql.DB]time.Time{},
		cur:              -1,
		maxWaitTimeOut:   14400,
		mutex:            &sync.Mutex{},
	}*/
	engine.Pool = &NoneConnectPool{}

	return engine, nil
}
