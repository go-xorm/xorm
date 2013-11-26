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

# Drivers Support

Drivers for Go's sql package which currently support database/sql includes:

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)


# Changelog

* **v0.2.2** : Postgres drivers now support lib/pq; Added method Iterate for record by record to handler；Added SetMaxConns(go1.2+) support; some bugs fixed.
* **v0.2.1** : Added database reverse tool, now support generate go & c++ codes, see [Xorm Tool README](https://github.com/lunny/xorm/blob/master/xorm/README.md); some bug fixed.
* **v0.2.0** : Added Cache supported, select is speeder up 3~5x; Added SameMapper for same name between struct and table; Added Sync method for auto added tables, columns, indexes;

[More changelogs ...](https://github.com/lunny/xorm/blob/master/docs/Changelog.md)


# Installation

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

# Discuss

Please visit [Xorm on Google Groups](https://groups.google.com/forum/#!forum/xorm)

# LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
