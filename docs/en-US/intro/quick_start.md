---
name: Quick Start
sort: 2
---

Quick Start
=====

* [1.Create ORM Engine](#10)
* [2.Define a struct](#20)
	* [2.1.Name mapping rule](#21)
	* [2.2.Use Table or Tag to change table or column name](#22)
	* [2.3.Column define](#23)
* [3. database schema operation](#30)
	* [3.1.Retrieve database schema infomation](#31)
	* [3.2.Table Operation](#32)
	* [3.3.Create indexes and uniques](#33)
	* [3.4.Sync database schema](#34)
* [4.Insert records](#40)
* [5.Query and Count records](#60)
	* [5.1.Query condition methods](#61)
	* [5.2.Temporory methods](#62)
	* [5.3.Get](#63)
	* [5.4.Find](#64)
	* [5.5.Iterate](#65)
	* [5.6.Count](#66)
* [6.Update records](#70)
* [6.1.Optimistic Locking](#71)
* [7.Delete records](#80)
* [8.Execute SQL command](#90)
* [9.Execute SQL query](#100)
* [10.Transaction](#110)
* [11.Cache](#120)
* [13.Examples](#140)

<a name="10" id="10"></a>
## 1.Create ORM Engine 

When using xorm, you can create multiple orm engines, an engine means a databse. So you can：

```Go
import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)
engine, err := xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
defer engine.Close()
```

or

```Go
import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/go-xorm/xorm"
	)
engine, err = xorm.NewEngine("sqlite3", "./test.db")
defer engine.Close()
```

You can create many engines for different databases.Generally, you just need create only one engine. Engine supports run on go routines.

xorm supports four drivers now:

* Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* MsSql: [github.com/lunny/godbc](https://githubcom/lunny/godbc)

NewEngine's parameters are the same as `sql.Open`. So you should read the drivers' document for parameters' usage.

After engine created, you can do some settings.

1.Logs

* `engine.ShowSQL = true`, Show SQL statement on standard output;
* `engine.ShowDebug = true`, Show debug infomation on standard output;
* `engine.ShowError = true`, Show error infomation on standard output;
* `engine.ShowWarn = true`, Show warnning information on standard output;

2.If want to record infomation with another method: use `engine.Logger` as `io.Writer`:

```Go
f, err := os.Create("sql.log")
	if err != nil {
		println(err.Error())
		return
	}
engine.Logger = xorm.NewSimpleLogger(f)
```

3.Engine provide DB connection pool settings.

* Use `engine.SetMaxIdleConns()` to set idle connections.
* Use `engine.SetMaxOpenConns()` to set Max connections. This methods support only Go 1.2+.

<a name="20" id="20"></a>
## 2.Define struct

xorm maps a struct to a database table, the rule is below.

<a name="21" id="21"></a>
### 2.1.name mapping rule

use xorm.IMapper interface to implement. There are two IMapper implemented: `SnakeMapper` and `SameMapper`. SnakeMapper means struct name is word by word and table name or column name as 下划线. SameMapper means same name between struct and table.

SnakeMapper is the default.

```Go
engine.SetMapper(SameMapper{})
```

And you should notice:

* If you want to use other mapping rule, implement IMapper
* Tables's mapping rule could be different from Columns':

```Go
engine.SetTableMapper(SameMapper{})
engine.SetColumnMapper(SnakeMapper{})
```

<a name="22" id="22"></a>
### 2.2.Prefix mapping, Suffix Mapping and Cache Mapping

* `engine.NewPrefixMapper(SnakeMapper{}, "prefix")` can add prefix string when naming based on SnakeMapper or SameMapper, or you custom Mapper.
* `engine.NewPrefixMapper(SnakeMapper{}, "suffix")` can add suffix string when naming based on SnakeMapper or SameMapper, or you custom Mapper.
* `engine.NewCacheMapper(SnakeMapper{})` add naming Mapper for memory cache.

Of course, you can implement IMapper to make custom naming strategy.

<a name="22" id="22"></a>
### 2.3.Tag mapping

It's idealized of using IMapper for all naming. But if table or column is not in rule, we need new method to archive.

* If struct or pointer of struct has `TableName() string` method, the return value will be the struct's table name.

* `engine.Table()` can change the database table name for struct. The struct tag `xorm:"'table_name'"` can set column name for struct field. Use a pair of single quotes to prevent confusion for column's definition in struct tag. If not in confusion, ignore single quotes.

<a name="23" id="23"></a>
### 2.4.Column definition

Struct tag defines something for column as basic sql concepts, such as :

```
type User struct {
	Id 	 int64
	Name string  `xorm:"varchar(25) not null unique 'usr_name'"`
}
```

Data types are different in different DBMS. So xorm makes own data types definition to keep compatible. Details is in document [Column Types](https://github.com/go-xorm/xorm/blob/master/docs/COLUMNTYPE.md).

The following table is field mapping rules, the keyword is not case sensitive except column name：

<table>
    <tr>
        <td>name or 'name'</td><td>Column Name, optional</td>
    </tr>
    <tr>
        <td>pk</td><td>If column is Primary Key</td>
    </tr>
    <tr>
        <td>support over 30 kinds of column types, details in [Column Types](https://github.com/go-xorm/xorm/blob/master/docs/COLUMNTYPE.md)</td><td>column type</td>
    </tr>
    <tr>
        <td>autoincr</td><td>If autoincrement column</td>
    </tr>
    <tr>
        <td>[not ]null | notnull</td><td>if column could be blank</td>
    </tr>
    <tr>
        <td>unique/unique(uniquename)</td><td>column is Unique index; if add (uniquename), the column is used for combined unique index with the field that defining same uniquename.</td>
    </tr>
    <tr>
        <td>index/index(indexname)</td><td>column is index. if add (indexname), the column is used for combined index with the field that defining same indexname.</td>
    </tr>
    <tr>
    	<td>extends</td><td>use for anonymous field, map the struct in anonymous field to database</td>
    </tr>
    <tr>
        <td>-</td><td>This field will not be mapping</td>
    </tr>
     <tr>
        <td>-></td><td>only write into database</td>
    </tr>
     <tr>
        <td>&lt;-</td><td>only read from database</td>
    </tr>
     <tr>
        <td>created</td><td>This field will be filled in current time on insert</td>
    </tr>
     <tr>
        <td>updated</td><td>This field will be filled in current time on insert or update</td>
    </tr>
    <tr>
        <td>version</td><td>This field will be filled 1 on insert and autoincrement on update</td>
    </tr>
    <tr>
        <td>default 0 | default 'name'</td><td>column default value</td>
    </tr>
</table>

Some default mapping rules：

- 1. If field is name of `Id` and type of `int64`, xorm makes it as auto increment primary key. If another field, use struct tag `xorm:"pk"`.

- 2. String is corresponding to varchar(255).

- 3. Support custom type as `type MyString string`，slice, map as field type. They are saving as Text column type and json-encode string. Support Blob column type with field type []byte or []uint8.

- 4. You can implement Conversion interface to define your custom mapping rule between field and database data.

```
type Conversion interface {
	FromDB([]byte) error
	ToDB() ([]byte, error)
}
```

- 5. If one struct has a Conversion field, so we need set an implementation to the field before get data from database. We can implement `BeforeSet(name string, cell xorm.Cell)` on struct to do this. For example: [testConversion](https://github.com/go-xorm/tests/blob/master/base.go#L1826)

<a name="30" id="30"></a>
## 3. database meta information

xorm provides methods to getting and setting table schema. For less schema changing production, `engine.Sync()` is enough.

<a name="31" id="31"></a>
## 3.1 retrieve database meta info

* DBMetas()
`engine.DBMetas()` returns all tables schema information.

<a name="31" id="31"></a>
## 3.2.directly table operation

* CreateTables()
`engine.CreateTables(struct)` creates table with struct or struct pointer.
`engine.Charset()` and `engine.StoreEngine()` can change charset or storage engine for **mysql** database.

* IsTableEmpty()
check table is empty or not.

* IsTableExist()
check table is existed or not.

* DropTables()
`engine.DropTables(struct)` drops table and indexes with struct or struct pointer. `engine.DropTables(string)` only drops table except indexes.

<a name="32" id="32"></a>
## 3.3.create indexes and uniques

* CreateIndexes
create indexes with struct.

* CreateUniques
create unique indexes with struct.

<a name="34" id="34"></a>
## 3.4.Synchronize database schema

xorm watches tables and indexes and sync schema:
1) use table name to create or drop table
2) use column name to alter column
3) use the indexes definition in struct field tag to create or drop indexes.

```Go
err := engine.Sync(new(User))
```

<a name="50" id="50"></a>
## 4.Insert data

Inserting records use Insert method. 

* Insert one record
```Go
user := new(User)
user.Name = "myname"
affected, err := engine.Insert(user)
```

After inseted, `user.Id` will be filled with primary key column value.
```Go
fmt.Println(user.Id)
```

* Insert multiple records by Slice on one table
```Go
users := make([]User, 0)
users[0].Name = "name0"
...
affected, err := engine.Insert(&users)
```

* Insert multiple records by Slice of pointer on one table
```Go
users := make([]*User, 0)
users[0] = new(User)
users[0].Name = "name0"
...
affected, err := engine.Insert(&users)
```

* Insert one record on two table.
```Go
user := new(User)
user.Name = "myname"
question := new(Question)
question.Content = "whywhywhwy?"
affected, err := engine.Insert(user, question)
```

* Insert multiple records on multiple tables.
```Go
users := make([]User, 0)
users[0].Name = "name0"
...
questions := make([]Question, 0)
questions[0].Content = "whywhywhwy?"
affected, err := engine.Insert(&users, &questions)
```

* Insert one or multple records on multiple tables.
```Go
user := new(User)
user.Name = "myname"
...
questions := make([]Question, 0)
questions[0].Content = "whywhywhwy?"
affected, err := engine.Insert(user, &questions)
```

Notice: If you want to use transaction on inserting, you should use session.Begin() before calling Insert.

<a name="60" id="60"></a>
## 5. Chainable APIs

<a name="61" id="61"></a>
### 5.1. Chainable APIs for Queries, Execusions and Aggregations

Queries and Aggregations is basically formed by using `Get`, `Find`, `Count` methods, with conjunction of following chainable APIs to form conditions, grouping and ordering:
查询和统计主要使用`Get`, `Find`, `Count`三个方法。在进行查询时可以使用多个方法来形成查询条件，条件函数如下：

* Id([]interface{})
Primary Key lookup

* Where(string, …interface{})
As SQL conditional WHERE clause

* And(string, …interface{})
Conditional AND 

* Or(string, …interface{})
Conditional OR

* Sql(string, …interface{})
Custom SQL query

* Asc(…string)
Ascending ordering on 1 or more fields

* Desc(…string)
Descending ordering on 1 or more fields

* OrderBy(string)
As SQL ORDER BY

* In(string, …interface{})
As SQL Conditional IN

* Cols(…string)
Explicity specify query or update columns. e.g.,:
```Go
engine.Cols("age", "name").Find(&users)
// SELECT age, name FROM user
engine.Cols("age", "name").Update(&user)
// UPDATE user SET age=? AND name=?
```

* Omit(...string)
Inverse function to Cols, to exclude specify query or update columns. Warning: Don't use with Cols()
```Go
engine.Omit("age").Update(&user)
// UPDATE user SET name = ? AND department = ?
```

* Distinct(…string)
As SQL DISTINCT
```Go
engine.Distinct("age", "department").Find(&users)
// SELECT DISTINCT age, department FROM user
```
Caution: this method will not lookup from caching store


* Table(nameOrStructPtr interface{})
Specify table name, or if struct pointer is passed into the name is extract from struct type name by IMapper conversion policy

* Limit(int, …int)
As SQL LIMIT with optional second param for OFFSET

* Top(int)
As SQL LIMIT

* Join(type, tableName, criteria string)
As SQL JOIN, support
type: either of these values [INNER, LEFT OUTER, CROSS] are supported now
tableName: joining table name
criteria: join criteria

* GroupBy(string)
As SQL GROUP BY

* Having(string)
As SQL HAVING

<a name="62" id="62"></a>
### 5.2. Override default behavior APIs

* NoAutoTime()
No auto timestamp for Created and Updated fields for INSERT and UPDATE

* NoCache()
Disable cache lookup


* UseBool(...string)
xorm's default behavior is fields with 0, "", nil, false, will not be used during query or update, use this method to explicit specify bool type fields for query or update 


* Cascade(bool)
Do cascade lookup for associations

<a name="50" id="50"></a>
### 5.3.Get one record
Fetch a single object by user

```Go
var user = User{Id:27}
has, err := engine.Get(&user)
// or has, err := engine.Id(27).Get(&user)

var user = User{Name:"xlw"}
has, err := engine.Get(&user)
```

<a name="60" id="60"></a>
### 5.4.Find
Fetch multipe objects into a slice or a map, use Find：

```Go
var everyone []Userinfo
err := engine.Find(&everyone)

users := make(map[int64]Userinfo)
err := engine.Find(&users)
```

* also you can use Where, Limit

```Go
var allusers []Userinfo
err := engine.Where("id > ?", "3").Limit(10,20).Find(&allusers) //Get id>3 limit 10 offset 20
```

* or you can use a struct query

```Go
var tenusers []Userinfo
err := engine.Limit(10).Find(&tenusers, &Userinfo{Name:"xlw"}) //Get All Name="xlw" limit 10 offset 0
```

* or In function

```Go
var tenusers []Userinfo
err := engine.In("id", 1, 3, 5).Find(&tenusers) //Get All id in (1, 3, 5)
```

* The default will query all columns of a table. Use Cols function if you want to select some columns

```Go
var tenusers []Userinfo
err := engine.Cols("id", "name").Find(&tenusers) //Find only id and name
```

<a name="70" id="70"></a>
### 5.5.Iterate records
Iterate, like find, but handle records one by one

```Go
err := engine.Where("age > ? or name=?)", 30, "xlw").Iterate(new(Userinfo), func(i int, bean interface{})error{
	user := bean.(*Userinfo)
	//do somthing use i and user
})
```

<a name="66" id="66"></a>
### 5.6.Count method usage

An ORM pointer struct is required for Count method in order to determine which table to retrieve from.
```Go
user := new(User)
total, err := engine.Where("id >?", 1).Count(user)
```

<a name="70" id="70"></a>
## 6.Update
    
更新数据使用`Update`方法，Update方法的第一个参数为需要更新的内容，可以为一个结构体指针或者一个Map[string]interface{}类型。当传入的为结构体指针时，只有非空和0的field才会被作为更新的字段。当传入的为Map类型时，key为数据库Column的名字，value为要更新的内容。

```Go
user := new(User)
user.Name = "myname"
affected, err := engine.Id(id).Update(user)
```

这里需要注意，Update会自动从user结构体中提取非0和非nil得值作为需要更新的内容，因此，如果需要更新一个值为0，则此种方法将无法实现，因此有两种选择：

1. 通过添加Cols函数指定需要更新结构体中的哪些值，未指定的将不更新，指定了的即使为0也会更新。
```Go
affected, err := engine.Id(id).Cols("age").Update(&user)
```

2. 通过传入map[string]interface{}来进行更新，但这时需要额外指定更新到哪个表，因为通过map是无法自动检测更新哪个表的。
```Go
affected, err := engine.Table(new(User)).Id(id).Update(map[string]interface{}{"age":0})
```


### 6.1.Optimistic Lock

To enable object optimistic lock, add 'version' tag value:
```Go
type User struct {
	Id int64
	Name string
	Version int `xorm:"version"`
}
```
The version starts with 1 when inserted to DB. For updating make sure originated version value is used for optimistic lock check.

```Go
var user User
engine.Id(1).Get(&user)
// SELECT * FROM user WHERE id = ?
engine.Id(1).Update(&user)
// UPDATE user SET ..., version = version + 1 WHERE id = ? AND version = ?
```


<a name="80" id="80"></a>
## 7.Delete one or more records
Delete one or more records

* delete by id

```Go
err := engine.Id(1).Delete(&User{})
```

* delete by other conditions

```Go
err := engine.Delete(&User{Name:"xlw"})
```

<a name="90" id="90"></a>
## 8.Execute SQL query

Of course, SQL execution is also provided.

If select then use Query

```Go
sql := "select * from userinfo"
results, err := engine.Query(sql)
```

<a name="100" id="100"></a>
## 9.Execute SQL command
If insert, update or delete then use Exec

```Go
sql = "update userinfo set username=? where id=?"
res, err := engine.Exec(sql, "xiaolun", 1) 
```

<a name="110" id="110"></a>
## 10.Transaction

```Go
session := engine.NewSession()
defer session.Close()

// add Begin() before any action
err := session.Begin()	
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

<a name="120" id="120"></a>
## 11.Built-in LRU memory cache provider

1. Global Cache
Xorm implements cache support. Defaultly, it's disabled. If enable it, use below code.

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.SetDefaultCacher(cacher)
```

If disable some tables' cache, then:

```Go
engine.MapCacher(&user, nil)
```

2. Table's Cache
If only some tables need cache, then:

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.MapCacher(&user, cacher)
```

Caution:

1. When use Cols methods on cache enabled, the system still return all the columns.

2. When using Exec method, you should clear cache：

```Go
engine.Exec("update user set name = ? where id = ?", "xlw", 1)
engine.ClearCache(new(User))
```

Cache implement theory below:

![cache design](https://raw.github.com/go-xorm/xorm/master/docs/cache_design.png)

<a name="140" id="140"></a>
## 13.Examples

Please visit [https://github.com/go-xorm/xorm/tree/master/examples](https://github.com/go-xorm/xorm/tree/master/examples)

