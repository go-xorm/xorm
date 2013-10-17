package xorm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Struct Session keep a pointer to sql.DB and provides all execution of all
// kind of database operations.
type Session struct {
	Db                     *sql.DB
	Engine                 *Engine
	Tx                     *sql.Tx
	Statement              Statement
	IsAutoCommit           bool
	IsCommitedOrRollbacked bool
	TransType              string
	IsAutoClose            bool
}

// Method Init reset the session as the init status.
func (session *Session) Init() {
	session.Statement = Statement{Engine: session.Engine}
	session.Statement.Init()
	session.IsAutoCommit = true
	session.IsCommitedOrRollbacked = false
	session.IsAutoClose = false
}

// Method Close release the connection from pool
func (session *Session) Close() {
	defer func() {
		if session.Db != nil {
			session.Engine.Pool.ReleaseDB(session.Engine, session.Db)
			session.Db = nil
			session.Tx = nil
			session.Init()
		}
	}()
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

// Method Id provides converting id as a query condition
func (session *Session) Id(id int64) *Session {
	session.Statement.Id(id)
	return session
}

// Method Table can input a string or pointer to struct for special a table to operate.
func (session *Session) Table(tableNameOrBean interface{}) *Session {
	session.Statement.Table(tableNameOrBean)
	return session
}

// Method In provides a query string like "id in (1, 2, 3)"
func (session *Session) In(column string, args ...interface{}) *Session {
	session.Statement.In(column, args...)
	return session
}

// Method Cols provides some columns to special
func (session *Session) Cols(columns ...string) *Session {
	session.Statement.Cols(columns...)
	return session
}

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
	sql := strings.Join(newColNames, session.Engine.Quote(" DESC, "))
	session.Statement.OrderStr += session.Engine.Quote(sql) + " DESC"
	return session
}

// Method Asc provide asc order by query condition, the input parameters are columns.
func (session *Session) Asc(colNames ...string) *Session {
	if session.Statement.OrderStr != "" {
		session.Statement.OrderStr += ", "
	}
	newColNames := col2NewCols(colNames...)
	sql := strings.Join(newColNames, session.Engine.Quote(" ASC, "))
	session.Statement.OrderStr += session.Engine.Quote(sql) + " ASC"
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

// Method Cascade
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

func (session *Session) GroupBy(keys string) *Session {
	session.Statement.GroupBy(keys)
	return session
}

func (session *Session) Having(conditions string) *Session {
	session.Statement.Having(conditions)
	return session
}

func (session *Session) newDb() error {
	if session.Db == nil {
		db, err := session.Engine.Pool.RetrieveDB(session.Engine)
		if err != nil {
			return err
		}
		session.Db = db
	}
	return nil
}

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

		session.Engine.LogSQL("BEGIN TRANSACTION")
	}
	return nil
}

func (session *Session) Rollback() error {
	if !session.IsAutoCommit && !session.IsCommitedOrRollbacked {
		session.Engine.LogSQL("ROLL BACK")
		session.IsCommitedOrRollbacked = true
		return session.Tx.Rollback()
	}
	return nil
}

func (session *Session) Commit() error {
	if !session.IsAutoCommit && !session.IsCommitedOrRollbacked {
		session.Engine.LogSQL("COMMIT")
		session.IsCommitedOrRollbacked = true
		return session.Tx.Commit()
	}
	return nil
}

func (session *Session) scanMapIntoStruct(obj interface{}, objMap map[string][]byte) error {
	dataStruct := reflect.Indirect(reflect.ValueOf(obj))
	if dataStruct.Kind() != reflect.Struct {
		return errors.New("Expected a pointer to a struct")
	}

	table := session.Engine.Tables[rType(obj)]

	for key, data := range objMap {
		if _, ok := table.Columns[key]; !ok {
			continue
		}
		col := table.Columns[key]
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
func (session *Session) innerExec(sql string, args ...interface{}) (sql.Result, error) {
	rs, err := session.Db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	res, err := rs.Exec(args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (session *Session) exec(sql string, args ...interface{}) (sql.Result, error) {
	for _, filter := range session.Engine.Filters {
		sql = filter.Do(sql, session)
	}

	session.Engine.LogSQL(sql)
	session.Engine.LogSQL(args)

	if session.IsAutoCommit {
		return session.innerExec(sql, args...)
	}
	return session.Tx.Exec(sql, args...)
}

func (session *Session) Exec(sql string, args ...interface{}) (sql.Result, error) {
	err := session.newDb()
	if err != nil {
		return nil, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	return session.exec(sql, args...)
}

// this function create a table according a bean
func (session *Session) CreateTable(bean interface{}) error {
	session.Statement.RefTable = session.Engine.AutoMap(bean)

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

func (session *Session) CreateIndexes(bean interface{}) error {
	session.Statement.RefTable = session.Engine.AutoMap(bean)

	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	sqls := session.Statement.genIndexSQL()
	for _, sql := range sqls {
		_, err = session.exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) CreateUniques(bean interface{}) error {
	session.Statement.RefTable = session.Engine.AutoMap(bean)

	err := session.newDb()
	if err != nil {
		return err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	sqls := session.Statement.genUniqueSQL()
	for _, sql := range sqls {
		_, err = session.exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) createOneTable() error {
	sql := session.Statement.genCreateSQL()
	_, err := session.exec(sql)
	return err
}

func (session *Session) CreateAll() error {
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
	for _, sql := range sqls {
		_, err = session.exec(sql)
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
		session.Statement.RefTable = session.Engine.AutoMap(bean)
	} else {
		return errors.New("Unsupported type")
	}

	sql := session.Statement.genDropSQL()
	_, err = session.exec(sql)
	return err
}

func (statement *Statement) convertIdSql(sql string) string {
	if statement.RefTable != nil {
		col := statement.RefTable.PKColumn()
		if col != nil {
			sqls := splitNNoCase(sql, "from", 2)
			if len(sqls) != 2 {
				return ""
			}
			newsql := fmt.Sprintf("SELECT %v.%v FROM %v", statement.Engine.Quote(statement.TableName()),
				statement.Engine.Quote(col.Name), sqls[1])
			return newsql
		}
	}
	return ""
}

func (session *Session) cacheGet(bean interface{}, sql string, args ...interface{}) (has bool, err error) {
	if session.Statement.RefTable == nil || session.Statement.RefTable.PrimaryKey == "" {
		return false, ErrCacheFailed
	}
	for _, filter := range session.Engine.Filters {
		sql = filter.Do(sql, session)
	}
	newsql := session.Statement.convertIdSql(sql)
	if newsql == "" {
		return false, ErrCacheFailed
	}

	cacher := session.Statement.RefTable.Cacher
	tableName := session.Statement.TableName()
	session.Engine.LogDebug("[xorm:cacheGet] find sql:", newsql, args)
	ids, err := getCacheSql(cacher, tableName, newsql, args)
	if err != nil {
		resultsSlice, err := session.query(newsql, args...)
		if err != nil {
			return false, err
		}
		session.Engine.LogDebug("[xorm:cacheGet] query ids:", resultsSlice)
		ids = make([]int64, 0)
		if len(resultsSlice) > 0 {
			data := resultsSlice[0]
			var id int64
			if v, ok := data[session.Statement.RefTable.PrimaryKey]; !ok {
				return false, ErrCacheFailed
			} else {
				id, err = strconv.ParseInt(string(v), 10, 64)
				if err != nil {
					return false, err
				}
			}
			ids = append(ids, id)
		}
		session.Engine.LogDebug("[xorm:cacheGet] cache ids:", newsql, ids)
		err = putCacheSql(cacher, ids, tableName, newsql, args)
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
		cacheBean := cacher.GetBean(tableName, id)
		if cacheBean == nil {
			newSession := session.Engine.NewSession()
			defer newSession.Close()
			cacheBean = reflect.New(structValue.Type()).Interface()
			has, err = newSession.Id(id).NoCache().Get(cacheBean)
			if err != nil {
				return has, err
			}

			session.Engine.LogDebug("[xorm:cacheGet] cache bean:", tableName, id, cacheBean)
			cacher.PutBean(tableName, id, cacheBean)
		} else {
			session.Engine.LogDebug("[xorm:cacheGet] cached bean:", tableName, id, cacheBean)
			has = true
		}
		structValue.Set(reflect.Indirect(reflect.ValueOf(cacheBean)))

		return has, nil
	}
	return false, nil
}

func (session *Session) cacheFind(t reflect.Type, sql string, rowsSlicePtr interface{}, args ...interface{}) (err error) {
	if session.Statement.RefTable == nil || session.Statement.RefTable.PrimaryKey == "" {
		return ErrCacheFailed
	}

	for _, filter := range session.Engine.Filters {
		sql = filter.Do(sql, session)
	}

	newsql := session.Statement.convertIdSql(sql)
	if newsql == "" {
		return ErrCacheFailed
	}

	table := session.Statement.RefTable
	cacher := table.Cacher
	ids, err := getCacheSql(cacher, session.Statement.TableName(), newsql, args)
	if err != nil {
		//session.Engine.LogError(err)
		resultsSlice, err := session.query(newsql, args...)
		if err != nil {
			return err
		}
		// 查询数目太大，采用缓存将不是一个很好的方式。
		if len(resultsSlice) > 100 {
			session.Engine.LogDebug("[xorm:cacheFind] ids length %v > 100, no cache", len(resultsSlice))
			return ErrCacheFailed
		}

		tableName := session.Statement.TableName()
		ids = make([]int64, 0)
		if len(resultsSlice) > 0 {
			for _, data := range resultsSlice {
				//fmt.Println(data)
				var id int64
				if v, ok := data[session.Statement.RefTable.PrimaryKey]; !ok {
					return errors.New("no id")
				} else {
					id, err = strconv.ParseInt(string(v), 10, 64)
					if err != nil {
						return err
					}
				}
				ids = append(ids, id)
			}
		}
		session.Engine.LogDebug("[xorm:cacheFind] cache ids:", ids, tableName, newsql, args)
		err = putCacheSql(cacher, ids, tableName, newsql, args)
		if err != nil {
			return err
		}
	} else {
		session.Engine.LogDebug("[xorm:cacheFind] cached sql:", newsql, args)
	}

	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	pkFieldName := session.Statement.RefTable.PKColumn().FieldName

	ididxes := make(map[int64]int)
	var ides []interface{} = make([]interface{}, 0)
	var temps []interface{} = make([]interface{}, len(ids))
	tableName := session.Statement.TableName()
	for idx, id := range ids {
		bean := cacher.GetBean(tableName, id)
		if bean == nil {
			ides = append(ides, id)
			ididxes[id] = idx
		} else {
			session.Engine.LogDebug("[xorm:cacheFind] cached bean:", tableName, id, bean)

			sid := reflect.Indirect(reflect.ValueOf(bean)).FieldByName(pkFieldName).Int()
			if sid != id {
				session.Engine.LogError("[xorm:cacheFind] error cache", id, sid, bean)
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
		err = newSession.In("(id)", ides...).NoCache().Find(beans)
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
			id := reflect.Indirect(reflect.ValueOf(bean)).FieldByName(pkFieldName).Int()
			//bean := vs.Index(i).Addr().Interface()
			temps[ididxes[id]] = bean
			//temps[idxes[i]] = bean
			session.Engine.LogDebug("[xorm:cacheFind] cache bean:", tableName, id, bean)
			cacher.PutBean(tableName, id, bean)
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
			var key int64
			if table.PrimaryKey != "" {
				key = ids[j]
			} else {
				key = int64(j)
			}
			if t.Kind() == reflect.Ptr {
				sliceValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(bean))
			} else {
				sliceValue.SetMapIndex(reflect.ValueOf(key), reflect.Indirect(reflect.ValueOf(bean)))
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

type IterFunc func(idx int, bean interface{}) error

func (session *Session) Iterate(bean interface{}, fun IterFunc) error {
	err := session.newDb()
	if err != nil {
		return err
	}

	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	var sql string
	var args []interface{}
	session.Statement.RefTable = session.Engine.AutoMap(bean)
	if session.Statement.RawSQL == "" {
		sql, args = session.Statement.genGetSql(bean)
	} else {
		sql = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	for _, filter := range session.Engine.Filters {
		sql = filter.Do(sql, session)
	}

	session.Engine.LogSQL(sql)
	session.Engine.LogSQL(args)

	s, err := session.Db.Prepare(sql)
	if err != nil {
		return err
	}
	defer s.Close()
	rows, err := s.Query(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	fields, err := rows.Columns()
	if err != nil {
		return err
	}
	t := reflect.Indirect(reflect.ValueOf(bean)).Type()
	b := reflect.New(t).Interface()
	i := 0
	for rows.Next() {
		result, err := row2map(rows, fields)
		if err == nil {
			err = session.scanMapIntoStruct(b, result)
		}
		if err == nil {
			err = fun(i, b)
			i = i + 1
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// get retrieve one record from database
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
	var sql string
	var args []interface{}
	session.Statement.RefTable = session.Engine.AutoMap(bean)
	if session.Statement.RawSQL == "" {
		sql, args = session.Statement.genGetSql(bean)
	} else {
		sql = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	if session.Statement.RefTable.Cacher != nil && session.Statement.UseCache {
		has, err := session.cacheGet(bean, sql, args...)
		if err != ErrCacheFailed {
			return has, err
		}
	}

	resultsSlice, err := session.query(sql, args...)
	if err != nil {
		return false, err
	}
	if len(resultsSlice) < 1 {
		return false, nil
	}

	err = session.scanMapIntoStruct(bean, resultsSlice[0])
	if err != nil {
		return true, err
	}
	if len(resultsSlice) == 1 {
		return true, nil
	} else {
		return true, errors.New("More than one record")
	}
}

func (session *Session) Count(bean interface{}) (int64, error) {
	err := session.newDb()
	if err != nil {
		return 0, err
	}

	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	var sql string
	var args []interface{}
	if session.Statement.RawSQL == "" {
		sql, args = session.Statement.genCountSql(bean)
	} else {
		sql = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	resultsSlice, err := session.query(sql, args...)
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
	var table *Table
	if sliceElementType.Kind() == reflect.Ptr {
		if sliceElementType.Elem().Kind() == reflect.Struct {
			table = session.Engine.AutoMapType(sliceElementType.Elem())
		} else {
			return errors.New("slice type")
		}
	} else if sliceElementType.Kind() == reflect.Struct {
		table = session.Engine.AutoMapType(sliceElementType)
	} else {
		return errors.New("slice type")
	}
	session.Statement.RefTable = table

	if len(condiBean) > 0 {
		colNames, args := buildConditions(session.Engine, table, condiBean[0])
		session.Statement.ConditionStr = strings.Join(colNames, " and ")
		session.Statement.BeanArgs = args
	}

	var sql string
	var args []interface{}
	if session.Statement.RawSQL == "" {
		var columnStr string = session.Statement.ColumnStr
		if columnStr == "" {
			columnStr = session.Statement.genColumnStr()
		}
		sql = session.Statement.genSelectSql(columnStr)
		args = append(session.Statement.Params, session.Statement.BeanArgs...)
	} else {
		sql = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	if table.Cacher != nil && session.Statement.UseCache {
		err = session.cacheFind(sliceElementType, sql, rowsSlicePtr, args...)
		if err != ErrCacheFailed {
			return err
		}
		session.Engine.LogWarn("Cache Find Failed")
	}

	resultsSlice, err := session.query(sql, args...)
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
		if sliceValue.Kind() == reflect.Slice {
			if sliceElementType.Kind() == reflect.Ptr {
				sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(newValue.Interface())))
			} else {
				sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
			}
		} else if sliceValue.Kind() == reflect.Map {
			var key int64
			if table.PrimaryKey != "" {
				x, err := strconv.ParseInt(string(results[table.PrimaryKey]), 10, 64)
				if err != nil {
					return errors.New("pk " + table.PrimaryKey + " as int64: " + err.Error())
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

// test if database is ok
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
	sql, args := session.Engine.dialect.ColumnCheckSql(tableName, colName)
	results, err := session.query(sql, args...)
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
	sql, args := session.Engine.dialect.TableCheckSql(tableName)
	results, err := session.query(sql, args...)
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
	sql, args := session.Engine.dialect.IndexCheckSql(tableName, idx)
	results, err := session.query(sql, args...)
	return len(results) > 0, err
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
	col := session.Statement.RefTable.Columns[colName]
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
	sql, args := session.Statement.genAddIndexStr(indexName(tableName, idxName), cols)
	_, err = session.exec(sql, args...)
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
	sql, args := session.Statement.genAddUniqueStr(uniqueName(tableName, uqeName), cols)
	_, err = session.exec(sql, args...)
	return err
}

func (session *Session) DropAll() error {
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
		sql := session.Statement.genDropSQL()
		_, err := session.exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

func row2map(rows *sql.Rows, fields []string) (resultsMap map[string][]byte, err error) {
	result := make(map[string][]byte)
	var scanResultContainers []interface{}
	for i := 0; i < len(fields); i++ {
		var scanResultContainer interface{}
		scanResultContainers = append(scanResultContainers, &scanResultContainer)
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
		aa := reflect.TypeOf(rawValue.Interface())
		vv := reflect.ValueOf(rawValue.Interface())
		var str string
		switch aa.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			str = strconv.FormatInt(vv.Int(), 10)
			result[key] = []byte(str)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			str = strconv.FormatUint(vv.Uint(), 10)
			result[key] = []byte(str)
		case reflect.Float32, reflect.Float64:
			str = strconv.FormatFloat(vv.Float(), 'f', -1, 64)
			result[key] = []byte(str)
		case reflect.Slice:
			switch aa.Elem().Kind() {
			case reflect.Uint8:
				result[key] = rawValue.Interface().([]byte)
			default:
				//session.Engine.LogError("Unsupported type")
			}
		case reflect.String:
			str = vv.String()
			result[key] = []byte(str)
		//时间类型
		case reflect.Struct:
			if aa.String() == "time.Time" {
				str = rawValue.Interface().(time.Time).Format("2006-01-02 15:04:05.000 -0700")
				result[key] = []byte(str)
			} else {
				//session.Engine.LogError("Unsupported struct type")
			}
		default:
			//session.Engine.LogError("Unsupported type")
		}
	}
	return result, nil
}

func rows2maps(rows *sql.Rows) (resultsSlice []map[string][]byte, err error) {
	fields, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		result, err := row2map(rows, fields)
		if err != nil {
			return nil, err
		}
		resultsSlice = append(resultsSlice, result)
	}

	return resultsSlice, nil
}

func query(db *sql.DB, sql string, params ...interface{}) (resultsSlice []map[string][]byte, err error) {
	s, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	rows, err := s.Query(params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows2maps(rows)
}

func (session *Session) query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	for _, filter := range session.Engine.Filters {
		sql = filter.Do(sql, session)
	}

	session.Engine.LogSQL(sql)
	session.Engine.LogSQL(paramStr)

	return query(session.Db, sql, paramStr...)

	/*s, err := session.Db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	res, err := s.Query(paramStr...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	fields, err := res.Columns()
	if err != nil {
		return nil, err
	}
	for res.Next() {
		result := make(map[string][]byte)
		var scanResultContainers []interface{}
		for i := 0; i < len(fields); i++ {
			var scanResultContainer interface{}
			scanResultContainers = append(scanResultContainers, &scanResultContainer)
		}
		if err := res.Scan(scanResultContainers...); err != nil {
			return nil, err
		}
		for ii, key := range fields {
			rawValue := reflect.Indirect(reflect.ValueOf(scanResultContainers[ii]))

			//if row is null then ignore
			if rawValue.Interface() == nil {
				//fmt.Println("ignore ...", key, rawValue)
				continue
			}
			aa := reflect.TypeOf(rawValue.Interface())
			vv := reflect.ValueOf(rawValue.Interface())
			var str string
			switch aa.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				str = strconv.FormatInt(vv.Int(), 10)
				result[key] = []byte(str)
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				str = strconv.FormatUint(vv.Uint(), 10)
				result[key] = []byte(str)
			case reflect.Float32, reflect.Float64:
				str = strconv.FormatFloat(vv.Float(), 'f', -1, 64)
				result[key] = []byte(str)
			case reflect.Slice:
				switch aa.Elem().Kind() {
				case reflect.Uint8:
					result[key] = rawValue.Interface().([]byte)
				default:
					session.Engine.LogError("Unsupported type")
				}
			case reflect.String:
				str = vv.String()
				result[key] = []byte(str)
			//时间类型
			case reflect.Struct:
				if aa.String() == "time.Time" {
					str = rawValue.Interface().(time.Time).Format("2006-01-02 15:04:05.000 -0700")
					result[key] = []byte(str)
				} else {
					session.Engine.LogError("Unsupported struct type")
				}
			default:
				session.Engine.LogError("Unsupported type")
			}
		}
		resultsSlice = append(resultsSlice, result)
	}
	return resultsSlice, nil*/
}

func (session *Session) Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	err = session.newDb()
	if err != nil {
		return nil, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	return session.query(sql, paramStr...)
}

// insert one or more beans
func (session *Session) Insert(beans ...interface{}) (int64, error) {
	var lastId int64 = -1
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
				lastId, err = session.innerInsertMulti(bean)
				if err != nil {
					return lastId, err
				}
			} else {
				size := sliceValue.Len()
				for i := 0; i < size; i++ {
					lastId, err = session.innerInsert(sliceValue.Index(i).Interface())
					if err != nil {
						return lastId, err
					}
				}
			}
		} else {
			lastId, err = session.innerInsert(bean)
			if err != nil {
				return lastId, err
			}
		}
	}

	return lastId, err
}

func (session *Session) innerInsertMulti(rowsSlicePtr interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return -1, errors.New("needs a pointer to a slice")
	}

	bean := sliceValue.Index(0).Interface()
	sliceElementType := rType(bean)

	table := session.Engine.AutoMapType(sliceElementType)
	session.Statement.RefTable = table

	size := sliceValue.Len()

	colNames := make([]string, 0)
	colMultiPlaces := make([]string, 0)
	var args = make([]interface{}, 0)
	cols := make([]*Column, 0)

	for i := 0; i < size; i++ {
		elemValue := sliceValue.Index(i).Interface()
		colPlaces := make([]string, 0)

		if i == 0 {
			for _, col := range table.Columns {
				fieldValue := reflect.Indirect(reflect.ValueOf(elemValue)).FieldByName(col.FieldName)
				if col.IsAutoIncrement && fieldValue.Int() == 0 {
					continue
				}
				if col.MapType == ONLYFROMDB {
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
				if col.MapType == ONLYFROMDB {
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

	statement := fmt.Sprintf("INSERT INTO %v%v%v (%v%v%v) VALUES (%v);",
		session.Engine.QuoteStr(),
		session.Statement.TableName(),
		session.Engine.QuoteStr(),
		session.Engine.QuoteStr(),
		strings.Join(colNames, session.Engine.QuoteStr()+", "+session.Engine.QuoteStr()),
		session.Engine.QuoteStr(),
		strings.Join(colMultiPlaces, "),("))

	res, err := session.exec(statement, args...)
	if err != nil {
		return -1, err
	}

	if table.Cacher != nil && session.Statement.UseCache {
		session.cacheInsert(session.Statement.TableName())
	}

	id, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return id, nil
}

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

// convert a db data([]byte) to a field value
func (session *Session) bytes2Value(col *Column, fieldValue *reflect.Value, data []byte) error {
	if structConvert, ok := fieldValue.Addr().Interface().(Conversion); ok {
		return structConvert.FromDB(data)
	}

	var v interface{}
	key := col.Name
	fieldType := fieldValue.Type()
	switch fieldType.Kind() {
	case reflect.Complex64, reflect.Complex128:
		x := reflect.New(fieldType)

		err := json.Unmarshal(data, x.Interface())
		if err != nil {
			session.Engine.LogSQL(err)
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
				session.Engine.LogSQL(err)
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
					session.Engine.LogSQL(err)
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
		v, err := strconv.ParseBool(string(data))
		if err != nil {
			return errors.New("arg " + key + " as bool: " + err.Error())
		}
		fieldValue.Set(reflect.ValueOf(v))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return errors.New("arg " + key + " as int: " + err.Error())
		}
		fieldValue.SetInt(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return errors.New("arg " + key + " as float64: " + err.Error())
		}
		fieldValue.SetFloat(x)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		x, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return errors.New("arg " + key + " as int: " + err.Error())
		}
		fieldValue.SetUint(x)
	//Now only support Time type
	case reflect.Struct:
		if fieldValue.Type().String() == "time.Time" {
			x, err := time.Parse("2006-01-02 15:04:05", string(data))
			if err != nil {
				x, err = time.Parse("2006-01-02 15:04:05.000 -0700", string(data))

				if err != nil {
					return errors.New("unsupported time format: " + string(data))
				}
			}

			v = x
			fieldValue.Set(reflect.ValueOf(v))
		} else if session.Statement.UseCascade {
			table := session.Engine.AutoMapType(fieldValue.Type())
			if table != nil {
				x, err := strconv.ParseInt(string(data), 10, 64)
				if err != nil {
					return errors.New("arg " + key + " as int: " + err.Error())
				}
				if x != 0 {
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
				return errors.New("unsupported struct type in Scan: " + fieldValue.Type().String())
			}
		}
	default:
		return errors.New("unsupported type in Scan: " + reflect.TypeOf(v).String())
	}

	return nil
}

// convert a field value of a struct to interface for put into db
func (session *Session) value2Interface(col *Column, fieldValue reflect.Value) (interface{}, error) {
	if fieldValue.CanAddr() {
		if fieldConvert, ok := fieldValue.Addr().Interface().(Conversion); ok {
			data, err := fieldConvert.ToDB()
			if err != nil {
				return 0, err
			} else {
				return string(data), nil
			}
		}
	}

	k := fieldValue.Type().Kind()
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
		if fieldValue.Type().String() == "time.Time" {
			//return fieldValue.Interface().(time.Time).Format(time.RFC3339Nano), nil
			//return fieldValue.Interface().(time.Time).Format("2006-01-02 15:04:05 -0700"), nil
			return fieldValue.Interface(), nil
		}
		if fieldTable, ok := session.Engine.Tables[fieldValue.Type()]; ok {
			if fieldTable.PrimaryKey != "" {
				pkField := reflect.Indirect(fieldValue).FieldByName(fieldTable.PKColumn().FieldName)
				return pkField.Interface(), nil
			} else {
				return 0, errors.New("no primary key")
			}
		} else {
			return 0, errors.New(fmt.Sprintf("Unsupported type %v", fieldValue.Type()))
		}
	case reflect.Complex64, reflect.Complex128:
		bytes, err := json.Marshal(fieldValue.Interface())
		if err != nil {
			session.Engine.LogSQL(err)
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
				session.Engine.LogSQL(err)
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
					session.Engine.LogSQL(err)
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
	table := session.Engine.AutoMap(bean)
	session.Statement.RefTable = table

	colNames, args, err := table.genCols(session, bean, false, false)
	if err != nil {
		return 0, err
	}

	colPlaces := strings.Repeat("?, ", len(colNames))
	colPlaces = colPlaces[0 : len(colPlaces)-2]

	sql := fmt.Sprintf("INSERT INTO %v%v%v (%v%v%v) VALUES (%v);",
		session.Engine.QuoteStr(),
		session.Statement.TableName(),
		session.Engine.QuoteStr(),
		session.Engine.QuoteStr(),
		strings.Join(colNames, session.Engine.Quote(", ")),
		session.Engine.QuoteStr(),
		colPlaces)

	res, err := session.exec(sql, args...)
	if err != nil {
		return 0, err
	}

	if table.Cacher != nil && session.Statement.UseCache {
		session.cacheInsert(session.Statement.TableName())
	}

	if table.PrimaryKey == "" {
		return 0, nil
	}

	var id int64 = 0
	pkValue := table.PKColumn().ValueOf(bean)
	if !pkValue.IsValid() || pkValue.Int() != 0 || !pkValue.CanSet() {
		return 0, nil
	}

	id, err = res.LastInsertId()
	if err != nil || id <= 0 {
		return 0, err
	}

	var v interface{} = id
	switch pkValue.Type().Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
		v = int(id)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		v = uint(id)
	}
	pkValue.Set(reflect.ValueOf(v))

	return id, nil
}

// Method InsertOne insert only one struct into database as a record.
// The in parameter bean must a struct or a point to struct. The return
// parameter is lastInsertId and error
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

func (statement *Statement) convertUpdateSql(sql string) (string, string) {
	if statement.RefTable == nil || statement.RefTable.PrimaryKey == "" {
		return "", ""
	}
	sqls := splitNNoCase(sql, "where", 2)
	if len(sqls) != 2 {
		return "", ""
	}

	var whereStr = sqls[1]

	//TODO: for postgres only, if any other database?
	if strings.Contains(sqls[1], "$") {
		dollers := strings.Split(sqls[1], "$")
		whereStr = dollers[0]
		for i, c := range dollers[1:] {
			ccs := strings.SplitN(c, " ", 2)
			whereStr += fmt.Sprintf("$%v %v", i+1, ccs[1])
		}
	}

	return sqls[0], fmt.Sprintf("SELECT %v FROM %v WHERE %v",
		statement.Engine.Quote(statement.RefTable.PrimaryKey), statement.Engine.Quote(statement.TableName()),
		whereStr)
}

func (session *Session) cacheInsert(tables ...string) error {
	if session.Statement.RefTable == nil || session.Statement.RefTable.PrimaryKey == "" {
		return ErrCacheFailed
	}

	table := session.Statement.RefTable
	cacher := table.Cacher

	for _, t := range tables {
		session.Engine.LogDebug("cache clear:", t)
		cacher.ClearIds(t)
	}

	return nil
}

func (session *Session) cacheUpdate(sql string, args ...interface{}) error {
	if session.Statement.RefTable == nil || session.Statement.RefTable.PrimaryKey == "" {
		return ErrCacheFailed
	}

	oldhead, newsql := session.Statement.convertUpdateSql(sql)
	if newsql == "" {
		return ErrCacheFailed
	}
	for _, filter := range session.Engine.Filters {
		newsql = filter.Do(newsql, session)
	}
	session.Engine.LogDebug("[xorm:cacheUpdate] new sql", oldhead, newsql)

	var nStart int
	if len(args) > 0 {
		if strings.Index(sql, "?") > -1 {
			nStart = strings.Count(oldhead, "?")
		} else {
			// only for pq, TODO: if any other databse?
			nStart = strings.Count(oldhead, "$")
		}
	}
	table := session.Statement.RefTable
	cacher := table.Cacher
	tableName := session.Statement.TableName()
	session.Engine.LogDebug("[xorm:cacheUpdate] get cache sql", newsql, args[nStart:])
	ids, err := getCacheSql(cacher, tableName, newsql, args[nStart:])
	if err != nil {
		resultsSlice, err := session.query(newsql, args[nStart:]...)
		if err != nil {
			return err
		}
		session.Engine.LogDebug("[xorm:cacheUpdate] find updated id", resultsSlice)

		ids = make([]int64, 0)
		if len(resultsSlice) > 0 {
			for _, data := range resultsSlice {
				var id int64
				if v, ok := data[session.Statement.RefTable.PrimaryKey]; !ok {
					return errors.New("no id")
				} else {
					id, err = strconv.ParseInt(string(v), 10, 64)
					if err != nil {
						return err
					}
				}
				ids = append(ids, id)
			}
		}
	} /*else {
		session.Engine.LogDebug("[xorm:cacheUpdate] del cached sql:", tableName, newsql, args)
		cacher.DelIds(tableName, genSqlKey(newsql, args))
	}*/

	for _, id := range ids {
		if bean := cacher.GetBean(tableName, id); bean != nil {
			sqls := splitNNoCase(sql, "where", 2)
			if len(sqls) != 2 {
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

				if col, ok := table.Columns[colName]; ok {
					fieldValue := col.ValueOf(bean)
					session.Engine.LogDebug("[xorm:cacheUpdate] set bean field", bean, colName, fieldValue.Interface())
					fieldValue.Set(reflect.ValueOf(args[idx]))
				}
			}

			session.Engine.LogDebug("[xorm:cacheUpdate] update cache", tableName, id, bean)
			cacher.PutBean(tableName, id, bean)
		}
	}
	session.Engine.LogDebug("[xorm:cacheUpdate] clear cached table sql:", tableName)
	cacher.ClearIds(tableName)
	return nil
}

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
	var table *Table

	if t.Kind() == reflect.Struct {
		table = session.Engine.AutoMap(bean)
		session.Statement.RefTable = table

		if session.Statement.ColumnStr == "" {
			colNames, args = buildConditions(session.Engine, table, bean)
		} else {
			colNames, args, err = table.genCols(session, bean, true, true)
			if err != nil {
				return 0, err
			}
		}
		if session.Statement.UseAutoTime && table.Updated != "" {
			colNames = append(colNames, session.Engine.Quote(table.Updated)+" = ?")
			args = append(args, time.Now())
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
		if session.Statement.UseAutoTime && table.Updated != "" {
			colNames = append(colNames, session.Engine.Quote(table.Updated)+" = ?")
			args = append(args, time.Now())
		}
	} else {
		return 0, ErrParamsType
	}

	var condiColNames []string
	var condiArgs []interface{}

	if len(condiBean) > 0 {
		condiColNames, condiArgs = buildConditions(session.Engine, session.Statement.RefTable, condiBean[0])
	}

	var condition = ""
	st := session.Statement
	defer session.Statement.Init()
	if st.WhereStr != "" {
		condition = fmt.Sprintf("WHERE %v", st.WhereStr)
	}

	if condition == "" {
		if len(condiColNames) > 0 {
			condition = fmt.Sprintf("WHERE %v ", strings.Join(condiColNames, " and "))
		}
	} else {
		if len(condiColNames) > 0 {
			condition = fmt.Sprintf("%v and %v", condition, strings.Join(condiColNames, " and "))
		}
	}

	sql := fmt.Sprintf("UPDATE %v SET %v %v",
		session.Engine.Quote(session.Statement.TableName()),
		strings.Join(colNames, ", "),
		condition)

	args = append(append(args, st.Params...), condiArgs...)

	res, err := session.exec(sql, args...)
	if err != nil {
		return 0, err
	}

	if table.Cacher != nil && session.Statement.UseCache {
		session.cacheUpdate(sql, args...)
	}

	return res.RowsAffected()
}

func (session *Session) cacheDelete(sql string, args ...interface{}) error {
	if session.Statement.RefTable == nil || session.Statement.RefTable.PrimaryKey == "" {
		return ErrCacheFailed
	}

	for _, filter := range session.Engine.Filters {
		sql = filter.Do(sql, session)
	}

	newsql := session.Statement.convertIdSql(sql)
	if newsql == "" {
		return ErrCacheFailed
	}

	cacher := session.Statement.RefTable.Cacher
	tableName := session.Statement.TableName()
	ids, err := getCacheSql(cacher, tableName, newsql, args)
	if err != nil {
		resultsSlice, err := session.query(newsql, args...)
		if err != nil {
			return err
		}
		ids = make([]int64, 0)
		if len(resultsSlice) > 0 {
			for _, data := range resultsSlice {
				var id int64
				if v, ok := data[session.Statement.RefTable.PrimaryKey]; !ok {
					return errors.New("no id")
				} else {
					id, err = strconv.ParseInt(string(v), 10, 64)
					if err != nil {
						return err
					}
				}
				ids = append(ids, id)
			}
		}
	} /*else {
		session.Engine.LogDebug("delete cache sql %v", newsql)
		cacher.DelIds(tableName, genSqlKey(newsql, args))
	}*/

	for _, id := range ids {
		session.Engine.LogDebug("[xorm:cacheDelete] delete cache obj", tableName, id)
		cacher.DelBean(tableName, id)
	}
	session.Engine.LogDebug("[xorm:cacheDelete] clear cache table", tableName)
	cacher.ClearIds(tableName)
	return nil
}

func (session *Session) Delete(bean interface{}) (int64, error) {
	err := session.newDb()
	if err != nil {
		return 0, err
	}
	defer session.Statement.Init()
	if session.IsAutoClose {
		defer session.Close()
	}

	table := session.Engine.AutoMap(bean)
	session.Statement.RefTable = table
	colNames, args := buildConditions(session.Engine, table, bean)

	var condition = ""
	if session.Statement.WhereStr != "" {
		condition = session.Statement.WhereStr
		if len(colNames) > 0 {
			condition += " and " + strings.Join(colNames, " and ")
		}
	} else {
		condition = strings.Join(colNames, " and ")
	}
	if len(condition) == 0 {
		return 0, ErrNeedDeletedCond
	}

	sql := fmt.Sprintf("DELETE FROM %v WHERE %v",
		session.Engine.Quote(session.Statement.TableName()), condition)

	args = append(session.Statement.Params, args...)

	if table.Cacher != nil && session.Statement.UseCache {
		session.cacheDelete(sql, args...)
	}

	res, err := session.exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
