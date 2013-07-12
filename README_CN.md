# xorm

[English](https://github.com/lunny/xorm/blob/master/README.md)

xorm是一个简单而强大的Go语言ORM库. 通过它可以使数据库操作非常简便。

[![Build Status](https://drone.io/github.com/lunny/xorm/status.png)](https://drone.io/github.com/lunny/xorm/latest)

## 驱动支持

目前支持的Go数据库驱动如下：

* Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

## 更新日志

* **v0.1.7** : 新增IConnectPool接口以及NoneConnectPool, SysConnectPool, SimpleConnectPool三种实现，可以选择不使用连接池，使用系统连接池和使用自带连接池三种实现，默认为SysConnectPool，即系统自带的连接池。同时支持自定义连接池。Engine新增Close方法，在系统退出时应调用此方法。
* **v0.1.6** : 新增Conversion，支持自定义类型到数据库类型的转换；新增查询结构体自动检测匿名成员支持；新增单向映射支持；
* **v0.1.5** : 新增对多线程的支持；新增Sql()函数；支持任意sql语句的struct查询；Get函数返回值变动；MakeSession和Create函数被NewSession和NewEngine函数替代；
* **v0.1.4** : Get函数和Find函数新增简单的级联载入功能；对更多的数据库类型支持。
* **v0.1.3** : Find函数现在支持传入Slice或者Map，当传入Map时，key为id；新增Table函数以为多表和临时表进行支持。
* **v0.1.2** : Insert函数支持混合struct和slice指针传入，并根据数据库类型自动批量插入，同时自动添加事务
* **v0.1.1** : 添加 Id, In 函数，改善 README 文档
* **v0.1.0** : 初始化工程

## 特性

* 支持Struct和数据库表之间的映射，映射方式支持命名约定和Tag两种方式，映射支持继承

* 事务支持

* 同时支持原始SQL语句和ORM操作的混合执行

* 使用连写来简化调用

* 支持使用Id, In, Where, Limit, Join, Having, Table, Sql等函数和结构体等方式作为条件

* 支持数据库连接池

* 支持级联加载struct


## 安装

	go get github.com/lunny/xorm

## 快速开始

1.创建数据库引擎，这个函数的参数和sql.Open相同，但不会立即创建连接 (例如: mysql),注意：Create方法将在后续版本中被弃用，请使用NewEngine方法。

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

1.1.默认将不会显示自动生成的SQL语句，如果要显示，则需要设置

```Go
engine.ShowSQL = true
```

1.2.如果要更换连接池实现，可使用SetPool方法
```Go
err = engine.SetPool(NewSimpleConnectPool())
```

2.所有的ORM操作都针对一个或多个结构体，一个结构体对应一张表，定义一个结构体如下：

```Go
type User struct {
    Id int
    Name string
    Age int    `xorm:"-"`
}
```

2.1 详细映射规则，请查看[映射规则](#mapping)

3.在程序初始化时，可能会需要创建表

```Go
err := engine.CreateTables(&User{})
```
	
4.然后，可以将一个结构体作为一条记录插入到表中， 如果插入成功，将会返回id，同时User对象的Id也会被自动赋值。  

```Go
user := &User{Name:"lunny"}
id, err := engine.Insert(user)
fmt.Println(user.Id)
```

或者执行更新操作：

```Go
user := User{Name:"xlw"}
rows, err := engine.Update(&user, &User{Id:1})
// rows, err := engine.Where("id = ?", 1).Update(&user)
// or rows, err := engine.Id(1).Update(&user)
```

5.获取单个对象，可以用Get方法：

```Go
var user = User{Id:27}
has, err := engine.Get(&user)
// or has, err := engine.Id(27).Get(&user)
var user = User{Name:"xlw"}
has, err := engine.Get(&user)
```
	
6.获取多个对象到一个Slice或一个Map对象中，可以用Find方法，如果传入的是map，则key中存放的是id：

```Go
var everyone []Userinfo
err := engine.Find(&everyone)

users := make(map[int64]Userinfo)
err := engine.Find(&users)
```

6.1 你也可以使用Where和Limit方法设定条件和查询数量

```Go
var allusers []Userinfo
err := engine.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20
```

6.2 用一个结构体作为查询条件也是允许的

```Go
var tenusers []Userinfo
err := engine.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10  offset 0
```

6.3 也可以调用In函数

```Go
var tenusers []Userinfo
err := engine.In("id", 1, 3, 5).Find(&tenusers) //Get All id in (1, 3, 5)
```

7.Delete方法

```Go
num, err := engine.Delete(&User{Id:1})
// or num, err := engine.Id(1).Delete(&User{})
```

8.Count方法

```Go
total, err := engine.Count(&User{Name:"xlw"})
```

## 直接执行SQL语句

当然，如果你想直接使用SQL语句进行操作，也是允许的。

如果执行Select，请用Query()

```Go
sql := "select * from userinfo"
results, err := engine.Query(sql)
```

如果执行Insert， Update， Delete 等操作，请用Exec()

```Go
sql = "update userinfo set username=? where id=?"
res, err := engine.Exec(sql, "xiaolun", 1) 
```

## 高级用法

<a name="mapping" id="mapping"></a>
更高级的用法，我们必须要使用session对象，session对象在创建时会立刻创建一个数据库连接。注意：MakeSession方法将会在后续版本中移除，请使用NewSession方法替代。

```Go
session := engine.NewSession()
defer session.Close()
```

1.session对象同样也可以查询

```Go
var user Userinfo
session.Where("id=?", 27).Get(&user)

var userJohn Userinfo
session.Where("name = ?", "john").Get(&userJohn) // more complex query

var userOldJohn Userinfo
session.Where("name = ? and age > ?", "john", 88).Get(&userOldJohn) // even more complex
```

2.获取多个对象

```Go
var allusers []Userinfo
err := session.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20

var tenusers []Userinfo
err := session.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10  if omit offset the default is 0

var everyone []Userinfo
err := session.Find(&everyone)
```
	
3.事务处理

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

4.混合型事务，这个事务中，既有直接的SQL语句，又有ORM方法：

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

5.匿名结构体继承：

请查看Examples中的derive.go文件。

## 映射规则

1.Struct 和 Struct 的field名字应该为Pascal式命名，默认的映射规则将转换成用下划线连接的命名规则，这个映射是自动进行的，当然，你可以通过修改Engine的成员Mapper来改变它。

例如：

结构体的名字UserInfo将会自动对应数据库中的名为user_info的表。	
UserInfo中的成员UserName将会自动对应名为user_name的字段。

2.当然你也可以改变这个规则，这有两种方法。一是实现你自己的IMapper，你可以在mapper.go中查看到这个接口。然后设置到 engine.Mapper。

另外一种方法就通过Field Tag来进行改变，关于Field Tag请参考Go的语言文档，如下列出了Tag中可用的关键字及其对应的意义：

<table>
    <tr>
        <td>name</td><td>当前field对应的字段的名称，可选</td>
    </tr>
    <tr>
        <td>pk</td><td>是否是Primary Key</td>
    </tr>
    <tr>
        <td>int(11)/varchar(50)/text/date/datetime/blob/decimal(26,2)</td><td>字段类型</td>
    </tr>
    <tr>
        <td>autoincr</td><td>是否是自增</td>
    </tr>
    <tr>
        <td>[not ]null</td><td>是否可以为空</td>
    </tr>
    <tr>
        <td>unique</td><td>是否是唯一</td>
    </tr>
    <tr>
    	<td>extends</td><td>应用于一个匿名结构体之上，表示此匿名结构体的成员也映射到数据库中</td>
    </tr>
    <tr>
        <td>-</td><td>这个Field将不进行字段映射</td>
    </tr>
</table>
例如：

```Go
type Userinfo struct {
	Uid        int `xorm:"id pk not null autoincr"`
	Username   string
	Departname string
	Alias      string `xorm:"-"`
	Created    time.Time
}
```
3.对于自定义的表名，可以用Table函数进行操作，比如：

```Go
// batch create tables
for i := 0; i < 10; i++ {
	engine.Table(fmt.Sprintf("user_%v", i)).CreateTable(&Userinfo{}) 
}

// insert into table according id
user := Userinfo{Uid: 25, Username:"sslfs"}
engine.Table(fmt.Sprintf("user_%v", user.Uid % 10)).Insert(&user)
```

## 文档

请访问 [GoWalker](http://gowalker.org/github.com/lunny/xorm) 查看详细文档

## FAQ

1.问：xorm的tag和json的tag如何同时起作用？
  
答案：使用空格分开

```Go
type User struct {
    Name string `json:"name" xorm:"name"`
}
```

2.问：xorm是否带有连接池

答案：database/sql默认就有连接池，因此xorm本身没有内建连接池，在使用过程中会自动调用database/sql的实现。

## LICENSE

BSD License
[http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
