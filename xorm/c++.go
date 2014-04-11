package main

import (
	//"fmt"
	"strings"
	"text/template"

	"github.com/go-xorm/core"
)

var (
	CPlusTmpl LangTmpl = LangTmpl{
		template.FuncMap{"Mapper": mapper.Table2Obj,
			"Type":    cPlusTypeStr,
			"UnTitle": unTitle,
		},
		nil,
		genCPlusImports,
	}
)

func cPlusTypeStr(col *core.Column) string {
	tp := col.SQLType
	name := strings.ToUpper(tp.Name)
	switch name {
	case core.Bit, core.TinyInt, core.SmallInt, core.MediumInt, core.Int, core.Integer, core.Serial:
		return "int"
	case core.BigInt, core.BigSerial:
		return "__int64"
	case core.Char, core.Varchar, core.TinyText, core.Text, core.MediumText, core.LongText:
		return "tstring"
	case core.Date, core.DateTime, core.Time, core.TimeStamp:
		return "time_t"
	case core.Decimal, core.Numeric:
		return "tstring"
	case core.Real, core.Float:
		return "float"
	case core.Double:
		return "double"
	case core.TinyBlob, core.Blob, core.MediumBlob, core.LongBlob, core.Bytea:
		return "tstring"
	case core.Bool:
		return "bool"
	default:
		return "tstring"
	}
	return ""
}

func genCPlusImports(tables []*core.Table) map[string]string {
	imports := make(map[string]string)

	for _, table := range tables {
		for _, col := range table.Columns() {
			switch cPlusTypeStr(col) {
			case "time_t":
				imports[`<time.h>`] = `<time.h>`
			case "tstring":
				imports["<string>"] = "<string>"
				//case "__int64":
				//    imports[""] = ""
			}
		}
	}
	return imports
}
