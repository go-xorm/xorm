---
name: FAQ
sort: 100
---

* 怎么同时使用xorm的tag和json的tag？
  
答：使用空格

```Go
type User struct {
    Name string `json:"name" xorm:"name"`
}
```

* 我的struct里面包含bool类型，为什么它不能作为条件也没法用Update更新？

答：默认bool类型因为无法判断是否为空，所以不会自动作为条件也不会作为Update的内容。可以使用UseBool函数，也可以使用Cols函数

```Go
engine.Cols("bool_field").Update(&Struct{BoolField:true})
// UPDATE struct SET bool_field = true
```

* 我的struct里面包含float64和float32类型，为什么用他们作为查询条件总是不正确？

答：默认float32和float64映射到数据库中为float,real,double这几种类型，这几种数据库类型数据库的实现一般都是非精确的。因此作为相等条件查询有可能不会返回正确的结果。如果一定要作为查询条件，请将数据库中的类型定义为Numeric或者Decimal。

```Go
type account struct {
money float64 `xorm:"Numeric"`
}
```

* 为什么Update时Sqlite3返回的affected和其它数据库不一样？

答：Sqlite3默认Update时返回的是update的查询条件的记录数条数，不管记录是否真的有更新。而Mysql和Postgres默认情况下都是只返回记录中有字段改变的记录数。

* xorm有几种命名映射规则？

答：目前支持SnakeMapper和SameMapper两种。SnakeMapper支持结构体和成员以驼峰式命名而数据库表和字段以下划线连接命名；SameMapper支持结构体和数据库的命名保持一致的映射。

* xorm支持复合主键吗？

答：支持。在定义时，如果有多个字段标记了pk，则这些字段自动成为复合主键，顺序为在struct中出现的顺序。在使用Id方法时，可以用`Id(xorm.PK{1, 2})`的方式来用。

* xorm如何使用Join？

答：一般我们配合Join()和extends标记来进行，比如我们要对两个表进行Join操作，我们可以这样：

	type Userinfo struct {
		Id int64
		Name string
		DetailId int64
	}

	type Userdetail struct {
		Id int64
		Gender int
	}

	type User struct {
		Userinfo `xorm:"extends"`
		Userdetail `xorm:"extends"`
	}

	var users = make([]User, 0)
	err := engine.Table(&Userinfo{}).Join("LEFT", "userdetail", "userinfo.detail_id = userdetail.id").Find(&users)

请注意这里的Userinfo在User中的位置必须在Userdetail的前面，因为他在join语句中的顺序在userdetail前面。如果顺序不对，那么对于同名的列，有可能会赋值出错。

当然，如果Join语句比较复杂，我们也可以直接用Sql函数

	err := engine.Sql("select * from userinfo, userdetail where userinfo.detail_id = userdetail.id").Find(&users)