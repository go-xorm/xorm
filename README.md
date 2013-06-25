# xorm

[中文](https://github.com/lunny/xorm/blob/master/README_CN.md)

xorm is a simple and powerful ORM for Go. It makes dabatabse operating simple. 

[![Build Status](https://drone.io/github.com/lunny/xorm/status.png)](https://drone.io/github.com/lunny/xorm/latest)


## Drivers Support

Drivers for Go's sql package which currently support database/sql includes:

* Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

## Changelog

* **v0.1.6** : Added conversion interface support; added struct derive support; added single mapping support
* **v0.1.5** : Added multi threads support; added Sql() function for struct query; Get function changed return inteface; MakeSession and Create are instead with NewSession and NewEngine.
* **v0.1.4** : Added simple cascade load support; added more data type supports.
* **v0.1.3** : Find function now supports both slice and map; Add Table function for multi tables and temperory tables support
* **v0.1.2** : Insert function now supports both struct and slice pointer parameters, batch inserting and auto transaction
* **v0.1.1** : Add Id, In functions and improved README
* **v0.1.0** : Inital release.

## Features

* Struct<->Table Mapping Supports, both name mapping and filed tags mapping

* Database Transaction Support

* Both ORM and SQL Operation Support

* Simply usage

* Support Id, In, Where, Limit, Join, Having, Sql functions and sturct as query conditions

* Support simple cascade load just like Hibernate for Java


## Installing xorm

	go get github.com/lunny/xorm

## Quick Start

1.Create a database engine just like sql.Open, commonly you just need create once. Please notice, Create function will be deprecated, use NewEngine instead.

```Go
import (
	_ "github.com/Go-SQL-Driver/MySQL"
	"github.com/lunny/xorm"
)
err, engine := xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
```

or

```Go
import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/lunny/xorm"
)
err, engine = xorm.NewEngine("sqlite3", "./test.db")
```

1.1.If you want to show all generated SQL

```Go
engine.ShowSQL = true
```

2.Define a struct

```Go
type User struct {
	Id int
    Name string
    Age int    `xorm:"-"`
}
```

2.1.More mapping rules, please see [Mapping Rules](#mapping)

3.When you set up your program, you can use CreateTables to create database tables.

```Go
err := engine.CreateTables(&User{})
// or err := engine.Map(&User{}, &Article{})
// err = engine.CreateAll()
```

4.then, insert a struct to table, if success, User.Id will be set to id

```Go
id, err := engine.Insert(&User{Name:"lunny"})
```

or if you want to update records

```Go
user := User{Name:"xlw"}
rows, err := engine.Update(&user, &User{Id:1})
// or rows, err := engine.Where("id = ?", 1).Update(&user)
// or rows, err := engine.Id(1).Update(&user)
```

5.Fetch a single object by user

```Go
var user = User{Id:27}
has, err := engine.Get(&user)
// or has, err := engine.Id(27).Get(&user)

var user = User{Name:"xlw"}
has, err := engine.Get(&user)
```
	
6.Fetch multipe objects into a slice or a map, use Find：

```Go
var everyone []Userinfo
err := engine.Find(&everyone)

users := make(map[int64]Userinfo)
err := engine.Find(&users)
```

6.1 also you can use Where, Limit

```Go
var allusers []Userinfo
err := engine.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20
```

6.2 or you can use a struct query

```Go
var tenusers []Userinfo
err := engine.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10 offset 0
```

6.3 or In function

```Go
var tenusers []Userinfo
err := engine.In("id", 1, 3, 5).Find(&tenusers) //Get All id in (1, 3, 5)
```

7.Delete

```Go
err := engine.Delete(&User{Id:1})
// or err := engine.Id(1).Delete(&User{})
```

8.Count

```Go
total, err := engine.Count(&User{Name:"xlw"})
```


## Execute SQL

Of course, SQL execution is also provided.

1.if select then use Query

```Go
sql := "select * from userinfo"
results, err := engine.Query(sql)
```

2.if insert, update or delete then use Exec

```Go
sql = "update userinfo set username=? where id=?"
res, err := engine.Exec(sql, "xiaolun", 1) 
```


## Advanced Usage

for deep usage, you should create a session, this func will create a database connection immediatelly.Please notice, MakeSession will be deprecated last, use NewSession instead

```Go
session := engine.NewSession()
defer session.Close()

```

1.Fetch a single object by where

```Go
var user Userinfo
session.Where("id=?", 27).Get(&user)

var user2 Userinfo
session.Where("name = ?", "john").Get(&user3) // more complex query

var user3 Userinfo
session.Where("name = ? and age < ?", "john", 88).Get(&user4) // even more complex
```

2.Fetch multiple objects

```Go
var allusers []Userinfo
err := session.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20

var tenusers []Userinfo
err := session.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10  if omit offset the default is 0

var everyone []Userinfo
err := session.Find(&everyone)
```
	
3.Transaction

```Go
// add Begin() before any action
err := session.Begin()	
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
```

4.Mixed Transaction

```Go
// add Begin() before any action
err := session.Begin()	
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
```

## Mapping Rules 

<a name="mapping" id="mapping"></a>
1.Struct and struct's fields name should be Pascal style, and the table and column's name default is SQL style.

For example: 

The struct's Name 'UserInfo' will turn into the table name 'user_info', the same as the keyname. If the keyname is 'UserName' will turn into the select colum 'user_name'

2.If You want change the mapping rules, you have two methods. One is to implement your own Map struct interface according IMapper, you can find the interface in mapper.go and set it to engine.Mapper

Another is use field tag, field tag support the below keywords which split with space:

<table>
    <tr>
        <td>name</td><td>column name</td>
    </tr>
    <tr>
        <td>pk</td><td>the field is a primary key</td>
    </tr>
    <tr>
        <td>int(11)/varchar(50)/text/date/datetime/blob/decimal(26,2)</td><td>column type</td>
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

For Example

```Go
type Userinfo struct {
	Uid        int `xorm:"id pk not null autoincr"`
	Username   string
	Departname string
	Alias      string `xorm:"-"`
	Created    time.Time
}
```
3.For customize table name, use Table() function, for example:

```Go
// batch create tables
for i := 0; i < 10; i++ {
	engine.Table(fmt.Sprintf("user_%v", i)).CreateTable(&Userinfo{}) 
}

// insert into table according id
user := Userinfo{Uid: 25, Username:"sslfs"}
engine.Table(fmt.Sprintf("user_%v", user.Uid % 10)).Insert(&user)
```

## Documents 

Please visit [GoWalker](http://gowalker.org/github.com/lunny/xorm)


## FAQ 

1.How the xorm tag use both with json?
  
  Use space.

```Go
type User struct {
    Name string `json:"name" xorm:"name"`
}
```

## LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
