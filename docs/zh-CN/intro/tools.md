---
root: false
name: xorm 工具
sort: 3
---

# xorm 工具

xorm 是一组数据库操作命令行工具。 

## 二进制安装

如果你安装了 [got](https://github.com/gobuild/got)，你可以输入如下命令安装：

```
got go-xorm/cmd/xorm
```

或者你可以从 [gobuild](http://gobuild.io/download/github.com/lunny/got) 下载后解压到可执行路径。

## 源码安装

`go get github.com/go-xorm/cmd/xorm`

同时你需要安装如下依赖:

* github.com/go-xorm/xorm

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) 

** 对于sqlite3的支持，你需要自己进行编译 `go build -tags sqlite3` 因为sqlite3需要cgo的支持。

## 命令列表

有如下可用的命令：

* **reverse**     反转一个数据库结构，生成代码
* **shell**       通用的数据库操作客户端，可对数据库结构和数据操作
* **dump**        Dump数据库中所有结构和数据到标准输出
* **source**      从标注输入中执行SQL文件
* **driver**      列出所有支持的数据库驱动

## reverse

Reverse command is a tool to convert your database struct to all kinds languages of structs or classes. After you installed the tool, you can type 

`xorm help reverse`

to get help

example:

sqlite:
`xorm reverse sqite3 test.db templates/goxorm`

mysql:
`xorm reverse mysql root:@/xorm_test?charset=utf8 templates/goxorm`

mymysql:
`xorm reverse mymysql xorm_test2/root/ templates/goxorm`

postgres:
`xorm reverse postgres "dbname=xorm_test sslmode=disable" templates/goxorm`

will generated go files in `./model` directory

### Template and Config

Now, xorm tool supports go and c++ two languages and have go, goxorm, c++ three of default templates. In template directory, we can put a config file to control how to generating.

````
lang=go
genJson=1
```

lang must be go or c++ now.
genJson can be 1 or 0, if 1 then the struct will have json tag.

## Shell

Shell command provides a tool to operate database. For example, you can create table, alter table, insert data, delete data and etc.

`xorm shell sqlite3 test.db` will connect to the sqlite3 database and you can type `help` to list all the shell commands.

## Dump

Dump command provides a tool to dump all database structs and data as SQL to your standard output.

`xorm dump sqlite3 test.db` could dump sqlite3 database test.db to standard output. If you want to save to file, just
type `xorm dump sqlite3 test.db > test.sql`.

## Source

`xorm source sqlite3 test.db < test.sql` will execute sql file on the test.db.

## Driver

List all supported drivers since default build will not include sqlite3.

## LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)
