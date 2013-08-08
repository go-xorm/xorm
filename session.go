// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

package xorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	Db                     *sql.DB
	Engine                 *Engine
	Tx                     *sql.Tx
	Statement              Statement
	IsAutoCommit           bool
	IsCommitedOrRollbacked bool
	TransType              string
}

func (session *Session) Init() {
	session.Statement = Statement{Engine: session.Engine}
	session.Statement.Init()
	session.IsAutoCommit = true
	session.IsCommitedOrRollbacked = false
}

func (session *Session) Close() {
	defer func() {
		if session.Db != nil {
			session.Engine.pool.ReleaseDB(session.Engine, session.Db)
			session.Db = nil
			session.Tx = nil
			session.Init()
		}
	}()
}

func (session *Session) Sql(querystring string, args ...interface{}) *Session {
	session.Statement.Sql(querystring, args...)
	return session
}

func (session *Session) Where(querystring string, args ...interface{}) *Session {
	session.Statement.Where(querystring, args...)
	return session
}

func (session *Session) Id(id int64) *Session {
	session.Statement.Id(id)
	return session
}

func (session *Session) Table(tableName string) *Session {
	session.Statement.Table(tableName)
	return session
}

func (session *Session) In(column string, args ...interface{}) *Session {
	session.Statement.In(column, args...)
	return session
}

func (session *Session) Cols(columns ...string) *Session {
	session.Statement.Cols(columns...)
	return session
}

func (session *Session) Trans(t string) *Session {
	session.TransType = t
	return session
}

func (session *Session) Limit(limit int, start ...int) *Session {
	session.Statement.Limit(limit, start...)
	return session
}

func (session *Session) OrderBy(order string) *Session {
	session.Statement.OrderBy(order)
	return session
}

func (session *Session) StoreEngine(storeEngine string) *Session {
	session.Statement.StoreEngine = storeEngine
	return session
}

func (session *Session) Charset(charset string) *Session {
	session.Statement.Charset = charset
	return session
}

func (session *Session) Cascade(trueOrFalse ...bool) *Session {
	if len(trueOrFalse) >= 1 {
		session.Statement.UseCascade = trueOrFalse[0]
	}
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
		db, err := session.Engine.pool.RetrieveDB(session.Engine)
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

	table := session.Engine.Tables[Type(obj)]

	for key, data := range objMap {
		if _, ok := table.Columns[key]; !ok {
			continue
		}
		fieldName := table.Columns[key].FieldName
		fieldPath := strings.Split(fieldName, ".")
		var structField reflect.Value
		if len(fieldPath) > 2 {
			session.Engine.LogError("Unsupported mutliderive", fieldName)
			continue
		} else if len(fieldPath) == 2 {
			parentField := dataStruct.FieldByName(fieldPath[0])
			if parentField.IsValid() {
				structField = parentField.FieldByName(fieldPath[1])
			}
		} else {
			structField = dataStruct.FieldByName(fieldName)
		}
		if !structField.IsValid() || !structField.CanSet() {
			continue
		}

		var v interface{}

		switch structField.Type().Kind() {
		case reflect.Slice:
			v = data
		case reflect.Array:
			if structField.Type().Elem() == reflect.TypeOf(b) {
				v = data
			}
		case reflect.String:
			v = string(data)
		case reflect.Bool:
			v = (string(data) == "1")
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			x, err := strconv.Atoi(string(data))
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		case reflect.Int64:
			x, err := strconv.ParseInt(string(data), 10, 64)
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(string(data), 64)
			if err != nil {
				return errors.New("arg " + key + " as float64: " + err.Error())
			}
			v = x
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x, err := strconv.ParseUint(string(data), 10, 64)
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		//Now only support Time type
		case reflect.Struct:
			if structField.Type().String() == "time.Time" {
				x, err := time.Parse("2006-01-02 15:04:05", string(data))
				if err != nil {
					x, err = time.Parse("2006-01-02 15:04:05.000 -0700", string(data))

					if err != nil {
						return errors.New("unsupported time format: " + string(data))
					}
				}

				v = x
			} else if structConvert, ok := structField.Addr().Interface().(Conversion); ok {
				err := structConvert.FromDB(data)
				if err != nil {
					return err
				}
				continue
			} else if session.Statement.UseCascade {
				table := session.Engine.AutoMapType(structField.Type())
				if table != nil {
					x, err := strconv.ParseInt(string(data), 10, 64)
					if err != nil {
						return errors.New("arg " + key + " as int: " + err.Error())
					}
					if x != 0 {
						structInter := reflect.New(structField.Type())
						newsession := session.Engine.NewSession()
						defer newsession.Close()
						has, err := newsession.Id(x).Get(structInter.Interface())
						if err != nil {
							return err
						}
						if has {
							v = structInter.Elem().Interface()
						} else {
							session.Engine.LogError("cascade obj is not exist!")
							continue
						}
					} else {
						continue
					}
				} else {
					session.Engine.LogError("unsupported struct type in Scan: " + structField.Type().String())
					continue
				}
			} else {
				continue
			}
		default:
			return errors.New("unsupported type in Scan: " + reflect.TypeOf(v).String())
		}

		structField.Set(reflect.ValueOf(v))
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

func (session *Session) Exec(sql string, args ...interface{}) (sql.Result, error) {
	err := session.newDb()
	if err != nil {
		return nil, err
	}

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

// this function create a table according a bean
func (session *Session) CreateTable(bean interface{}) error {
	session.Statement.RefTable = session.Engine.AutoMap(bean)

	err := session.newDb()
	if err != nil {
		return err
	}

	return session.createOneTable()
}

func (session *Session) createOneTable() error {
	sql := session.Statement.genCreateSQL()
	_, err := session.Exec(sql)
	if err == nil {
		sqls := session.Statement.genIndexSQL()
		for _, sql := range sqls {
			_, err = session.Exec(sql)
			if err != nil {
				return err
			}
		}
	}
	if err == nil {
		sqls := session.Statement.genUniqueSQL()
		for _, sql := range sqls {
			_, err = session.Exec(sql)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (session *Session) CreateAll() error {
	err := session.newDb()
	if err != nil {
		return err
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

func (session *Session) DropTable(bean interface{}) error {
	err := session.newDb()
	if err != nil {
		return err
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
	_, err = session.Exec(sql)
	return err
}

func (session *Session) Get(bean interface{}) (bool, error) {
	err := session.newDb()
	if err != nil {
		return false, err
	}

	defer session.Statement.Init()
	session.Statement.Limit(1)
	var sql string
	var args []interface{}
	if session.Statement.RawSQL == "" {
		sql, args = session.Statement.genGetSql(bean)
	} else {
		sql = session.Statement.RawSQL
		args = session.Statement.RawParams
		session.Engine.AutoMap(bean)
	}
	resultsSlice, err := session.Query(sql, args...)
	if err != nil {
		return false, err
	}
	if len(resultsSlice) < 1 {
		return false, nil
	}

	results := resultsSlice[0]
	err = session.scanMapIntoStruct(bean, results)
	if err != nil {
		return false, err
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
	var sql string
	var args []interface{}
	if session.Statement.RawSQL == "" {
		sql, args = session.Statement.genCountSql(bean)
	} else {
		sql = session.Statement.RawSQL
		args = session.Statement.RawParams
	}

	resultsSlice, err := session.Query(sql, args...)
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
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice && sliceValue.Kind() != reflect.Map {
		return errors.New("needs a pointer to a slice or a map")
	}

	sliceElementType := sliceValue.Type().Elem()
	table := session.Engine.AutoMapType(sliceElementType)
	session.Statement.RefTable = table

	if len(condiBean) > 0 {
		colNames, args := BuildConditions(session.Engine, table, condiBean[0])
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

	resultsSlice, err := session.Query(sql, args...)

	if err != nil {
		return err
	}

	for i, results := range resultsSlice {
		newValue := reflect.New(sliceElementType)
		err := session.scanMapIntoStruct(newValue.Interface(), results)
		if err != nil {
			return err
		}
		if sliceValue.Kind() == reflect.Slice {
			sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
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
			sliceValue.SetMapIndex(reflect.ValueOf(key), reflect.Indirect(reflect.ValueOf(newValue.Interface())))
		}
	}
	return nil
}

func (session *Session) Ping() error {
	err := session.newDb()
	if err != nil {
		return err
	}

	return session.Db.Ping()
}

func (session *Session) DropAll() error {
	err := session.newDb()
	if err != nil {
		return err
	}

	for _, table := range session.Engine.Tables {
		session.Statement.Init()
		session.Statement.RefTable = table
		sql := session.Statement.genDropSQL()
		_, err := session.Exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	err = session.newDb()
	if err != nil {
		return nil, err
	}

	for _, filter := range session.Engine.Filters {
		sql = filter.Do(sql, session)
	}

	session.Engine.LogSQL(sql)
	session.Engine.LogSQL(paramStr)

	s, err := session.Db.Prepare(sql)
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
		//scanResultContainers := make([]interface{}, len(fields))
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
				if aa.Elem().Kind() == reflect.Uint8 {
					result[key] = rawValue.Interface().([]byte)
					break
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
			}
			//default:

		}
		resultsSlice = append(resultsSlice, result)
	}
	return resultsSlice, nil
}

func (session *Session) Insert(beans ...interface{}) (int64, error) {
	var lastId int64 = -1
	var err error = nil
	isInTransaction := !session.IsAutoCommit

	if !isInTransaction {
		err = session.Begin()
		defer session.Close()
		if err != nil {
			return 0, err
		}
	}

	for _, bean := range beans {
		sliceValue := reflect.Indirect(reflect.ValueOf(bean))
		if sliceValue.Kind() == reflect.Slice {
			if session.Engine.SupportInsertMany() {
				lastId, err = session.innerInsertMulti(bean)
				if err != nil {
					if !isInTransaction {
						err1 := session.Rollback()
						if err1 == nil {
							return lastId, err
						}
						err = err1
					}
					return lastId, err
				}
			} else {
				size := sliceValue.Len()
				for i := 0; i < size; i++ {
					lastId, err = session.innerInsert(sliceValue.Index(i).Interface())
					if err != nil {
						if !isInTransaction {
							err1 := session.Rollback()
							if err1 == nil {
								return lastId, err
							}
							err = err1
						}
						return lastId, err
					}
				}
			}
		} else {
			lastId, err = session.innerInsert(bean)
			if err != nil {
				if !isInTransaction {
					err1 := session.Rollback()
					if err1 == nil {
						return lastId, err
					}
					err = err1
				}
				return lastId, err
			}
		}
	}
	if !isInTransaction {
		err = session.Commit()
	}
	return lastId, err
}

func (session *Session) innerInsertMulti(rowsSlicePtr interface{}) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return -1, errors.New("needs a pointer to a slice")
	}

	bean := sliceValue.Index(0).Interface()
	sliceElementType := Type(bean)

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
				arg, err := session.value2Interface(fieldValue)
				if err != nil {
					return 0, err
				}

				args = append(args, arg)
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
				arg, err := session.value2Interface(fieldValue)
				if err != nil {
					return 0, err
				}

				args = append(args, arg)
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

	res, err := session.Exec(statement, args...)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return id, nil
}

func (session *Session) InsertMulti(rowsSlicePtr interface{}) (int64, error) {
	err := session.newDb()
	if session.IsAutoCommit {
		defer session.Close()
	}
	if err != nil {
		return 0, err
	}

	return session.innerInsertMulti(rowsSlicePtr)
}

func (session *Session) value2Interface(fieldValue reflect.Value) (interface{}, error) {
	if fieldValue.Type().Kind() == reflect.Bool {
		if fieldValue.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	} else if fieldValue.Type().String() == "time.Time" {
		return fieldValue.Interface(), nil
	} else if fieldValue.Type().Kind() == reflect.Struct {
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
	} else {
		return fieldValue.Interface(), nil
	}
}

func (session *Session) innerInsert(bean interface{}) (int64, error) {
	table := session.Engine.AutoMap(bean)

	session.Statement.RefTable = table
	colNames := make([]string, 0)
	colPlaces := make([]string, 0)
	var args = make([]interface{}, 0)

	for _, col := range table.Columns {
		if col.MapType == ONLYFROMDB {
			continue
		}

		fieldValue := col.ValueOf(bean)
		if col.IsAutoIncrement && fieldValue.Int() == 0 {
			continue
		}

		if session.Statement.ColumnStr != "" {
			if _, ok := session.Statement.columnMap[col.Name]; !ok {
				continue
			}
		}

		arg, err := session.value2Interface(fieldValue)
		if err != nil {
			return 0, err
		}

		args = append(args, arg)
		colNames = append(colNames, col.Name)
		colPlaces = append(colPlaces, "?")
	}

	sql := fmt.Sprintf("INSERT INTO %v%v%v (%v%v%v) VALUES (%v);",
		session.Engine.QuoteStr(),
		session.Statement.TableName(),
		session.Engine.QuoteStr(),
		session.Engine.QuoteStr(),
		strings.Join(colNames, session.Engine.Quote(", ")),
		session.Engine.QuoteStr(),
		strings.Join(colPlaces, ", "))

	res, err := session.Exec(sql, args...)
	if err != nil {
		return 0, err
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
	case reflect.Int8, reflect.Int16, reflect.Int32:
		v = int(id)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = uint(id)

	}
	pkValue.Set(reflect.ValueOf(v))

	return id, nil
}

func (session *Session) InsertOne(bean interface{}) (int64, error) {
	err := session.newDb()
	if session.IsAutoCommit {
		defer session.Close()
	}
	if err != nil {
		return 0, err
	}

	return session.innerInsert(bean)
}

func (session *Session) Update(bean interface{}, condiBean ...interface{}) (int64, error) {
	err := session.newDb()
	if session.IsAutoCommit {
		defer session.Close()
	}
	if err != nil {
		return 0, err
	}

	table := session.Engine.AutoMap(bean)
	session.Statement.RefTable = table
	colNames, args := BuildConditions(session.Engine, table, bean)
	var condiColNames []string
	var condiArgs []interface{}

	if len(condiBean) > 0 {
		condiColNames, condiArgs = BuildConditions(session.Engine, table, condiBean[0])
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

	eargs := append(append(args, st.Params...), condiArgs...)
	res, err := session.Exec(sql, eargs...)
	if err != nil {
		return -1, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return -1, err
	}
	return rows, nil
}

func (session *Session) Delete(bean interface{}) (int64, error) {
	err := session.newDb()
	if session.IsAutoCommit {
		defer session.Close()
	}
	if err != nil {
		return 0, err
	}

	table := session.Engine.AutoMap(bean)
	session.Statement.RefTable = table
	colNames, args := BuildConditions(session.Engine, table, bean)

	var condition = ""
	st := session.Statement
	defer session.Statement.Init()
	if st.WhereStr != "" {
		condition = fmt.Sprintf("WHERE %v", st.WhereStr)
		if len(colNames) > 0 {
			condition += " and "
			condition += strings.Join(colNames, " and ")
		}
	} else {
		condition = "WHERE " + strings.Join(colNames, " and ")
	}

	statement := fmt.Sprintf("DELETE FROM %v%v%v %v",
		session.Engine.QuoteStr(),
		session.Statement.TableName(),
		session.Engine.QuoteStr(),
		condition)

	res, err := session.Exec(statement, append(st.Params, args...)...)

	if err != nil {
		return -1, err
	}

	id, err := res.RowsAffected()

	if err != nil {
		return -1, err
	}
	return id, nil
}
