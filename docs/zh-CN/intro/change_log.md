---
name: 变更日志
sort: 1
---

## 更新日志

* **v0.4.0 RC1** 
	新特性:
	* 移动xorm cmd [github.com/go-xorm/cmd](github.com/go-xorm/cmd)
	* 在重构一般DB操作核心库 [github.com/go-xorm/core](https://github.com/go-xorm/core)
	* 移动测试github.com/XORM/tests [github.com/go-xorm/tests](github.com/go-xorm/tests)

	改进：
	* Prepared statement 缓存
	* 添加 Incr API
	* 指定时区位置

* **v0.3.2** 
	改进:
	* Add AllCols & MustCols function
	* Add TableName for custom table name

	Bug 修复:
	* #46
	* #51
	* #53
	* #89
	* #86
	* #92
	
* **v0.3.1**

	新特性:
	* 支持 MSSQL DB 通过 ODBC 驱动 ([github.com/lunny/godbc](https://github.com/lunny/godbc));
	* 通过多个pk标记支持联合主键; 
	* 新增 Rows() API 用来遍历查询结果，该函数提供了类似sql.Rows的相似用法，可作为 Iterate() API 的可选替代；
	* ORM 结构体现在允许内建类型的指针作为成员，使得数据库为null成为可能；
	* Before 和 After 支持

	改进:
	* 允许 int/int32/int64/uint/uint32/uint64/string 作为主键类型
	* 查询函数 Get()/Find()/Iterate() 在性能上的改进

* **v0.2.3** : 改善了文档；提供了乐观锁支持；添加了带时区时间字段支持；Mapper现在分成表名Mapper和字段名Mapper，同时实现了表或字段的自定义前缀后缀；Insert方法的返回值含义从id, err更改为 affected, err，请大家注意；添加了UseBool 和 Distinct函数。
* **v0.2.2** : Postgres驱动新增了对lib/pq的支持；新增了逐条遍历方法Iterate；新增了SetMaxConns(go1.2+)支持，修复了bug若干；
* **v0.2.1** : 新增数据库反转工具，当前支持go和c++代码的生成，详见 [Xorm Tool README](https://github.com/go-xorm/xorm/blob/master/xorm/README.md); 修复了一些bug.
* **v0.2.0** : 新增 [缓存](https://github.com/go-xorm/xorm/blob/master/docs/QuickStart.md#120)支持，查询速度提升3-5倍； 新增数据库表和Struct同名的映射方式； 新增Sync同步表结构；
* **v0.1.9** : 新增 postgres 和 mymysql 驱动支持; 在Postgres中支持原始SQL语句中使用 ` 和 ? 符号; 新增Cols, StoreEngine, Charset 函数；SQL语句打印支持io.Writer接口，默认打印到控制台；新增更多的字段类型支持，详见 [映射规则](https://github.com/go-xorm/xorm/blob/master/docs/QuickStartCn.md#21)；删除废弃的MakeSession和Create函数。
* **v0.1.8** : 新增联合index，联合unique支持，请查看 [映射规则](https://github.com/go-xorm/xorm/blob/master/docs/QuickStartCn.md#21)。
* **v0.1.7** : 新增IConnectPool接口以及NoneConnectPool, SysConnectPool, SimpleConnectPool三种实现，可以选择不使用连接池，使用系统连接池和使用自带连接池三种实现，默认为SysConnectPool，即系统自带的连接池。同时支持自定义连接池。Engine新增Close方法，在系统退出时应调用此方法。
* **v0.1.6** : 新增Conversion，支持自定义类型到数据库类型的转换；新增查询结构体自动检测匿名成员支持；新增单向映射支持；
* **v0.1.5** : 新增对多线程的支持；新增Sql()函数；支持任意sql语句的struct查询；Get函数返回值变动；MakeSession和Create函数被NewSession和NewEngine函数替代；
* **v0.1.4** : Get函数和Find函数新增简单的级联载入功能；对更多的数据库类型支持。
* **v0.1.3** : Find函数现在支持传入Slice或者Map，当传入Map时，key为id；新增Table函数以为多表和临时表进行支持。
* **v0.1.2** : Insert函数支持混合struct和slice指针传入，并根据数据库类型自动批量插入，同时自动添加事务
* **v0.1.1** : 添加 Id, In 函数，改善 README 文档
* **v0.1.0** : 初始化工程
