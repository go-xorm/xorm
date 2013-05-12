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

type Userdetail struct {
	Uid     int `xorm:"id pk not null"`
	Intro   string
	Profile string
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

	err = engine.Map(&Userinfo{}, &Userdetail{})
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

func query(t *testing.T) {
	sql := "select * from userinfo"
	results, err := engine.Query(sql)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(results)
}

func exec(t *testing.T) {
	sql := "update userinfo set username=? where id=?"
	res, err := engine.Exec(sql, "xiaolun", 1)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
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
	users := []*Userinfo{
		{Username: "xlw", Departname: "dev", Alias: "lunny2", Created: time.Now()},
		{Username: "xlw2", Departname: "dev", Alias: "lunny3", Created: time.Now()},
	}
	_, err := engine.Insert(&users)
	if err != nil {
		t.Error(err)
	}

	engine.InsertMany = false

	users = []*Userinfo{
		{Username: "xlw9", Departname: "dev", Alias: "lunny9", Created: time.Now()},
		{Username: "xlw10", Departname: "dev", Alias: "lunny10", Created: time.Now()},
	}
	_, err = engine.Insert(&users)
	if err != nil {
		t.Error(err)
	}

	engine.InsertMany = true
}

func insertTwoTable(t *testing.T) {
	userinfo := Userinfo{Username: "xlw3", Departname: "dev", Alias: "lunny4", Created: time.Now()}
	userdetail := Userdetail{Uid: 1, Intro: "I'm a very beautiful women.", Profile: "sfsaf"}
	_, err := engine.Insert(&userinfo, &userdetail)
	if err != nil {
		t.Error(err)
	}
}

func update(t *testing.T) {
	// update by id
	user := Userinfo{Username: "xxx"}
	_, err := engine.Id(1).Update(&user)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = engine.Update(&Userinfo{Username: "yyy"}, &user)
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

func in(t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.In("id", 1, 2, 3).Find(&users)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(users)

	ids := []interface{}{1, 2, 3}
	err = engine.Where("id > ?", 2).In("id", ids...).Find(&users)
	if err != nil {
		t.Error(err)
		return
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

func join(t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.Join("LEFT", "userdetail", "userinfo.id=userdetail.id").Find(&users)
	if err != nil {
		t.Error(err)
	}
}

func having(t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.GroupBy("username").Having("username='xlw'").Find(&users)
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
	defer counter()
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		t.Error(err)
		return
	}

	session.Begin()
	//session.IsAutoRollback = false
	user1 := Userinfo{Username: "xiaoxiao", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		t.Error(err)
		return
	}
	user2 := Userinfo{Username: "yyy"}
	_, err = session.Where("uid = ?", 0).Update(&user2)
	if err != nil {
		session.Rollback()
		fmt.Println(err)
		//t.Error(err)
		return
	}

	_, err = session.Delete(&user2)
	if err != nil {
		session.Rollback()
		t.Error(err)
		return
	}

	err = session.Commit()
	if err != nil {
		t.Error(err)
		return
	}
}

func combineTransaction(t *testing.T) {
	counter := func() {
		total, err := engine.Count(&Userinfo{})
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("----now total %v records\n", total)
	}

	counter()
	defer counter()
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		t.Error(err)
		return
	}

	session.Begin()
	//session.IsAutoRollback = false
	user1 := Userinfo{Username: "xiaoxiao2", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		t.Error(err)
		return
	}
	user2 := Userinfo{Username: "zzz"}
	_, err = session.Where("id = ?", 0).Update(&user2)
	if err != nil {
		session.Rollback()
		t.Error(err)
		return
	}

	_, err = session.Exec("delete from userinfo where username = ?", user2.Username)
	if err != nil {
		session.Rollback()
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
	engine = xorm.Create("mysql", "root:123@/test?charset=utf8")
	engine.ShowSQL = true

	directCreateTable(t)
	mapper(t)
	insert(t)
	query(t)
	exec(t)
	insertAutoIncr(t)
	insertMulti(t)
	insertTwoTable(t)
	update(t)
	delete(t)
	get(t)
	find(t)
	count(t)
	where(t)
	in(t)
	limit(t)
	order(t)
	join(t)
	having(t)
	transaction(t)
	combineTransaction(t)
}

func TestSqlite(t *testing.T) {
	engine = xorm.Create("sqlite3", "./test.db")
	engine.ShowSQL = true

	directCreateTable(t)
	mapper(t)
	insert(t)
	query(t)
	exec(t)
	insertAutoIncr(t)
	insertMulti(t)
	insertTwoTable(t)
	update(t)
	delete(t)
	get(t)
	find(t)
	count(t)
	where(t)
	in(t)
	limit(t)
	order(t)
	join(t)
	having(t)
	transaction(t)
	combineTransaction(t)
}
