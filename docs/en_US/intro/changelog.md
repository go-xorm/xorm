---
name: Changelog
sort: 1
---

## Changelog

* **v0.4.0 RC1** 
	Changes:
	* moved xorm cmd to [github.com/go-xorm/cmd](github.com/go-xorm/cmd)
	* refactored general DB operation a core lib at [github.com/go-xorm/core](https://github.com/go-xorm/core)
	* moved tests to github.com/go-xorm/tests [github.com/go-xorm/tests](github.com/go-xorm/tests)

	Improvements:
	* Prepared statement cache
	* Add Incr API
	* Specify Timezone Location

* **v0.3.2** 
	Improvements:
	* Add AllCols & MustCols function
	* Add TableName for custom table name

	Bug Fixes:
	* #46
	* #51
	* #53
	* #89
	* #86
	* #92

* **v0.3.1** 

	Features:
	* Support MSSQL DB via ODBC driver ([github.com/lunny/godbc](https://github.com/lunny/godbc));
	* Composite Key, using multiple pk xorm tag 
	* Added Row() API as alternative to Iterate() API for traversing result set, provide similar usages to sql.Rows type
	* ORM struct allowed declaration of pointer builtin type as members to allow null DB fields 
	* Before and After Event processors

	Improvements:
	* Allowed int/int32/int64/uint/uint32/uint64/string as Primary Key type
	* Performance improvement for Get()/Find()/Iterate()


* **v0.2.3** : Improved documents; Optimistic Locking support; Timestamp with time zone support; Mapper change to tableMapper and columnMapper & added PrefixMapper & SuffixMapper support custom table or column name's prefix and suffix;Insert now return affected, err instead of id, err; Added UseBool & Distinct;
* **v0.2.2** : Postgres drivers now support lib/pq; Added method Iterate for record by record to handler；Added SetMaxConns(go1.2+) support; some bugs fixed.
* **v0.2.1** : Added database reverse tool, now support generate go & c++ codes, see [Xorm Tool README](https://github.com/go-xorm/xorm/blob/master/xorm/README.md); some bug fixed.
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
* **v0.1.0** : Initial release.

