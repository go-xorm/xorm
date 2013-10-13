package main

import (
	"fmt"
	_ "github.com/bylevel/pq"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"xorm"
)

type SyncUser struct {
	Id      int64
	Name    string `xorm:"unique"`
	Age     int    `xorm:"index"`
	Title   string
	Address string
	Genre   string
	Area    string
	Date    int
}

type SyncLoginInfo struct {
	Id       int64
	IP       string `xorm:"index"`
	UserId   int64
	AddedCol int
	// timestamp should be updated by database, so only allow get from db
	TimeStamp string
	// assume
	Nonuse int    `xorm:"unique"`
	Newa   string `xorm:"index"`
}

func sync(engine *xorm.Engine) error {
	return engine.Sync(&SyncLoginInfo{}, &SyncUser{})
}

func sqliteEngine() (*xorm.Engine, error) {
	f := "sync.db"
	//os.Remove(f)

	return xorm.NewEngine("sqlite3", f)
}

func mysqlEngine() (*xorm.Engine, error) {
	return xorm.NewEngine("mysql", "root:@/test?charset=utf8")
}

func postgresEngine() (*xorm.Engine, error) {
	return xorm.NewEngine("postgres", "dbname=xorm_test sslmode=disable")
}

type engineFunc func() (*xorm.Engine, error)

func main() {
	engines := []engineFunc{sqliteEngine, mysqlEngine, postgresEngine}
	for _, enginefunc := range engines {
		Orm, err := enginefunc()
		fmt.Println("--------", Orm.DriverName, "----------")
		if err != nil {
			fmt.Println(err)
			return
		}
		Orm.ShowSQL = true
		err = sync(Orm)
		if err != nil {
			fmt.Println(err)
		}

		user := &SyncUser{
			Name:    "testsdf",
			Age:     15,
			Title:   "newsfds",
			Address: "fasfdsafdsaf",
			Genre:   "fsafd",
			Area:    "fafdsafd",
			Date:    1000,
		}
		_, err = Orm.Insert(user)
		if err != nil {
			fmt.Println(err)
		}
	}
}
