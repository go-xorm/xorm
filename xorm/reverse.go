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
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
		"-m": false,
	}
}

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
		fmt.Println("no")
		return
	}

	var isMultiFile bool
	if _, ok := cmd.Flags["-m"]; ok {
		isMultiFile = true
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

	dir, err := filepath.Abs(args[2])
	if err != nil {
		logging.Error("%v", err)
		return
	}

	m := &xorm.SnakeMapper{}

	filepath.Walk(dir, func(f string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		bs, err := ioutil.ReadFile(f)
		if err != nil {
			logging.Error("%v", err)
			return err
		}

		t := template.New(f)
		t.Funcs(template.FuncMap{"Mapper": m.Table2Obj,
			"Type": typestring,
			"Tag":  tag,
		})

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

			imports := make(map[string]string)
			tbls := make([]*xorm.Table, 0)
			for _, table := range tables {
				for _, col := range table.Columns {
					if typestring(col.SQLType) == "time.Time" {
						imports["time"] = "time"
					}
				}
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
			source, err := format.Source(tplcontent)
			if err != nil {
				logging.Error("%v", err)
				return err
			}

			w.WriteString(string(source))
			w.Close()
		} else {
			for _, table := range tables {
				// imports
				imports := make(map[string]string)
				for _, col := range table.Columns {
					if typestring(col.SQLType) == "time.Time" {
						imports["time"] = "time"
					}
				}

				w, err := os.OpenFile(path.Join(genDir, unTitle(m.Table2Obj(table.Name))+ext), os.O_RDWR|os.O_CREATE, 0600)
				if err != nil {
					logging.Error("%v", err)
					return err
				}

				newbytes := bytes.NewBufferString("")

				t := &Tmpl{Tables: []*xorm.Table{table}, Imports: imports, Model: model}
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
				source, err := format.Source(tplcontent)
				if err != nil {
					logging.Error("%v", err)
					return err
				}

				w.WriteString(string(source))
				w.Close()
			}
		}

		return nil
	})

}
