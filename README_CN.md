# xorm
[English](https://github.com/lunny/xorm/blob/master/README.md)

xorm是一个Go语言的ORM库. 通过它可以使数据库操作非常简便。

目前没有正式的项目来使用此库，如果有，我们将会把它列出来。

## 驱动支持
目前支持的Go数据库驱动如下：

* Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

## 更新日志

* **v0.1.2** : Insert函数支持混合struct和slice指针传入，并根据数据库类型自动批量插入，同时自动添加事务
* **v0.1.1** : 添加 Id, In 函数，改善 README 文档
* **v0.1.0** : 初始化工程

## 特性
* 支持Struct和数据库表之间的映射，映射方式支持命名约定和Tag两种方式

* 事务支持

* 同时支持原始SQL语句和ORM操作的混合执行

* 使用连写来简化调用

* 支持使用Id, In, Where, Limit, Join, Having等函数和结构体等方式作为条件

## 安装

	go get github.com/lunny/xorm

## 快速开始

1.创建数据库引擎，这个函数的参数和sql.Open相同，但不会立即创建连接 (例如: mysql)

```Go
import (
	_ "github.com/Go-SQL-Driver/MySQL"
	"github.com/lunny/xorm"
)
engine := xorm.Create("mysql", "root:123@/test?charset=utf8")
```

or

```Go
import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/lunny/xorm"
	)
engine = xorm.Create("sqlite3", "./test.db")
```

1.1.默认将不会显示自动生成的SQL语句，如果要显示，则需要设置

```Go
engine.ShowSQL = true
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
	
4.然后，可以将一个结构体作为一条记录插入到表中。  

```Go
id, err := engine.Insert(&User{Name:"lunny"})
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
err := engine.Get(&user)
// or err := engine.Id(27).Get(&user)
var user = User{Name:"xlw"}
err := engine.Get(&user)
```
	
6.获取多个对象，可以用Find方法：

```Go
var everyone []Userinfo
err := engine.Find(&everyone)
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
err := engine.Delete(&User{Id:1})
// or err := engine.Id(1).Delete(&User{})
```

8.Count方法

```Go
total, err := engine.Count(&User{Name:"xlw"})
```

##直接执行SQL语句
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

##高级用法
<a name="mapping" id="mapping"></a>
更高级的用法，我们必须要使用session对象，session对象在创建时会立刻创建一个数据库连接。

```Go
session, err := engine.MakeSession()
defer session.Close()
if err != nil {
    return
}
```

1.session对象同样也可以查询

```Go
var user Userinfo
session.Where("id=?", 27).Get(&user)

var user2 Userinfo
session.Where("name = ?", "john").Get(&user3) // more complex query

var user3 Userinfo
session.Where("name = ? and age < ?", "john", 88).Get(&user4) // even more complex
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
```

4.混合型事务，这个事务中，既有直接的SQL语句，又有ORM方法：

```Go
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
```

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
        <td>int(11)/varchar(50)</td><td>字段类型</td>
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

##文档
请访问 [GoWalker](http://gowalker.org/github.com/lunny/xorm) 查看详细文档

##FAQ
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
