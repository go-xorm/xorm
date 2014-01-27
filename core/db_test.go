package core

import (
	"errors"
	"fmt"
	"os"
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

func BenchmarkOriQuery(b *testing.B) {
	b.StopTimer()
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		b.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			b.Error(err)
		}
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("select * from user")
		if err != nil {
			b.Error(err)
		}

		for rows.Next() {
			var Id int64
			var Name, Title, Alias, NickName string
			var Age float32
			err = rows.Scan(&Id, &Name, &Title, &Age, &Alias, &NickName)
			if err != nil {
				b.Error(err)
			}
			//fmt.Println(Id, Name, Title, Age, Alias, NickName)
		}
		rows.Close()
	}
}

func BenchmarkStructQuery(b *testing.B) {
	b.StopTimer()
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		b.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			b.Error(err)
		}
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("select * from user")
		if err != nil {
			b.Error(err)
		}

		for rows.Next() {
			var user User
			err = rows.ScanStruct(&user)
			if err != nil {
				b.Error(err)
			}
			if user.Name != "xlw" {
				fmt.Println(user)
				b.Error(errors.New("name should be xlw"))
			}
		}
		rows.Close()
	}
}

func BenchmarkStruct2Query(b *testing.B) {
	b.StopTimer()
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		b.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			b.Error(err)
		}
	}

	db.Mapper = NewCacheMapper(&SnakeMapper{})
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("select * from user")
		if err != nil {
			b.Error(err)
		}

		for rows.Next() {
			var user User
			err = rows.ScanStruct2(&user)
			if err != nil {
				b.Error(err)
			}
			if user.Name != "xlw" {
				fmt.Println(user)
				b.Error(errors.New("name should be xlw"))
			}
		}
		rows.Close()
	}
}

func BenchmarkSliceQuery(b *testing.B) {
	b.StopTimer()
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		b.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			b.Error(err)
		}
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("select * from user")
		if err != nil {
			b.Error(err)
		}

		cols, err := rows.Columns()
		if err != nil {
			b.Error(err)
		}

		for rows.Next() {
			slice := make([]interface{}, len(cols))
			err = rows.ScanSlice(&slice)
			if err != nil {
				b.Error(err)
			}
			if slice[1].(string) != "xlw" {
				fmt.Println(slice)
				b.Error(errors.New("name should be xlw"))
			}
		}

		rows.Close()
	}
}

func BenchmarkMapInterfaceQuery(b *testing.B) {
	b.StopTimer()
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		b.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			b.Error(err)
		}
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("select * from user")
		if err != nil {
			b.Error(err)
		}

		for rows.Next() {
			m := make(map[string]interface{})
			err = rows.ScanMap(&m)
			if err != nil {
				b.Error(err)
			}
			if m["name"].(string) != "xlw" {
				fmt.Println(m)
				b.Error(errors.New("name should be xlw"))
			}
		}

		rows.Close()
	}
}

func BenchmarkMapBytesQuery(b *testing.B) {
	b.StopTimer()
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		b.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			b.Error(err)
		}
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("select * from user")
		if err != nil {
			b.Error(err)
		}

		for rows.Next() {
			m := make(map[string][]byte)
			err = rows.ScanMap(&m)
			if err != nil {
				b.Error(err)
			}
			if string(m["name"]) != "xlw" {
				fmt.Println(m)
				b.Error(errors.New("name should be xlw"))
			}
		}

		rows.Close()
	}
}

func BenchmarkMapStringQuery(b *testing.B) {
	b.StopTimer()
	os.Remove("./test.db")
	db, err := Open("sqlite3", "./test.db")
	if err != nil {
		b.Error(err)
	}

	_, err = db.Exec(createTableSqlite3)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 50; i++ {
		_, err = db.Exec("insert into user (name, title, age, alias, nick_name) values (?,?,?,?,?)",
			"xlw", "tester", 1.2, "lunny", "lunny xiao")
		if err != nil {
			b.Error(err)
		}
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query("select * from user")
		if err != nil {
			b.Error(err)
		}

		for rows.Next() {
			m := make(map[string]string)
			err = rows.ScanMap(&m)
			if err != nil {
				b.Error(err)
			}
			if m["name"] != "xlw" {
				fmt.Println(m)
				b.Error(errors.New("name should be xlw"))
			}
		}

		rows.Close()
	}
}
