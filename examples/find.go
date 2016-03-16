package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-xorm/xorm"
)

type User struct {
	Id      int64
	Name    string
	Created time.Time `xorm:"created"`
	Updated time.Time `xorm:"updated"`
}

func main() {
	f := "conversion.db"
	os.Remove(f)

	Orm, err := xorm.NewEngine("sqlite3", f)
	if err != nil {
		fmt.Println(err)
		return
	}
	Orm.ShowSQL(true)

	err = Orm.CreateTables(&User{})
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = Orm.Insert(&User{Id: 1, Name: "xlw"})
	if err != nil {
		fmt.Println(err)
		return
	}

	users := make([]User, 0)
	err = Orm.Find(&users)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(users)
}
