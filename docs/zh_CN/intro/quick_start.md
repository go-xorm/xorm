---
name: 快速开始
sort: 2
---

xorm 快速入门
=====

* [1.创建Orm引擎](#10)
* [2.定义表结构体](#20)
	* [2.1.名称映射规则](#21)
	* [2.2.前缀映射，后缀映射和缓存映射](#22)
	* [2.3.使用Table和Tag改变名称映射](#23)
	* [2.4.Column属性定义](#24)
	* [2.5.Go与字段类型对应表](#25)
* [3.表结构操作](#30)
	* [3.1 获取数据库信息](#31)
	* [3.2 表操作](#32)
	* [3.3 创建索引和唯一索引](#33)
	* [3.4 同步数据库结构](#34)
* [4.插入数据](#50)
* [5.查询和统计数据](#60)
	* [5.1.查询条件方法](#61)
	* [5.2.临时开关方法](#62)
	* [5.3.Get方法](#63)
	* [5.4.Find方法](#64)
	* [5.5.Iterate方法](#65)
	* [5.6.Count方法](#66)
	* [5.7.Rows方法](#67)
* [6.更新数据](#70)
	* [6.1.乐观锁](#71)
* [7.删除数据](#80)
* [8.执行SQL查询](#90)
* [9.执行SQL命令](#100)
* [10.事务处理](#110)
* [11.缓存](#120)
* [12.事件](#125)
* [14.Examples](#140)

<a name="10" id="10"></a>
## 1.创建Orm引擎

在xorm里面，可以同时存在多个Orm引擎，一个Orm引擎称为Engine。因此在使用前必须调用NewEngine，如：

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

你可以创建一个或多个engine, 不过一般如果操作一个数据库，只需要创建一个Engine即可。Engine支持在多GoRutine下使用。

xorm当前支持五种驱动四个数据库如下：

* Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* MsSql: [github.com/lunny/godbc](https://githubcom/lunny/godbc)

NewEngine传入的参数和`sql.Open`传入的参数完全相同，因此，使用哪个驱动前，请查看此驱动中关于传入参数的说明文档。

在engine创建完成后可以进行一些设置，如：

1.错误显示设置，默认如下均为`false`

* `engine.ShowSQL = true`，则会在控制台打印出生成的SQL语句；
* `engine.ShowDebug = true`，则会在控制台打印调试信息；
* `engine.ShowError = true`，则会在控制台打印错误信息；
* `engine.ShowWarn = true`，则会在控制台打印警告信息；

2.如果希望用其它方式记录，则可以`engine.Logger`赋值为一个`io.Writer`的实现。比如记录到Log文件，则可以：

```Go
f, err := os.Create("sql.log")
	if err != nil {
		println(err.Error())
		return
	}
engine.Logger = xorm.NewSimpleLogger(f)
```

3.engine内部支持连接池接口。

* 如果需要设置连接池的空闲数大小，可以使用`engine.SetMaxIdleConns()`来实现。
* 如果需要设置最大打开连接数，则可以使用`engine.SetMaxOpenConns()`来实现。

<a name="20" id="20"></a>
## 2.定义表结构体

xorm支持将一个struct映射为数据库中对应的一张表。映射规则如下：

<a name="21" id="21"></a>
### 2.1.名称映射规则

名称映射规则主要负责结构体名称到表名和结构体field到表字段的名称映射。由xorm.IMapper接口的实现者来管理，xorm内置了两种IMapper实现：`SnakeMapper` 和 `SameMapper`。SnakeMapper支持struct为驼峰式命名，表结构为下划线命名之间的转换；SameMapper支持结构体名称和对应的表名称以及结构体field名称与对应的表字段名称相同的命名。

当前SnakeMapper为默认值，如果需要改变时，在engine创建完成后使用

```Go
engine.SetMapper(SameMapper{})
```

同时需要注意的是：

* 如果你使用了别的命名规则映射方案，也可以自己实现一个IMapper。
* 表名称和字段名称的映射规则默认是相同的，当然也可以设置为不同，如：

```Go
engine.SetTableMapper(SameMapper{})
engine.SetColumnMapper(SnakeMapper{})
```

<a name="22" id="22"></a>
### 2.2.前缀映射，后缀映射和缓存映射

* 通过`engine.NewPrefixMapper(SnakeMapper{}, "prefix")`可以在SnakeMapper的基础上在命名中添加统一的前缀，当然也可以把SnakeMapper{}换成SameMapper或者你自定义的Mapper。
* 通过`engine.NewSufffixMapper(SnakeMapper{}, "suffix")`可以在SnakeMapper的基础上在命名中添加统一的后缀，当然也可以把SnakeMapper{}换成SameMapper或者你自定义的Mapper。
* 通过`eneing.NewCacheMapper(SnakeMapper{})`可以组合其它的映射规则，起到在内存中缓存曾经映射过的命名映射。

<a name="23" id="23"></a>
### 2.3.使用Table和Tag改变名称映射

如果所有的命名都是按照IMapper的映射来操作的，那当然是最理想的。但是如果碰到某个表名或者某个字段名跟映射规则不匹配时，我们就需要别的机制来改变。

* 如果struct拥有`TableName() string`的成员方法，那么此方法的返回值即是该struct默认对应的数据库表名。

* 通过`engine.Table()`方法可以改变struct对应的数据库表的名称，通过sturct中field对应的Tag中使用`xorm:"'column_name'"`可以使该field对应的Column名称为指定名称。这里使用两个单引号将Column名称括起来是为了防止名称冲突，因为我们在Tag中还可以对这个Column进行更多的定义。如果名称不冲突的情况，单引号也可以不使用。

<a name="23" id="23"></a>
### 2.4.Column属性定义
我们在field对应的Tag中对Column的一些属性进行定义，定义的方法基本和我们写SQL定义表结构类似，比如：

```
type User struct {
	Id 	 int64
	Name string  `xorm:"varchar(25) not null unique 'usr_name'"`
}
```

对于不同的数据库系统，数据类型其实是有些差异的。因此xorm中对数据类型有自己的定义，基本的原则是尽量兼容各种数据库的字段类型，具体的字段对应关系可以查看[字段类型对应表](https://github.com/go-xorm/xorm/blob/master/docs/COLUMNTYPE.md)。对于使用者，一般只要使用自己熟悉的数据库字段定义即可。

具体的映射规则如下，另Tag中的关键字均不区分大小写，字段名区分大小写：

<table>
    <tr>
        <td>name</td><td>当前field对应的字段的名称，可选，如不写，则自动根据field名字和转换规则命名，如与其它关键字冲突，请使用单引号括起来。</td>
    </tr>
    <tr>
        <td>pk</td><td>是否是Primary Key，如果在一个struct中有多个字段都使用了此标记，则这多个字段构成了复合主键，单主键当前支持int32,int,int64,uint32,uint,uint64,string这7种Go的数据类型，复合主键支持这7种Go的数据类型的组合。</td>
    </tr>
    <tr>
        <td>当前支持30多种字段类型，详情参见 [字段类型](https://github.com/go-xorm/xorm/blob/master/docs/COLUMNTYPE.md)</td><td>字段类型</td>
    </tr>
    <tr>
        <td>autoincr</td><td>是否是自增</td>
    </tr>
    <tr>
        <td>[not ]null 或 notnull</td><td>是否可以为空</td>
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
        <td>&lt;-</td><td>这个Field将只从数据库读取，而不写入到数据库</td>
    </tr>
     <tr>
        <td>created</td><td>这个Field将在Insert时自动赋值为当前时间</td>
    </tr>
     <tr>
        <td>updated</td><td>这个Field将在Insert或Update时自动赋值为当前时间</td>
    </tr>
     <tr>
        <td>version</td><td>这个Field将会在insert时默认为1，每次更新自动加1</td>
    </tr>
    <tr>
        <td>default 0</td><td>设置默认值，紧跟的内容如果是Varchar等需要加上单引号</td>
    </tr>
</table>

另外有如下几条自动映射的规则：

- 1.如果field名称为`Id`而且类型为`int64`并且没有定义tag，则会被xorm视为主键，并且拥有自增属性。如果想用`Id`以外的名字或非int64类型做为主键名，必须在对应的Tag上加上`xorm:"pk"`来定义主键，加上`xorm:"autoincr"`作为自增。这里需要注意的是，有些数据库并不允许非主键的自增属性。

- 2.string类型默认映射为varchar(255)，如果需要不同的定义，可以在tag中自定义

- 3.支持`type MyString string`等自定义的field，支持Slice, Map等field成员，这些成员默认存储为Text类型，并且默认将使用Json格式来序列化和反序列化。也支持数据库字段类型为Blob类型，如果是Blob类型，则先使用Json格式序列化再转成[]byte格式。当然[]byte或者[]uint8默认为Blob类型并且都以二进制方式存储。具体参见 [go类型<->数据库类型对应表](https://github.com/go-xorm/xorm/blob/master/docs/AutoMap.md)

- 4.实现了Conversion接口的类型或者结构体，将根据接口的转换方式在类型和数据库记录之间进行相互转换。
```Go
type Conversion interface {
	FromDB([]byte) error
	ToDB() ([]byte, error)
}
```

- 5.如果一个结构体包含一个Conversion的接口类型，那么在获取数据时，必须要预先设置一个实现此接口的struct或者struct的指针。此时可以在此struct中实现`BeforeSet(name string, cell xorm.Cell)`方法来进行预先给Conversion赋值。例子参见 [testConversion](https://github.com/go-xorm/tests/blob/master/base.go#L1826)

<a name="24" id="24"></a>
### 2.4.Go与字段类型对应表

如果不使用tag来定义field对应的数据库字段类型，那么系统会自动给出一个默认的字段类型，对应表如下：

[go类型<->数据库类型对应表](https://github.com/go-xorm/xorm/blob/master/docs/AutoMap.md)

<a name="30" id="30"></a>
## 3.表结构操作

xorm提供了一些动态获取和修改表结构的方法。对于一般的应用，很少动态修改表结构，则只需调用Sync()同步下表结构即可。

<a name="31" id="31"></a>
## 3.1 获取数据库信息

* DBMetas()

xorm支持获取表结构信息，通过调用`engine.DBMetas()`可以获取到所有的表，字段，索引的信息。

<a name="31" id="31"></a>
## 3.2.表操作

* CreateTables()

创建表使用`engine.CreateTables()`，参数为一个或多个空的对应Struct的指针。同时可用的方法有Charset()和StoreEngine()，如果对应的数据库支持，这两个方法可以在创建表时指定表的字符编码和使用的引擎。当前仅支持Mysql数据库。

* IsTableEmpty()

判断表是否为空，参数和CreateTables相同

* IsTableExist()

判断表是否存在

* DropTables()

删除表使用`engine.DropTables()`，参数为一个或多个空的对应Struct的指针或者表的名字。如果为string传入，则只删除对应的表，如果传入的为Struct，则删除表的同时还会删除对应的索引。

<a name="32" id="32"></a>
## 3.3.创建索引和唯一索引

* CreateIndexes

根据struct中的tag来创建索引

* CreateUniques

根据struct中的tag来创建唯一索引

<a name="34" id="34"></a>
## 3.4.同步数据库结构

同步能够部分智能的根据结构体的变动检测表结构的变动，并自动同步。目前能够实现：

* 1) 自动检测和创建表，这个检测是根据表的名字
* 2）自动检测和新增表中的字段，这个检测是根据字段名
* 3）自动检测和创建索引和唯一索引，这个检测是根据一个或多个字段名，而不根据索引名称

调用方法如下：

```Go
err := engine.Sync(new(User))
```

<a name="50" id="50"></a>
## 4.插入数据

插入数据使用Insert方法，Insert方法的参数可以是一个或多个Struct的指针，一个或多个Struct的Slice的指针。
如果传入的是Slice并且当数据库支持批量插入时，Insert会使用批量插入的方式进行插入。

* 插入一条数据

```Go
user := new(User)
user.Name = "myname"
affected, err := engine.Insert(user)
```

在插入单条数据成功后，如果该结构体有自增字段，则自增字段会被自动赋值为数据库中的id

```Go
fmt.Println(user.Id)
```

* 插入同一个表的多条数据

```Go
users := make([]User, 0)
users[0].Name = "name0"
...
affected, err := engine.Insert(&users)
```

* 使用指针Slice插入多条记录

```Go
users := make([]*User, 0)
users[0] = new(User)
users[0].Name = "name0"
...
affected, err := engine.Insert(&users)
```

* 插入不同表的一条记录

```Go
user := new(User)
user.Name = "myname"
question := new(Question)
question.Content = "whywhywhwy?"
affected, err := engine.Insert(user, question)
```

* 插入不同表的多条记录

```Go
users := make([]User, 0)
users[0].Name = "name0"
...
questions := make([]Question, 0)
questions[0].Content = "whywhywhwy?"
affected, err := engine.Insert(&users, &questions)
```

* 插入不同表的一条或多条记录
```Go
user := new(User)
user.Name = "myname"
...
questions := make([]Question, 0)
questions[0].Content = "whywhywhwy?"
affected, err := engine.Insert(user, &questions)
```

这里需要注意以下几点：
* 这里虽然支持同时插入，但这些插入并没有事务关系。因此有可能在中间插入出错后，后面的插入将不会继续。
* 多条插入会自动生成`Insert into table values (),(),()`的语句，因此这样的语句有一个最大的记录数，根据经验测算在150条左右。大于150条后，生成的sql语句将太长可能导致执行失败。因此在插入大量数据时，目前需要自行分割成每150条插入一次。

<a name="60" id="60"></a>
## 5.查询和统计数据

所有的查询条件不区分调用顺序，但必须在调用Get，Find，Count, Iterate, Rows这几个函数之前调用。同时需要注意的一点是，在调用的参数中，如果采用默认的`SnakeMapper`所有的字符字段名均为映射后的数据库的字段名，而不是field的名字。

<a name="61" id="61"></a>
### 5.1.查询条件方法

查询和统计主要使用`Get`, `Find`, `Count`, `Rows`, `Iterate`这几个方法。在进行查询时可以使用多个方法来形成查询条件，条件函数如下：

* Id(interface{})
传入一个PK字段的值，作为查询条件，如果是复合主键，则
`Id(xorm.PK{1, 2})`
传入的两个参数按照struct中pk标记字段出现的顺序赋值。

* Where(string, …interface{})
和SQL中Where语句中的条件基本相同，作为条件

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
某字段在一些值中，这里需要注意必须是[]interface{}才可以展开，由于Go语言的限制，[]int64等均不可以展开。

* Cols(…string)
只查询或更新某些指定的字段，默认是查询所有映射的字段或者根据Update的第一个参数来判断更新的字段。例如：
```Go
engine.Cols("age", "name").Find(&users)
// SELECT age, name FROM user
engine.Cols("age", "name").Update(&user)
// UPDATE user SET age=? AND name=?
```

* AllCols()
查询或更新所有字段。

* MustCols(…string)
某些字段必须更新。

* Omit(...string)
和cols相反，此函数指定排除某些指定的字段。注意：此方法和Cols方法不可同时使用
```Go
engine.Omit("age").Update(&user)
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

* NoCascade()
是否自动关联查询field中的数据，如果struct的field也是一个struct并且映射为某个Id，则可以在查询时自动调用Get方法查询出对应的数据。

<a name="63" id="63"></a>
### 5.3.Get方法

查询单条数据使用`Get`方法，在调用Get方法时需要传入一个对应结构体的指针，同时结构体中的非空field自动成为查询的条件和前面的方法条件组合在一起查询。

如：

1) 根据Id来获得单条数据:

```Go
user := new(User)
has, err := engine.Id(id).Get(user)
// 复合主键的获取方法
// has, errr := engine.Id(xorm.PK{1,2}).Get(user)
```

2) 根据Where来获得单条数据：

```Go
user := new(User)
has, err := engine.Where("name=?", "xlw").Get(user)
```

3) 根据user结构体中已有的非空数据来获得单条数据：

```Go
user := &User{Id:1}
has, err := engine.Get(user)
```

或者其它条件

```Go
user := &User{Name:"xlw"}
has, err := engine.Get(user)
```

返回的结果为两个参数，一个`has`为该条记录是否存在，第二个参数`err`为是否有错误。不管err是否为nil，has都有可能为true或者false。

<a name="64" id="64"></a>
### 5.4.Find方法

查询多条数据使用`Find`方法，Find方法的第一个参数为`slice`的指针或`Map`指针，即为查询后返回的结果，第二个参数可选，为查询的条件struct的指针。

1) 传入Slice用于返回数据

```Go
everyone := make([]Userinfo, 0)
err := engine.Find(&everyone)

pEveryOne := make([]*Userinfo, 0)
err := engine.Find(&pEveryOne)
```

2) 传入Map用户返回数据，map必须为`map[int64]Userinfo`的形式，map的key为id，因此对于复合主键无法使用这种方式。

```Go
users := make(map[int64]Userinfo)
err := engine.Find(&users)

pUsers := make(map[int64]*Userinfo)
err := engine.Find(&pUsers)
```

3) 也可以加入各种条件

```Go
users := make([]Userinfo, 0)
err := engine.Where("age > ? or name = ?", 30, "xlw").Limit(20, 10).Find(&users)
```

<a name="65" id="65"></a>
### 5.5.Iterate方法

Iterate方法提供逐条执行查询到的记录的方法，他所能使用的条件和Find方法完全相同

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

<a name="67" id="67"></a>
### 5.7.Rows方法

Rows方法和Iterate方法类似，提供逐条执行查询到的记录的方法，不过Rows更加灵活好用。
```Go
user := new(User)
rows, err := engine.Where("id >?", 1).Rows(user)
if err != nil {
}
defer rows.Close()
for rows.Next() {
	err = rows.Scan(user)
	//...
}
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

* 1.通过添加Cols函数指定需要更新结构体中的哪些值，未指定的将不更新，指定了的即使为0也会更新。

```Go
affected, err := engine.Id(id).Cols("age").Update(&user)
```

* 2.通过传入map[string]interface{}来进行更新，但这时需要额外指定更新到哪个表，因为通过map是无法自动检测更新哪个表的。

```Go
affected, err := engine.Table(new(User)).Id(id).Update(map[string]interface{}{"age":0})
```

<a name="71" id="71"></a>
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
## 7.删除数据

删除数据`Delete`方法，参数为struct的指针并且成为查询条件。

```Go
user := new(User)
affected, err := engine.Id(id).Delete(user)
```

`Delete`的返回值第一个参数为删除的记录数，第二个参数为错误。

注意：当删除时，如果user中包含有bool,float64或者float32类型，有可能会使删除失败。具体请查看 <a href="#160">FAQ</a>

<a name="90" id="90"></a>
## 8.执行SQL查询

也可以直接执行一个SQL查询，即Select命令。在Postgres中支持原始SQL语句中使用 ` 和 ? 符号。

```Go
sql := "select * from userinfo"
results, err := engine.Query(sql)
```

当调用`Query`时，第一个返回值`results`为`[]map[string][]byte`的形式。

<a name="100" id="100"></a>
## 9.执行SQL命令

也可以直接执行一个SQL命令，即执行Insert， Update， Delete 等操作。此时不管数据库是何种类型，都可以使用 ` 和 ? 符号。

```Go
sql = "update `userinfo` set username=? where id=?"
res, err := engine.Exec(sql, "xiaolun", 1) 
```

<a name="110" id="110"></a>
## 10.事务处理
当使用事务处理时，需要创建Session对象。在进行事物处理时，可以混用ORM方法和RAW方法，如下代码所示：

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

* 注意如果您使用的是mysql，数据库引擎为innodb事务才有效，myisam引擎是不支持事务的。

<a name="120" id="120"></a>
## 11.缓存

xorm内置了一致性缓存支持，不过默认并没有开启。要开启缓存，需要在engine创建完后进行配置，如：
启用一个全局的内存缓存

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.SetDefaultCacher(cacher)
```

上述代码采用了LRU算法的一个缓存，缓存方式是存放到内存中，缓存struct的记录数为1000条，缓存针对的范围是所有具有主键的表，没有主键的表中的数据将不会被缓存。
如果只想针对部分表，则：

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.MapCacher(&user, cacher)
```

如果要禁用某个表的缓存，则：

```Go
engine.MapCacher(&user, nil)
```

设置完之后，其它代码基本上就不需要改动了，缓存系统已经在后台运行。

当前实现了内存存储的CacheStore接口MemoryStore，如果需要采用其它设备存储，可以实现CacheStore接口。

不过需要特别注意不适用缓存或者需要手动编码的地方：

1. 当使用了`Distinct`,`Having`,`GroupBy`方法将不会使用缓存

2. 在`Get`或者`Find`时使用了`Cols`,`Omit`方法，则在开启缓存后此方法无效，系统仍旧会取出这个表中的所有字段。

3. 在使用Exec方法执行了方法之后，可能会导致缓存与数据库不一致的地方。因此如果启用缓存，尽量避免使用Exec。如果必须使用，则需要在使用了Exec之后调用ClearCache手动做缓存清除的工作。比如：

```Go
engine.Exec("update user set name = ? where id = ?", "xlw", 1)
engine.ClearCache(new(User))
```

缓存的实现原理如下图所示：

![cache design](https://raw.github.com/go-xorm/xorm/master/docs/cache_design.png)

<a name="125" id="125"></a>
## 12.事件
xorm支持两种方式的事件，一种是在Struct中的特定方法来作为事件的方法，一种是在执行语句的过程中执行事件。

在Struct中作为成员方法的事件如下：

* BeforeInsert()

* BeforeUpdate()

* BeforeDelete()

* AfterInsert()

* AfterUpdate()

* AfterDelete()

在语句执行过程中的事件方法为：

* Before(beforeFunc interface{})

* After(afterFunc interface{})

其中beforeFunc和afterFunc的原型为func(bean interface{}).

<a name="140" id="140"></a>
## 14.Examples

请访问[https://github.com/go-xorm/xorm/tree/master/examples](https://github.com/go-xorm/xorm/tree/master/examples)
