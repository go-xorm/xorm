Quick Start
=====

* [1.Create database engine](#10)
* [2.Define a struct](#20)
* [3.Create tables](#30)
	* [3.1.Sync database schema](#31)
* [4.Insert, update one or more records](#40)
* [5.Get one record](#50)
* [6.Find many records](#60)
* [7.Iterate records](#70)
* [8.Delete records](#80)
* [9.Count records](#90)
* [10.Cache](#100)
* [11.Execute SQL](#110)
* [12.Advanced Usage](#120)
* [13.Mapping Rules](#130)

<a name="10" id="10"></a>
## 1.Create database engine
Create a database engine just like sql.Open, commonly you just need create once. Please notice, Create function will be deprecated, use NewEngine instead.

```Go
import (
	_ "github.com/Go-SQL-Driver/MySQL"
	"github.com/lunny/xorm"
)
engine, err := xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
defer engine.Close()
```

or

```Go
import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/lunny/xorm"
)
engine, err = xorm.NewEngine("sqlite3", "./test.db")
defer engine.Close()
```

1.1.If you want to show all generated SQL

```Go
engine.ShowSQL = true
```
1.2 Defaultly, xorm use go's connection pool. If you want to use your own connection pool, you can

```Go
err = engine.SetPool(NewSimpleConnectPool())
```

1.3 If you want to enable cache system

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
Engine.SetDefaultCacher(cacher)
```

<a name="20" id="20"></a>
## 2.Define a struct

```Go
type User struct {
	Id int
    Name string
    Age int    `xorm:"-"`
}
```

2.1.More mapping rules, please see [Mapping Rules](#mapping)

<a name="30" id="30"></a>
## 3.Create tables
When you set up your program, you can use CreateTables to create database tables.

```Go
err := engine.CreateTables(&User{})
// or err := engine.Map(&User{}, &Article{})
// err = engine.CreateAll()
```

3.1 If you want to auto sync database schema

```Go
err = engine.Sync(new(User), new(Category))
```

<a name="40" id="40"></a>
## 4.Insert, Update records
then, insert a struct to table, if success, User.Id will be set to id

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

<a name="50" id="50"></a>
## 5.Get one record
Fetch a single object by user

```Go
var user = User{Id:27}
has, err := engine.Get(&user)
// or has, err := engine.Id(27).Get(&user)

var user = User{Name:"xlw"}
has, err := engine.Get(&user)
```

<a name="60" id="60"></a>
## 6.Find many records
Fetch multipe objects into a slice or a map, use Find：

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

6.4 The default will query all columns of a table. Use Cols function if you want to select some columns

```Go
var tenusers []Userinfo
err := engine.Cols("id", "name").Find(&tenusers) //Find only id and name
```

<a name="70" id="70"></a>
## 7.Iterate records
Iterate, like find, but handle records one by one

```Go
err := engine.Where("age > ? or name=?)", 30, "xlw").Iterate(new(Userinfo), func(i int, bean interface{})error{
	user := bean.(*Userinfo)
	//do somthing use i and user
})
```

<a name="80" id="80"></a>
## 8.Delete one or more records
Delete one or more records

8.1 deleted by id

```Go
err := engine.Id(1).Delete(&User{})
```

8.2 deleted by other conditions

```Go
err := engine.Delete(&User{Name:"xlw"})
```

<a name="90" id="90"></a>
## 9.Count records
9.Count

```Go
total, err := engine.Where("id > ?", 5).Count(&User{Name:"xlw"})
```

<a name="100" id="100"></a>
## 10.Cache
```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.SetDefaultCacher(cacher)
```

<a name="110" id="110"></a>
## 11.Execute SQL

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

<a name="120" id="120"></a>
## 12.Advanced Usage

For deep usage, you should create a session.

```Go
session := engine.NewSession()
defer session.Close()
```

1.Fetch a single object by where, these methods are same to engine.

```Go
var user Userinfo
session.Where("id=?", 27).Get(&user)

var userJohn Userinfo
session.Where("name = ?", "john").Get(&userJohn) // more complex query

var userOldJohn Userinfo
session.Where("name = ? and age > ?", "john", 88).Get(&userOldJohn) // even more complex
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
5.Derive mapping
Please see derive.go in examples folder.

<a name="130" id="130"></a>
## 13.Mapping Rules 

<a name="mapping" id="mapping"></a>
1.Struct and struct's fields name should be Pascal style, and the table and column's name default is SQL style.

For example: 

The struct's Name 'UserInfo' will turn into the table name 'user_info', the same as the keyname. If the keyname is 'UserName' will turn into the select colum 'user_name'

2.If You want change the mapping rules, you have two methods. One is to implement your own Map struct interface according IMapper, you can find the interface in mapper.go and set it to engine.Mapper

Another is use field tag, field tag support the below keywords which split with space:

<table>
    <tr>
        <td>name</td><td>column name, if no this name, the name is auto generated according field name and mapper rule.</td>
    </tr>
    <tr>
        <td>pk</td><td>the field is a primary key</td>
    </tr>
    <tr>
        <td>more than 30 column type supported, please see [Column Type](https://github.com/lunny/xorm/blob/master/docs/COLUMNTYPE.md)</td><td>column type</td>
    </tr>
    <tr>
        <td>autoincr</td><td>auto incrment</td>
    </tr>
    <tr>
        <td>[not ]null</td><td>if column can be null value</td>
    </tr>
    <tr>
        <td>unique or unique(uniquename)</td><td>unique or union unique as uniquename</td>
    </tr>
    <tr>
        <td>index or index(indexname)</td><td>index or union index as indexname</td>
    </tr>
     <tr>
        <td>extends</td><td>used in anonymous struct means mapping this struct's fields to table</td>
    </tr>
    <tr>
        <td>-</td><td>this field is not map as a table column</td>
    </tr>
    <tr>
        <td>-></td><td>this field only write to db and not read from db</td>
    </tr>
     <tr>
        <td>&lt;-</td><td>this field only read from db and not write to db</td>
    </tr>
     <tr>
        <td>created</td><td>this field will auto fill current time when insert</td>
    </tr>
     <tr>
        <td>updated</td><td>this field will auto fill current time when update</td>
    </tr>
    <tr>
        <td>default 0 or default 'abc'</td><td>default value, use single quote for string</td>
    </tr>
</table>

For Example

```Go
type Userinfo struct {
	Uid        int `xorm:"id pk not null autoincr"`
	Username   string `xorm:"unique"`
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