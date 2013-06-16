package main

import (
	//xorm "github.com/lunny/xorm"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	//"time"
	xorm "xorm"
)

type User struct {
	Id   int
	Name string
}

func sqliteEngine() (*xorm.Engine, error) {
	os.Remove("./test.db")
	return xorm.NewEngine("sqlite3", "./goroutine.db")
}

func mysqlEngine() (*xorm.Engine, error) {
	return xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
}

func main() {
	engine, err := sqliteEngine()
	// engine, err := mysqlEngine()

	if err != nil {
		fmt.Println(err)
		return
	}

	u := &User{}

	err = engine.CreateTables(u)
	if err != nil {
		fmt.Println(err)
		return
	}

	size := 10
	queue := make(chan int, size)

	for i := 0; i < size; i++ {
		go func(x int) {
			//x := i
			err := engine.Test()
			if err != nil {
				fmt.Println(err)
			} else {
				err = engine.Map(u)
				if err != nil {
					fmt.Println("Map user failed")
				} else {
					for j := 0; j < 10; j++ {
						if x+j < 2 {
							_, err = engine.Get(u)
						} else if x+j < 4 {
							users := make([]User, 0)
							err = engine.Find(&users)
						} else if x+j < 8 {
							_, err = engine.Count(u)
						} else if x+j < 16 {
							_, err = engine.Insert(&User{Name: "xlw"})
						} else if x+j < 32 {
							_, err = engine.Id(1).Delete(u)
						}
						if err != nil {
							fmt.Println(err)
							queue <- x
							return
						}
					}
					fmt.Printf("%v success!\n", x)
				}
			}
			queue <- x
		}(i)
	}

	for i := 0; i < size; i++ {
		<-queue
	}
	fmt.Println("end")
}
