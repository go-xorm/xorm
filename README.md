# xorm

[中文](https://github.com/lunny/xorm/blob/master/README_CN.md)

Xorm is a simple and powerful ORM for Go. It makes dabatabse operating simple. 

[![Build Status](https://drone.io/github.com/lunny/xorm/status.png)](https://drone.io/github.com/lunny/xorm/latest)  [![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/lunny/xorm) [![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/lunny/xorm/trend.png)](https://bitdeli.com/free "Bitdeli Badge")

## Discuss

Message me on [G+](https://plus.google.com/u/0/106406879480103142585)

## Drivers Support

Drivers for Go's sql package which currently support database/sql includes:

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/bylevel/pq](https://github.com/bylevel/pq)


## Changelog

* **v0.2.1** : Added database reverse tool, now support generate go & c++ codes, see [Xorm Tool README](https://github.com/lunny/xorm/blob/master/xorm/README.md); some bug fixed.
* **v0.2.0** : Added Cache supported, select is speeder up 3~5x; Added SameMapper for same name between struct and table; Added Sync method for auto added tables, columns, indexes;
* **v0.1.9** : Added postgres and mymysql supported; Added ` and ? supported on Raw SQL even if postgres; Added Cols, StoreEngine, Charset function, Added many column data type supported, please see [Mapping Rules](#mapping).
* **v0.1.8** : Added union index and union unique supported, please see [Mapping Rules](#mapping).
* **v0.1.7** : Added IConnectPool interface and NoneConnectPool, SysConnectPool, SimpleConnectPool the three implements. You can choose one of them and the default is SysConnectPool. You can customrize your own connection pool. struct Engine added Close method, It should be invoked before system exit.
* **v0.1.6** : Added conversion interface support; added struct derive support; added single mapping support
* **v0.1.5** : Added multi threads support; added Sql() function for struct query; Get function changed return inteface; MakeSession and Create are instead with NewSession and NewEngine.
* **v0.1.4** : Added simple cascade load support; added more data type supports.
* **v0.1.3** : Find function now supports both slice and map; Add Table function for multi tables and temperory tables support
* **v0.1.2** : Insert function now supports both struct and slice pointer parameters, batch inserting and auto transaction
* **v0.1.1** : Add Id, In functions and improved README
* **v0.1.0** : Inital release.

## Features

* Struct<->Table Mapping Supports, both name mapping and filed tags mapping

* Database Transaction Support

* Both ORM and SQL Operation Support

* Simply usage

* Support Id, In, Where, Limit, Join, Having, Sql functions and sturct as query conditions

* Support simple cascade load just like Hibernate for Java

* Code generator support, See [Xorm Tool README](https://github.com/lunny/xorm/blob/master/xorm/README.md)


## Installing xorm

	go get github.com/lunny/xorm

## Documents 

[Quick Start](https://github.com/lunny/xorm/blob/master/docs/QuickStartEn.md)

[GoDoc](http://godoc.org/github.com/lunny/xorm)

[GoWalker](http://gowalker.org/github.com/lunny/xorm)

## Cases

* [Sudo China](http://sudochina.com) - [github.com/insionng/toropress](http://github.com/insionng/toropress)

* [Godaily](http://godaily.org) - [github.com/govc/godaily](http://github.com/govc/godaily)

* [Very Hour](http://veryhour.com/)

## FAQ 

1.How the xorm tag use both with json?
  
  Use space.

```Go
type User struct {
    Name string `json:"name" xorm:"name"`
}
```

## LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
