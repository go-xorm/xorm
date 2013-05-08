package xorm

import (
	"reflect"
	"strings"
)

// 'sqlite:///foo.db'
// 'sqlite:////Uses/lunny/foo.db'
// 'sqlite:///:memory:'
// '<protocol>://<username>:<passwd>@<host>/<dbname>?charset=<encoding>'
func Create(schema string) Engine {
	engine := Engine{}
	engine.Mapper = SnakeMapper{}
	engine.Tables = make(map[reflect.Type]Table)
	engine.Statement.Engine = &engine
	l := strings.Split(schema, "://")
	if len(l) == 2 {
		engine.Protocol = l[0]
		if l[0] == "sqlite" {
			engine.Charset = "utf8"
			engine.AutoIncrement = "AUTOINCREMENT"
			if l[1] == "/:memory:" {
				engine.Others = l[1]
			} else if strings.Index(l[1], "//") == 0 {
				engine.Others = l[1][1:]
			} else if strings.Index(l[1], "/") == 0 {
				engine.Others = "." + l[1]
			}
		} else {
			engine.AutoIncrement = "AUTO_INCREMENT"
			x := strings.Split(l[1], ":")
			engine.UserName = x[0]
			y := strings.Split(x[1], "@")
			engine.Password = y[0]
			z := strings.Split(y[1], "/")
			engine.Host = z[0]
			a := strings.Split(z[1], "?")
			engine.DBName = a[0]
			if len(a) == 2 {
				engine.Charset = strings.Split(a[1], "=")[1]
			} else {
				engine.Charset = "utf8"
			}
		}
	}

	return engine
}
