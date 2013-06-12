package xorm

import (
	"reflect"
)

func NewEngine(driverName string, dataSourceName string) *Engine {
	engine := &Engine{ShowSQL: false, DriverName: driverName, Mapper: SnakeMapper{},
		DataSourceName: dataSourceName}

	engine.Tables = make(map[reflect.Type]Table)
	engine.Statement.Engine = engine
	engine.InsertMany = true
	engine.TagIdentifier = "xorm"
	if driverName == SQLITE {
		engine.Dialect = sqlite3{}
		engine.AutoIncrement = "AUTOINCREMENT"
	} else {
		engine.Dialect = mysql{}
		engine.AutoIncrement = "AUTO_INCREMENT"
	}

	engine.QuoteIdentifier = "`"

	return engine
}
