[中文](https://github.com/lunny/xorm/blob/master/README_CN.md)

Xorm is a simple and powerful ORM for Go.

[![Build Status](https://drone.io/github.com/lunny/xorm/status.png)](https://drone.io/github.com/lunny/xorm/latest)  [![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/lunny/xorm) [![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/lunny/xorm/trend.png)](https://bitdeli.com/free "Bitdeli Badge")

# Features

* Struct <-> Table Mapping Support

* Chainable APIs
 
* Transaction Support

* Both ORM and raw SQL operation Support

* Sync database sechmea Support

* Query Cache speed up

* Database Reverse support, See [Xorm Tool README](https://github.com/lunny/xorm/blob/master/xorm/README.md)

* Simple cascade loading support

* Optimistic Locking support


# Drivers Support

Drivers for Go's sql package which currently support database/sql includes:

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* MsSql: [github.com/lunny/godbc](https://github.com/lunny/godbc)

# Changelog

* **v0.3.1** 

	Features:
	* Support MSSQL DB via ODBC driver ([github.com/lunny/godbc](https://github.com/lunny/godbc));
	* Composite Key, using multiple pk xorm tag 
	* Added Row() API as alternative to Iterate() API for traversing result set, provide similar usages to sql.Rows type
	* ORM struct allowed declaration of pointer builtin type as members to allow null DB fields 
	* Before and After Event processors

	Improvements:
	* Allowed int/int32/int64/uint/uint32/uint64/string as Primary Key type
	* Performance improvement for Get()/Find()/Iterate()

[More changelogs ...](https://github.com/lunny/xorm/blob/master/docs/Changelog.md)

# Installation

If you have [gopm](https://github.com/gpmgo/gopm) installed, 

	gopm get github.com/lunny/xorm
	
Or

	go get github.com/lunny/xorm

# Documents

* [GoDoc](http://godoc.org/github.com/lunny/xorm)

* [GoWalker](http://gowalker.org/github.com/lunny/xorm)

* [Quick Start](https://github.com/lunny/xorm/blob/master/docs/QuickStartEn.md)

# Cases

* [Gowalker](http://gowalker.org) - [github.com/Unknwon/gowalker](http://github.com/Unknwon/gowalker)

* [Sudo China](http://sudochina.com) - [github.com/insionng/toropress](http://github.com/insionng/toropress)

* [Godaily](http://godaily.org) - [github.com/govc/godaily](http://github.com/govc/godaily)

* [Very Hour](http://veryhour.com/)

# Todo

[Todo List](https://trello.com/b/IHsuAnhk/xorm)

# Discuss

Please visit [Xorm on Google Groups](https://groups.google.com/forum/#!forum/xorm)

# Contributors

If you want to pull request, please see [CONTRIBUTING](https://github.com/lunny/xorm/blob/master/CONTRIBUTING.md)

* [Lunny](https://github.com/lunny)
* [Nashtsai](https://github.com/nashtsai)

# LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
