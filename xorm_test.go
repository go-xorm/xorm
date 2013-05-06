package xorm_test

import (
	"fmt"
	_ "github.com/Go-SQL-Driver/MySQL"
	_ "github.com/mattn/go-sqlite3"
	"testing"
	"time"
	"xorm"
)

/*
CREATE TABLE `userinfo` (
	`uid` INT(10) NULL AUTO_INCREMENT,
	`username` VARCHAR(64) NULL,
	`departname` VARCHAR(64) NULL,
	`created` DATE NULL,
	PRIMARY KEY (`uid`)
);
CREATE TABLE `userdeatail` (
	`uid` INT(10) NULL,
	`intro` TEXT NULL,
	`profile` TEXT NULL,
	PRIMARY KEY (`uid`)
);
*/

type Userinfo struct {
	Uid        int `xorm:"id pk not null autoincr"`
	Username   string
	Departname string
	Alias      string `xorm:"-"`
	Created    time.Time
}

var engine xorm.Engine

func directCreateTable(t *testing.T) {
	err := engine.CreateTables(&Userinfo{})
	if err != nil {
		t.Error(err)
	}
}

func mapper(t *testing.T) {
	err := engine.UnMap(&Userinfo{})
	if err != nil {
		t.Error(err)
	}

	err = engine.Map(&Userinfo{})
	if err != nil {
		t.Error(err)
	}

	err = engine.DropAll()
	if err != nil {
		t.Error(err)
	}

	err = engine.CreateAll()
	if err != nil {
		t.Error(err)
	}
}

func insert(t *testing.T) {
	user := Userinfo{1, "xiaolunwen", "dev", "lunny", time.Now()}
	_, err := engine.Insert(&user)
	if err != nil {
		t.Error(err)
	}
}

func insertAutoIncr(t *testing.T) {
	// auto increment insert
	user := Userinfo{Username: "xiaolunwen", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err := engine.Insert(&user)
	if err != nil {
		t.Error(err)
	}
}

func insertMulti(t *testing.T) {
	user1 := Userinfo{Username: "xlw", Departname: "dev", Alias: "lunny2", Created: time.Now()}
	user2 := Userinfo{Username: "xlw2", Departname: "dev", Alias: "lunny3", Created: time.Now()}
	_, err := engine.Insert(&user1, &user2)
	if err != nil {
		t.Error(err)
	}
}

func update(t *testing.T) {
	// update by id
	user := Userinfo{Uid: 1, Username: "xxx"}
	_, err := engine.Update(&user)
	if err != nil {
		t.Error(err)
	}
}

func delete(t *testing.T) {
	user := Userinfo{Uid: 1}
	_, err := engine.Delete(&user)
	if err != nil {
		t.Error(err)
	}
}

func get(t *testing.T) {
	user := Userinfo{Uid: 2}

	err := engine.Get(&user)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(user)
}

func find(t *testing.T) {
	users := make([]Userinfo, 0)

	err := engine.Find(&users)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(users)
}

func count(t *testing.T) {
	user := Userinfo{Departname: "dev"}
	total, err := engine.Count(&user)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Total %d records!!!", total)
}

func where(t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.Where("id > ?", 2).Find(&users)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(users)
}

func limit(t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.Limit(2, 1).Find(&users)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(users)
}

func order(t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.OrderBy("id desc").Find(&users)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(users)
}

func transaction(t *testing.T) {
	counter := func() {
		total, err := engine.Count(&Userinfo{})
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("----now total %v records\n", total)
	}

	counter()

	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		t.Error(err)
		return
	}

	defer counter()

	session.Begin()
	session.IsAutoRollback = true
	user1 := Userinfo{Username: "xiaoxiao", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		t.Error(err)
		return
	}
	user2 := Userinfo{Username: "yyy"}
	_, err = session.Where("id = ?", 2).Update(&user2)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = session.Delete(&user2)
	if err != nil {
		t.Error(err)
		return
	}

	err = session.Commit()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestMysql(t *testing.T) {
	engine = xorm.Create("mysql://root:123@localhost/test")
	engine.ShowSQL = true

	directCreateTable(t)
	mapper(t)
	insert(t)
	insertAutoIncr(t)
	insertMulti(t)
	update(t)
	delete(t)
	get(t)
	find(t)
	count(t)
	where(t)
	limit(t)
	order(t)
	transaction(t)
}

func TestSqlite(t *testing.T) {
	engine = xorm.Create("sqlite:///test.db")
	engine.ShowSQL = true

	directCreateTable(t)
	mapper(t)
	insert(t)
	insertAutoIncr(t)
	insertMulti(t)
	update(t)
	delete(t)
	get(t)
	find(t)
	count(t)
	where(t)
	limit(t)
	order(t)
	transaction(t)
}
