// Copyright 2013 The XORM Authors. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// Package xorm provides is a simple and powerful ORM for Go. It makes your
// database operation simple.

package xorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	PQSQL   = "pqsql"
	MSSQL   = "mssql"
	SQLITE  = "sqlite3"
	MYSQL   = "mysql"
	MYMYSQL = "mymysql"
)

type dialect interface {
	SqlType(t *Column) string
	SupportInsertMany() bool
	QuoteIdentifier() string
	AutoIncrIdentifier() string
}

type Engine struct {
	Mapper         IMapper
	TagIdentifier  string
	DriverName     string
	DataSourceName string
	Dialect        dialect
	Tables         map[reflect.Type]*Table
	mutex          *sync.Mutex
	ShowSQL        bool
	pool           IConnectPool
	CacheMapping   bool
}

func (engine *Engine) SupportInsertMany() bool {
	return engine.Dialect.SupportInsertMany()
}

func (engine *Engine) QuoteIdentifier() string {
	return engine.Dialect.QuoteIdentifier()
}

func (engine *Engine) AutoIncrIdentifier() string {
	return engine.Dialect.AutoIncrIdentifier()
}

func (engine *Engine) SetPool(pool IConnectPool) error {
	engine.pool = pool
	return engine.pool.Init(engine)
}

func Type(bean interface{}) reflect.Type {
	sliceValue := reflect.Indirect(reflect.ValueOf(bean))
	return reflect.TypeOf(sliceValue.Interface())
}

func StructName(v reflect.Type) string {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Name()
}

func (e *Engine) OpenDB() (*sql.DB, error) {
	return sql.Open(e.DriverName, e.DataSourceName)
}

func (engine *Engine) NewSession() *Session {
	session := &Session{Engine: engine}
	session.Init()
	return session
}

func (engine *Engine) Close() error {
	return engine.pool.Close(engine)
}

func (engine *Engine) Test() error {
	session := engine.NewSession()
	defer session.Close()
	if engine.ShowSQL {
		fmt.Printf("PING DATABASE %v\n", engine.DriverName)
	}
	return session.Ping()
}

func (engine *Engine) Sql(querystring string, args ...interface{}) *Session {
	session := engine.NewSession()
	session.Sql(querystring, args...)
	return session
}

func (engine *Engine) Where(querystring string, args ...interface{}) *Session {
	session := engine.NewSession()
	session.Where(querystring, args...)
	return session
}

func (engine *Engine) Id(id int64) *Session {
	session := engine.NewSession()
	session.Id(id)
	return session
}

func (engine *Engine) In(column string, args ...interface{}) *Session {
	session := engine.NewSession()
	session.In(column, args...)
	return session
}

func (engine *Engine) Table(tableName string) *Session {
	session := engine.NewSession()
	session.Table(tableName)
	return session
}

func (engine *Engine) Limit(limit int, start ...int) *Session {
	session := engine.NewSession()
	session.Limit(limit, start...)
	return session
}

func (engine *Engine) OrderBy(order string) *Session {
	session := engine.NewSession()
	session.OrderBy(order)
	return session
}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (engine *Engine) Join(join_operator, tablename, condition string) *Session {
	session := engine.NewSession()
	session.Join(join_operator, tablename, condition)
	return session
}

func (engine *Engine) GroupBy(keys string) *Session {
	session := engine.NewSession()
	session.GroupBy(keys)
	return session
}

func (engine *Engine) Having(conditions string) *Session {
	session := engine.NewSession()
	session.Having(conditions)
	return session
}

// some lock needed
func (engine *Engine) AutoMapType(t reflect.Type) *Table {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	table, ok := engine.Tables[t]
	if !ok {
		table = engine.MapType(t)
		engine.Tables[t] = table
	}
	return table
}

func (engine *Engine) AutoMap(bean interface{}) *Table {
	t := Type(bean)
	return engine.AutoMapType(t)
}

func (engine *Engine) MapType(t reflect.Type) *Table {
	table := &Table{Name: engine.Mapper.Obj2Table(t.Name()), Type: t}
	table.Columns = make(map[string]Column)

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		ormTagStr := tag.Get(engine.TagIdentifier)
		var col Column
		fieldType := t.Field(i).Type

		if ormTagStr != "" {
			col = Column{FieldName: t.Field(i).Name, Nullable: true, IsPrimaryKey: false,
				IsAutoIncrement: false, MapType: TWOSIDES}
			ormTagStr = strings.ToLower(ormTagStr)
			tags := strings.Split(ormTagStr, " ")
			// TODO:
			if len(tags) > 0 {
				if tags[0] == "-" {
					continue
				}
				if (tags[0] == "extends") &&
					(fieldType.Kind() == reflect.Struct) &&
					t.Field(i).Anonymous {
					parentTable := engine.MapType(fieldType)
					for name, col := range parentTable.Columns {
						col.FieldName = fmt.Sprintf("%v.%v", fieldType.Name(), col.FieldName)
						table.Columns[name] = col
					}
				}
				for j, key := range tags {
					k := strings.ToLower(key)
					switch {
					case k == "<-":
						col.MapType = ONLYFROMDB
					case k == "->":
						col.MapType = ONLYTODB
					case k == "pk":
						col.IsPrimaryKey = true
						col.Nullable = false
					case k == "null":
						col.Nullable = (tags[j-1] != "not")
					case k == "autoincr":
						col.IsAutoIncrement = true
					case k == "default":
						col.Default = tags[j+1]
					case k == "text":
						col.SQLType = Text
					case strings.HasPrefix(k, "int"):
						if k == "int" {
							col.SQLType = Int
							col.Length = Int.DefaultLength
							col.Length2 = Int.DefaultLength2
						} else {
							col.SQLType = Int
							lens := k[len("int")+1 : len(k)-1]
							col.Length, _ = strconv.Atoi(lens)
						}
					case strings.HasPrefix(k, "varchar"):
						col.SQLType = Varchar
						lens := k[len("varchar")+1 : len(k)-1]
						col.Length, _ = strconv.Atoi(lens)
					case strings.HasPrefix(k, "decimal"):
						col.SQLType = Decimal
						lens := k[len("decimal")+1 : len(k)-1]
						twolen := strings.Split(lens, ",")
						col.Length, _ = strconv.Atoi(twolen[0])
						col.Length2, _ = strconv.Atoi(twolen[1])
					case k == "date":
						col.SQLType = Date
					case k == "datetime":
						col.SQLType = DateTime
					case k == "timestamp":
						col.SQLType = TimeStamp
					case k == "unique":
						col.IsUnique = true
					case k == "not":
					default:
						if k != col.Default {
							col.Name = k
						}
					}
				}
				if col.SQLType.Name == "" {
					col.SQLType = Type2SQLType(fieldType)
				}

				if col.Length == 0 {
					col.Length = col.SQLType.DefaultLength
				}
				if col.Length2 == 0 {
					col.Length2 = col.SQLType.DefaultLength2
				}

				if col.Name == "" {
					col.Name = engine.Mapper.Obj2Table(t.Field(i).Name)
				}
				if col.IsPrimaryKey {
					table.PrimaryKey = col.Name
				}
			}
		} else {
			sqlType := Type2SQLType(fieldType)
			col = Column{engine.Mapper.Obj2Table(t.Field(i).Name), t.Field(i).Name, sqlType,
				sqlType.DefaultLength, sqlType.DefaultLength2, true, "", false, false, false, TWOSIDES}

			if col.Name == "id" {
				col.IsPrimaryKey = true
				col.IsAutoIncrement = true
				col.Nullable = false
				table.PrimaryKey = col.Name
			}
		}
		table.Columns[col.Name] = col
	}

	return table
}

// Map should use after all operation because it's not thread safe
func (engine *Engine) Map(beans ...interface{}) (e error) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	for _, bean := range beans {
		t := Type(bean)
		engine.Tables[t] = engine.MapType(t)
	}
	return
}

func (engine *Engine) UnMap(beans ...interface{}) (e error) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	for _, bean := range beans {
		t := Type(bean)
		if _, ok := engine.Tables[t]; ok {
			delete(engine.Tables, t)
		}
	}
	return
}

func (e *Engine) DropAll() error {
	session := e.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}
	err = session.DropAll()
	if err != nil {
		return session.Rollback()
	}
	return session.Commit()
}

func (e *Engine) CreateTables(beans ...interface{}) error {
	session := e.NewSession()
	err := session.Begin()
	defer session.Close()
	if err != nil {
		return err
	}

	for _, bean := range beans {
		err = session.CreateTable(bean)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return session.Commit()
}

func (e *Engine) CreateAll() error {
	session := e.NewSession()
	err := session.Begin()
	defer session.Close()
	if err != nil {
		return err
	}

	err = session.CreateAll()
	if err != nil {
		return session.Rollback()
	}
	return session.Commit()
}

func (engine *Engine) Exec(sql string, args ...interface{}) (sql.Result, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Exec(sql, args...)
}

func (engine *Engine) Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Query(sql, paramStr...)
}

func (engine *Engine) Insert(beans ...interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Insert(beans...)
}

func (engine *Engine) Update(bean interface{}, condiBeans ...interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Update(bean, condiBeans...)
}

func (engine *Engine) Delete(bean interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Delete(bean)
}

func (engine *Engine) Get(bean interface{}) (bool, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Get(bean)
}

func (engine *Engine) Find(beans interface{}, condiBeans ...interface{}) error {
	session := engine.NewSession()
	defer session.Close()
	return session.Find(beans, condiBeans...)
}

func (engine *Engine) Count(bean interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Count(bean)
}
