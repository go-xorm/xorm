package main

import (
	"github.com/lunny/xorm"
	"strings"
)

func unTitle(src string) string {
	if src == "" {
		return ""
	}

	return strings.ToLower(string(src[0])) + src[1:]
}

func typestring(st xorm.SQLType) string {
	t := xorm.SQLType2Type(st)
	s := t.String()
	if s == "[]uint8" {
		return "[]byte"
	}
	return s
}

func tag(table *xorm.Table, col *xorm.Column) string {
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
	for name, _ := range col.Indexes {
		index := table.Indexes[name]
		var uistr string
		if index.Type == xorm.UniqueType {
			uistr = "unique"
		} else if index.Type == xorm.IndexType {
			uistr = "index"
		}
		if index.Name != col.Name {
			uistr += "(" + index.Name + ")"
		}
		res = append(res, uistr)
	}

	if len(res) > 0 {
		return "`xorm:\"" + strings.Join(res, " ") + "\"`"
	}
	return ""
}
