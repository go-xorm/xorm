{{ range .Imports}}
#include {{.}}
{{ end }}

{{range .Tables}}class {{Mapper .Name}} {
{{$table := .}}
public:
{{range .Columns}}{{$name := Mapper .Name}}	{{Type .}} Get{{Mapper .Name}}() {
		return this->m_{{UnTitle $name}};
	}

	void Set{{$name}}({{Type .}} {{UnTitle $name}}) {
		this->m_{{UnTitle $name}} = {{UnTitle $name}};
	}

{{end}}private:
{{range .Columns}}{{$name := Mapper .Name}}	{{Type .}} m_{{UnTitle $name}};	
{{end}}
}

{{end}}