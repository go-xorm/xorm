# xorm

[English](https://github.com/lunny/xorm/blob/master/README.md)

xorm是一个简单而强大的Go语言ORM库. 通过它可以使数据库操作非常简便。

[![Build Status](https://drone.io/github.com/lunny/xorm/status.png)](https://drone.io/github.com/lunny/xorm/latest)  [![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/lunny/xorm)

## 特性

* 支持Struct和数据库表之间的灵活映射，并支持自动同步

* 事务支持

* 同时支持原始SQL语句和ORM操作的混合执行

* 使用连写来简化调用

* 支持使用Id, In, Where, Limit, Join, Having, Table, Sql, Cols等函数和结构体等方式作为条件

* 支持级联加载Struct 

* 支持缓存

* 支持根据数据库自动生成xorm的结构体

## 驱动支持

目前支持的Go数据库驱动如下：

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* Postgres: [github.com/bylevel/pq](https://github.com/bylevel/pq)

## 更新日志
* **v0.2.2** : Postgres驱动新增了对lib/pq的支持；新增了逐条遍历方法Iterate；新增了SetMaxConns(go1.2+)支持，修复了bug若干；
* **v0.2.1** : 新增数据库反转工具，当前支持go和c++代码的生成，详见 [Xorm Tool README](https://github.com/lunny/xorm/blob/master/xorm/README.md); 修复了一些bug.
* **v0.2.0** : 新增 [缓存](https://github.com/lunny/xorm/blob/master/docs/QuickStart.md#120)支持，查询速度提升3-5倍； 新增数据库表和Struct同名的映射方式； 新增Sync同步表结构；

[更多更新日志...](https://github.com/lunny/xorm/blob/master/docs/ChangelogCN.md)

## 安装

	go get github.com/lunny/xorm

## 文档

* [快速开始](https://github.com/lunny/xorm/blob/master/docs/QuickStart.md)

* [GoWalker代码文档](http://gowalker.org/github.com/lunny/xorm)

* [Godoc代码文档](http://godoc.org/github.com/lunny/xorm)


## 案例

* [Gowalker](http://gowalker.org) - [github.com/Unknwon/gowalker](http://github.com/Unknwon/gowalker)

* [Sudo China](http://sudochina.com) - [github.com/insionng/toropress](http://github.com/insionng/toropress)

* [Godaily](http://godaily.org) - [github.com/govc/godaily](http://github.com/govc/godaily)

* [Very Hour](http://veryhour.com/)

## 讨论

请加入QQ群：280360085 进行讨论。


## LICENSE

BSD License
[http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
