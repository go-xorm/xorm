# xorm
===========

[中文](README_CN.md)

xorm is an ORM for Go. It lets you map Go structs to tables in a database. 

Right now, it interfaces with Mysql/SQLite. The goal however is to add support for PostgreSQL/DB2/MS ADODB/ODBC/Oracle in the future. 

All in all, it's not entirely ready for product use yet, but it's getting there.

Drivers for Go's sql package which support database/sql includes:

Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

## Installing xorm

	go get github.com/lunny/xorm

## Quick Start

1.Create a database engine (for example: mysql)

	engine := xorm.Create("mysql://root:123@localhost/test")


2.Define your struct


	type User struct {
    	Id int
    	Name string
    	Age int    `xorm:"-"`
	}


for Simple Task, just use engine's functions:

before beginning, you should create a database in mysql and then we will create the tables.


	err := engine.CreateTables(&User{})

	
then, insert an struct to table
  

	id, err := engine.Insert(&User{Name:"lunny"})


or you want to update this struct


	user := User{Name:"xlw"}
	rows, err := engine.Update(&user, &User{Id:1})
	// rows, err := engine.Where("id = ?", 1).Update(&user)


3.Fetch a single object by user


	var user = User{Id:27}
	err := engine.Get(&user)

	var user = User{Name:"xlw"}
	err := engine.Get(&user)
	
4.Fetch multipe objects, use Find：

	var allusers []Userinfo
	err := engine.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20

	var tenusers []Userinfo
	err := engine.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10  if omit offset the default is 0

	var everyone []Userinfo
	err := engine.Find(&everyone)

5.Delete and Count：

	err := engine.Delete(&User{Id:1})
	
	total, err := engine.Count(&User{Name:"xlw"})

##Origin Use
Of course, the basic usage is also provided.

	sql := "select * from userinfo"
	results, err := engine.Query(sql)
	
	sql = "update userinfo set username=? where id=?"
	res, err := engine.Exec(sql, "xiaolun", 1) 

##Deep Use
for deep use, you should create a session, this func will create a connection to db


	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
    	return
	}


1.Fetch a single object by where

	var user Userinfo
	session.Where("id=?", 27).Get(&user)

	var user2 Userinfo
	session.Where("name = ?", "john").Get(&user3) // more complex query

	var user3 Userinfo
	session.Where("name = ? and age < ?", "john", 88).Get(&user4) // even more complex


2.Fetch multiple objects

	var allusers []Userinfo
	err := session.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20

	var tenusers []Userinfo
	err := session.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10  if omit offset the default is 0

	var everyone []Userinfo
	err := session.Find(&everyone)
	
3.Transaction

	// add Begin() before any action
	session.Begin()	
	user1 := Userinfo{Username: "xiaoxiao", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		return
	}
	user2 := Userinfo{Username: "yyy"}
	_, err = session.Where("id = ?", 2).Update(&user2)
	if err != nil {
		session.Rollback()
		return
	}

	_, err = session.Delete(&user2)
	if err != nil {
		session.Rollback()
		return
	}

    // add Commit() after all actions
	err = session.Commit()
	if err != nil {
		return
	}

4.Mixed Transaction

	// add Begin() before any action
	session.Begin()	
	user1 := Userinfo{Username: "xiaoxiao", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		return
	}
	user2 := Userinfo{Username: "yyy"}
	_, err = session.Where("id = ?", 2).Update(&user2)
	if err != nil {
		session.Rollback()
		return
	}

	_, err = session.Exec("delete from userinfo where username = ?", user2.Username)
	if err != nil {
		session.Rollback()
		return
	}

    // add Commit() after all actions
	err = session.Commit()
	if err != nil {
		return
	}

##Mapping Rules
1.Struct and struct's fields name should be Pascal style, and the table and column's name default is us

For example: 
The structs Name 'UserInfo' will turn into the table name 'user_info', the same as the keyname.	
If the keyname is 'UserName' will turn into the select colum 'user_name'

2.You have two method to change the rule. One is implement your own Map interface according IMapper, you can find the interface in mapper.go and set it to engine.Mapper

another is use field tag, field tag support the below keywords:

<table>
    <tr>
        <td>name</td><td>column name</td>
    </tr>
    <tr>
        <td>pk</td><td>the field is a primary key</td>
    </tr>
    <tr>
        <td>int(11)/varchar(50)</td><td>column type</td>
    </tr>
    <tr>
        <td>autoincr</td><td>auto incrment</td>
    </tr>
    <tr>
        <td>[not ]null</td><td>if column can be null value</td>
    </tr>
    <tr>
        <td>unique</td><td>unique</td>
    </tr>
    <tr>
        <td>-</td><td>this field is not map as a table column</td>
    </tr>
</table>


##FAQ
1.How the xorm tag use both with json?
  
  use space

	type User struct {
    	User string `json:"user" orm:"user_id"`
	}

## LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
