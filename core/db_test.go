package core

import (
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var (
	createTableSqlite3 = "CREATE TABLE IF NOT EXISTS `user` (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `name` TEXT NULL, `title` TEXT NULL, `age` FLOAT NULL, `alias` TEXT NULL, `nick_name` TEXT NULL);"
)

type User struct {
	Id       int64
	Name     string
	Title    string
	Age      float32
	Alias    string
	NickName string
}

func TestQuery(t *testing.T) {
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
		"xlw", "tester", 1.2, "lunny", "lunny xiao")
	if err != nil {
		t.Error(err)
	}

	rows, err := db.Query("select * from user")
	if err != nil {
		t.Error(err)
	}

	for rows.Next() {
		var user User
		err = rows.Scan(&user)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(user)
	}
}
