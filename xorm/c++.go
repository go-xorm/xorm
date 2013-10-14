package main

import (
	//"fmt"
	"github.com/lunny/xorm"
	"strings"
	"text/template"
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

func cPlusTypeStr(col *xorm.Column) string {
	tp := col.SQLType
	name := strings.ToUpper(tp.Name)
	switch name {
	case xorm.Bit, xorm.TinyInt, xorm.SmallInt, xorm.MediumInt, xorm.Int, xorm.Integer, xorm.Serial:
		return "int"
	case xorm.BigInt, xorm.BigSerial:
		return "__int64"
	case xorm.Char, xorm.Varchar, xorm.TinyText, xorm.Text, xorm.MediumText, xorm.LongText:
		return "tstring"
	case xorm.Date, xorm.DateTime, xorm.Time, xorm.TimeStamp:
		return "time_t"
	case xorm.Decimal, xorm.Numeric:
		return "tstring"
	case xorm.Real, xorm.Float:
		return "float"
	case xorm.Double:
		return "double"
	case xorm.TinyBlob, xorm.Blob, xorm.MediumBlob, xorm.LongBlob, xorm.Bytea:
		return "tstring"
	case xorm.Bool:
		return "bool"
	default:
		return "tstring"
	}
	return ""
}

func genCPlusImports(tables []*xorm.Table) map[string]string {
	imports := make(map[string]string)

	for _, table := range tables {
		for _, col := range table.Columns {
			switch cPlusTypeStr(col) {
			case "time_t":
				imports[`<time.h>`] = `<time.h>`
			case "tstring":
				imports["<string>"] = "<string>"
				//case "__int64":
				//	imports[""] = ""
			}
		}
	}
	return imports
}
