xorm 快速入门
=====

* [创建Orm引擎](#10)
* [定义表结构体](#20)
* [创建表](#30)
* [删除表](#40)
* [插入数据](#50)
* [查询和统计数据](#60)
* [更新数据](#70)
* [删除数据](#80)
* [执行SQL查询](#90)
* [执行SQL命令](#100)
* [事务处理](#110)

<a name="10" id="10"></a>
### 创建Orm引擎

在xorm里面，可以同时存在多个Orm引擎，一个Orm引擎称为Engine。因此在使用前必须调用NewEngine，如：

```Go
Orm, err := xorm.NewEngine("sqlite3", "./test.db")
```

一般如果只针对一个数据库进行操作，只需要创建一个Engine即可。Engine支持在多GoRutine下使用。

xorm当前支持四种驱动如下：

* Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/bylevel/pg](https://github.com/bylevel/pg)

NewEngine传入的参数和`sql.Open`传入的参数完全相同，因此，使用哪个驱动前，请查看此驱动中关于传入参数的说明文档。

在engine创建完成后可以进行一些设置，如：
1.设置`engine.ShowSQL = true`，则会在控制台打印出生成的SQL语句；如果希望用其它方式记录，则可以`engine.Logger`赋值为一个`io.Writer`的实现。比如记录到Log文件，则可以：

```Go
f, err := os.Create("sql.log")
	if err != nil {
		println(err.Error())
		return
	}
engine.Logger = f
```

2.engine内部支持连接池接口，默认使用的Go所实现的连接池，同时自带了另外两种实现，一种是不使用连接池，另一种为一个自实现的连接池。推荐使用Go所实现的连接池。如果要使用自己实现的连接池，可以实现`xorm.IConnectPool`并通过`engine.SetPool`进行设置。
如果需要设置连接池的空闲数大小，可以使用`engine.Pool.SetIdleConns()`来实现。

<a name="20" id="20"></a>
### 定义表结构体

xorm支持将一个struct映射为数据库中对应的一张表。映射规则如下：

1.名称映射规则

名称映射规则主要负责结构体名称到表名和结构体field到表字段的名称映射。由xorm.IMapper接口的实现者来管理，xorm内置了两种IMapper实现：`SnakeMapper` 和 `SameMapper`。SnakeMapper支持struct为驼峰式命名，表结构为下划线命名之间的转换；SameMapper支持相同的命名。当前SnakeMapper为默认值，当需要改变时，在engine创建完成后使用

```Go
engine.Mapper = SameMapper{}
```

当然，如果你使用了别的命名规则映射方案，也可以自己实现一个IMapper。

2.使用Table和Tag来临时改变名称映射

如果所有的命名都是按照IMapper的映射来操作的，那当然是最理想的。但是如果碰到某个表名或者某个字段名跟映射规则不匹配时，我们就需要别的机制来改变。

通过`engine.Table()`方法可以改变struct对应的数据库表的名称，通过sturct中field对应的Tag中使用`xorm:"'table_name'"`可以使该field对应的Column名称为指定名称。这里使用两个单引号将Column名称括起来是为了防止名称冲突，因为我们在Tag中还可以对这个Column进行更多的定义。如果名称不冲突的情况，单引号也可以不使用。

3.Column属性定义
我们在field对应的Tag中对Column的一些属性进行定义，定义的方法基本和我们写SQL定义表结构类似，比如：

```
type User struct {
	Id 	 int64
	Name string  `xorm:"varchar(25) not null unique 'usr_name'"`
}
```

对于不同的数据库系统，数据类型其实是有些差异的。因此xorm中对数据类型有自己的定义，基本的原则是尽量兼容各种数据库的字段类型，具体的字段对应关系可以查看[字段类型对应表](https://github.com/lunny/xorm/blob/master/COLUMNTYPE.md)。

具体的映射规则如下，另Tag中的关键字均不区分大小写，字段名区分大小写：

<table>
    <tr>
        <td>name</td><td>当前field对应的字段的名称，可选，如不写，则自动根据field名字和转换规则命名</td>
    </tr>
    <tr>
        <td>pk</td><td>是否是Primary Key，当前仅支持int64类型</td>
    </tr>
    <tr>
        <td>当前支持30多种字段类型，详情参见 [字段类型](https://github.com/lunny/xorm/blob/master/COLUMNTYPE.md)</td><td>字段类型</td>
    </tr>
    <tr>
        <td>autoincr</td><td>是否是自增</td>
    </tr>
    <tr>
        <td>[not ]null</td><td>是否可以为空</td>
    </tr>
    <tr>
        <td>unique或unique(uniquename)</td><td>是否是唯一，如不加括号则该字段不允许重复；如加上括号，则括号中为联合唯一索引的名字，此时如果有另外一个或多个字段和本unique的uniquename相同，则这些uniquename相同的字段组成联合唯一索引</td>
    </tr>
    <tr>
        <td>index或index(indexname)</td><td>是否是索引，如不加括号则该字段自身为索引，如加上括号，则括号中为联合索引的名字，此时如果有另外一个或多个字段和本index的indexname相同，则这些indexname相同的字段组成联合索引</td>
    </tr>
    <tr>
    	<td>extends</td><td>应用于一个匿名结构体之上，表示此匿名结构体的成员也映射到数据库中</td>
    </tr>
    <tr>
        <td>-</td><td>这个Field将不进行字段映射</td>
    </tr>
     <tr>
        <td>-></td><td>这个Field将只写入到数据库而不从数据库读取</td>
    </tr>
     <tr>
        <td><-</td><td>这个Field将只从数据库读取，而不写入到数据库</td>
    </tr>
     <tr>
        <td>created</td><td>这个Field将在Insert时自动赋值为当前时间</td>
    </tr>
     <tr>
        <td>updated</td><td>这个Field将在Insert或Update时自动赋值为当前时间</td>
    </tr>
    <tr>
        <td>default 0</td><td>设置默认值，紧跟的内容如过是Varchar等需要加上单引号</td>
    </tr>
</table>

另外有如下几条自动映射的规则：

- 如果field名称为`Id`而且类型为`int64`的话，会被xorm视为主键，并且拥有自增属性。如果想用`Id`以外的名字做为主键名，可以在对应的Tag上加上`xorm:"pk"`来定义主键。

- string类型默认映射为varchar(255)，如果需要不同的定义，可以在tag中自定义

- 支持`type MyString string`等自定义的field，支持Slice, Map等field成员，这些成员默认存储为Text类型，并且默认将使用Json格式来序列化和反序列化。也支持数据库字段类型为Blob类型，如果是Blob类型，则先使用Jsong格式序列化再转成[]byte格式。当然[]byte或者[]uint8默认为Blob类型并且都已二进制方式存储。

<a name="30" id="30"></a>
### 创建表

创建表使用engine.CreateTables()，参数为一个空的对应Struct的指针。同时可用的方法有Charset()和StoreEngine()，如果对应的数据库支持，这两个方法可以在创建表时指定表的字符编码和使用的引擎。当前仅支持Mysql数据库。
在创建表时会判断表是否已经创建，如果已经创建则不再创建。

<a name="40" id="40"></a>
### 删除表
删除表使用engine.DropTables()，参数为一个空的对应Struct的指针或者表的名字。

<a name="50" id="50"></a>
### 插入数据
插入数据使用Insert方法，Insert方法的参数可以是一个或多个Struct的指针，一个或多个Struct的Slice的指针。
如果传入的是Slice并且当数据库支持批量插入时，Insert会使用批量插入的方式进行插入。

```Go
user := new(User)
user.Name = "myname"
affcted, err := engine.Insert(user)
```

在插入成功后，如果该结构体有PK字段，则PK字段会被自动赋值为数据库中的id
```Go
fmt.Println(user.Id)
```

<a name="60" id="60"></a>
### 查询和统计数据

查询和统计主要使用`Get`, `Find`, `Count`三个方法。在进行查询时可以使用多个方法来形成查询条件，条件函数如下：

* Id(int64)
传入一个PK字段的值，作为查询条件

* Where(string, …interface{})
和Where语句中的条件基本相同，作为条件

* Cols(…string)
只查询某些指定的字段，默认是查询所有映射的字段

* Sql(string, …interface{})
执行指定的Sql语句，并把结果映射到结构体

* Table()
指定特殊的Table名，如不加此函数，则根据系统的IMapper自动映射的表名进行查询

* Asc(…string)
指定字段名正序排序

* Desc(…string)
指定字段名逆序排序

* OrderBy()
按照指定的顺序进行排序

* In(string, …interface{})
某字段在一些值中

* Table(nameOrStructPtr interface{})
传入表名称或者结构体指针，如果传入的是结构体指针，则按照IMapper的规则提取出表名

* Limit(int, …int)
限制获取的数目，第一个参数为条数，第二个参数为可选，表示开始位置

* Join(string,string,string)
第一个参数为连接类型，当前支持INNER, LEFT OUTER, CROSS中的一个值，第二个参数为表名，第三个参数为连接条件

* GroupBy(string)
Groupby的参数字符串

* Having(string)
Having的参数字符串

* Cascade(bool)
是否自动关联查询field中的数据，如果struct的field也是一个struct并且映射为某个Id，则可以在查询时自动调用Get方法查询出对应的数据。

所有的查询条件不区分调用顺序，但Get，Find，Count这三个函数必须在最后再调用。同时需要注意的一点是，在调用的参数中，所有的字符字段名均为映射后的数据库的字段名，而不是field的名字。

1.查询单条数据使用`Get`方法，在调用Get方法时需要传入一个对应结构体的指针，同时结构体中的非空field自动成为查询的条件和前面的方法条件组合在一起查询。

```Go
user := new(User)
has, err := engine.Id(id).Get(user)
```

返回的结果为两个参数，一个为该条记录是否存在，第二个参数为是否有错误。如果不管err是否为nil，has都有可能为true或者false。


2.查询多条数据使用`Find`方法，Find方法的第一个参数为slice的指针，第二个参数可选，为查询的条件struct的指针。
```Go
users := make([]User, 0)
err := engine.Where("id >?", 1).Find(&users)
```

3.统计数据使用`Count`方法，Count方法的参数为struct的指针并且成为查询条件。
```Go
user := new(User)
total, err := engine.Where("id >?", 1).Count(user)
```

<a name="70" id="70"></a>
### 更新数据：
- 更新数据使用`Update`方法，Update方法的第一个参数为需要更新的内容，可以为一个结构体指针或者一个Map[string]interface{}类型。当传入的为结构体指针时，只有非空和0的field才会被作为更新的字段。当传入的为Map类型时，key为数据库Column的名字，value为要更新的内容。

```Go
user := new(User)
user.Name = "myname"
affected, err := engine.Id(id).Update(&user)
```

<a name="80" id="80"></a>
### 删除数据
- 删除数据`Delete`方法，参数为struct的指针并且成为查询条件。
```Go
user := new(User)
engine.Id(id).Delete(user)
```

<a name="90" id="90"></a>
###执行SQL查询
也可以直接执行一个SQL查询
```Go
engine.Query()
```

<a name="100" id="100"></a>
###执行SQL命令
```Go
engine.Exec()
```

<a name="110" id="110"></a>
###事务处理
当使用事务处理时，需要创建Session对象。

```Go
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
