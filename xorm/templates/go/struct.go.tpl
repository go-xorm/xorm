package {{.Model}}

import (
	"github.com/lunny/xorm"
	{{range .Imports}}"{{.}}"{{end}}
)

type {{Mapper .Table.Name}} struct {
{{range .Table.Columns}}	{{Mapper .Name}}	{{Type .SQLType}} {{Tag .}}
{{end}}
}