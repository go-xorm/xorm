package main

import (
	"fmt"
	//"github.com/lunny/xorm"
	"bytes"
	_ "github.com/bylevel/pq"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/ziutek/mymysql/godrv"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"xorm"
)

var CmdReverse = &Command{
	UsageLine: "reverse -m driverName datasourceName tmplpath",
	Short:     "reverse a db to codes",
	Long: `
according database's tables and columns to generate codes for Go, C++ and etc.
`,
}

func init() {
	CmdReverse.Run = runReverse
	CmdReverse.Flags = map[string]bool{}
}

func printReversePrompt(flag string) {
}

type Tmpl struct {
	Table   *xorm.Table
	Imports map[string]string
	Model   string
}

func runReverse(cmd *Command, args []string) {
	if len(args) < 3 {
		fmt.Println("no")
		return
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
		fmt.Println(err)
		return
	}

	tables, err := Orm.DBMetas()
	if err != nil {
		fmt.Println(err)
		return
	}

	dir, err := filepath.Abs(args[2])
	if err != nil {
		fmt.Println(curPath)
		return
	}

	var isMultiFile bool = true
	m := &xorm.SnakeMapper{}

	filepath.Walk(dir, func(f string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		bs, err := ioutil.ReadFile(f)
		if err != nil {
			fmt.Println(err)
			return err
		}

		t := template.New(f)
		t.Funcs(template.FuncMap{"Mapper": m.Table2Obj,
			"Type": typestring,
			"Tag":  tag,
		})

		tmpl, err := t.Parse(string(bs))
		if err != nil {
			fmt.Println(err)
			return err
		}

		var w *os.File
		fileName := info.Name()
		newFileName := fileName[:len(fileName)-4]
		ext := path.Ext(newFileName)

		if !isMultiFile {
			w, err = os.OpenFile(path.Join(genDir, newFileName), os.O_RDWR|os.O_CREATE, 0700)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

		for _, table := range tables {
			// imports
			imports := make(map[string]string)
			for _, col := range table.Columns {
				if typestring(col.SQLType) == "time.Time" {
					imports["time.Time"] = "time.Time"
				}
			}

			if isMultiFile {
				w, err = os.OpenFile(path.Join(genDir, m.Table2Obj(table.Name)+ext), os.O_RDWR|os.O_CREATE, 0700)
				if err != nil {
					fmt.Println(err)
					return err
				}
			}

			newbytes := bytes.NewBufferString("")

			t := &Tmpl{Table: table, Imports: imports, Model: model}
			err = tmpl.Execute(newbytes, t)
			if err != nil {
				fmt.Println(err)
				return err
			}

			tplcontent, err := ioutil.ReadAll(newbytes)
			if err != nil {
				fmt.Println(err)
				return err
			}
			source, err := format.Source(tplcontent)
			if err != nil {
				fmt.Println(err)
				return err
			}

			w.WriteString(string(source))
			if isMultiFile {
				w.Close()
			}
		}
		if !isMultiFile {
			w.Close()
		}

		return nil
	})

}
