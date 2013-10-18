# xorm tools


xorm tools is a set of  tools for database operation. 

## Install

`go get github.com/lunny/xorm/xorm`

and you should install the depends below:

* github.com/lunny/xorm

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/bylevel/pq](https://github.com/bylevel/pq)


## Reverse

After you installed the tool, you can type 

`xorm help reverse`

to get help

example:

`xorm reverse sqite3 test.db templates/goxorm`

will generated go files in `./model` directory

## Template and Config

Now, xorm tool supports go and c++ two languages and have go, goxorm, c++ three of default templates. In template directory, we can put a config file to control how to generating.

````
lang=go
genJson=1
```

lang must be go or c++ now.
genJson can be 1 or 0, if 1 then the struct will have json tag.

## LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
