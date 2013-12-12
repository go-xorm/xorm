package main

import (
    "fmt"
    "github.com/lunny/xorm"
    _ "github.com/mattn/go-sqlite3"
    "os"
)

type User struct {
    Id   int64
    Name string
}

func main() {
    f := "pool.db"
    os.Remove(f)

    Orm, err := NewEngine("sqlite3", f)
    if err != nil {
        fmt.Println(err)
        return
    }
    err = Orm.SetPool(NewSimpleConnectPool())
    if err != nil {
        fmt.Println(err)
        return
    }

    Orm.ShowSQL = true
    err = Orm.CreateTables(&User{})
    if err != nil {
        fmt.Println(err)
        return
    }

    for i := 0; i < 10; i++ {
        _, err = Orm.Get(&User{})
        if err != nil {
            fmt.Println(err)
            break
        }

    }
}
