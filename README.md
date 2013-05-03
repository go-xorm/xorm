xorm
=====

[中文](README_CN.md)

xorm is an ORM for Go. It lets you map Go structs to tables in a database. 

Right now, it interfaces with Mysql/SQLite. The goal however is to add support for PostgreSQL/DB2/MS ADODB/ODBC/Oracle in the future. 

All in all, it's not entirely ready for advanced use yet, but it's getting there.

Drivers for Go's sql package which support database/sql includes:

Mysql:[github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

Mysql:[github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

SQLite:[github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

### Installing xorm
    go get github.com/lunny/xorm

### Quick Start

1.Create an database engine (for example: mysql)

```go
engine := xorm.Create("mysql://root:123@localhost/test")
```

2.Define your struct

```go
type User struct {
    Id int
    Name string
    Age int    `xorm:"-"`
}
```

for Simple Task, just use engine's functions:

begin start, you should create a database and then we create the tables

```go
err := engine.CreateTables(&User{})
```
	
then, insert an struct to table
  
```go
id, err := engine.Insert(&User{Name:"lunny"})
```

or you want to update this struct

```go
user := User{Id:1, Name:"xlw"}
rows, err := engine.Update(&user)
```

3.Fetch a single object by user

```go
var user = User{Id:27}
engine.Get(&user)

var user = User{Name:"xlw"}
engine.Get(&user)
```

for deep use, you should create a session, this func will create a connection to db

```go
session, err := engine.MakeSession()
defer session.Close()
if err != nil {
    return
}
```

1.Fetch a single object by where

```go
var user Userinfo
session.Where("id=?", 27).Get(&user)

var user2 Userinfo
session.Where(3).Get(&user2) // this is shorthand for the version above

var user3 Userinfo
session.Where("name = ?", "john").Get(&user3) // more complex query

var user4 Userinfo
session.Where("name = ? and age < ?", "john", 88).Get(&user4) // even more complex
```

2.Fetch multiple objects

```go
var allusers []Userinfo
err := session.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20

var tenusers []Userinfo
err := session.Where("id > ?", "3").Limit(10).Find(&tenusers) //Get id>3 limit 10  if omit offset the default is 0

var everyone []Userinfo
err := session.Find(&everyone)
```

###***About Map Rules***
1.Struct and struct's fields name should be Pascal style, and the table and column's name default is us
for example: 
The structs Name 'UserInfo' will turn into the table name 'user_info', the same as the keyname.	
If the keyname is 'UserName' will turn into the select colum 'user_name'

2.You have two method to change the rule. One is implement your own Map interface according IMapper, you can find the interface in mapper.go and set it to engine.Mapper

another is use field tag, field tag support the below keywords:
[name]                  column name
pk                      the field is a primary key
int(11)/varchar(50)     column type
autoincr                auto incrment
[not ]null              if column can be null value
unique                  unique
-                       this field is not map as a table column

## LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
