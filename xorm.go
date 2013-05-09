package xorm

import (
	"reflect"
)

func Create(driverName string, dataSourceName string) Engine {
	engine := Engine{ShowSQL: false, DriverName: driverName, Mapper: SnakeMapper{},
		DataSourceName: dataSourceName}

	engine.Tables = make(map[reflect.Type]Table)
	engine.Statement.Engine = &engine
	if driverName == SQLITE {
		engine.AutoIncrement = "AUTOINCREMENT"
	} else {
		engine.AutoIncrement = "AUTO_INCREMENT"
	}

	if engine.DriverName == PQSQL {
		engine.QuoteIdentifier = "\""
	} else if engine.DriverName == MSSQL {
		engine.QuoteIdentifier = ""
	} else {
		engine.QuoteIdentifier = "`"
	}
	return engine
}
