# xorm

[English](https://github.com/lunny/xorm/blob/master/README.md)

xorm是一个简单而强大的Go语言ORM库. 通过它可以使数据库操作非常简便。

[![Build Status](https://drone.io/github.com/lunny/xorm/status.png)](https://drone.io/github.com/lunny/xorm/latest)  [![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/lunny/xorm)

## 讨论

请加入QQ群：280360085 进行讨论。

## 驱动支持

目前支持的Go数据库驱动如下：

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/bylevel/pq](https://github.com/bylevel/pq)

## 更新日志

* **v0.2.1** : 新增数据库反转工具，当前支持go和c++代码的生成，详见 [Xorm Tool README](https://github.com/lunny/xorm/blob/master/xorm/README.md); 修复了一些bug.
* **v0.2.0** : 新增 [缓存](https://github.com/lunny/xorm/blob/master/docs/QuickStart.md#120)支持，查询速度提升3-5倍； 新增数据库表和Struct同名的映射方式； 新增Sync同步表结构；
* **v0.1.9** : 新增 postgres 和 mymysql 驱动支持; 在Postgres中支持原始SQL语句中使用 ` 和 ? 符号; 新增Cols, StoreEngine, Charset 函数；SQL语句打印支持io.Writer接口，默认打印到控制台；新增更多的字段类型支持，详见 [映射规则](https://github.com/lunny/xorm/blob/master/docs/QuickStart.md#21)；删除废弃的MakeSession和Create函数。
* **v0.1.8** : 新增联合index，联合unique支持，请查看 [映射规则](https://github.com/lunny/xorm/blob/master/docs/QuickStart.md#21)。
* **v0.1.7** : 新增IConnectPool接口以及NoneConnectPool, SysConnectPool, SimpleConnectPool三种实现，可以选择不使用连接池，使用系统连接池和使用自带连接池三种实现，默认为SysConnectPool，即系统自带的连接池。同时支持自定义连接池。Engine新增Close方法，在系统退出时应调用此方法。
* **v0.1.6** : 新增Conversion，支持自定义类型到数据库类型的转换；新增查询结构体自动检测匿名成员支持；新增单向映射支持；
* **v0.1.5** : 新增对多线程的支持；新增Sql()函数；支持任意sql语句的struct查询；Get函数返回值变动；MakeSession和Create函数被NewSession和NewEngine函数替代；
* **v0.1.4** : Get函数和Find函数新增简单的级联载入功能；对更多的数据库类型支持。
* **v0.1.3** : Find函数现在支持传入Slice或者Map，当传入Map时，key为id；新增Table函数以为多表和临时表进行支持。
* **v0.1.2** : Insert函数支持混合struct和slice指针传入，并根据数据库类型自动批量插入，同时自动添加事务
* **v0.1.1** : 添加 Id, In 函数，改善 README 文档
* **v0.1.0** : 初始化工程

## 特性

* 支持Struct和数据库表之间的灵活映射，并支持自动同步

* 事务支持

* 同时支持原始SQL语句和ORM操作的混合执行

* 使用连写来简化调用

* 支持使用Id, In, Where, Limit, Join, Having, Table, Sql, Cols等函数和结构体等方式作为条件

* 支持级联加载Struct 

* 支持缓存

* 支持根据数据库自动生成xorm的结构体



## 安装

	go get github.com/lunny/xorm

## 文档

* [快速开始](https://github.com/lunny/xorm/blob/master/docs/QuickStart.md)

* [GoWalker代码文档](http://gowalker.org/github.com/lunny/xorm)

## 案例

* [Sudo China](http://sudochina.com) - [github.com/insionng/toropress](http://github.com/insionng/toropress)

* [Godaily](http://godaily.org) - [github.com/govc/godaily](http://github.com/govc/godaily)

* [Very Hour](http://veryhour.com/)


## FAQ

1.问：xorm的tag和json的tag如何同时起作用？
  
答案：使用空格分开

```Go
type User struct {
    Name string `json:"name" xorm:"name"`
}
```
2.问：xorm有几种命名映射规则？
答案：目前支持SnakeMapper和SameMapper两种。SnakeMapper支持结构体和成员以驼峰式命名而数据库表和字段以下划线连接命名；SameMapper支持结构体和数据库的命名保持一致的映射。

## LICENSE

BSD License
[http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
