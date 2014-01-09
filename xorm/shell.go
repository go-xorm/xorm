package main

import (
	"fmt"
	"github.com/lunny/xorm"
	"strings"
)

var CmdShell = &Command{
	UsageLine: "shell driverName datasourceName",
	Short:     "a general shell to operate all kinds of database",
	Long: `
general database's shell for sqlite3, mysql, postgres.

    driverName        Database driver name, now supported four: mysql mymysql sqlite3 postgres
    datasourceName    Database connection uri, for detail infomation please visit driver's project page
`,
}

func init() {
	CmdShell.Run = runShell
	CmdShell.Flags = map[string]bool{}
}

var engine *xorm.Engine

func shellHelp() {
	fmt.Println(`
        show tables                    show all tables
        columns <table_name>         show table's column info
        indexes <table_name>        show table's index info
        exit                         exit shell
        source <sql_file>            exec sql file to current database
        dump [-nodata] <sql_file>    dump structs or records to sql file
        help                        show this document
        <statement>                    SQL statement
    `)
}

func runShell(cmd *Command, args []string) {
	if len(args) != 2 {
		fmt.Println("params error, please see xorm help shell")
		return
	}

	var err error
	engine, err = xorm.NewEngine(args[0], args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	err = engine.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}

	var scmd string
	fmt.Print("xorm$ ")
	for {
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if strings.ToLower(input) == "exit" {
			fmt.Println("bye")
			return
		}
		if !strings.HasSuffix(input, ";") {
			scmd = scmd + " " + input
			continue
		}
		scmd = scmd + " " + input
		lcmd := strings.TrimSpace(strings.ToLower(scmd))
		if strings.HasPrefix(lcmd, "select") {
			res, err := engine.Query(scmd + "\n")
			if err != nil {
				fmt.Println(err)
			} else {
				if len(res) <= 0 {
					fmt.Println("no records")
				} else {
					columns := make(map[string]int)
					for k, _ := range res[0] {
						columns[k] = len(k)
					}

					for _, m := range res {
						for k, s := range m {
							l := len(string(s))
							if l > columns[k] {
								columns[k] = l
							}
						}
					}

					var maxlen = 0
					for _, l := range columns {
						maxlen = maxlen + l + 3
					}
					maxlen = maxlen + 1

					fmt.Println(strings.Repeat("-", maxlen))
					fmt.Print("|")
					slice := make([]string, 0)
					for k, l := range columns {
						fmt.Print(" " + k + " ")
						fmt.Print(strings.Repeat(" ", l-len(k)))
						fmt.Print("|")
						slice = append(slice, k)
					}
					fmt.Print("\n")
					for _, r := range res {
						fmt.Print("|")
						for _, k := range slice {
							fmt.Print(" " + string(r[k]) + " ")
							fmt.Print(strings.Repeat(" ", columns[k]-len(string(r[k]))))
							fmt.Print("|")
						}
						fmt.Print("\n")
					}
					fmt.Println(strings.Repeat("-", maxlen))
					//fmt.Println(res)
				}
			}
		} else if lcmd == "show tables;" {
			/*tables, err := engine.DBMetas()
			  if err != nil {
			      fmt.Println(err)
			  } else {

			  }*/
		} else {
			cnt, err := engine.Exec(scmd)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%d records changed.\n", cnt)
			}
		}
		scmd = ""
		fmt.Print("xorm$ ")
	}
}
