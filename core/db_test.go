package core

import (
	"fmt"
	"os"
	"testing"
	"time"

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

func TestOriQuery(t *testing.T) {
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			t.Error(err)
		}
	}

	rows, err := db.Query("select * from user")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	start := time.Now()

	for rows.Next() {
		var Id int64
		var Name, Title, Alias, NickName string
		var Age float32
		err = rows.Scan(&Id, &Name, &Title, &Age, &Alias, &NickName)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(Id, Name, Title, Age, Alias, NickName)
	}

	fmt.Println("ori ------", time.Now().Sub(start), "ns")
}

func TestStructQuery(t *testing.T) {
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			t.Error(err)
		}
	}

	rows, err := db.Query("select * from user")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()
	start := time.Now()

	for rows.Next() {
		var user User
		err = rows.ScanStruct(&user)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(user)
	}
	fmt.Println("struct ------", time.Now().Sub(start))
}

func TestStruct2Query(t *testing.T) {
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			t.Error(err)
		}
	}

	db.Mapper = &SnakeMapper{}

	rows, err := db.Query("select * from user")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()
	start := time.Now()

	for rows.Next() {
		var user User
		err = rows.ScanStruct2(&user)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(user)
	}
	fmt.Println("struct2 ------", time.Now().Sub(start))
}

func TestSliceQuery(t *testing.T) {
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			t.Error(err)
		}
	}

	rows, err := db.Query("select * from user")
	if err != nil {
		t.Error(err)
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Error(err)
	}

	start := time.Now()

	for rows.Next() {
		slice := make([]interface{}, len(cols))
		err = rows.ScanSlice(&slice)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(slice)
	}

	fmt.Println("slice ------", time.Now().Sub(start))
}

func TestMapQuery(t *testing.T) {
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			t.Error(err)
		}
	}

	rows, err := db.Query("select * from user")
	if err != nil {
		t.Error(err)
	}

	defer rows.Close()

	start := time.Now()

	for rows.Next() {
		m := make(map[string]interface{})
		err = rows.ScanMap(&m)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(m)
	}

	fmt.Println("map ------", time.Now().Sub(start))
}
