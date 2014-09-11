---
name: FAQ
sort: 100
---

* How the xorm tag use both with json?
  
  Use space.

```Go
type User struct {
    Name string `json:"name" xorm:"name"`
}
```

* Does xorm support composite primary key?

  Yes. You can use pk tag. All fields have tag will as one primary key by fields order on struct. When use, you can use xorm.PK{1, 2}. For example: `Id(xorm.PK{1, 2})`.

* How to use join？

  We can use Join() and extends tag to do join operation. For example:

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

    //assert(User.Userinfo.Id != 0 && User.Userdetail.Id != 0)

Please notice that Userinfo field on User should be before Userdetail because of the order on join SQL stsatement. If the order is wrong, the same name field may be set a wrong value.

Of course, If join statment is very long, you could directly use Sql():

    err := engine.Sql("select * from userinfo, userdetail where userinfo.detail_id = userdetail.id").Find(&users)

    //assert(User.Userinfo.Id != 0 && User.Userdetail.Id != 0)