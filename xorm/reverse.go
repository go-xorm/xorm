package main

import (
	"bytes"
	"fmt"
	_ "github.com/bylevel/pq"
	"github.com/dvirsky/go-pylog/logging"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/ziutek/mymysql/godrv"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"text/template"
)

var CmdReverse = &Command{
	UsageLine: "reverse [-m] driverName datasourceName tmplPath [generatedPath]",
	Short:     "reverse a db to codes",
	Long: `
according database's tables and columns to generate codes for Go, C++ and etc.

	-m 				Generated one go file for every table
	driverName		Database driver name, now supported four: mysql mymysql sqlite3 postgres
	datasourceName	Database connection uri, for detail infomation please visit driver's project page
	tmplPath		Template dir for generated. the default templates dir has provide 1 template
	generatedPath	This parameter is optional, if blank, the default value is model, then will
					generated all codes in model dir
`,
}

func init() {
	CmdReverse.Run = runReverse
	CmdReverse.Flags = map[string]bool{
		"-s": false,
		"-l": false,
	}
}

var (
	genJson bool = false
)

func printReversePrompt(flag string) {
}

type Tmpl struct {
	Tables  []*xorm.Table
	Imports map[string]string
	Model   string
}

func runReverse(cmd *Command, args []string) {
	num := checkFlags(cmd.Flags, args, printReversePrompt)
	if num == -1 {
		return
	}
	args = args[num:]

	if len(args) < 3 {
		fmt.Println("params error, please see xorm help reverse")
		return
	}

	var isMultiFile bool = true
	if use, ok := cmd.Flags["-s"]; ok {
		isMultiFile = !use
	}

	curPath, err := os.Getwd()
	if err != nil {
		fmt.Println(curPath)
		return
	}

	var genDir string
	var model string
	if len(args) == 4 {
		genDir, err = filepath.Abs(args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		model = path.Base(genDir)
	} else {
		model = "model"
		genDir = path.Join(curPath, model)
	}

	dir, err := filepath.Abs(args[2])
	if err != nil {
		logging.Error("%v", err)
		return
	}

	var langTmpl LangTmpl
	var ok bool
	var lang string = "go"

	cfgPath := path.Join(dir, "config")
	info, err := os.Stat(cfgPath)
	var configs map[string]string
	if err == nil && !info.IsDir() {
		configs = loadConfig(cfgPath)
		if l, ok := configs["lang"]; ok {
			lang = l
		}
		if j, ok := configs["genJson"]; ok {
			genJson, err = strconv.ParseBool(j)
		}
	}

	if langTmpl, ok = langTmpls[lang]; !ok {
		fmt.Println("Unsupported lang", lang)
		return
	}

	os.MkdirAll(genDir, os.ModePerm)

	Orm, err := xorm.NewEngine(args[0], args[1])
	if err != nil {
		logging.Error("%v", err)
		return
	}

	tables, err := Orm.DBMetas()
	if err != nil {
		logging.Error("%v", err)
		return
	}

	filepath.Walk(dir, func(f string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if info.Name() == "config" {
			return nil
		}

		bs, err := ioutil.ReadFile(f)
		if err != nil {
			logging.Error("%v", err)
			return err
		}

		t := template.New(f)
		t.Funcs(langTmpl.Funcs)

		tmpl, err := t.Parse(string(bs))
		if err != nil {
			logging.Error("%v", err)
			return err
		}

		var w *os.File
		fileName := info.Name()
		newFileName := fileName[:len(fileName)-4]
		ext := path.Ext(newFileName)

		if !isMultiFile {
			w, err = os.OpenFile(path.Join(genDir, newFileName), os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				logging.Error("%v", err)
				return err
			}

			imports := langTmpl.GenImports(tables)

			tbls := make([]*xorm.Table, 0)
			for _, table := range tables {
				tbls = append(tbls, table)
			}

			newbytes := bytes.NewBufferString("")

			t := &Tmpl{Tables: tbls, Imports: imports, Model: model}
			err = tmpl.Execute(newbytes, t)
			if err != nil {
				logging.Error("%v", err)
				return err
			}

			tplcontent, err := ioutil.ReadAll(newbytes)
			if err != nil {
				logging.Error("%v", err)
				return err
			}
			var source string
			if langTmpl.Formater != nil {
				source, err = langTmpl.Formater(string(tplcontent))
				if err != nil {
					logging.Error("%v", err)
					return err
				}
			} else {
				source = string(tplcontent)
			}

			w.WriteString(source)
			w.Close()
		} else {
			for _, table := range tables {
				// imports
				tbs := []*xorm.Table{table}
				imports := langTmpl.GenImports(tbs)

				w, err := os.OpenFile(path.Join(genDir, unTitle(mapper.Table2Obj(table.Name))+ext), os.O_RDWR|os.O_CREATE, 0600)
				if err != nil {
					logging.Error("%v", err)
					return err
				}

				newbytes := bytes.NewBufferString("")

				t := &Tmpl{Tables: tbs, Imports: imports, Model: model}
				err = tmpl.Execute(newbytes, t)
				if err != nil {
					logging.Error("%v", err)
					return err
				}

				tplcontent, err := ioutil.ReadAll(newbytes)
				if err != nil {
					logging.Error("%v", err)
					return err
				}
				var source string
				if langTmpl.Formater != nil {
					source, err = langTmpl.Formater(string(tplcontent))
					if err != nil {
						logging.Error("%v-%v", err, string(tplcontent))
						return err
					}
				} else {
					source = string(tplcontent)
				}

				w.WriteString(source)
				w.Close()
			}
		}

		return nil
	})

}
