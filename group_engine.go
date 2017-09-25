// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"database/sql"
	"io"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/go-xorm/builder"
	"github.com/go-xorm/core"
)

type GroupEngine struct {
	engines []*Engine
	count   uint64
}

func (ge *GroupEngine) Master() *Engine {
	return ge.engines[0]
}

// Slave returns one of the physical databases which is a slave
func (ge *GroupEngine) Slave() *Engine {
	return ge.engines[ge.slave(len(ge.engines))]
}

func (ge *GroupEngine) GetEngine(i int) *Engine {
	if i >= len(ge.engines) {
		return ge.engines[0]
	}
	return ge.engines[i]
}

func (ge *GroupEngine) slave(n int) int {
	if n <= 1 {
		return 0
	}
	return int(1 + (atomic.AddUint64(&ge.count, 1) % uint64(n-1)))
}

// ShowSQL show SQL statement or not on logger if log level is great than INFO
func (ge *GroupEngine) ShowSQL(show ...bool) {
	for i, _ := range ge.engines {
		ge.engines[i].ShowSQL(show...)
	}
}

// ShowExecTime show SQL statement and execute time or not on logger if log level is great than INFO
func (ge *GroupEngine) ShowExecTime(show ...bool) {
	for i, _ := range ge.engines {
		ge.engines[i].ShowExecTime(show...)
	}
}

// SetMapper set the name mapping rules
func (ge *GroupEngine) SetMapper(mapper core.IMapper) {
	for i, _ := range ge.engines {
		ge.engines[i].SetTableMapper(mapper)
		ge.engines[i].SetColumnMapper(mapper)
	}
}

// SetTableMapper set the table name mapping rule
func (ge *GroupEngine) SetTableMapper(mapper core.IMapper) {
	for i, _ := range ge.engines {
		ge.engines[i].TableMapper = mapper
	}
}

// SetColumnMapper set the column name mapping rule
func (ge *GroupEngine) SetColumnMapper(mapper core.IMapper) {
	for i, _ := range ge.engines {
		ge.engines[i].ColumnMapper = mapper
	}
}

// SetMaxOpenConns is only available for go 1.2+
func (ge *GroupEngine) SetMaxOpenConns(conns int) {
	for i, _ := range ge.engines {
		ge.engines[i].db.SetMaxOpenConns(conns)
	}
}

// SetMaxIdleConns set the max idle connections on pool, default is 2
func (ge *GroupEngine) SetMaxIdleConns(conns int) {
	for i, _ := range ge.engines {
		ge.engines[i].db.SetMaxIdleConns(conns)
	}
}

// NoCascade If you do not want to auto cascade load object
func (ge *GroupEngine) NoCascade() *GESession {
	ges := ge.NewGESession()
	return ges.NoCascade()
}

// Close the engine
func (ge *GroupEngine) Close() error {
	for i, _ := range ge.engines {
		err := ge.engines[i].db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Ping tests if database is alive
func (ge *GroupEngine) Ping() error {
	return scatter(len(ge.engines), func(i int) error {
		return ge.engines[i].Ping()
	})
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
func (ge *GroupEngine) SetConnMaxLifetime(d time.Duration) {
	for i, _ := range ge.engines {
		ge.engines[i].db.SetConnMaxLifetime(d)
	}
}

func scatter(n int, fn func(i int) error) error {
	errors := make(chan error, n)

	var i int
	for i = 0; i < n; i++ {
		go func(i int) { errors <- fn(i) }(i)
	}

	var err, innerErr error
	for i = 0; i < cap(errors); i++ {
		if innerErr = <-errors; innerErr != nil {
			err = innerErr
		}
	}

	return err
}

// SqlType will be deprecated, please use SQLType instead
//
// Deprecated: use SQLType instead
func (ge *GroupEngine) SqlType(c *core.Column) string {
	return ge.Master().SQLType(c)
}

// SQLType A simple wrapper to dialect's core.SqlType method
func (ge *GroupEngine) SQLType(c *core.Column) string {
	return ge.Master().dialect.SqlType(c)
}

// NewSession New a session
func (ge *GroupEngine) NewSession() *Session {
	return ge.Master().NewSession()
}

// NewSession New a session
func (ge *GroupEngine) NewGESession() *GESession {
	args := make(map[string]interface{})
	ges := &GESession{ge: ge, operation: []string{}, args: args}
	return ges
}

type SqlArgs struct {
	query interface{}
	args  []interface{}
}

func (ge *GroupEngine) Sql(query interface{}, args ...interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.Sql(query, args...)
}

func (ge *GroupEngine) SQL(query interface{}, args ...interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.SQL(query, args...)
}

// NoAutoTime Default if your struct has "created" or "updated" filed tag, the fields
// will automatically be filled with current time when Insert or Update
// invoked. Call NoAutoTime if you dont' want to fill automatically.
func (ge *GroupEngine) NoAutoTime() *GESession {
	ges := ge.NewGESession()
	return ges.NoAutoTime()
}

type NoAutoConditionArgs struct {
	no []bool
}

// NoAutoCondition disable auto generate Where condition from bean or not
func (ge *GroupEngine) NoAutoCondition(no ...bool) *GESession {
	ges := ge.NewGESession()
	return ges.NoAutoCondition(no...)
}

// DBMetas Retrieve all tables, columns, indexes' informations from database.
func (ge *GroupEngine) DBMetas() ([]*core.Table, error) {
	return ge.Master().DBMetas()
}

// DumpAllToFile dump database all table structs and data to a file
func (ge *GroupEngine) DumpAllToFile(fp string, tp ...core.DbType) error {
	return ge.Master().DumpAllToFile(fp, tp...)
}

// DumpAll dump database all table structs and data to w
func (ge *GroupEngine) DumpAll(w io.Writer, tp ...core.DbType) error {
	return ge.Master().DumpAll(w, tp...)
}

// DumpTablesToFile dump specified tables to SQL file.
func (ge *GroupEngine) DumpTablesToFile(tables []*core.Table, fp string, tp ...core.DbType) error {
	return ge.Master().DumpTablesToFile(tables, fp, tp...)
}

// DumpTables dump specify tables to io.Writer
func (ge *GroupEngine) DumpTables(tables []*core.Table, w io.Writer, tp ...core.DbType) error {
	return ge.Master().DumpTables(tables, w, tp...)
}

type CascadeArgs struct {
	trueOrFalse []bool
}

// Cascade use cascade or not
func (ge *GroupEngine) Cascade(trueOrFalse ...bool) *GESession {
	ges := ge.NewGESession()
	return ges.Cascade(trueOrFalse...)
}

type WhereArgs struct {
	query interface{}
	args  []interface{}
}

// Where method provide a condition query
func (ge *GroupEngine) Where(query interface{}, args ...interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.Where(query, args...)
}

type IdArgs struct {
	id interface{}
}

// Id will be deprecated, please use ID instead
func (ge *GroupEngine) Id(id interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.Id(id)
}

type IDArgs struct {
	id interface{}
}

// ID method provoide a condition as (id) = ?
func (ge *GroupEngine) ID(id interface{}) *GESession {
	ges := ge.NewGESession()
	ges.operation = append(ges.operation, "ID")
	args := IDArgs{
		id: id,
	}
	ges.args["ID"] = args
	return ges
}

type BeforeArgs struct {
	closures func(interface{})
}

// Before apply before Processor, affected bean is passed to closure arg
func (ge *GroupEngine) Before(closures func(interface{})) *GESession {
	ges := ge.NewGESession()
	return ges.Before(closures)
}

type AfterArgs struct {
	closures func(interface{})
}

// After apply after insert Processor, affected bean is passed to closure arg
func (ge *GroupEngine) After(closures func(interface{})) *GESession {
	ges := ge.NewGESession()
	return ges.After(closures)
}

type CharsetArgs struct {
	charset string
}

// Charset set charset when create table, only support mysql now
func (ge *GroupEngine) Charset(charset string) *GESession {
	ges := ge.NewGESession()
	return ges.Charset(charset)
}

type StoreEngineArgs struct {
	storeEngine string
}

// StoreEngine set store engine when create table, only support mysql now
func (ge *GroupEngine) StoreEngine(storeEngine string) *GESession {
	ges := ge.NewGESession()
	return ges.StoreEngine(storeEngine)
}

type DistinctArgs struct {
	columns []string
}

// Distinct use for distinct columns. Caution: when you are using cache,
// distinct will not be cached because cache system need id,
// but distinct will not provide id
func (ge *GroupEngine) Distinct(columns ...string) *GESession {
	ges := ge.NewGESession()
	return ges.Distinct(columns...)
}

type SelectArgs struct {
	str string
}

// Select customerize your select columns or contents
func (ge *GroupEngine) Select(str string) *GESession {
	ges := ge.NewGESession()
	return ges.Select(str)
}

type ColsArgs struct {
	columns []string
}

// Cols only use the parameters as select or update columns
func (ge *GroupEngine) Cols(columns ...string) *GESession {
	ges := ge.NewGESession()
	return ges.Cols(columns...)
}

// AllCols indicates that all columns should be use
func (ge *GroupEngine) AllCols() *GESession {
	ges := ge.NewGESession()
	return ges.AllCols()
}

type MustColsArgs struct {
	columns []string
}

// MustCols specify some columns must use even if they are empty
func (ge *GroupEngine) MustCols(columns ...string) *GESession {
	ges := ge.NewGESession()
	return ges.MustCols(columns...)
}

type UseBoolArgs struct {
	columns []string
}

// UseBool xorm automatically retrieve condition according struct, but
// if struct has bool field, it will ignore them. So use UseBool
// to tell system to do not ignore them.
// If no parameters, it will use all the bool field of struct, or
// it will use parameters's columns
func (ge *GroupEngine) UseBool(columns ...string) *GESession {
	ges := ge.NewGESession()
	return ges.UseBool(columns...)
}

type OmitArgs struct {
	columns []string
}

// Omit only not use the parameters as select or update columns
func (ge *GroupEngine) Omit(columns ...string) *GESession {
	ges := ge.NewGESession()
	return ges.Omit(columns...)
}

type NullableArgs struct {
	columns []string
}

// Nullable set null when column is zero-value and nullable for update
func (ge *GroupEngine) Nullable(columns ...string) *GESession {
	ges := ge.NewGESession()
	return ges.Nullable(columns...)
}

type InArgs struct {
	column string
	args   []interface{}
}

// In will generate "column IN (?, ?)"
func (ge *GroupEngine) In(column string, args ...interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.In(column, args...)
}

type NotInArgs struct {
	column string
	args   []interface{}
}

// NotIn will generate "column NOT IN (?, ?)"
func (ge *GroupEngine) NotIn(column string, args ...interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.NotIn(column, args...)
}

type IncrArgs struct {
	column string
	args   []interface{}
}

// Incr provides a update string like "column = column + ?"
func (ge *GroupEngine) Incr(column string, args ...interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.Incr(column, args...)
}

type DecrArgs struct {
	column string
	args   []interface{}
}

// Decr provides a update string like "column = column - ?"
func (ge *GroupEngine) Decr(column string, args ...interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.Decr(column, args...)
}

type SetExprArgs struct {
	column     string
	expression string
}

// SetExpr provides a update string like "column = {expression}"
func (ge *GroupEngine) SetExpr(column string, expression string) *GESession {
	ges := ge.NewGESession()
	return ges.SetExpr(column, expression)
}

type TableArgs struct {
	tableNameOrBean interface{}
}

// Table temporarily change the Get, Find, Update's table
func (ge *GroupEngine) Table(tableNameOrBean interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.Table(tableNameOrBean)
}

type AliasArgs struct {
	alias string
}

// Alias set the table alias
func (ge *GroupEngine) Alias(alias string) *GESession {
	ges := ge.NewGESession()
	return ges.Alias(alias)
}

type LimitArgs struct {
	limit int
	start []int
}

// Limit will generate "LIMIT start, limit"
func (ge *GroupEngine) Limit(limit int, start ...int) *GESession {
	ges := ge.NewGESession()
	return ges.Limit(limit, start...)
}

type DescArgs struct {
	colNames []string
}

// Desc will generate "ORDER BY column1 DESC, column2 DESC"
func (ge *GroupEngine) Desc(colNames ...string) *GESession {
	ges := ge.NewGESession()
	return ges.Desc(colNames...)
}

type AscArgs struct {
	colNames []string
}

// Asc will generate "ORDER BY column1,column2 Asc"
// This method can chainable use.
//
//        engine.Desc("name").Asc("age").Find(&users)
//        // SELECT * FROM user ORDER BY name DESC, age ASC
//
func (ge *GroupEngine) Asc(colNames ...string) *GESession {
	ges := ge.NewGESession()
	return ges.Asc(colNames...)
}

type OrderByArgs struct {
	order string
}

// OrderBy will generate "ORDER BY order"
func (ge *GroupEngine) OrderBy(order string) *GESession {
	ges := ge.NewGESession()
	return ges.OrderBy(order)
}

type JoinArgs struct {
	joinOperator string
	tablename    interface{}
	condition    string
	args         []interface{}
}

// Join the join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (ge *GroupEngine) Join(joinOperator string, tablename interface{}, condition string, args ...interface{}) *GESession {
	ges := ge.NewGESession()
	return ges.Join(joinOperator, tablename, condition, args...)
}

type GroupByArgs struct {
	keys string
}

// GroupBy generate group by statement
func (ge *GroupEngine) GroupBy(keys string) *GESession {
	ges := ge.NewGESession()
	return ges.GroupBy(keys)
}

type HavingArgs struct {
	conditions string
}

// Having generate having statement
func (ge *GroupEngine) Having(conditions string) *GESession {
	ges := ge.NewGESession()
	return ges.Having(conditions)
}

// IdOf get id from one struct
//
// Deprecated: use IDOf instead.
func (ge *GroupEngine) IdOf(bean interface{}) core.PK {
	return ge.Master().IdOf(bean)
}

// IDOf get id from one struct
func (ge *GroupEngine) IDOf(bean interface{}) core.PK {
	return ge.Master().IDOf(bean)
}

// IdOfV get id from one value of struct
//
// Deprecated: use IDOfV instead.
func (ge *GroupEngine) IdOfV(rv reflect.Value) core.PK {
	return ge.Master().IdOfV(rv)
}

// IDOfV get id from one value of struct
func (ge *GroupEngine) IDOfV(rv reflect.Value) core.PK {
	return ge.Master().IDOfV(rv)
}

// CreateIndexes create indexes
func (ge *GroupEngine) CreateIndexes(bean interface{}) error {
	return ge.Master().CreateIndexes(bean)
}

// CreateUniques create uniques
func (ge *GroupEngine) CreateUniques(bean interface{}) error {
	return ge.Master().CreateUniques(bean)
}

// Sync the new struct changes to database, this method will automatically add
// table, column, index, unique. but will not delete or change anything.
// If you change some field, you should change the database manually.
func (ge *GroupEngine) Sync(beans ...interface{}) error {
	return ge.Master().Sync(beans...)
}

// Sync2 synchronize structs to database tables
func (ge *GroupEngine) Sync2(beans ...interface{}) error {
	return ge.Master().Sync2(beans...)
}

// CreateTables create tabls according bean
func (ge *GroupEngine) CreateTables(beans ...interface{}) error {
	return ge.Master().CreateTables(beans...)
}

// DropTables drop specify tables
func (ge *GroupEngine) DropTables(beans ...interface{}) error {
	return ge.Master().DropTables(beans...)
}

// DropIndexes drop indexes of a table
func (ge *GroupEngine) DropIndexes(bean interface{}) error {
	return ge.Master().DropIndexes(bean)
}

func (ge *GroupEngine) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return ge.Master().Exec(sql, args...)
}

// Query a raw sql and return records as []map[string][]byte
func (ge *GroupEngine) Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	return ge.Slave().Query(sql, paramStr...)
}

// QueryString runs a raw sql and return records as []map[string]string
func (ge *GroupEngine) QueryString(sqlStr string, args ...interface{}) ([]map[string]string, error) {
	return ge.Slave().QueryString(sqlStr, args...)
}

// QueryInterface runs a raw sql and return records as []map[string]interface{}
func (ge *GroupEngine) QueryInterface(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	return ge.Slave().QueryInterface(sqlStr, args...)
}

// Insert one or more records
func (ge *GroupEngine) Insert(beans ...interface{}) (int64, error) {
	return ge.Master().Insert(beans...)
}

// InsertOne insert only one record
func (ge *GroupEngine) InsertOne(bean interface{}) (int64, error) {
	return ge.Master().InsertOne(bean)
}

// IsTableEmpty if a table has any reocrd
func (ge *GroupEngine) IsTableEmpty(bean interface{}) (bool, error) {
	return ge.Master().IsTableEmpty(bean)
}

// IsTableExist if a table is exist
func (ge *GroupEngine) IsTableExist(beanOrTableName interface{}) (bool, error) {
	return ge.Master().IsTableExist(beanOrTableName)
}

func (ge *GroupEngine) Update(bean interface{}, condiBeans ...interface{}) (int64, error) {
	return ge.Master().Update(bean, condiBeans...)
}

// Delete records, bean's non-empty fields are conditions
func (ge *GroupEngine) Delete(bean interface{}) (int64, error) {
	return ge.Master().Delete(bean)
}

// Get retrieve one record from table, bean's non-empty fields
// are conditions
func (ge *GroupEngine) Get(bean interface{}) (bool, error) {
	return ge.Slave().Get(bean)
}

// Exist returns true if the record exist otherwise return false
func (ge *GroupEngine) Exist(bean ...interface{}) (bool, error) {
	return ge.Slave().Exist(bean...)
}

// Iterate record by record handle records from table, bean's non-empty fields
// are conditions.
func (ge *GroupEngine) Iterate(bean interface{}, fun IterFunc) error {
	return ge.Master().Iterate(bean, fun)
}

func (ge *GroupEngine) Find(beans interface{}, condiBeans ...interface{}) error {
	return ge.Slave().Find(beans, condiBeans...)
}

// Rows return sql.Rows compatible Rows obj, as a forward Iterator object for iterating record by record, bean's non-empty fields
// are conditions.
func (ge *GroupEngine) Rows(bean interface{}) (*Rows, error) {
	return ge.Slave().Rows(bean)
}

// Count counts the records. bean's non-empty fields are conditions.
func (ge *GroupEngine) Count(bean ...interface{}) (int64, error) {
	return ge.Slave().Count(bean...)
}

// Sum sum the records by some column. bean's non-empty fields are conditions.
func (ge *GroupEngine) Sum(bean interface{}, colName string) (float64, error) {
	return ge.Slave().Sum(bean, colName)
}

// SumInt sum the records by some column. bean's non-empty fields are conditions.
func (ge *GroupEngine) SumInt(bean interface{}, colName string) (int64, error) {
	return ge.Slave().SumInt(bean, colName)
}

// Sums sum the records by some columns. bean's non-empty fields are conditions.
func (ge *GroupEngine) Sums(bean interface{}, colNames ...string) ([]float64, error) {
	return ge.Slave().Sums(bean, colNames...)
}

// SumsInt like Sums but return slice of int64 instead of float64.
func (ge *GroupEngine) SumsInt(bean interface{}, colNames ...string) ([]int64, error) {
	return ge.Slave().SumsInt(bean, colNames...)
}

// ImportFile SQL DDL file
func (ge *GroupEngine) ImportFile(ddlPath string) ([]sql.Result, error) {
	return ge.Master().ImportFile(ddlPath)
}

// Import SQL DDL from io.Reader
func (ge *GroupEngine) Import(r io.Reader) ([]sql.Result, error) {
	return ge.Master().Import(r)
}

// NowTime2 return current time
func (ge *GroupEngine) NowTime2(sqlTypeName string) (interface{}, time.Time) {
	return ge.Master().NowTime2(sqlTypeName)
}

// Unscoped always disable struct tag "deleted"
func (ge *GroupEngine) Unscoped() *GESession {
	ges := ge.NewGESession()
	return ges.Unscoped()
}

// CondDeleted returns the conditions whether a record is soft deleted.
func (ge *GroupEngine) CondDeleted(colName string) builder.Cond {
	return ge.Master().CondDeleted(colName)
}

type BufferSizeArgs struct {
	size int
}

// BufferSize sets buffer size for iterate
func (ge *GroupEngine) BufferSize(size int) *GESession {
	ges := ge.NewGESession()
	return ges.BufferSize(size)
}
