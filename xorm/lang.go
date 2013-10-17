package main

import (
	"github.com/lunny/xorm"
	"io/ioutil"
	"strings"
	"text/template"
)

type LangTmpl struct {
	Funcs      template.FuncMap
	Formater   func(string) (string, error)
	GenImports func([]*xorm.Table) map[string]string
}

var (
	mapper    = &xorm.SnakeMapper{}
	langTmpls = map[string]LangTmpl{
		"go":  GoLangTmpl,
		"c++": CPlusTmpl,
	}
)

func loadConfig(f string) map[string]string {
	bts, err := ioutil.ReadFile(f)
	if err != nil {
		return nil
	}
	configs := make(map[string]string)
	lines := strings.Split(string(bts), "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		vs := strings.Split(line, "=")
		if len(vs) == 2 {
			configs[strings.TrimSpace(vs[0])] = strings.TrimSpace(vs[1])
		}
	}
	return configs
}

func unTitle(src string) string {
	if src == "" {
		return ""
	}

	if len(src) == 1 {
		return strings.ToLower(string(src[0]))
	} else {
		return strings.ToLower(string(src[0])) + src[1:]
	}
}
