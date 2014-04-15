package xorm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-xorm/core"
)

// Struct Session keep a pointer to sql.DB and provides all execution of all
// kind of database operations.
type Session struct {
	Db                     *core.DB
	Engine                 *Engine
	Tx                     *core.Tx
	Statement              Statement
	IsAutoCommit           bool
	IsCommitedOrRollbacked bool
	TransType              string
	IsAutoClose            bool

	// !nashtsai! storing these beans due to yet committed tx
	afterInsertBeans map[interface{}]*[]func(interface{})
	afterUpdateBeans map[interface{}]*[]func(interface{})
	afterDeleteBeans map[interface{}]*[]func(interface{})
	// --

	beforeClosures []func(interface{})
	afterClosures  []func(interface{})

	stmtCache map[uint32]*core.Stmt //key: hash.Hash32 of (queryStr, len(queryStr))
}

// Method Init reset the session as the init status.
func (session *Session) Init() {
	session.Statement = Statement{Engine: session.Engine}
	session.Statement.Init()
	session.IsAutoCommit = true
	session.IsCommitedOrRollbacked = false
	session.IsAutoClose = false

	// !nashtsai! is lazy init better?
	session.afterInsertBeans = make(map[interface{}]*[]func(interface{}), 0)
	session.afterUpdateBeans = make(map[interface{}]*[]func(interface{}), 0)
	session.afterDeleteBeans = make(map[interface{}]*[]func(interface{}), 0)
	session.beforeClosures = make([]func(interface{}), 0)
	session.afterClosures = make([]func(interface{}), 0)
}

// Method Close release the connection from pool
func (session *Session) Close() {
	for _, v := range session.stmtCache {
		v.Close()
	}

	if session.Db != nil {
		//session.Engine.Pool.ReleaseDB(session.Engine, session.Db)
		session.Db = nil
		session.Tx = nil
		session.stmtCache = nil
		session.Init()
	}
}

// Method Sql provides raw sql input parameter. When you have a complex SQL statement
// and cannot use Where, Id, In and etc. Methods to describe, you can use Sql.
func (session *Session) Sql(querystring string, args ...interface{}) *Session {
	session.Statement.Sql(querystring, args...)
	return session
}

// Method Where provides custom query condition.
func (session *Session) Where(querystring string, args ...interface{}) *Session {
	session.Statement.Where(querystring, args...)
	return session
}

// Method Where provides custom query condition.
func (session *Session) And(querystring string, args ...interface{}) *Session {
	session.Statement.And(querystring, args...)
	return session
}

// Method Where provides custom query condition.
func (session *Session) Or(querystring string, args ...interface{}) *Session {
	session.Statement.Or(querystring, args...)
	return session
}

// Method Id provides converting id as a query condition
func (session *Session) Id(id interface{}) *Session {
	session.Statement.Id(id)
	return session
}

// Apply before Processor, affected bean is passed to closure arg
func (session *Session) Before(closures func(interface{})) *Session {
	if closures != nil {
		session.beforeClosures = append(session.beforeClosures, closures)
	}
	return session
}

// Apply after Processor, affected bean is passed to closure arg
func (session *Session) After(closures func(interface{})) *Session {
	if closures != nil {
		session.afterClosures = append(session.afterClosures, closures)
	}
	return session
}

// Method core.Table can input a string or pointer to struct for special a table to operate.
func (session *Session) Table(tableNameOrBean interface{}) *Session {
	session.Statement.Table(tableNameOrBean)
	return session
}

// Method In provides a query string like "id in (1, 2, 3)"
func (session *Session) In(column string, args ...interface{}) *Session {
	session.Statement.In(column, args...)
	return session
}

// Method In provides a query string like "count = count + 1"
func (session *Session) Incr(column string, arg ...interface{}) *Session {
	session.Statement.Incr(column, arg...)
	return session
}

// Method Cols provides some columns to special
func (session *Session) Cols(columns ...string) *Session {
	session.Statement.Cols(columns...)
	return session
}

func (session *Session) AllCols() *Session {
	session.Statement.AllCols()
	return session
}

func (session *Session) MustCols(columns ...string) *Session {
	session.Statement.MustCols(columns...)
	return session
}

func (session *Session) NoCascade() *Session {
	session.Statement.UseCascade = false
	return session
}

/*
func (session *Session) MustCols(columns ...string) *Session {
	session.Statement.Must()
}*/

// Xorm automatically retrieve condition according struct, but
// if struct has bool field, it will ignore them. So use UseBool
// to tell system to do not ignore them.
// If no paramters, it will use all the bool field of struct, or
// it will use paramters's columns
func (session *Session) UseBool(columns ...string) *Session {
	session.Statement.UseBool(columns...)
	return session
}

// use for distinct columns. Caution: when you are using cache,
// distinct will not be cached because cache system need id,
// but distinct will not provide id
func (session *Session) Distinct(columns ...string) *Session {
	session.Statement.Distinct(columns...)
	return session
}

// Only not use the paramters as select or update columns
func (session *Session) Omit(columns ...string) *Session {
	session.Statement.Omit(columns...)
	return session
}

// Method NoAutoTime means do not automatically give created field and updated field
// the current time on the current session temporarily
func (session *Session) NoAutoTime() *Session {
	session.Statement.UseAutoTime = false
	return session
}

// Method Limit provide limit and offset query condition
func (session *Session) Limit(limit int, start ...int) *Session {
	session.Statement.Limit(limit, start...)
	return session
}

// Method OrderBy provide order by query condition, the input parameter is the content
// after order by on a sql statement.
func (session *Session) OrderBy(order string) *Session {
	session.Statement.OrderBy(order)
	return session
}

// Method Desc provide desc order by query condition, the input parameters are columns.
func (session *Session) Desc(colNames ...string) *Session {
	if session.Statement.OrderStr != "" {
		session.Statement.OrderStr += ", "
	}
	newColNames := col2NewCols(colNames...)
	sqlStr := strings.Join(newColNames, session.Engine.Quote(" DESC, "))
	session.Statement.OrderStr += session.Engine.Quote(sqlStr) + " DESC"
	return session
}

// Method Asc provide asc order by query condition, the input parameters are columns.
func (session *Session) Asc(colNames ...string) *Session {
	if session.Statement.OrderStr != "" {
		session.Statement.OrderStr += ", "
	}
	newColNames := col2NewCols(colNames...)
	sqlStr := strings.Join(newColNames, session.Engine.Quote(" ASC, "))
	session.Statement.OrderStr += session.Engine.Quote(sqlStr) + " ASC"
	return session
}

// Method StoreEngine is only avialble mysql dialect currently
func (session *Session) StoreEngine(storeEngine string) *Session {
	session.Statement.StoreEngine = storeEngine
	return session
}

// Method StoreEngine is only avialble charset dialect currently
func (session *Session) Charset(charset string) *Session {
	session.Statement.Charset = charset
	return session
}

// Method Cascade indicates if loading sub Struct
func (session *Session) Cascade(trueOrFalse ...bool) *Session {
	if len(trueOrFalse) >= 1 {
		session.Statement.UseCascade = trueOrFalse[0]
	}
	return session
}

// Method NoCache ask this session do not retrieve data from cache system and
// get data from database directly.
func (session *Session) NoCache() *Session {
	session.Statement.UseCache = false
	return session
}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (session *Session) Join(join_operator, tablename, condition string) *Session {
	session.Statement.Join(join_operator, tablename, condition)
	return session
}

// Generate Group By statement
func (session *Session) GroupBy(keys string) *Session {
	session.Statement.GroupBy(keys)
	return session
}

// Generate Having statement
func (session *Session) Having(conditions string) *Session {
	session.Statement.Having(conditions)
	return session
}

func (session *Session) newDb() error {
	if session.Db == nil {
		/*db, err := session.Engine.Pool.RetrieveDB(session.Engine)
		if err != nil {
			return err
		}*/
		session.Db = session.Engine.db
		session.stmtCache = make(map[uint32]*core.Stmt, 0)
	}
	return nil
}

// Begin a transaction
func (session *Session) Begin() error {
	err := session.newDb()
	if err != nil {
		return err
	}
	if session.IsAutoCommit {
		tx, err := session.Db.Begin()
		if err != nil {
			return err
		}
		session.IsAutoCommit = false
		session.IsCommitedOrRollbacked = false
		session.Tx = tx

		session.Engine.logSQL("BEGIN TRANSACTION")
	}
	return nil
}

// When using transaction, you can rollback if any error
func (session *Session) Rollback() error {
	if !session.IsAutoCommit && !session.IsCommitedOrRollbacked {
		session.Engine.logSQL(session.Engine.dialect.RollBackStr())
		session.IsCommitedOrRollbacked = true
		return session.Tx.Rollback()
	}
	return nil
}

// When using transaction, Commit will commit all operations.
func (session *Session) Commit() error {
	if !session.IsAutoCommit && !session.IsCommitedOrRollbacked {
		session.Engine.logSQL("COMMIT")
		session.IsCommitedOrRollbacked = true
		var err error
		if err = session.Tx.Commit(); err == nil {
			// handle processors after tx committed

			closureCallFunc := func(closuresPtr *[]func(interface{}), bean interface{}) {

				if closuresPtr != nil {
					for _, closure := range *closuresPtr {
						closure(bean)
					}
				}
			}

			for bean, closuresPtr := range session.afterInsertBeans {
				closureCallFunc(closuresPtr, bean)

				if processor, ok := interface{}(bean).(AfterInsertProcessor); ok {
					processor.AfterInsert()
				}
			}
			for bean, closuresPtr := range session.afterUpdateBeans {
				closureCallFunc(closuresPtr, bean)

				if processor, ok := interface{}(bean).(AfterUpdateProcessor); ok {
					processor.AfterUpdate()
				}
			}
			for bean, closuresPtr := range session.afterDeleteBeans {
				closureCallFunc(closuresPtr, bean)

				if processor, ok := interface{}(bean).(AfterDeleteProcessor); ok {
					processor.AfterDelete()
				}
			}
			cleanUpFunc := func(slices *map[interface{}]*[]func(interface{})) {
				if len(*slices) > 0 {
					*slices = make(map[interface{}]*[]func(interface{}), 0)
				}
			}
			cleanUpFunc(&session.afterInsertBeans)
			cleanUpFunc(&session.afterUpdateBeans)
			cleanUpFunc(&session.afterDeleteBeans)
		}
		return err
	}
	return nil
}

func cleanupProcessorsClosures(slices *[]func(interface{})) {
	if len(*slices) > 0 {
		*slices = make([]func(interface{}), 0)
	}
}

func (session *Session) scanMapIntoStruct(obj interface{}, objMap map[string][]byte) error {
	dataStruct := rValue(obj)
	if dataStruct.Kind() != reflect.Struct {
		return errors.New("Expected a pointer to a struct")
	}

	var col *core.Column
	table := session.Engine.autoMapType(dataStruct)

	for key, data := range objMap {
		if col = table.GetColumn(key); col == nil {
			session.Engine.LogWarn(fmt.Sprintf("table %v's has not column %v. %v", table.Name, key, table.Columns()))
			continue
		}

		fieldName := col.FieldName
		fieldPath := strings.Split(fieldName, ".")
		var fieldValue reflect.Value
		if len(fieldPath) > 2 {
			session.Engine.LogError("Unsupported mutliderive", fieldName)
			continue
		} else if len(fieldPath) == 2 {
			parentField := dataStruct.FieldByName(fieldPath[0])
			if parentField.IsValid() {
				fieldValue = parentField.FieldByName(fieldPath[1])
			}
		} else {
			fieldValue = dataStruct.FieldByName(fieldName)
		}
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			session.Engine.LogWarn("table %v's column %v is not valid or cannot set",
				table.Name, key)
			continue
		}

		err := session.bytes2Value(col, &fieldValue, data)
		if err != nil {
			return err
		}
	}

	return nil
}

//Execute sql
func (session *Session) innerExec(sqlStr string, args ...interface{}) (sql.Result, error) {
	stmt, err := session.doPrepare(sqlStr)
	if err != nil {
		return nil, err
	}
	//defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (session *Session) exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	for _, filter := range session.Engine.dialect.Filters() {
		sqlStr = filter.Do(sqlStr, session.Engine.dialect, session.Statement.RefTable)
	}

	session.Engine.logSQL(sqlStr, args)

	if session.IsAutoCommit {
		return session.innerExec(sqlStr, args...)
	}
	return session.Tx.Exec(sqlStr, args...)
}

// Exec raw sql
func (session *Session) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	err := session.newDb()
	if err != nil {
		return nil, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	return session.exec(sqlStr, args...)
}

// this function create a table according a bean
func (session *Session) CreateTable(bean interface{}) error {
	session.Statement.RefTable = session.Engine.autoMap(bean)

	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	return session.createOneTable()
}

// create indexes
func (session *Session) CreateIndexes(bean interface{}) error {
	session.Statement.RefTable = session.Engine.autoMap(bean)

	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	sqls := session.Statement.genIndexSQL()
	for _, sqlStr := range sqls {
		_, err = session.exec(sqlStr)
		if err != nil {
			return err
		}
	}
	return nil
}

// create uniques
func (session *Session) CreateUniques(bean interface{}) error {
	session.Statement.RefTable = session.Engine.autoMap(bean)

	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	sqls := session.Statement.genUniqueSQL()
	for _, sqlStr := range sqls {
		_, err = session.exec(sqlStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createOneTable() error {
	sqlStr := session.Statement.genCreateTableSQL()
	session.Engine.LogDebug("create table sql: [", sqlStr, "]")
	_, err := session.exec(sqlStr)
	return err
}

// to be deleted
func (session *Session) createAll() error {
	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	for _, table := range session.Engine.Tables {
		session.Statement.RefTable = table
		err := session.createOneTable()
		if err != nil {
			return err
		}
	}
	return nil
}

// drop indexes
func (session *Session) DropIndexes(bean interface{}) error {
	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	sqls := session.Statement.genDelIndexSQL()
	for _, sqlStr := range sqls {
		_, err = session.exec(sqlStr)
		if err != nil {
			return err
		}
	}
	return nil
}

// DropTable drop a table and all indexes of the table
func (session *Session) DropTable(bean interface{}) error {
	err := session.newDb()
	if err != nil {
		return err
	}

	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	t := reflect.Indirect(reflect.ValueOf(bean)).Type()
	defer session.Statement.Init()
	if t.Kind() == reflect.String {
		session.Statement.AltTableName = bean.(string)
	} else if t.Kind() == reflect.Struct {
		session.Statement.RefTable = session.Engine.autoMap(bean)
	} else {
		return errors.New("Unsupported type")
	}

	sqlStr := session.Statement.genDropSQL()
	_, err = session.exec(sqlStr)
	return err
}

func (statement *Statement) convertIdSql(sqlStr string) string {
	if statement.RefTable != nil {
		col := statement.RefTable.PKColumns()[0]
		if col != nil {
			sqls := splitNNoCase(sqlStr, "from", 2)
			if len(sqls) != 2 {
				return ""
			}
			//fmt.Println("-----", col)
			newsql := fmt.Sprintf("SELECT %v.%v FROM %v", statement.Engine.Quote(statement.TableName()),
				statement.Engine.Quote(col.Name), sqls[1])
			return newsql
		}
	}
	return ""
}

func (session *Session) cacheGet(bean interface{}, sqlStr string, args ...interface{}) (has bool, err error) {
	// if has no reftable or number of pks is not equal to 1, then don't use cache currently
	if session.Statement.RefTable == nil || len(session.Statement.RefTable.PrimaryKeys) != 1 {
		return false, ErrCacheFailed
	}
	for _, filter := range session.Engine.dialect.Filters() {
		sqlStr = filter.Do(sqlStr, session.Engine.dialect, session.Statement.RefTable)
	}
	newsql := session.Statement.convertIdSql(sqlStr)
	if newsql == "" {
		return false, ErrCacheFailed
	}

	cacher := session.Engine.getCacher2(session.Statement.RefTable)
	tableName := session.Statement.TableName()
	session.Engine.LogDebug("[xorm:cacheGet] find sql:", newsql, args)
	ids, err := core.GetCacheSql(cacher, tableName, newsql, args)
	if err != nil {
		resultsSlice, err := session.query(newsql, args...)
		if err != nil {
			return false, err
		}
		session.Engine.LogDebug("[xorm:cacheGet] query ids:", resultsSlice)
		ids = make([]core.PK, 0)
		if len(resultsSlice) > 0 {
			data := resultsSlice[0]
			var id int64
			if v, ok := data[session.Statement.RefTable.PrimaryKeys[0]]; !ok {
				return false, ErrCacheFailed
			} else {
				id, err = strconv.ParseInt(string(v), 10, 64)
				if err != nil {
					return false, err
				}
			}
			ids = append(ids, core.PK{id})
		}
		session.Engine.LogDebug("[xorm:cacheGet] cache ids:", newsql, ids)
		err = core.PutCacheSql(cacher, ids, tableName, newsql, args)
		if err != nil {
			return false, err
		}
	} else {
		session.Engine.LogDebug("[xorm:cacheGet] cached sql:", newsql)
	}

	if len(ids) > 0 {
		structValue := reflect.Indirect(reflect.ValueOf(bean))
		id := ids[0]
		session.Engine.LogDebug("[xorm:cacheGet] get bean:", tableName, id)
		sid, err := id.ToString()
		if err != nil {
			return false, err
		}
		cacheBean := cacher.GetBean(tableName, sid)
		if cacheBean == nil {
			newSession := session.Engine.NewSession()
			defer newSession.Close()
			cacheBean = reflect.New(structValue.Type()).Interface()
			newSession.Id(id).NoCache()
			if session.Statement.AltTableName != "" {
				newSession.Table(session.Statement.AltTableName)
			}
			if !session.Statement.UseCascade {
				newSession.NoCascade()
			}
			has, err = newSession.Get(cacheBean)
			if err != nil || !has {
				return has, err
			}

			session.Engine.LogDebug("[xorm:cacheGet] cache bean:", tableName, id, cacheBean)
			cacher.PutBean(tableName, sid, cacheBean)
		} else {
			session.Engine.LogDebug("[xorm:cacheGet] cached bean:", tableName, id, cacheBean)
			has = true
		}
		structValue.Set(reflect.Indirect(reflect.ValueOf(cacheBean)))

		return has, nil
	}
	return false, nil
}

func (session *Session) cacheFind(t reflect.Type, sqlStr string, rowsSlicePtr interface{}, args ...interface{}) (err error) {
	if session.Statement.RefTable == nil ||
		len(session.Statement.RefTable.PrimaryKeys) != 1 ||
		indexNoCase(sqlStr, "having") != -1 ||
		indexNoCase(sqlStr, "group by") != -1 {
		return ErrCacheFailed
	}

	for _, filter := range session.Engine.dialect.Filters() {
		sqlStr = filter.Do(sqlStr, session.Engine.dialect, session.Statement.RefTable)
	}

	newsql := session.Statement.convertIdSql(sqlStr)
	if newsql == "" {
		return ErrCacheFailed
	}

	table := session.Statement.RefTable
	cacher := session.Engine.getCacher2(table)
	ids, err := core.GetCacheSql(cacher, session.Statement.TableName(), newsql, args)
	if err != nil {
		//session.Engine.LogError(err)
		resultsSlice, err := session.query(newsql, args...)
		if err != nil {
			return err
		}
		// 查询数目太大，采用缓存将不是一个很好的方式。
		if len(resultsSlice) > 500 {
			session.Engine.LogDebug("[xorm:cacheFind] ids length %v > 500, no cache", len(resultsSlice))
			return ErrCacheFailed
		}

		tableName := session.Statement.TableName()
		ids = make([]core.PK, 0)
		if len(resultsSlice) > 0 {
			for _, data := range resultsSlice {
				//fmt.Println(data)
				var id int64
				if v, ok := data[session.Statement.RefTable.PrimaryKeys[0]]; !ok {
					return errors.New("no id")
				} else {
					id, err = strconv.ParseInt(string(v), 10, 64)
					if err != nil {
						return err
					}
				}
				ids = append(ids, core.PK{id})
			}
		}
		session.Engine.LogDebug("[xorm:cacheFind] cache ids:", ids, tableName, newsql, args)
		err = core.PutCacheSql(cacher, ids, tableName, newsql, args)
		if err != nil {
			return err
		}
	} else {
		session.Engine.LogDebug("[xorm:cacheFind] cached sql:", newsql, args)
	}

	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	//pkFieldName := session.Statement.RefTable.PKColumns()[0].FieldName

	ididxes := make(map[string]int)
	var ides []core.PK = make([]core.PK, 0)
	var temps []interface{} = make([]interface{}, len(ids))
	tableName := session.Statement.TableName()
	for idx, id := range ids {
		sid, err := id.ToString()
		if err != nil {
			return err
		}
		bean := cacher.GetBean(tableName, sid)
		if bean == nil {
			ides = append(ides, id)
			ididxes[sid] = idx
		} else {
			session.Engine.LogDebug("[xorm:cacheFind] cached bean:", tableName, id, bean)

			pk := session.Engine.IdOf(bean)
			xid, err := pk.ToString()
			if err != nil {
				return err
			}

			if sid != xid {
				session.Engine.LogError("[xorm:cacheFind] error cache", xid, sid, bean)
				return ErrCacheFailed
			}
			temps[idx] = bean
		}
	}

	if len(ides) > 0 {
		newSession := session.Engine.NewSession()
		defer newSession.Close()

		slices := reflect.New(reflect.SliceOf(t))
		beans := slices.Interface()
		//beans := reflect.New(sliceValue.Type()).Interface()
		//err = newSession.In("(id)", ides...).OrderBy(session.Statement.OrderStr).NoCache().Find(beans)
		ff := make([][]interface{}, len(table.PrimaryKeys))
		for i, _ := range table.PrimaryKeys {
			ff[i] = make([]interface{}, 0)
		}
		for _, ie := range ides {
			for i, _ := range table.PrimaryKeys {
				ff[i] = append(ff[i], ie[i])
			}
		}
		for i, name := range table.PrimaryKeys {
			newSession.In(name, ff[i]...)
		}
		err = newSession.NoCache().Find(beans)
		if err != nil {
			return err
		}

		vs := reflect.Indirect(reflect.ValueOf(beans))
		for i := 0; i < vs.Len(); i++ {
			rv := vs.Index(i)
			if rv.Kind() != reflect.Ptr {
				rv = rv.Addr()
			}
			bean := rv.Interface()
			id := session.Engine.IdOf(bean)
			sid, err := id.ToString()
			if err != nil {
				return err
			}
			//bean := vs.Index(i).Addr().Interface()
			temps[ididxes[sid]] = bean
			//temps[idxes[i]] = bean
			session.Engine.LogDebug("[xorm:cacheFind] cache bean:", tableName, id, bean)
			cacher.PutBean(tableName, sid, bean)
		}
	}

	for j := 0; j < len(temps); j++ {
		bean := temps[j]
		if bean == nil {
			session.Engine.LogError("[xorm:cacheFind] cache error:", tableName, ides[j], bean)
			return errors.New("cache error")
		}
		if sliceValue.Kind() == reflect.Slice {
			if t.Kind() == reflect.Ptr {
				sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(bean)))
			} else {
				sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(bean))))
			}
		} else if sliceValue.Kind() == reflect.Map {
			var key core.PK
			if table.PrimaryKeys[0] != "" {
				key = ids[j]
			}

			if len(key) == 1 {
				ikey, err := strconv.ParseInt(fmt.Sprintf("%v", key[0]), 10, 64)
				if err != nil {
					return err
				}
				if t.Kind() == reflect.Ptr {
					sliceValue.SetMapIndex(reflect.ValueOf(ikey), reflect.ValueOf(bean))
				} else {
					sliceValue.SetMapIndex(reflect.ValueOf(ikey), reflect.Indirect(reflect.ValueOf(bean)))
				}
			}
		}
		/*} else {
		    session.Engine.LogDebug("[xorm:cacheFind] cache delete:", tableName, ides[j])
		    cacher.DelBean(tableName, ids[j])

		    session.Engine.LogDebug("[xorm:cacheFind] cache clear:", tableName)
		    cacher.ClearIds(tableName)
		}*/
	}

	return nil
}

// IterFunc only use by Iterate
type IterFunc func(idx int, bean interface{}) error

// Return sql.Rows compatible Rows obj, as a forward Iterator object for iterating record by record, bean's non-empty fields
// are conditions.
func (session *Session) Rows(bean interface{}) (*Rows, error) {
	return newRows(session, bean)
}

// Iterate record by record handle records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct
func (session *Session) Iterate(bean interface{}, fun IterFunc) error {

	rows, err := session.Rows(bean)
	if err != nil {
		return err
	} else {
		defer rows.Close()
		//b := reflect.New(iterator.beanType).Interface()
		i := 0
		for rows.Next() {
			b := reflect.New(rows.beanType).Interface()
			err = rows.Scan(b)
			if err != nil {
				return err
			}
			err = fun(i, b)
			if err != nil {
				return err
			}
			i++
		}
		return err
	}
	return nil
}

func (session *Session) doPrepare(sqlStr string) (stmt *core.Stmt, err error) {
	crc := crc32.ChecksumIEEE([]byte(sqlStr))
	// TODO try hash(sqlStr+len(sqlStr))
	var has bool
	stmt, has = session.stmtCache[crc]
	if !has {
		stmt, err = session.Db.Prepare(sqlStr)
		if err != nil {
			return nil, err
		}
		session.stmtCache[crc] = stmt
	}
	return
}

// get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (session *Session) Get(bean interface{}) (bool, error) {
	err := session.newDb()
	if err != nil {
		return false, err
	}

	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	session.Statement.Limit(1)
	var sqlStr string
	var args []interface{}

	session.Statement.RefTable = session.Engine.autoMap(bean)

	if session.Statement.RawSQL == "" {
		sqlStr, args = session.Statement.genGetSql(bean)
	} else {
		sqlStr = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	if cacher := session.Engine.getCacher2(session.Statement.RefTable); cacher != nil && session.Statement.UseCache {
		has, err := session.cacheGet(bean, sqlStr, args...)
		if err != ErrCacheFailed {
			return has, err
		}
	}

	var rawRows *core.Rows
	session.queryPreprocess(&sqlStr, args...)
	if session.IsAutoCommit {
		stmt, err := session.doPrepare(sqlStr)
		if err != nil {
			return false, err
		}
		// defer stmt.Close() // !nashtsai! don't close due to stmt is cached and bounded to this session
		rawRows, err = stmt.Query(args...)
	} else {
		rawRows, err = session.Tx.Query(sqlStr, args...)
	}
	if err != nil {
		return false, err
	}

	defer rawRows.Close()

	if rawRows.Next() {
		if fields, err := rawRows.Columns(); err == nil {
			err = session.row2Bean(rawRows, fields, len(fields), bean)
		}
		return true, err
	} else {
		return false, nil
	}

	// resultsSlice, err := session.query(sqlStr, args...)
	// if err != nil {
	// 	return false, err
	// }
	// if len(resultsSlice) < 1 {
	// 	return false, nil
	// }

	// err = session.scanMapIntoStruct(bean, resultsSlice[0])
	// if err != nil {
	// 	return true, err
	// }
	// if len(resultsSlice) == 1 {
	// 	return true, nil
	// } else {
	// 	return true, errors.New("More than one record")
	// }
}

// Count counts the records. bean's non-empty fields
// are conditions.
func (session *Session) Count(bean interface{}) (int64, error) {
	err := session.newDb()
	if err != nil {
		return 0, err
	}

	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	var sqlStr string
	var args []interface{}
	if session.Statement.RawSQL == "" {
		sqlStr, args = session.Statement.genCountSql(bean)
	} else {
		sqlStr = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	resultsSlice, err := session.query(sqlStr, args...)
	if err != nil {
		return 0, err
	}

	var total int64 = 0
	if len(resultsSlice) > 0 {
		results := resultsSlice[0]
		for _, value := range results {
			total, err = strconv.ParseInt(string(value), 10, 64)
			break
		}
	}

	return int64(total), err
}

// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct
func (session *Session) Find(rowsSlicePtr interface{}, condiBean ...interface{}) error {
	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice && sliceValue.Kind() != reflect.Map {
		return errors.New("needs a pointer to a slice or a map")
	}

	sliceElementType := sliceValue.Type().Elem()
	var table *core.Table
	if session.Statement.RefTable == nil {
		if sliceElementType.Kind() == reflect.Ptr {
			if sliceElementType.Elem().Kind() == reflect.Struct {
				pv := reflect.New(sliceElementType.Elem())
				table = session.Engine.autoMapType(pv.Elem())
			} else {
				return errors.New("slice type")
			}
		} else if sliceElementType.Kind() == reflect.Struct {
			pv := reflect.New(sliceElementType)
			table = session.Engine.autoMapType(pv.Elem())
		} else {
			return errors.New("slice type")
		}
		session.Statement.RefTable = table
	} else {
		table = session.Statement.RefTable
	}

	if len(condiBean) > 0 {
		colNames, args := buildConditions(session.Engine, table, condiBean[0], true, true,
			false, true, session.Statement.allUseBool, session.Statement.useAllCols,
			session.Statement.mustColumnMap)
		session.Statement.ConditionStr = strings.Join(colNames, " AND ")
		session.Statement.BeanArgs = args
	}

	var sqlStr string
	var args []interface{}
	if session.Statement.RawSQL == "" {
		var columnStr string = session.Statement.ColumnStr
		if columnStr == "" {
			columnStr = session.Statement.genColumnStr()
		}

		session.Statement.attachInSql()

		sqlStr = session.Statement.genSelectSql(columnStr)
		args = append(session.Statement.Params, session.Statement.BeanArgs...)
	} else {
		sqlStr = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	if cacher := session.Engine.getCacher2(table); cacher != nil &&
		session.Statement.UseCache &&
		!session.Statement.IsDistinct {
		err = session.cacheFind(sliceElementType, sqlStr, rowsSlicePtr, args...)
		if err != ErrCacheFailed {
			return err
		}
		err = nil // !nashtsai! reset err to nil for ErrCacheFailed
		session.Engine.LogWarn("Cache Find Failed")
	}

	if sliceValue.Kind() != reflect.Map {
		var rawRows *core.Rows
		var stmt *core.Stmt

		session.queryPreprocess(&sqlStr, args...)
		// err = session.queryRows(&stmt, &rawRows, sqlStr, args...)
		// if err != nil {
		// 	return err
		// }
		// if stmt != nil {
		// 	defer stmt.Close()
		// }
		// defer rawRows.Close()

		if session.IsAutoCommit {
			stmt, err = session.doPrepare(sqlStr)
			if err != nil {
				return err
			}
			rawRows, err = stmt.Query(args...)
		} else {
			rawRows, err = session.Tx.Query(sqlStr, args...)
		}
		if err != nil {
			return err
		}
		defer rawRows.Close()

		fields, err := rawRows.Columns()
		if err != nil {
			return err
		}

		fieldsCount := len(fields)

		var newElemFunc func() reflect.Value
		if sliceElementType.Kind() == reflect.Ptr {
			newElemFunc = func() reflect.Value {
				return reflect.New(sliceElementType.Elem())
			}
		} else {
			newElemFunc = func() reflect.Value {
				return reflect.New(sliceElementType)
			}
		}

		var sliceValueSetFunc func(*reflect.Value)

		if sliceValue.Kind() == reflect.Slice {
			if sliceElementType.Kind() == reflect.Ptr {
				sliceValueSetFunc = func(newValue *reflect.Value) {
					sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(newValue.Interface())))
				}
			} else {
				sliceValueSetFunc = func(newValue *reflect.Value) {
					sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
				}
			}
		}

		for rawRows.Next() {
			var newValue reflect.Value = newElemFunc()
			if sliceValueSetFunc != nil {
				err := session.row2Bean(rawRows, fields, fieldsCount, newValue.Interface())
				if err != nil {
					return err
				}
				sliceValueSetFunc(&newValue)
			}
		}
	} else {
		resultsSlice, err := session.query(sqlStr, args...)
		if err != nil {
			return err
		}

		for i, results := range resultsSlice {
			var newValue reflect.Value
			if sliceElementType.Kind() == reflect.Ptr {
				newValue = reflect.New(sliceElementType.Elem())
			} else {
				newValue = reflect.New(sliceElementType)
			}
			err := session.scanMapIntoStruct(newValue.Interface(), results)
			if err != nil {
				return err
			}
			var key int64
			// if there is only one pk, we can put the id as map key.
			// TODO: should know if the column is ints
			if len(table.PrimaryKeys) == 1 {
				x, err := strconv.ParseInt(string(results[table.PrimaryKeys[0]]), 10, 64)
				if err != nil {
					return errors.New("pk " + table.PrimaryKeys[0] + " as int64: " + err.Error())
				}
				key = x
			} else {
				key = int64(i)
			}
			if sliceElementType.Kind() == reflect.Ptr {
				sliceValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(newValue.Interface()))
			} else {
				sliceValue.SetMapIndex(reflect.ValueOf(key), reflect.Indirect(reflect.ValueOf(newValue.Interface())))
			}
		}
	}
	return nil
}

// func (session *Session) queryRows(rawStmt **sql.Stmt, rawRows **sql.Rows, sqlStr string, args ...interface{}) error {
// 	var err error
// 	if session.IsAutoCommit {
// 		*rawStmt, err = session.doPrepare(sqlStr)
// 		if err != nil {
// 			return err
// 		}
// 		*rawRows, err = (*rawStmt).Query(args...)
// 	} else {
// 		*rawRows, err = session.Tx.Query(sqlStr, args...)
// 	}
// 	return err
// }

// Test if database is ok
func (session *Session) Ping() error {
	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	return session.Db.Ping()
}

func (session *Session) isColumnExist(tableName, colName string) (bool, error) {
	err := session.newDb()
	if err != nil {
		return false, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}
	sqlStr, args := session.Engine.dialect.ColumnCheckSql(tableName, colName)
	results, err := session.query(sqlStr, args...)
	return len(results) > 0, err
}

func (session *Session) isTableExist(tableName string) (bool, error) {
	err := session.newDb()
	if err != nil {
		return false, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}
	sqlStr, args := session.Engine.dialect.TableCheckSql(tableName)
	results, err := session.query(sqlStr, args...)
	return len(results) > 0, err
}

func (session *Session) isIndexExist(tableName, idxName string, unique bool) (bool, error) {
	err := session.newDb()
	if err != nil {
		return false, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}
	var idx string
	if unique {
		idx = uniqueName(tableName, idxName)
	} else {
		idx = indexName(tableName, idxName)
	}
	sqlStr, args := session.Engine.dialect.IndexCheckSql(tableName, idx)
	results, err := session.query(sqlStr, args...)
	return len(results) > 0, err
}

// find if index is exist according cols
func (session *Session) isIndexExist2(tableName string, cols []string, unique bool) (bool, error) {
	indexes, err := session.Engine.dialect.GetIndexes(tableName)
	if err != nil {
		return false, err
	}

	for _, index := range indexes {
		if sliceEq(index.Cols, cols) {
			if unique {
				return index.Type == core.UniqueType, nil
			} else {
				return index.Type == core.IndexType, nil
			}
		}
	}
	return false, nil
}

func (session *Session) addColumn(colName string) error {
	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}
	//fmt.Println(session.Statement.RefTable)

	col := session.Statement.RefTable.GetColumn(colName)
	sql, args := session.Statement.genAddColumnStr(col)
	_, err = session.exec(sql, args...)
	return err
}

func (session *Session) addIndex(tableName, idxName string) error {
	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}
	//fmt.Println(idxName)
	cols := session.Statement.RefTable.Indexes[idxName].Cols
	sqlStr, args := session.Statement.genAddIndexStr(indexName(tableName, idxName), cols)
	_, err = session.exec(sqlStr, args...)
	return err
}

func (session *Session) addUnique(tableName, uqeName string) error {
	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}
	//fmt.Println(uqeName, session.Statement.RefTable.Uniques)
	cols := session.Statement.RefTable.Indexes[uqeName].Cols
	sqlStr, args := session.Statement.genAddUniqueStr(uniqueName(tableName, uqeName), cols)
	_, err = session.exec(sqlStr, args...)
	return err
}

// To be deleted
func (session *Session) dropAll() error {
	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	for _, table := range session.Engine.Tables {
		session.Statement.Init()
		session.Statement.RefTable = table
		sqlStr := session.Statement.genDropSQL()
		_, err := session.exec(sqlStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func row2map(rows *core.Rows, fields []string) (resultsMap map[string][]byte, err error) {
	result := make(map[string][]byte)
	scanResultContainers := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		var scanResultContainer interface{}
		scanResultContainers[i] = &scanResultContainer
	}
	if err := rows.Scan(scanResultContainers...); err != nil {
		return nil, err
	}

	for ii, key := range fields {
		rawValue := reflect.Indirect(reflect.ValueOf(scanResultContainers[ii]))
		//if row is null then ignore
		if rawValue.Interface() == nil {
			//fmt.Println("ignore ...", key, rawValue)
			continue
		}

		if data, err := value2Bytes(&rawValue); err == nil {
			result[key] = data
		} else {
			return nil, err // !nashtsai! REVIEW, should return err or just error log?
		}
	}
	return result, nil
}

func (session *Session) getField(dataStruct *reflect.Value, key string, table *core.Table) *reflect.Value {
	var col *core.Column
	if col = table.GetColumn(key); col == nil {
		session.Engine.LogWarn(fmt.Sprintf("table %v's has not column %v. %v", table.Name, key, table.Columns()))
		return nil
	}

	fieldValue, err := col.ValueOfV(dataStruct)
	if err != nil {
		session.Engine.LogError(err)
		return nil
	}

	if !fieldValue.IsValid() || !fieldValue.CanSet() {
		session.Engine.LogWarn("table %v's column %v is not valid or cannot set",
			table.Name, key)
		return nil
	}
	return fieldValue
}

func (session *Session) row2Bean(rows *core.Rows, fields []string, fieldsCount int, bean interface{}) error {
	dataStruct := rValue(bean)
	if dataStruct.Kind() != reflect.Struct {
		return errors.New("Expected a pointer to a struct")
	}

	table := session.Engine.autoMapType(dataStruct)

	scanResultContainers := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		var scanResultContainer interface{}
		scanResultContainers[i] = &scanResultContainer
	}
	if err := rows.Scan(scanResultContainers...); err != nil {
		return err
	}

	for ii, key := range fields {
		if fieldValue := session.getField(&dataStruct, key, table); fieldValue != nil {
			rawValue := reflect.Indirect(reflect.ValueOf(scanResultContainers[ii]))

			//if row is null then ignore
			if rawValue.Interface() == nil {
				//fmt.Println("ignore ...", key, rawValue)
				continue
			}

			if structConvert, ok := fieldValue.Addr().Interface().(core.Conversion); ok {
				if data, err := value2Bytes(&rawValue); err == nil {
					structConvert.FromDB(data)
				} else {
					session.Engine.LogError(err)
				}
				continue
			}

			rawValueType := reflect.TypeOf(rawValue.Interface())
			vv := reflect.ValueOf(rawValue.Interface())

			fieldType := fieldValue.Type()

			//fmt.Println("column name:", key, ", fieldType:", fieldType.String())

			hasAssigned := false

			switch fieldType.Kind() {

			case reflect.Complex64, reflect.Complex128:
				if rawValueType.Kind() == reflect.String {
					hasAssigned = true
					x := reflect.New(fieldType)
					err := json.Unmarshal([]byte(vv.String()), x.Interface())
					if err != nil {
						session.Engine.LogError(err)
						return err
					}
					fieldValue.Set(x.Elem())
				}
			case reflect.Slice, reflect.Array:
				switch rawValueType.Kind() {
				case reflect.Slice, reflect.Array:
					switch rawValueType.Elem().Kind() {
					case reflect.Uint8:
						if fieldType.Elem().Kind() == reflect.Uint8 {
							hasAssigned = true
							fieldValue.Set(vv)
						}
					}
				}
			case reflect.String:
				if rawValueType.Kind() == reflect.String {
					hasAssigned = true
					fieldValue.SetString(vv.String())
				}
			case reflect.Bool:
				if rawValueType.Kind() == reflect.Bool {
					hasAssigned = true
					fieldValue.SetBool(vv.Bool())
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch rawValueType.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					hasAssigned = true
					fieldValue.SetInt(vv.Int())
				}
			case reflect.Float32, reflect.Float64:
				switch rawValueType.Kind() {
				case reflect.Float32, reflect.Float64:
					hasAssigned = true
					fieldValue.SetFloat(vv.Float())
				}
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
				switch rawValueType.Kind() {
				case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
					hasAssigned = true
					fieldValue.SetUint(vv.Uint())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					hasAssigned = true
					fieldValue.SetUint(uint64(vv.Int()))
				}
			case reflect.Struct:
				if fieldType == core.TimeType {
					if rawValueType == core.TimeType {
						hasAssigned = true
						fieldValue.Set(vv)
					}
				} else if session.Statement.UseCascade {
					table := session.Engine.autoMapType(*fieldValue)
					if table != nil {
						var x int64
						if rawValueType.Kind() == reflect.Int64 {
							x = vv.Int()
						}
						if x != 0 {
							// !nashtsai! TODO for hasOne relationship, it's preferred to use join query for eager fetch
							// however, also need to consider adding a 'lazy' attribute to xorm tag which allow hasOne
							// property to be fetched lazily
							structInter := reflect.New(fieldValue.Type())
							newsession := session.Engine.NewSession()
							defer newsession.Close()
							has, err := newsession.Id(x).Get(structInter.Interface())
							if err != nil {
								return err
							}
							if has {
								v := structInter.Elem().Interface()
								fieldValue.Set(reflect.ValueOf(v))
							} else {
								return errors.New("cascade obj is not exist!")
							}
						}
					} else {
						session.Engine.LogError("unsupported struct type in Scan: ", fieldValue.Type().String())
					}
				}
			case reflect.Ptr:
				// !nashtsai! TODO merge duplicated codes above
				//typeStr := fieldType.String()
				switch fieldType {
				// following types case matching ptr's native type, therefore assign ptr directly
				case core.PtrStringType:
					if rawValueType.Kind() == reflect.String {
						x := vv.String()
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrBoolType:
					if rawValueType.Kind() == reflect.Bool {
						x := vv.Bool()
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrTimeType:
					if rawValueType == core.PtrTimeType {
						hasAssigned = true
						var x time.Time = rawValue.Interface().(time.Time)
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrFloat64Type:
					if rawValueType.Kind() == reflect.Float64 {
						x := vv.Float()
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrUint64Type:
					if rawValueType.Kind() == reflect.Int64 {
						var x uint64 = uint64(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrInt64Type:
					if rawValueType.Kind() == reflect.Int64 {
						x := vv.Int()
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrFloat32Type:
					if rawValueType.Kind() == reflect.Float64 {
						var x float32 = float32(vv.Float())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrIntType:
					if rawValueType.Kind() == reflect.Int64 {
						var x int = int(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrInt32Type:
					if rawValueType.Kind() == reflect.Int64 {
						var x int32 = int32(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrInt8Type:
					if rawValueType.Kind() == reflect.Int64 {
						var x int8 = int8(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrInt16Type:
					if rawValueType.Kind() == reflect.Int64 {
						var x int16 = int16(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrUintType:
					if rawValueType.Kind() == reflect.Int64 {
						var x uint = uint(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.PtrUint32Type:
					if rawValueType.Kind() == reflect.Int64 {
						var x uint32 = uint32(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.Uint8Type:
					if rawValueType.Kind() == reflect.Int64 {
						var x uint8 = uint8(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.Uint16Type:
					if rawValueType.Kind() == reflect.Int64 {
						var x uint16 = uint16(vv.Int())
						hasAssigned = true
						fieldValue.Set(reflect.ValueOf(&x))
					}
				case core.Complex64Type:
					var x complex64
					err := json.Unmarshal([]byte(vv.String()), &x)
					if err != nil {
						session.Engine.LogError(err)
					} else {
						fieldValue.Set(reflect.ValueOf(&x))
					}
					hasAssigned = true
				case core.Complex128Type:
					var x complex128
					err := json.Unmarshal([]byte(vv.String()), &x)
					if err != nil {
						session.Engine.LogError(err)
					} else {
						fieldValue.Set(reflect.ValueOf(&x))
					}
					hasAssigned = true
				} // switch fieldType
				// default:
				// 	session.Engine.LogError("unsupported type in Scan: ", reflect.TypeOf(v).String())
			} // switch fieldType.Kind()

			// !nashtsai! for value can't be assigned directly fallback to convert to []byte then back to value
			if !hasAssigned {
				data, err := value2Bytes(&rawValue)
				if err == nil {
					session.bytes2Value(table.GetColumn(key), fieldValue, data)
				} else {
					session.Engine.LogError(err.Error())
				}
			}
		}
	}
	return nil

}

func (session *Session) queryPreprocess(sqlStr *string, paramStr ...interface{}) {
	for _, filter := range session.Engine.dialect.Filters() {
		*sqlStr = filter.Do(*sqlStr, session.Engine.dialect, session.Statement.RefTable)
	}

	session.Engine.logSQL(*sqlStr, paramStr)
}

func (session *Session) query(sqlStr string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	session.queryPreprocess(&sqlStr, paramStr...)

	if session.IsAutoCommit {
		return query(session.Db, sqlStr, paramStr...)
	}
	return txQuery(session.Tx, sqlStr, paramStr...)
}

func txQuery(tx *core.Tx, sqlStr string, params ...interface{}) (resultsSlice []map[string][]byte, err error) {
	rows, err := tx.Query(sqlStr, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows2maps(rows)
}

func query(db *core.DB, sqlStr string, params ...interface{}) (resultsSlice []map[string][]byte, err error) {
	s, err := db.Prepare(sqlStr)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	rows, err := s.Query(params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//fmt.Println(rows)
	return rows2maps(rows)
}

// Exec a raw sql and return records as []map[string][]byte
func (session *Session) Query(sqlStr string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	err = session.newDb()
	if err != nil {
		return nil, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	return session.query(sqlStr, paramStr...)
}

// insert one or more beans
func (session *Session) Insert(beans ...interface{}) (int64, error) {
	var affected int64 = 0
	var err error = nil
	err = session.newDb()
	if err != nil {
		return 0, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	for _, bean := range beans {
		sliceValue := reflect.Indirect(reflect.ValueOf(bean))
		if sliceValue.Kind() == reflect.Slice {
			if session.Engine.SupportInsertMany() {
				cnt, err := session.innerInsertMulti(bean)
				if err != nil {
					return affected, err
				}
				affected += cnt
			} else {
				size := sliceValue.Len()
				for i := 0; i < size; i++ {
					cnt, err := session.innerInsert(sliceValue.Index(i).Interface())
					if err != nil {
						return affected, err
					}
					affected += cnt
				}
			}
		} else {
			cnt, err := session.innerInsert(bean)
			if err != nil {
				return affected, err
			}
			affected += cnt
		}
	}

	return affected, err
}

func (session *Session) innerInsertMulti(rowsSlicePtr interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return 0, errors.New("needs a pointer to a slice")
	}

	bean := sliceValue.Index(0).Interface()
	elementValue := rValue(bean)
	//sliceElementType := elementValue.Type()

	table := session.Engine.autoMapType(elementValue)
	session.Statement.RefTable = table

	size := sliceValue.Len()

	colNames := make([]string, 0)
	colMultiPlaces := make([]string, 0)
	var args = make([]interface{}, 0)
	cols := make([]*core.Column, 0)

	for i := 0; i < size; i++ {
		elemValue := sliceValue.Index(i).Interface()
		colPlaces := make([]string, 0)

		// handle BeforeInsertProcessor
		// !nashtsai! does user expect it's same slice to passed closure when using Before()/After() when insert multi??
		for _, closure := range session.beforeClosures {
			closure(elemValue)
		}

		if processor, ok := interface{}(elemValue).(BeforeInsertProcessor); ok {
			processor.BeforeInsert()
		}
		// --

		if i == 0 {
			for _, col := range table.Columns() {
				fieldValue := reflect.Indirect(reflect.ValueOf(elemValue)).FieldByName(col.FieldName)
				if col.IsAutoIncrement && fieldValue.Int() == 0 {
					continue
				}
				if col.MapType == core.ONLYFROMDB {
					continue
				}
				if session.Statement.ColumnStr != "" {
					if _, ok := session.Statement.columnMap[col.Name]; !ok {
						continue
					}
				}
				if (col.IsCreated || col.IsUpdated) && session.Statement.UseAutoTime {
					args = append(args, time.Now())
				} else {
					arg, err := session.value2Interface(col, fieldValue)
					if err != nil {
						return 0, err
					}
					args = append(args, arg)
				}

				colNames = append(colNames, col.Name)
				cols = append(cols, col)
				colPlaces = append(colPlaces, "?")
			}
		} else {
			for _, col := range cols {
				fieldValue := reflect.Indirect(reflect.ValueOf(elemValue)).FieldByName(col.FieldName)
				if col.IsAutoIncrement && fieldValue.Int() == 0 {
					continue
				}
				if col.MapType == core.ONLYFROMDB {
					continue
				}
				if session.Statement.ColumnStr != "" {
					if _, ok := session.Statement.columnMap[col.Name]; !ok {
						continue
					}
				}
				if (col.IsCreated || col.IsUpdated) && session.Statement.UseAutoTime {
					args = append(args, time.Now())
				} else {
					arg, err := session.value2Interface(col, fieldValue)
					if err != nil {
						return 0, err
					}
					args = append(args, arg)
				}

				colPlaces = append(colPlaces, "?")
			}
		}
		colMultiPlaces = append(colMultiPlaces, strings.Join(colPlaces, ", "))
	}
	cleanupProcessorsClosures(&session.beforeClosures)

	statement := fmt.Sprintf("INSERT INTO %v%v%v (%v%v%v) VALUES (%v)",
		session.Engine.QuoteStr(),
		session.Statement.TableName(),
		session.Engine.QuoteStr(),
		session.Engine.QuoteStr(),
		strings.Join(colNames, session.Engine.QuoteStr()+", "+session.Engine.QuoteStr()),
		session.Engine.QuoteStr(),
		strings.Join(colMultiPlaces, "),("))

	res, err := session.exec(statement, args...)
	if err != nil {
		return 0, err
	}

	if cacher := session.Engine.getCacher2(table); cacher != nil && session.Statement.UseCache {
		session.cacheInsert(session.Statement.TableName())
	}

	lenAfterClosures := len(session.afterClosures)
	for i := 0; i < size; i++ {
		elemValue := sliceValue.Index(i).Interface()
		// handle AfterInsertProcessor
		if session.IsAutoCommit {
			// !nashtsai! does user expect it's same slice to passed closure when using Before()/After() when insert multi??
			for _, closure := range session.afterClosures {
				closure(elemValue)
			}
			if processor, ok := interface{}(elemValue).(AfterInsertProcessor); ok {
				processor.AfterInsert()
			}
		} else {
			if lenAfterClosures > 0 {
				if value, has := session.afterInsertBeans[elemValue]; has && value != nil {
					*value = append(*value, session.afterClosures...)
				} else {
					afterClosures := make([]func(interface{}), lenAfterClosures)
					copy(afterClosures, session.afterClosures)
					session.afterInsertBeans[elemValue] = &afterClosures
				}

			} else {
				if _, ok := interface{}(elemValue).(AfterInsertProcessor); ok {
					session.afterInsertBeans[elemValue] = nil
				}
			}
		}
	}
	cleanupProcessorsClosures(&session.afterClosures)
	return res.RowsAffected()
}

// Insert multiple records
func (session *Session) InsertMulti(rowsSlicePtr interface{}) (int64, error) {
	err := session.newDb()
	if err != nil {
		return 0, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	return session.innerInsertMulti(rowsSlicePtr)
}

func (session *Session) byte2Time(col *core.Column, data []byte) (outTime time.Time, outErr error) {
	sdata := strings.TrimSpace(string(data))
	var x time.Time
	var err error

	if sdata == "0000-00-00 00:00:00" ||
		sdata == "0001-01-01 00:00:00" {
	} else if !strings.ContainsAny(sdata, "- :") {
		// time stamp
		sd, err := strconv.ParseInt(sdata, 10, 64)
		if err == nil {
			x = time.Unix(0, sd)
		}
	} else if len(sdata) > 19 {
		x, err = time.Parse(time.RFC3339Nano, sdata)
		if err != nil {
			x, err = time.Parse("2006-01-02 15:04:05.999999999", sdata)
		}
		if err != nil {
			x, err = time.Parse("2006-01-02 15:04:05.9999999 Z07:00", sdata)
		}
	} else if len(sdata) == 19 {
		x, err = time.Parse("2006-01-02 15:04:05", sdata)
	} else if len(sdata) == 10 && sdata[4] == '-' && sdata[7] == '-' {
		x, err = time.Parse("2006-01-02", sdata)
	} else if col.SQLType.Name == core.Time {
		if strings.Contains(sdata, " ") {
			ssd := strings.Split(sdata, " ")
			sdata = ssd[1]
		}

		sdata = strings.TrimSpace(sdata)
		//fmt.Println(sdata)
		if session.Engine.dialect.DBType() == core.MYSQL && len(sdata) > 8 {
			sdata = sdata[len(sdata)-8:]
		}
		//fmt.Println(sdata)

		st := fmt.Sprintf("2006-01-02 %v", sdata)
		x, err = time.Parse("2006-01-02 15:04:05", st)
	} else {
		outErr = errors.New(fmt.Sprintf("unsupported time format %v", sdata))
		return
	}
	if err != nil {
		outErr = errors.New(fmt.Sprintf("unsupported time format %v: %v", sdata, err))
		return
	}
	outTime = x
	return
}

// convert a db data([]byte) to a field value
func (session *Session) bytes2Value(col *core.Column, fieldValue *reflect.Value, data []byte) error {
	if structConvert, ok := fieldValue.Addr().Interface().(core.Conversion); ok {
		return structConvert.FromDB(data)
	}

	var v interface{}
	key := col.Name
	fieldType := fieldValue.Type()

	//fmt.Println("column name:", key, ", fieldType:", fieldType.String())
	switch fieldType.Kind() {
	case reflect.Complex64, reflect.Complex128:
		x := reflect.New(fieldType)

		err := json.Unmarshal(data, x.Interface())
		if err != nil {
			session.Engine.LogError(err)
			return err
		}
		fieldValue.Set(x.Elem())
	case reflect.Slice, reflect.Array, reflect.Map:
		v = data
		t := fieldType.Elem()
		k := t.Kind()
		if col.SQLType.IsText() {
			x := reflect.New(fieldType)
			err := json.Unmarshal(data, x.Interface())
			if err != nil {
				session.Engine.LogError(err)
				return err
			}
			fieldValue.Set(x.Elem())
		} else if col.SQLType.IsBlob() {
			if k == reflect.Uint8 {
				fieldValue.Set(reflect.ValueOf(v))
			} else {
				x := reflect.New(fieldType)
				err := json.Unmarshal(data, x.Interface())
				if err != nil {
					session.Engine.LogError(err)
					return err
				}
				fieldValue.Set(x.Elem())
			}
		} else {
			return ErrUnSupportedType
		}
	case reflect.String:
		fieldValue.SetString(string(data))
	case reflect.Bool:
		d := string(data)
		v, err := strconv.ParseBool(d)
		if err != nil {
			return fmt.Errorf("arg %v as bool: %s", key, err.Error())
		}
		fieldValue.Set(reflect.ValueOf(v))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sdata := string(data)
		var x int64
		var err error
		// for mysql, when use bit, it returned \x01
		if col.SQLType.Name == core.Bit &&
			session.Engine.dialect.DBType() == core.MYSQL {
			if len(data) == 1 {
				x = int64(data[0])
			} else {
				x = 0
			}
			//fmt.Println("######", x, data)
		} else if strings.HasPrefix(sdata, "0x") {
			x, err = strconv.ParseInt(sdata, 16, 64)
		} else if strings.HasPrefix(sdata, "0") {
			x, err = strconv.ParseInt(sdata, 8, 64)
		} else if strings.ToLower(sdata) == "true" {
			x = 1
		} else if strings.ToLower(sdata) == "false" {
			x = 0
		} else {
			x, err = strconv.ParseInt(sdata, 10, 64)
		}
		if err != nil {
			return fmt.Errorf("arg %v as int: %s", key, err.Error())
		}
		fieldValue.SetInt(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return fmt.Errorf("arg %v as float64: %s", key, err.Error())
		}
		fieldValue.SetFloat(x)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		x, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return fmt.Errorf("arg %v as int: %s", key, err.Error())
		}
		fieldValue.SetUint(x)
	//Currently only support Time type
	case reflect.Struct:
		if fieldType == core.TimeType {
			x, err := session.byte2Time(col, data)
			if err != nil {
				return err
			}
			v = x
			fieldValue.Set(reflect.ValueOf(v))
		} else if session.Statement.UseCascade {
			table := session.Engine.autoMapType(*fieldValue)
			if table != nil {
				x, err := strconv.ParseInt(string(data), 10, 64)
				if err != nil {
					return fmt.Errorf("arg %v as int: %s", key, err.Error())
				}
				if x != 0 {
					// !nashtsai! TODO for hasOne relationship, it's preferred to use join query for eager fetch
					// however, also need to consider adding a 'lazy' attribute to xorm tag which allow hasOne
					// property to be fetched lazily
					structInter := reflect.New(fieldValue.Type())
					newsession := session.Engine.NewSession()
					defer newsession.Close()
					has, err := newsession.Id(x).Get(structInter.Interface())
					if err != nil {
						return err
					}
					if has {
						v = structInter.Elem().Interface()
						fieldValue.Set(reflect.ValueOf(v))
					} else {
						return errors.New("cascade obj is not exist!")
					}
				}
			} else {
				return fmt.Errorf("unsupported struct type in Scan: %s", fieldValue.Type().String())
			}
		}
	case reflect.Ptr:
		// !nashtsai! TODO merge duplicated codes above
		//typeStr := fieldType.String()
		switch fieldType {
		// case "*string":
		case core.PtrStringType:
			x := string(data)
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*bool":
		case core.PtrBoolType:
			d := string(data)
			v, err := strconv.ParseBool(d)
			if err != nil {
				return fmt.Errorf("arg %v as bool: %s", key, err.Error())
			}
			fieldValue.Set(reflect.ValueOf(&v))
		// case "*complex64":
		case core.PtrComplex64Type:
			var x complex64
			err := json.Unmarshal(data, &x)
			if err != nil {
				session.Engine.LogError(err)
				return err
			}
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*complex128":
		case core.PtrComplex128Type:
			var x complex128
			err := json.Unmarshal(data, &x)
			if err != nil {
				session.Engine.LogError(err)
				return err
			}
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*float64":
		case core.PtrFloat64Type:
			x, err := strconv.ParseFloat(string(data), 64)
			if err != nil {
				return fmt.Errorf("arg %v as float64: %s", key, err.Error())
			}
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*float32":
		case core.PtrFloat32Type:
			var x float32
			x1, err := strconv.ParseFloat(string(data), 32)
			if err != nil {
				return fmt.Errorf("arg %v as float32: %s", key, err.Error())
			}
			x = float32(x1)
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*time.Time":
		case core.PtrTimeType:
			x, err := session.byte2Time(col, data)
			if err != nil {
				return err
			}
			v = x
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*uint64":
		case core.PtrUint64Type:
			var x uint64
			x, err := strconv.ParseUint(string(data), 10, 64)
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*uint":
		case core.PtrUintType:
			var x uint
			x1, err := strconv.ParseUint(string(data), 10, 64)
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			x = uint(x1)
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*uint32":
		case core.PtrUint32Type:
			var x uint32
			x1, err := strconv.ParseUint(string(data), 10, 64)
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			x = uint32(x1)
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*uint8":
		case core.PtrUint8Type:
			var x uint8
			x1, err := strconv.ParseUint(string(data), 10, 64)
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			x = uint8(x1)
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*uint16":
		case core.PtrUint16Type:
			var x uint16
			x1, err := strconv.ParseUint(string(data), 10, 64)
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			x = uint16(x1)
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*int64":
		case core.PtrInt64Type:
			sdata := string(data)
			var x int64
			var err error
			// for mysql, when use bit, it returned \x01
			if col.SQLType.Name == core.Bit &&
				strings.Contains(session.Engine.DriverName(), "mysql") {
				if len(data) == 1 {
					x = int64(data[0])
				} else {
					x = 0
				}
				//fmt.Println("######", x, data)
			} else if strings.HasPrefix(sdata, "0x") {
				x, err = strconv.ParseInt(sdata, 16, 64)
			} else if strings.HasPrefix(sdata, "0") {
				x, err = strconv.ParseInt(sdata, 8, 64)
			} else {
				x, err = strconv.ParseInt(sdata, 10, 64)
			}
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*int":
		case core.PtrIntType:
			sdata := string(data)
			var x int
			var x1 int64
			var err error
			// for mysql, when use bit, it returned \x01
			if col.SQLType.Name == core.Bit &&
				strings.Contains(session.Engine.DriverName(), "mysql") {
				if len(data) == 1 {
					x = int(data[0])
				} else {
					x = 0
				}
				//fmt.Println("######", x, data)
			} else if strings.HasPrefix(sdata, "0x") {
				x1, err = strconv.ParseInt(sdata, 16, 64)
				x = int(x1)
			} else if strings.HasPrefix(sdata, "0") {
				x1, err = strconv.ParseInt(sdata, 8, 64)
				x = int(x1)
			} else {
				x1, err = strconv.ParseInt(sdata, 10, 64)
				x = int(x1)
			}
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*int32":
		case core.PtrInt32Type:
			sdata := string(data)
			var x int32
			var x1 int64
			var err error
			// for mysql, when use bit, it returned \x01
			if col.SQLType.Name == core.Bit &&
				session.Engine.dialect.DBType() == core.MYSQL {
				if len(data) == 1 {
					x = int32(data[0])
				} else {
					x = 0
				}
				//fmt.Println("######", x, data)
			} else if strings.HasPrefix(sdata, "0x") {
				x1, err = strconv.ParseInt(sdata, 16, 64)
				x = int32(x1)
			} else if strings.HasPrefix(sdata, "0") {
				x1, err = strconv.ParseInt(sdata, 8, 64)
				x = int32(x1)
			} else {
				x1, err = strconv.ParseInt(sdata, 10, 64)
				x = int32(x1)
			}
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*int8":
		case core.PtrInt8Type:
			sdata := string(data)
			var x int8
			var x1 int64
			var err error
			// for mysql, when use bit, it returned \x01
			if col.SQLType.Name == core.Bit &&
				strings.Contains(session.Engine.DriverName(), "mysql") {
				if len(data) == 1 {
					x = int8(data[0])
				} else {
					x = 0
				}
				//fmt.Println("######", x, data)
			} else if strings.HasPrefix(sdata, "0x") {
				x1, err = strconv.ParseInt(sdata, 16, 64)
				x = int8(x1)
			} else if strings.HasPrefix(sdata, "0") {
				x1, err = strconv.ParseInt(sdata, 8, 64)
				x = int8(x1)
			} else {
				x1, err = strconv.ParseInt(sdata, 10, 64)
				x = int8(x1)
			}
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			fieldValue.Set(reflect.ValueOf(&x))
		// case "*int16":
		case core.PtrInt16Type:
			sdata := string(data)
			var x int16
			var x1 int64
			var err error
			// for mysql, when use bit, it returned \x01
			if col.SQLType.Name == core.Bit &&
				strings.Contains(session.Engine.DriverName(), "mysql") {
				if len(data) == 1 {
					x = int16(data[0])
				} else {
					x = 0
				}
				//fmt.Println("######", x, data)
			} else if strings.HasPrefix(sdata, "0x") {
				x1, err = strconv.ParseInt(sdata, 16, 64)
				x = int16(x1)
			} else if strings.HasPrefix(sdata, "0") {
				x1, err = strconv.ParseInt(sdata, 8, 64)
				x = int16(x1)
			} else {
				x1, err = strconv.ParseInt(sdata, 10, 64)
				x = int16(x1)
			}
			if err != nil {
				return fmt.Errorf("arg %v as int: %s", key, err.Error())
			}
			fieldValue.Set(reflect.ValueOf(&x))
		default:
			return fmt.Errorf("unsupported type in Scan: %s", reflect.TypeOf(v).String())
		}
	default:
		return fmt.Errorf("unsupported type in Scan: %s", reflect.TypeOf(v).String())
	}

	return nil
}

// convert a field value of a struct to interface for put into db
func (session *Session) value2Interface(col *core.Column, fieldValue reflect.Value) (interface{}, error) {
	if fieldValue.CanAddr() {
		if fieldConvert, ok := fieldValue.Addr().Interface().(core.Conversion); ok {
			data, err := fieldConvert.ToDB()
			if err != nil {
				return 0, err
			} else {
				return string(data), nil
			}
		}
	}
	fieldType := fieldValue.Type()
	k := fieldType.Kind()
	if k == reflect.Ptr {
		if fieldValue.IsNil() {
			return nil, nil
		} else if !fieldValue.IsValid() {
			session.Engine.LogWarn("the field[", col.FieldName, "] is invalid")
			return nil, nil
		} else {
			// !nashtsai! deference pointer type to instance type
			fieldValue = fieldValue.Elem()
			fieldType = fieldValue.Type()
			k = fieldType.Kind()
		}
	}

	switch k {
	case reflect.Bool:
		if fieldValue.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	case reflect.String:
		return fieldValue.String(), nil
	case reflect.Struct:
		if fieldType == core.TimeType {
			t := fieldValue.Interface().(time.Time)
			if session.Engine.dialect.DBType() == core.MSSQL {
				if t.IsZero() {
					return nil, nil
				}
			}
			if col.SQLType.Name == core.Time {
				//s := fieldValue.Interface().(time.Time).Format("2006-01-02 15:04:05 -0700")
				s := fieldValue.Interface().(time.Time).Format(time.RFC3339)
				return s[11:19], nil
			} else if col.SQLType.Name == core.Date {
				return fieldValue.Interface().(time.Time).Format("2006-01-02"), nil
			} else if col.SQLType.Name == core.TimeStampz {
				if session.Engine.dialect.DBType() == core.MSSQL {
					tf := t.Format("2006-01-02T15:04:05.9999999Z07:00")
					return tf, nil
				}
				return fieldValue.Interface().(time.Time).Format(time.RFC3339Nano), nil
			}
			return fieldValue.Interface(), nil
		}
		if fieldTable, ok := session.Engine.Tables[fieldValue.Type()]; ok {
			if len(fieldTable.PrimaryKeys) == 1 {
				pkField := reflect.Indirect(fieldValue).FieldByName(fieldTable.PKColumns()[0].FieldName)
				return pkField.Interface(), nil
			} else {
				return 0, fmt.Errorf("no primary key for col %v", col.Name)
			}
		} else {
			return 0, fmt.Errorf("Unsupported type %v\n", fieldValue.Type())
		}
	case reflect.Complex64, reflect.Complex128:
		bytes, err := json.Marshal(fieldValue.Interface())
		if err != nil {
			session.Engine.LogError(err)
			return 0, err
		}
		return string(bytes), nil
	case reflect.Array, reflect.Slice, reflect.Map:
		if !fieldValue.IsValid() {
			return fieldValue.Interface(), nil
		}

		if col.SQLType.IsText() {
			bytes, err := json.Marshal(fieldValue.Interface())
			if err != nil {
				session.Engine.LogError(err)
				return 0, err
			}
			return string(bytes), nil
		} else if col.SQLType.IsBlob() {
			var bytes []byte
			var err error
			if (k == reflect.Array || k == reflect.Slice) &&
				(fieldValue.Type().Elem().Kind() == reflect.Uint8) {
				bytes = fieldValue.Bytes()
			} else {
				bytes, err = json.Marshal(fieldValue.Interface())
				if err != nil {
					session.Engine.LogError(err)
					return 0, err
				}
			}
			return bytes, nil
		} else {
			return nil, ErrUnSupportedType
		}
	default:
		return fieldValue.Interface(), nil
	}
}

func (session *Session) innerInsert(bean interface{}) (int64, error) {
	table := session.Engine.autoMap(bean)
	session.Statement.RefTable = table

	// handle BeforeInsertProcessor
	for _, closure := range session.beforeClosures {
		closure(bean)
	}
	cleanupProcessorsClosures(&session.beforeClosures) // cleanup after used

	if processor, ok := interface{}(bean).(BeforeInsertProcessor); ok {
		processor.BeforeInsert()
	}
	// --

	colNames, args, err := genCols(table, session, bean, false, false)
	if err != nil {
		return 0, err
	}

	colPlaces := strings.Repeat("?, ", len(colNames))
	//fmt.Println(colNames, args)
	colPlaces = colPlaces[0 : len(colPlaces)-2]

	sqlStr := fmt.Sprintf("INSERT INTO %v%v%v (%v%v%v) VALUES (%v)",
		session.Engine.QuoteStr(),
		session.Statement.TableName(),
		session.Engine.QuoteStr(),
		session.Engine.QuoteStr(),
		strings.Join(colNames, session.Engine.Quote(", ")),
		session.Engine.QuoteStr(),
		colPlaces)

	handleAfterInsertProcessorFunc := func(bean interface{}) {

		if session.IsAutoCommit {
			for _, closure := range session.afterClosures {
				closure(bean)
			}
			if processor, ok := interface{}(bean).(AfterInsertProcessor); ok {
				processor.AfterInsert()
			}
		} else {
			lenAfterClosures := len(session.afterClosures)
			if lenAfterClosures > 0 {
				if value, has := session.afterInsertBeans[bean]; has && value != nil {
					*value = append(*value, session.afterClosures...)
				} else {
					afterClosures := make([]func(interface{}), lenAfterClosures)
					copy(afterClosures, session.afterClosures)
					session.afterInsertBeans[bean] = &afterClosures
				}

			} else {
				if _, ok := interface{}(bean).(AfterInsertProcessor); ok {
					session.afterInsertBeans[bean] = nil
				}
			}
		}
		cleanupProcessorsClosures(&session.afterClosures) // cleanup after used
	}

	// for postgres, many of them didn't implement lastInsertId, so we should
	// implemented it ourself.

	if session.Engine.DriverName() != core.POSTGRES || table.AutoIncrement == "" {
		res, err := session.exec(sqlStr, args...)
		if err != nil {
			return 0, err
		} else {
			handleAfterInsertProcessorFunc(bean)
		}

		if cacher := session.Engine.getCacher2(table); cacher != nil && session.Statement.UseCache {
			session.cacheInsert(session.Statement.TableName())
		}

		if table.Version != "" && session.Statement.checkVersion {
			verValue, err := table.VersionColumn().ValueOf(bean)
			if err != nil {
				session.Engine.LogError(err)
			} else if verValue.IsValid() && verValue.CanSet() {
				verValue.SetInt(1)
			}
		}

		if table.AutoIncrement == "" {
			return res.RowsAffected()
		}

		var id int64 = 0
		id, err = res.LastInsertId()
		if err != nil || id <= 0 {
			return res.RowsAffected()
		}

		aiValue, err := table.AutoIncrColumn().ValueOf(bean)
		if err != nil {
			session.Engine.LogError(err)
		}

		if aiValue == nil || !aiValue.IsValid() /*|| aiValue.Int() != 0*/ || !aiValue.CanSet() {
			return res.RowsAffected()
		}

		var v interface{} = id
		switch aiValue.Type().Kind() {
		case reflect.Int32:
			v = int32(id)
		case reflect.Int:
			v = int(id)
		case reflect.Uint32:
			v = uint32(id)
		case reflect.Uint64:
			v = uint64(id)
		case reflect.Uint:
			v = uint(id)
		}
		aiValue.Set(reflect.ValueOf(v))

		return res.RowsAffected()
	} else {
		//assert table.AutoIncrement != ""
		sqlStr = sqlStr + " RETURNING " + session.Engine.Quote(table.AutoIncrement)
		res, err := session.query(sqlStr, args...)

		if err != nil {
			return 0, err
		} else {
			handleAfterInsertProcessorFunc(bean)
		}

		if cacher := session.Engine.getCacher2(table); cacher != nil && session.Statement.UseCache {
			session.cacheInsert(session.Statement.TableName())
		}

		if table.Version != "" && session.Statement.checkVersion {
			verValue, err := table.VersionColumn().ValueOf(bean)
			if err != nil {
				session.Engine.LogError(err)
			} else if verValue.IsValid() && verValue.CanSet() {
				verValue.SetInt(1)
			}
		}

		if len(res) < 1 {
			return 0, errors.New("insert no error but not returned id")
		}

		idByte := res[0][table.AutoIncrement]
		id, err := strconv.ParseInt(string(idByte), 10, 64)
		if err != nil {
			return 1, err
		}

		aiValue, err := table.AutoIncrColumn().ValueOf(bean)
		if err != nil {
			session.Engine.LogError(err)
		}

		if aiValue == nil || !aiValue.IsValid() /*|| aiValue. != 0*/ || !aiValue.CanSet() {
			return 1, nil
		}

		var v interface{} = id
		switch aiValue.Type().Kind() {
		case reflect.Int32:
			v = int32(id)
		case reflect.Int:
			v = int(id)
		case reflect.Uint32:
			v = uint32(id)
		case reflect.Uint64:
			v = uint64(id)
		case reflect.Uint:
			v = uint(id)
		}
		aiValue.Set(reflect.ValueOf(v))

		return 1, nil
	}
}

// Method InsertOne insert only one struct into database as a record.
// The in parameter bean must a struct or a point to struct. The return
// parameter is inserted and error
func (session *Session) InsertOne(bean interface{}) (int64, error) {
	err := session.newDb()
	if err != nil {
		return 0, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	return session.innerInsert(bean)
}

func (statement *Statement) convertUpdateSql(sqlStr string) (string, string) {
	if statement.RefTable == nil || len(statement.RefTable.PrimaryKeys) != 1 {
		return "", ""
	}
	sqls := splitNNoCase(sqlStr, "where", 2)
	if len(sqls) != 2 {
		if len(sqls) == 1 {
			return sqls[0], fmt.Sprintf("SELECT %v FROM %v",
				statement.Engine.Quote(statement.RefTable.PrimaryKeys[0]),
				statement.Engine.Quote(statement.RefTable.Name))
		}
		return "", ""
	}

	var whereStr = sqls[1]

	//TODO: for postgres only, if any other database?
	var paraStr string
	if statement.Engine.dialect.DBType() == core.POSTGRES {
		paraStr = "$"
	} else if statement.Engine.dialect.DBType() == core.MSSQL {
		paraStr = ":"
	}

	if paraStr != "" {
		if strings.Contains(sqls[1], paraStr) {
			dollers := strings.Split(sqls[1], paraStr)
			whereStr = dollers[0]
			for i, c := range dollers[1:] {
				ccs := strings.SplitN(c, " ", 2)
				whereStr += fmt.Sprintf(paraStr+"%v %v", i+1, ccs[1])
			}
		}
	}

	return sqls[0], fmt.Sprintf("SELECT %v FROM %v WHERE %v",
		statement.Engine.Quote(statement.RefTable.PrimaryKeys[0]), statement.Engine.Quote(statement.TableName()),
		whereStr)
}

func (session *Session) cacheInsert(tables ...string) error {
	if session.Statement.RefTable == nil || len(session.Statement.RefTable.PrimaryKeys) != 1 {
		return ErrCacheFailed
	}

	table := session.Statement.RefTable
	cacher := session.Engine.getCacher2(table)

	for _, t := range tables {
		session.Engine.LogDebug("cache clear:", t)
		cacher.ClearIds(t)
	}

	return nil
}

func (session *Session) cacheUpdate(sqlStr string, args ...interface{}) error {
	if session.Statement.RefTable == nil || len(session.Statement.RefTable.PrimaryKeys) != 1 {
		return ErrCacheFailed
	}

	oldhead, newsql := session.Statement.convertUpdateSql(sqlStr)
	if newsql == "" {
		return ErrCacheFailed
	}
	for _, filter := range session.Engine.dialect.Filters() {
		newsql = filter.Do(newsql, session.Engine.dialect, session.Statement.RefTable)
	}
	session.Engine.LogDebug("[xorm:cacheUpdate] new sql", oldhead, newsql)

	var nStart int
	if len(args) > 0 {
		if strings.Index(sqlStr, "?") > -1 {
			nStart = strings.Count(oldhead, "?")
		} else {
			// only for pq, TODO: if any other databse?
			nStart = strings.Count(oldhead, "$")
		}
	}
	table := session.Statement.RefTable
	cacher := session.Engine.getCacher2(table)
	tableName := session.Statement.TableName()
	session.Engine.LogDebug("[xorm:cacheUpdate] get cache sql", newsql, args[nStart:])
	ids, err := core.GetCacheSql(cacher, tableName, newsql, args[nStart:])
	if err != nil {
		resultsSlice, err := session.query(newsql, args[nStart:]...)
		if err != nil {
			return err
		}
		session.Engine.LogDebug("[xorm:cacheUpdate] find updated id", resultsSlice)

		ids = make([]core.PK, 0)
		if len(resultsSlice) > 0 {
			for _, data := range resultsSlice {
				var id int64
				if v, ok := data[session.Statement.RefTable.PrimaryKeys[0]]; !ok {
					return errors.New("no id")
				} else {
					id, err = strconv.ParseInt(string(v), 10, 64)
					if err != nil {
						return err
					}
				}
				ids = append(ids, core.PK{id})
			}
		}
	} /*else {
	    session.Engine.LogDebug("[xorm:cacheUpdate] del cached sql:", tableName, newsql, args)
	    cacher.DelIds(tableName, genSqlKey(newsql, args))
	}*/

	for _, id := range ids {
		sid, err := id.ToString()
		if err != nil {
			return err
		}
		if bean := cacher.GetBean(tableName, sid); bean != nil {
			sqls := splitNNoCase(sqlStr, "where", 2)
			if len(sqls) == 0 || len(sqls) > 2 {
				return ErrCacheFailed
			}

			sqls = splitNNoCase(sqls[0], "set", 2)
			if len(sqls) != 2 {
				return ErrCacheFailed
			}
			kvs := strings.Split(strings.TrimSpace(sqls[1]), ",")
			for idx, kv := range kvs {
				sps := strings.SplitN(kv, "=", 2)
				sps2 := strings.Split(sps[0], ".")
				colName := sps2[len(sps2)-1]
				if strings.Contains(colName, "`") {
					colName = strings.TrimSpace(strings.Replace(colName, "`", "", -1))
				} else if strings.Contains(colName, session.Engine.QuoteStr()) {
					colName = strings.TrimSpace(strings.Replace(colName, session.Engine.QuoteStr(), "", -1))
				} else {
					session.Engine.LogDebug("[xorm:cacheUpdate] cannot find column", tableName, colName)
					return ErrCacheFailed
				}

				if col := table.GetColumn(colName); col != nil {
					fieldValue, err := col.ValueOf(bean)
					if err != nil {
						session.Engine.LogError(err)
					} else {
						session.Engine.LogDebug("[xorm:cacheUpdate] set bean field", bean, colName, fieldValue.Interface())
						if col.IsVersion && session.Statement.checkVersion {
							fieldValue.SetInt(fieldValue.Int() + 1)
						} else {
							fieldValue.Set(reflect.ValueOf(args[idx]))
						}
					}
				} else {
					session.Engine.LogError("[xorm:cacheUpdate] ERROR: column %v is not table %v's",
						colName, table.Name)
				}
			}

			session.Engine.LogDebug("[xorm:cacheUpdate] update cache", tableName, id, bean)
			cacher.PutBean(tableName, sid, bean)
		}
	}
	session.Engine.LogDebug("[xorm:cacheUpdate] clear cached table sql:", tableName)
	cacher.ClearIds(tableName)
	return nil
}

// Update records, bean's non-empty fields are updated contents,
// condiBean' non-empty filds are conditions
// CAUTION:
//        1.bool will defaultly be updated content nor conditions
//         You should call UseBool if you have bool to use.
//        2.float32 & float64 may be not inexact as conditions
func (session *Session) Update(bean interface{}, condiBean ...interface{}) (int64, error) {
	err := session.newDb()
	if err != nil {
		return 0, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	t := rType(bean)

	var colNames []string
	var args []interface{}
	var table *core.Table

	// handle before update processors
	for _, closure := range session.beforeClosures {
		closure(bean)
	}
	cleanupProcessorsClosures(&session.beforeClosures) // cleanup after used
	if processor, ok := interface{}(bean).(BeforeUpdateProcessor); ok {
		processor.BeforeUpdate()
	}
	// --

	if t.Kind() == reflect.Struct {
		table = session.Engine.autoMap(bean)
		session.Statement.RefTable = table

		if session.Statement.ColumnStr == "" {
			colNames, args = buildConditions(session.Engine, table, bean, false, false,
				false, false, session.Statement.allUseBool, session.Statement.useAllCols,
				session.Statement.mustColumnMap)
		} else {
			colNames, args, err = genCols(table, session, bean, true, true)
			if err != nil {
				return 0, err
			}
		}
	} else if t.Kind() == reflect.Map {
		if session.Statement.RefTable == nil {
			return 0, ErrTableNotFound
		}
		table = session.Statement.RefTable
		colNames = make([]string, 0)
		args = make([]interface{}, 0)
		bValue := reflect.Indirect(reflect.ValueOf(bean))

		for _, v := range bValue.MapKeys() {
			colNames = append(colNames, session.Engine.Quote(v.String())+" = ?")
			args = append(args, bValue.MapIndex(v).Interface())
		}
	} else {
		return 0, ErrParamsType
	}

	if session.Statement.UseAutoTime && table.Updated != "" {
		colNames = append(colNames, session.Engine.Quote(table.Updated)+" = ?")
		args = append(args, time.Now())
	}

	//for update action to like "column = column + ?"
	incColumns := session.Statement.getInc()
	for k, v := range incColumns {
		colNames = append(colNames, k+" = "+k+" + ?")
		args = append(args, v)
	}
	var condiColNames []string
	var condiArgs []interface{}

	if len(condiBean) > 0 {
		condiColNames, condiArgs = buildConditions(session.Engine, session.Statement.RefTable, condiBean[0], true, true,
			false, true, session.Statement.allUseBool, session.Statement.useAllCols,
			session.Statement.mustColumnMap)
	}

	var condition = ""
	session.Statement.processIdParam()
	st := session.Statement
	defer session.Statement.Init()
	if st.WhereStr != "" {
		condition = fmt.Sprintf("%v", st.WhereStr)
	}

	if condition == "" {
		if len(condiColNames) > 0 {
			condition = fmt.Sprintf("%v", strings.Join(condiColNames, " AND "))
		}
	} else {
		if len(condiColNames) > 0 {
			condition = fmt.Sprintf("(%v) AND (%v)", condition, strings.Join(condiColNames, " AND "))
		}
	}

	var sqlStr, inSql string
	var inArgs []interface{}
	doIncVer := false
	var verValue *reflect.Value
	if table.Version != "" && session.Statement.checkVersion {
		if condition != "" {
			condition = fmt.Sprintf("WHERE (%v) AND %v = ?", condition,
				session.Engine.Quote(table.Version))
		} else {
			condition = fmt.Sprintf("WHERE %v = ?", session.Engine.Quote(table.Version))
		}
		inSql, inArgs = session.Statement.genInSql()
		if len(inSql) > 0 {
			if condition != "" {
				condition += " AND " + inSql
			} else {
				condition = "WHERE " + inSql
			}
		}

		sqlStr = fmt.Sprintf("UPDATE %v SET %v, %v %v",
			session.Engine.Quote(session.Statement.TableName()),
			strings.Join(colNames, ", "),
			session.Engine.Quote(table.Version)+" = "+session.Engine.Quote(table.Version)+" + 1",
			condition)

		verValue, err = table.VersionColumn().ValueOf(bean)
		if err != nil {
			return 0, err
		}

		condiArgs = append(condiArgs, verValue.Interface())
		doIncVer = true
	} else {
		if condition != "" {
			condition = "WHERE " + condition
		}
		inSql, inArgs = session.Statement.genInSql()
		if len(inSql) > 0 {
			if condition != "" {
				condition += " AND " + inSql
			} else {
				condition = "WHERE " + inSql
			}
		}

		sqlStr = fmt.Sprintf("UPDATE %v SET %v %v",
			session.Engine.Quote(session.Statement.TableName()),
			strings.Join(colNames, ", "),
			condition)
	}

	args = append(args, st.Params...)
	args = append(args, inArgs...)
	args = append(args, condiArgs...)

	res, err := session.exec(sqlStr, args...)
	if err != nil {
		return 0, err
	} else if doIncVer {
		verValue.SetInt(verValue.Int() + 1)
	}

	if cacher := session.Engine.getCacher2(table); cacher != nil && session.Statement.UseCache {
		//session.cacheUpdate(sqlStr, args...)
		cacher.ClearIds(session.Statement.TableName())
		cacher.ClearBeans(session.Statement.TableName())
	}

	// handle after update processors
	if session.IsAutoCommit {
		for _, closure := range session.afterClosures {
			closure(bean)
		}
		if processor, ok := interface{}(bean).(AfterUpdateProcessor); ok {
			session.Engine.LogDebug(session.Statement.TableName(), " has after update processor")
			processor.AfterUpdate()
		}
	} else {
		lenAfterClosures := len(session.afterClosures)
		if lenAfterClosures > 0 {
			if value, has := session.afterUpdateBeans[bean]; has && value != nil {
				*value = append(*value, session.afterClosures...)
			} else {
				afterClosures := make([]func(interface{}), lenAfterClosures)
				copy(afterClosures, session.afterClosures)
				session.afterUpdateBeans[bean] = &afterClosures
			}

		} else {
			if _, ok := interface{}(bean).(AfterInsertProcessor); ok {
				session.afterUpdateBeans[bean] = nil
			}
		}
	}
	cleanupProcessorsClosures(&session.afterClosures) // cleanup after used
	// --

	return res.RowsAffected()
}

func (session *Session) cacheDelete(sqlStr string, args ...interface{}) error {
	if session.Statement.RefTable == nil || len(session.Statement.RefTable.PrimaryKeys) != 1 {
		return ErrCacheFailed
	}

	for _, filter := range session.Engine.dialect.Filters() {
		sqlStr = filter.Do(sqlStr, session.Engine.dialect, session.Statement.RefTable)
	}

	newsql := session.Statement.convertIdSql(sqlStr)
	if newsql == "" {
		return ErrCacheFailed
	}

	cacher := session.Engine.getCacher2(session.Statement.RefTable)
	tableName := session.Statement.TableName()
	ids, err := core.GetCacheSql(cacher, tableName, newsql, args)
	if err != nil {
		resultsSlice, err := session.query(newsql, args...)
		if err != nil {
			return err
		}
		ids = make([]core.PK, 0)
		if len(resultsSlice) > 0 {
			for _, data := range resultsSlice {
				var id int64
				if v, ok := data[session.Statement.RefTable.PrimaryKeys[0]]; !ok {
					return errors.New("no id")
				} else {
					id, err = strconv.ParseInt(string(v), 10, 64)
					if err != nil {
						return err
					}
				}
				ids = append(ids, core.PK{id})
			}
		}
	} /*else {
	    session.Engine.LogDebug("delete cache sql %v", newsql)
	    cacher.DelIds(tableName, genSqlKey(newsql, args))
	}*/

	for _, id := range ids {
		session.Engine.LogDebug("[xorm:cacheDelete] delete cache obj", tableName, id)
		sid, err := id.ToString()
		if err != nil {
			return err
		}
		cacher.DelBean(tableName, sid)
	}
	session.Engine.LogDebug("[xorm:cacheDelete] clear cache table", tableName)
	cacher.ClearIds(tableName)
	return nil
}

// Delete records, bean's non-empty fields are conditions
func (session *Session) Delete(bean interface{}) (int64, error) {
	err := session.newDb()
	if err != nil {
		return 0, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	// handle before delete processors
	for _, closure := range session.beforeClosures {
		closure(bean)
	}
	cleanupProcessorsClosures(&session.beforeClosures)

	if processor, ok := interface{}(bean).(BeforeDeleteProcessor); ok {
		processor.BeforeDelete()
	}
	// --

	table := session.Engine.autoMap(bean)
	session.Statement.RefTable = table
	colNames, args := buildConditions(session.Engine, table, bean, true, true,
		false, true, session.Statement.allUseBool, session.Statement.useAllCols,
		session.Statement.mustColumnMap)

	var condition = ""

	session.Statement.processIdParam()
	if session.Statement.WhereStr != "" {
		condition = session.Statement.WhereStr
		if len(colNames) > 0 {
			condition += " AND " + strings.Join(colNames, " AND ")
		}
	} else {
		condition = strings.Join(colNames, " AND ")
	}
	inSql, inArgs := session.Statement.genInSql()
	if len(inSql) > 0 {
		if len(condition) > 0 {
			condition += " AND "
		}
		condition += inSql
		args = append(args, inArgs...)
	}
	if len(condition) == 0 {
		return 0, ErrNeedDeletedCond
	}

	sqlStr := fmt.Sprintf("DELETE FROM %v WHERE %v",
		session.Engine.Quote(session.Statement.TableName()), condition)

	args = append(session.Statement.Params, args...)

	if cacher := session.Engine.getCacher2(session.Statement.RefTable); cacher != nil && session.Statement.UseCache {
		session.cacheDelete(sqlStr, args...)
	}

	res, err := session.exec(sqlStr, args...)
	if err != nil {
		return 0, err
	}

	// handle after delete processors
	if session.IsAutoCommit {
		for _, closure := range session.afterClosures {
			closure(bean)
		}
		if processor, ok := interface{}(bean).(AfterDeleteProcessor); ok {
			processor.AfterDelete()
		}
	} else {
		lenAfterClosures := len(session.afterClosures)
		if lenAfterClosures > 0 {
			if value, has := session.afterDeleteBeans[bean]; has && value != nil {
				*value = append(*value, session.afterClosures...)
			} else {
				afterClosures := make([]func(interface{}), lenAfterClosures)
				copy(afterClosures, session.afterClosures)
				session.afterDeleteBeans[bean] = &afterClosures
			}

		} else {
			if _, ok := interface{}(bean).(AfterInsertProcessor); ok {
				session.afterDeleteBeans[bean] = nil
			}
		}
	}
	cleanupProcessorsClosures(&session.afterClosures)
	// --

	return res.RowsAffected()
}

func genCols(table *core.Table, session *Session, bean interface{}, useCol bool, includeQuote bool) ([]string, []interface{}, error) {
	colNames := make([]string, 0)
	args := make([]interface{}, 0)

	for _, col := range table.Columns() {
		lColName := strings.ToLower(col.Name)
		if useCol && !col.IsVersion && !col.IsCreated && !col.IsUpdated {
			if _, ok := session.Statement.columnMap[lColName]; !ok {
				continue
			}
		}
		if col.MapType == core.ONLYFROMDB {
			continue
		}

		fieldValuePtr, err := col.ValueOf(bean)
		if err != nil {
			session.Engine.LogError(err)
			continue
		}
		fieldValue := *fieldValuePtr

		if col.IsAutoIncrement {
			switch fieldValue.Type().Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
				if fieldValue.Int() == 0 {
					continue
				}
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
				if fieldValue.Uint() == 0 {
					continue
				}
			case reflect.String:
				if len(fieldValue.String()) == 0 {
					continue
				}
			}
		}

		if session.Statement.ColumnStr != "" {
			if _, ok := session.Statement.columnMap[lColName]; !ok {
				continue
			}
		}
		if session.Statement.OmitStr != "" {
			if _, ok := session.Statement.columnMap[lColName]; ok {
				continue
			}
		}

		if (col.IsCreated || col.IsUpdated) && session.Statement.UseAutoTime {
			args = append(args, time.Now())
		} else if col.IsVersion && session.Statement.checkVersion {
			args = append(args, 1)
		} else {
			arg, err := session.value2Interface(col, fieldValue)
			if err != nil {
				return colNames, args, err
			}
			args = append(args, arg)
		}

		if includeQuote {
			colNames = append(colNames, session.Engine.Quote(col.Name)+" = ?")
		} else {
			colNames = append(colNames, col.Name)
		}
	}
	return colNames, args, nil
}
