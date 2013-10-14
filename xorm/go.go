package main

import (
	"github.com/lunny/xorm"
	"go/format"
	"strings"
	"text/template"
)

var (
	GoLangTmpl LangTmpl = LangTmpl{
		template.FuncMap{"Mapper": mapper.Table2Obj,
			"Type":    typestring,
			"Tag":     tag,
			"UnTitle": unTitle,
		},
		formatGo,
		genGoImports,
	}
)

func formatGo(src string) (string, error) {
	source, err := format.Source([]byte(src))
	if err != nil {
		return "", err
	}
	return string(source), nil
}

func genGoImports(tables []*xorm.Table) map[string]string {
	imports := make(map[string]string)

	for _, table := range tables {
		for _, col := range table.Columns {
			if typestring(col) == "time.Time" {
				imports["time"] = "time"
			}
		}
	}
	return imports
}

func typestring(col *xorm.Column) string {
	st := col.SQLType
	if col.IsPrimaryKey {
		return "int64"
	}
	t := xorm.SQLType2Type(st)
	s := t.String()
	if s == "[]uint8" {
		return "[]byte"
	}
	return s
}

func tag(table *xorm.Table, col *xorm.Column) string {
	isNameId := (mapper.Table2Obj(col.Name) == "Id")
	res := make([]string, 0)
	if !col.Nullable {
		if !isNameId {
			res = append(res, "not null")
		}
	}
	if col.IsPrimaryKey {
		if !isNameId {
			res = append(res, "pk")
		}
	}
	if col.Default != "" {
		res = append(res, "default "+col.Default)
	}
	if col.IsAutoIncrement {
		if !isNameId {
			res = append(res, "autoincr")
		}
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

	var tags []string
	if genJson {
		tags = append(tags, "json:\""+col.Name+"\"")
	}
	if len(res) > 0 {
		tags = append(tags, "xorm:\""+strings.Join(res, " ")+"\"")
	}
	if len(tags) > 0 {
		return "`" + strings.Join(tags, " ") + "`"
	} else {
		return ""
	}
}
