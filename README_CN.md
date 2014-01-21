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

* 支持记录版本（即乐观锁）

## 驱动支持

目前支持的Go数据库驱动和对应的数据库如下：

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* MsSql: [github.com/lunny/godbc](https://github.com/lunny/godbc)

## 更新日志

* **v0.3.1**

	新特性:
	* 支持 MSSQL DB 通过 ODBC 驱动 ([github.com/lunny/godbc](https://github.com/lunny/godbc));
	* 通过多个pk标记支持联合主键; 
	* 新增 Rows() API 用来遍历查询结果，该函数提供了类似sql.Rows的相似用法，可作为 Iterate() API 的可选替代；
	* ORM 结构体现在允许内建类型的指针作为成员，使得数据库为null成为可能；
	* Before 和 After 支持

	改进:
	* 允许 int/int32/int64/uint/uint32/uint64/string 作为主键类型
	* 查询函数 Get()/Find()/Iterate() 在性能上的改进


[更多更新日志...](https://github.com/lunny/xorm/blob/master/docs/ChangelogCN.md)

## 安装

推荐使用 [gopm](https://github.com/gpmgo/gopm) 进行安装： 

	gopm get github.com/lunny/xorm
	
或者您也可以使用go工具进行安装：

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

## Todo

[开发计划](https://trello.com/b/IHsuAnhk/xorm)

## 讨论

请加入QQ群：280360085 进行讨论。

# 贡献者

如果您也想为Xorm贡献您的力量，请查看 [CONTRIBUTING](https://github.com/lunny/xorm/blob/master/CONTRIBUTING.md)

* [Lunny](https://github.com/lunny)
* [Nashtsai](https://github.com/nashtsai)

## LICENSE

BSD License
[http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
