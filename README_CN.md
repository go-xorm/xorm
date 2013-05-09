# xorm
===========

[English](README.md)

xorm是一个Go语言的ORM库. 通过它可以简化对数据库的操作。

目前仅支持Mysql和SQLite，当然我们的目标是支持PostgreSQL/DB2/MS ADODB/ODBC/Oracle等等。

但是，目前的版本还不可用于正式版本。

目前支持的Go数据库驱动如下：

Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

## 安装

	go get github.com/lunny/xorm

## 快速开始

1.创建数据库引擎，这个函数的参数和sql.OpenDB相同，但不会立即创建连接 (例如: mysql)

	engine := xorm.Create("mysql", "root:123@/test?charset=utf8")

or

	engine = xorm.Create("sqlite3", "./test.db")


2.定义你的Struct


	type User struct {
    	Id int
    	Name string
    	Age int    `xorm:"-"`
	}


对于简单的任务，可以只用engine一个对象就可以完成操作。
首先，需要创建一个数据库，然后使用以下语句创建一个Struct对应的表。


	err := engine.CreateTables(&User{})

	
然后，可以将一个结构体作为一条记录插入到表中。
  

	id, err := engine.Insert(&User{Name:"lunny"})


或者执行更新操作：


	user := User{Name:"xlw"}
	rows, err := engine.Update(&user, &User{Id:1})
	// rows, err := engine.Where("id = ?", 1).Update(&user)


3.获取单个对象，可以用Get方法：


	var user = User{Id:27}
	err := engine.Get(&user)

	var user = User{Name:"xlw"}
	err := engine.Get(&user)
	
4.获取多个对象，可以用Find方法：

	var allusers []Userinfo
	err := engine.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20

	var tenusers []Userinfo
	err := engine.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10  if omit offset the default is 0

	var everyone []Userinfo
	err := engine.Find(&everyone)

5.另外还有Delete和Count方法：

	err := engine.Delete(&User{Id:1})
	
	total, err := engine.Count(&User{Name:"xlw"})

##Origin Use
当然，如果你想直接使用SQL语句进行操作，也是允许的。

	sql := "select * from userinfo"
	results, err := engine.Query(sql)
	
	sql = "update userinfo set username=? where id=?"
	res, err := engine.Exec(sql, "xiaolun", 1) 

##Deep Use
更高级的用法，我们必须要使用session对象，session对象在创建时会创建一个数据库连接。


	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
    	return
	}


1.session对象同样也可以查询

	var user Userinfo
	session.Where("id=?", 27).Get(&user)

	var user2 Userinfo
	session.Where("name = ?", "john").Get(&user3) // more complex query

	var user3 Userinfo
	session.Where("name = ? and age < ?", "john", 88).Get(&user4) // even more complex


2.获取多个对象

	var allusers []Userinfo
	err := session.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20

	var tenusers []Userinfo
	err := session.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10  if omit offset the default is 0

	var everyone []Userinfo
	err := session.Find(&everyone)
	
3.事务处理

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

4.混合型事务，这个事务中，既有直接的SQL语句，又有其它方法：

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


##FAQ
1.xorm的tag和json的tag如何同时起作用？
  
  使用空格分开

	type User struct {
    	User string `json:"user" orm:"user_id"`
	}

## LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
