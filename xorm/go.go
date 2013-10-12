package main

import (
	//"github.com/lunny/xorm"
	"strings"
	"xorm"
)

func typestring(st xorm.SQLType) string {
	t := xorm.SQLType2Type(st)
	s := t.String()
	if s == "[]uint8" {
		return "[]byte"
	}
	return s
}

func tag(col *xorm.Column) string {
	res := make([]string, 0)
	if !col.Nullable {
		res = append(res, "not null")
	}
	if col.IsPrimaryKey {
		res = append(res, "pk")
	}
	if col.Default != "" {
		res = append(res, "default "+col.Default)
	}
	if col.IsAutoIncrement {
		res = append(res, "autoincr")
	}
	if col.IsCreated {
		res = append(res, "created")
	}
	if col.IsUpdated {
		res = append(res, "updated")
	}

	if len(res) > 0 {
		return "`xorm:\"" + strings.Join(res, " ") + "\"`"
	}
	return ""
}
