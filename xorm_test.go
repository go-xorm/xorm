package xorm_test

import (
	"fmt"
	_ "github.com/Go-SQL-Driver/MySQL"
	//_ "github.com/ziutek/mymysql/godrv"
	//_ "github.com/mattn/go-sqlite3"
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

func TestCreateEngine(t *testing.T) {
	engine = xorm.Create("mysql://root:123@localhost/test")
	//engine = orm.Create("mymysql://root:123@localhost/test")
	//engine = orm.Create("sqlite:///test.db")
	engine.ShowSQL = true
}

func TestDirectCreateTable(t *testing.T) {
	err := engine.CreateTables(&Userinfo{})
	if err != nil {
		t.Error(err)
	}
}

func TestMapper(t *testing.T) {
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

func TestInsert(t *testing.T) {
	user := Userinfo{1, "xiaolunwen", "dev", "lunny", time.Now()}
	_, err := engine.Insert(&user)
	if err != nil {
		t.Error(err)
	}
}

func TestInsertAutoIncr(t *testing.T) {
	// auto increment insert
	user := Userinfo{Username: "xiaolunwen", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err := engine.Insert(&user)
	if err != nil {
		t.Error(err)
	}
}

func TestInsertMulti(t *testing.T) {
	user1 := Userinfo{Username: "xlw", Departname: "dev", Alias: "lunny2", Created: time.Now()}
	user2 := Userinfo{Username: "xlw2", Departname: "dev", Alias: "lunny3", Created: time.Now()}
	_, err := engine.Insert(&user1, &user2)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdate(t *testing.T) {
	// update by id
	user := Userinfo{Uid: 1, Username: "xxx"}
	_, err := engine.Update(&user)
	if err != nil {
		t.Error(err)
	}
}

func TestDelete(t *testing.T) {
	user := Userinfo{Uid: 1}
	_, err := engine.Delete(&user)
	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	user := Userinfo{Uid: 2}

	err := engine.Get(&user)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(user)
}

func TestFind(t *testing.T) {
	users := make([]Userinfo, 0)

	err := engine.Find(&users)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(users)
}

func TestCount(t *testing.T) {
	user := Userinfo{}
	total, err := engine.Count(&user)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Total %d records!!!", total)
}

func TestWhere(t *testing.T) {
	users := make([]Userinfo, 0)
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	err = session.Where("id > ?", 2).Find(&users)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(users)
}

func TestLimit(t *testing.T) {
	users := make([]Userinfo, 0)
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	err = session.Limit(2, 1).Find(&users)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(users)
}

func TestOrder(t *testing.T) {
	users := make([]Userinfo, 0)
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	err = session.OrderBy("id desc").Find(&users)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(users)
}

func TestTransaction(*testing.T) {
}
