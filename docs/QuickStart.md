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
* [12.Xorm Tool](#130)
	* [12.1.Reverse command](#131)
* [13.Examples](#140)
* [14.Cases](#150)
* [15.FAQ](#160)
* [16.Discuss](#170)

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

Generally, you can only create one engine. Engine supports run on go rutines.

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
engine.Logger = f
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

* 通过`engine.NewPrefixMapper(SnakeMapper{}, "prefix")`可以在SnakeMapper的基础上在命名中添加统一的前缀，当然也可以把SnakeMapper{}换成SameMapper或者你自定义的Mapper。
* 通过`engine.NewSufffixMapper(SnakeMapper{}, "suffix")`可以在SnakeMapper的基础上在命名中添加统一的后缀，当然也可以把SnakeMapper{}换成SameMapper或者你自定义的Mapper。
* 通过`eneing.NewCacheMapper(SnakeMapper{})`可以起到在内存中缓存曾经映射过的命名映射。

当然，如果你使用了别的命名规则映射方案，也可以自己实现一个IMapper。

<a name="22" id="22"></a>
### 2.3.Tag mapping

如果所有的命名都是按照IMapper的映射来操作的，那当然是最理想的。但是如果碰到某个表名或者某个字段名跟映射规则不匹配时，我们就需要别的机制来改变。

通过`engine.Table()`方法可以改变struct对应的数据库表的名称，通过sturct中field对应的Tag中使用`xorm:"'table_name'"`可以使该field对应的Column名称为指定名称。这里使用两个单引号将Column名称括起来是为了防止名称冲突，因为我们在Tag中还可以对这个Column进行更多的定义。如果名称不冲突的情况，单引号也可以不使用。

<a name="23" id="23"></a>
### 2.4.Column defenition

我们在field对应的Tag中对Column的一些属性进行定义，定义的方法基本和我们写SQL定义表结构类似，比如：

```
type User struct {
	Id 	 int64
	Name string  `xorm:"varchar(25) not null unique 'usr_name'"`
}
```

For different DBMS, data types对于不同的数据库系统，数据类型其实是有些差异的。因此xorm中对数据类型有自己的定义，基本的原则是尽量兼容各种数据库的字段类型，具体的字段对应关系可以查看[字段类型对应表](https://github.com/go-xorm/xorm/blob/master/docs/COLUMNTYPE.md)。

具体的映射规则如下，另Tag中的关键字均不区分大小写，字段名区分大小写：

<table>
    <tr>
        <td>name or 'name'</td><td>Column Name, optional</td>
    </tr>
    <tr>
        <td>pk</td><td>If column is Primary Key</td>
    </tr>
    <tr>
        <td>当前支持30多种字段类型，详情参见 [字段类型](https://github.com/go-xorm/xorm/blob/master/docs/COLUMNTYPE.md)</td><td>字段类型</td>
    </tr>
    <tr>
        <td>autoincr</td><td>If autoincrement column</td>
    </tr>
    <tr>
        <td>[not ]null | notnull</td><td>if column could be blank</td>
    </tr>
    <tr>
        <td>unique/unique(uniquename)</td><td>是否是唯一，如不加括号则该字段不允许重复；如加上括号，则括号中为联合唯一索引的名字，此时如果有另外一个或多个字段和本unique的uniquename相同，则这些uniquename相同的字段组成联合唯一索引</td>
    </tr>
    <tr>
        <td>index/index(indexname)</td><td>是否是索引，如不加括号则该字段自身为索引，如加上括号，则括号中为联合索引的名字，此时如果有另外一个或多个字段和本index的indexname相同，则这些indexname相同的字段组成联合索引</td>
    </tr>
    <tr>
    	<td>extends</td><td>应用于一个匿名结构体之上，表示此匿名结构体的成员也映射到数据库中</td>
    </tr>
    <tr>
        <td>-</td><td>This field will not be mapping</td>
    </tr>
     <tr>
        <td>-></td><td>这个Field将只写入到数据库而不从数据库读取</td>
    </tr>
     <tr>
        <td>&lt;-</td><td>这个Field将只从数据库读取，而不写入到数据库</td>
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

另外有如下几条自动映射的规则：

- 1.如果field名称为`Id`而且类型为`int64`的话，会被xorm视为主键，并且拥有自增属性。如果想用`Id`以外的名字做为主键名，可以在对应的Tag上加上`xorm:"pk"`来定义主键。

- 2.string类型默认映射为varchar(255)，如果需要不同的定义，可以在tag中自定义

- 3.支持`type MyString string`等自定义的field，支持Slice, Map等field成员，这些成员默认存储为Text类型，并且默认将使用Json格式来序列化和反序列化。也支持数据库字段类型为Blob类型，如果是Blob类型，则先使用Json格式序列化再转成[]byte格式。当然[]byte或者[]uint8默认为Blob类型并且都以二进制方式存储。

- 4.实现了Conversion接口的类型或者结构体，将根据接口的转换方式在类型和数据库记录之间进行相互转换。
```Go
type Conversion interface {
	FromDB([]byte) error
	ToDB() ([]byte, error)
}
```

<a name="30" id="30"></a>
## 3.表结构操作

xorm提供了一些动态获取和修改表结构的方法。对于一般的应用，很少动态修改表结构，则只需调用Sync()同步下表结构即可。

<a name="31" id="31"></a>
## 3.1 retrieve database meta info

* DBMetas()
xorm支持获取表结构信息，通过调用`engine.DBMetas()`可以获取到所有的表的信息

<a name="31" id="31"></a>
## 3.2.directly table operation

* CreateTables()
创建表使用`engine.CreateTables()`，参数为一个或多个空的对应Struct的指针。同时可用的方法有Charset()和StoreEngine()，如果对应的数据库支持，这两个方法可以在创建表时指定表的字符编码和使用的引擎。当前仅支持Mysql数据库。

* IsTableEmpty()
判断表是否为空，参数和CreateTables相同

* IsTableExist()
判断表是否存在

* DropTables()
删除表使用`engine.DropTables()`，参数为一个或多个空的对应Struct的指针或者表的名字。如果为string传入，则只删除对应的表，如果传入的为Struct，则删除表的同时还会删除对应的索引。

<a name="32" id="32"></a>
## 3.3.create indexes and uniques

* CreateIndexes
根据struct中的tag来创建索引

* CreateUniques
根据struct中的tag来创建唯一索引

<a name="34" id="34"></a>
## 3.4.同步数据库结构

同步能够部分智能的根据结构体的变动检测表结构的变动，并自动同步。目前能够实现：
1) 自动检测和创建表，这个检测是根据表的名字
2）自动检测和新增表中的字段，这个检测是根据字段名
3）自动检测和创建索引和唯一索引，这个检测是根据一个或多个字段名，而不根据索引名称

调用方法如下：
```Go
err := engine.Sync(new(User))
```

<a name="50" id="50"></a>
## 4.插入数据

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
## 5.Query and count

所有的查询条件不区分调用顺序，但必须在调用Get，Find，Count这三个函数之前调用。同时需要注意的一点是，在调用的参数中，所有的字符字段名均为映射后的数据库的字段名，而不是field的名字。

<a name="61" id="61"></a>
### 5.1.查询条件方法

查询和统计主要使用`Get`, `Find`, `Count`三个方法。在进行查询时可以使用多个方法来形成查询条件，条件函数如下：

* Id(int64)
传入一个PK字段的值，作为查询条件

* Where(string, …interface{})
和Where语句中的条件基本相同，作为条件

* And(string, …interface{})
和Where函数中的条件基本相同，作为条件

* Or(string, …interface{})
和Where函数中的条件基本相同，作为条件

* Sql(string, …interface{})
执行指定的Sql语句，并把结果映射到结构体

* Asc(…string)
指定字段名正序排序

* Desc(…string)
指定字段名逆序排序

* OrderBy(string)
按照指定的顺序进行排序

* In(string, …interface{})
某字段在一些值中

* Cols(…string)
只查询或更新某些指定的字段，默认是查询所有映射的字段或者根据Update的第一个参数来判断更新的字段。例如：
```Go
engine.Cols("age", "name").Find(&users)
// SELECT age, name FROM user
engine.Cols("age", "name").Update(&user)
// UPDATE user SET age=? AND name=?
```

其中的参数"age", "name"也可以写成"age, name"，两种写法均可

* Omit(...string)
和cols相反，此函数指定排除某些指定的字段。注意：此方法和Cols方法不可同时使用
```Go
engine.Cols("age").Update(&user)
// UPDATE user SET name = ? AND department = ?
```

* Distinct(…string)
按照参数中指定的字段归类结果
```Go
engine.Distinct("age", "department").Find(&users)
// SELECT DISTINCT age, department FROM user
```
注意：当开启了缓存时，此方法的调用将在当前查询中禁用缓存。因为缓存系统当前依赖Id，而此时无法获得Id

* Table(nameOrStructPtr interface{})
传入表名称或者结构体指针，如果传入的是结构体指针，则按照IMapper的规则提取出表名

* Limit(int, …int)
限制获取的数目，第一个参数为条数，第二个参数为可选，表示开始位置

* Top(int)
相当于Limit(int, 0)

* Join(string,string,string)
第一个参数为连接类型，当前支持INNER, LEFT OUTER, CROSS中的一个值，第二个参数为表名，第三个参数为连接条件

* GroupBy(string)
Groupby的参数字符串

* Having(string)
Having的参数字符串

<a name="62" id="62"></a>
### 5.2.临时开关方法

* NoAutoTime()
如果此方法执行，则此次生成的语句中Created和Updated字段将不自动赋值为当前时间

* NoCache()
如果此方法执行，则此次生成的语句则在非缓存模式下执行

* UseBool(...string)
当从一个struct来生成查询条件或更新字段时，xorm会判断struct的field是否为0,"",nil，如果为以上则不当做查询条件或者更新内容。因为bool类型只有true和false两种值，因此默认所有bool类型不会作为查询条件或者更新字段。如果可以使用此方法，如果默认不传参数，则所有的bool字段都将会被使用，如果参数不为空，则参数中指定的为字段名，则这些字段对应的bool值将被使用。

* Cascade(bool)
是否自动关联查询field中的数据，如果struct的field也是一个struct并且映射为某个Id，则可以在查询时自动调用Get方法查询出对应的数据。

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
### 5.6.Count方法

统计数据使用`Count`方法，Count方法的参数为struct的指针并且成为查询条件。
```Go
user := new(User)
total, err := engine.Where("id >?", 1).Count(user)
```

<a name="70" id="70"></a>
## 6.更新数据
    
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


### 6.1.乐观锁

要使用乐观锁，需要使用version标记
type User struct {
	Id int64
	Name string
	Version int `xorm:"version"`
}

在Insert时，version标记的字段将会被设置为1，在Update时，Update的内容必须包含version原来的值。

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
## 11.缓存

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

<a name="130" id="130"></a>
## 12.xorm tool
xorm工具提供了xorm命令，能够帮助做很多事情。

### 12.1.Reverse command
Please visit [xorm tool](https://github.com/go-xorm/xorm/tree/master/xorm)

<a name="140" id="140"></a>
## 13.Examples

请访问[https://github.com/go-xorm/xorm/tree/master/examples](https://github.com/go-xorm/xorm/tree/master/examples)

<a name="150" id="150"></a>
## 14.Cases

* [Gowalker](http://gowalker.org)，source [github.com/Unknwon/gowalker](http://github.com/Unknwon/gowalker)

* [GoDaily](http://godaily.org)，source [github.com/govc/godaily](http://github.com/govc/godaily)

* [Sudochina](http://sudochina.com) source [github.com/insionng/toropress](http://github.com/insionng/toropress)

* [VeryHour](http://veryhour.com)

<a name="160"></a>
## 15.FAQ 

1.How the xorm tag use both with json?
  
  Use space.

```Go
type User struct {
    Name string `json:"name" xorm:"name"`
}
```
