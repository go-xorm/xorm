// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"database/sql"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/go-xorm/builder"
	"github.com/go-xorm/core"
)

type EngineGroup struct {
	master  *Engine
	slaves  []*Engine
	weight  []int
	count   int
	s_count int
	policy  Policy
	p       int
}

func NewGroup(args1 interface{}, args2 interface{}, policy ...Policy) (*EngineGroup, error) {
	driverName, ok1 := args1.(string)
	dataSourceNames, ok2 := args2.(string)
	if ok1 && ok2 {
		return newGroup1(driverName, dataSourceNames, policy...)
	}

	Master, ok3 := args1.(*Engine)
	Slaves, ok4 := args2.([]*Engine)
	if ok3 && ok4 {
		return newGroup2(Master, Slaves, policy...)
	}
	return nil, ErrParamsType
}

func newGroup1(driverName string, dataSourceNames string, policy ...Policy) (*EngineGroup, error) {
	conns := strings.Split(dataSourceNames, ";")
	engines := make([]*Engine, len(conns))
	for i, _ := range conns {
		engine, err := NewEngine(driverName, conns[i])
		if err != nil {
			return nil, err
		}
		engines[i] = engine
	}

	n := len(policy)
	if n > 1 {
		return nil, ErrParamsType
	} else if n == 1 {
		eg := &EngineGroup{
			master:  engines[0],
			slaves:  engines[1:],
			count:   len(engines),
			s_count: len(engines[1:]),
			policy:  policy[0],
		}
		eg.policy.Init()
		return eg, nil
	} else {
		xPolicy := new(XormEngineGroupPolicy)
		eg := &EngineGroup{
			master:  engines[0],
			slaves:  engines[1:],
			count:   len(engines),
			s_count: len(engines[1:]),
			policy:  xPolicy,
		}
		xPolicy.Init()
		return eg, nil
	}

}

func newGroup2(Master *Engine, Slaves []*Engine, policy ...Policy) (*EngineGroup, error) {
	n := len(policy)
	if n > 1 {
		return nil, ErrParamsType
	} else if n == 1 {
		eg := &EngineGroup{
			master:  Master,
			slaves:  Slaves,
			count:   1 + len(Slaves),
			s_count: len(Slaves),
			policy:  policy[0],
		}
		eg.policy.Init()
		return eg, nil
	} else {
		xPolicy := new(XormEngineGroupPolicy)
		eg := &EngineGroup{
			master:  Master,
			slaves:  Slaves,
			count:   1 + len(Slaves),
			s_count: len(Slaves),
			policy:  xPolicy,
		}
		xPolicy.Init()
		return eg, nil
	}
}

func (eg *EngineGroup) SetPolicy(policy Policy) *EngineGroup {
	eg.policy = policy
	return eg
}

func (eg *EngineGroup) UsePolicy(policy int) *EngineGroup {
	eg.p = policy
	return eg
}

func (eg *EngineGroup) SetWeight(weight ...interface{}) *EngineGroup {
	l := len(weight)
	if l == 1 {
		switch weight[0].(type) {
		case []int:
			eg.weight = weight[0].([]int)
		}
	} else if l > 1 {
		s := make([]int, 0)
		for i, _ := range weight {
			switch weight[i].(type) {
			case int:
				s = append(s, weight[i].(int))
			default:
				s = append(s, 1)
			}
		}
		eg.weight = s
	}

	return eg
}

func (eg *EngineGroup) Master() *Engine {
	return eg.master
}

// Slave returns one of the physical databases which is a slave
func (eg *EngineGroup) Slave() *Engine {
	if eg.count == 1 {
		return eg.master
	}
	return eg.policy.Slave(eg)
}

func (eg *EngineGroup) Slaves() []*Engine {
	if eg.count == 1 {
		return []*Engine{eg.master}
	}
	return eg.slaves
}

func (eg *EngineGroup) GetSlave(i int) *Engine {
	if eg.count == 1 || i == 0 {
		return eg.master
	}
	if i > eg.s_count {
		return eg.slaves[0]
	}
	return eg.slaves[i]
}

func (eg *EngineGroup) GetEngine(i int) *Engine {
	if i >= eg.count || i == 0 {
		return eg.master
	}
	return eg.slaves[i-1]
}

// ShowSQL show SQL statement or not on logger if log level is great than INFO
func (eg *EngineGroup) ShowSQL(show ...bool) {
	eg.master.ShowSQL(show...)
	for i, _ := range eg.slaves {
		eg.slaves[i].ShowSQL(show...)
	}
}

// ShowExecTime show SQL statement and execute time or not on logger if log level is great than INFO
func (eg *EngineGroup) ShowExecTime(show ...bool) {
	eg.master.ShowExecTime(show...)
	for i, _ := range eg.slaves {
		eg.slaves[i].ShowExecTime(show...)
	}
}

// SetMapper set the name mapping rules
func (eg *EngineGroup) SetMapper(mapper core.IMapper) {
	eg.master.SetTableMapper(mapper)
	eg.master.SetColumnMapper(mapper)
	for i, _ := range eg.slaves {
		eg.slaves[i].SetTableMapper(mapper)
		eg.slaves[i].SetColumnMapper(mapper)
	}
}

// SetTableMapper set the table name mapping rule
func (eg *EngineGroup) SetTableMapper(mapper core.IMapper) {
	eg.master.TableMapper = mapper
	for i, _ := range eg.slaves {
		eg.slaves[i].TableMapper = mapper
	}
}

// SetColumnMapper set the column name mapping rule
func (eg *EngineGroup) SetColumnMapper(mapper core.IMapper) {
	eg.master.ColumnMapper = mapper
	for i, _ := range eg.slaves {
		eg.slaves[i].ColumnMapper = mapper
	}
}

// SetMaxOpenConns is only available for go 1.2+
func (eg *EngineGroup) SetMaxOpenConns(conns int) {
	eg.master.db.SetMaxOpenConns(conns)
	for i, _ := range eg.slaves {
		eg.slaves[i].db.SetMaxOpenConns(conns)
	}
}

// SetMaxIdleConns set the max idle connections on pool, default is 2
func (eg *EngineGroup) SetMaxIdleConns(conns int) {
	eg.master.db.SetMaxIdleConns(conns)
	for i, _ := range eg.slaves {
		eg.slaves[i].db.SetMaxIdleConns(conns)
	}
}

// NoCascade If you do not want to auto cascade load object
func (eg *EngineGroup) NoCascade() *EGSession {
	egs := eg.NewEGSession()
	return egs.NoCascade()
}

// Close the engine
func (eg *EngineGroup) Close() error {
	err := eg.master.db.Close()
	if err != nil {
		return err
	}

	for i, _ := range eg.slaves {
		err := eg.slaves[i].db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Ping tests if database is alive
func (eg *EngineGroup) Ping() error {
	eg.master.Ping()
	return scatter(eg.s_count, func(i int) error {
		return eg.slaves[i].Ping()
	})
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
func (eg *EngineGroup) SetConnMaxLifetime(d time.Duration) {
	eg.master.db.SetConnMaxLifetime(d)
	for i, _ := range eg.slaves {
		eg.slaves[i].db.SetConnMaxLifetime(d)
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
func (eg *EngineGroup) SqlType(c *core.Column) string {
	return eg.Master().SQLType(c)
}

// SQLType A simple wrapper to dialect's core.SqlType method
func (eg *EngineGroup) SQLType(c *core.Column) string {
	return eg.Master().dialect.SqlType(c)
}

// NewSession New a session
func (eg *EngineGroup) NewSession() *Session {
	return eg.Master().NewSession()
}

// NewSession New a session
func (eg *EngineGroup) NewEGSession() *EGSession {
	args := make(map[string]interface{})
	egs := &EGSession{eg: eg, operation: []string{}, args: args}
	return egs
}

type SqlArgs struct {
	query string
	args  []interface{}
}

func (eg *EngineGroup) Sql(query string, args ...interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.Sql(query, args...)
}

func (eg *EngineGroup) SQL(query interface{}, args ...interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.SQL(query, args...)
}

// NoAutoTime Default if your struct has "created" or "updated" filed tag, the fields
// will automatically be filled with current time when Insert or Update
// invoked. Call NoAutoTime if you dont' want to fill automatically.
func (eg *EngineGroup) NoAutoTime() *EGSession {
	egs := eg.NewEGSession()
	return egs.NoAutoTime()
}

type NoAutoConditionArgs struct {
	no []bool
}

// NoAutoCondition disable auto generate Where condition from bean or not
func (eg *EngineGroup) NoAutoCondition(no ...bool) *EGSession {
	egs := eg.NewEGSession()
	return egs.NoAutoCondition(no...)
}

// DBMetas Retrieve all tables, columns, indexes' informations from database.
func (eg *EngineGroup) DBMetas() ([]*core.Table, error) {
	return eg.Master().DBMetas()
}

// DumpAllToFile dump database all table structs and data to a file
func (eg *EngineGroup) DumpAllToFile(fp string, tp ...core.DbType) error {
	return eg.Master().DumpAllToFile(fp, tp...)
}

// DumpAll dump database all table structs and data to w
func (eg *EngineGroup) DumpAll(w io.Writer, tp ...core.DbType) error {
	return eg.Master().DumpAll(w, tp...)
}

// DumpTablesToFile dump specified tables to SQL file.
func (eg *EngineGroup) DumpTablesToFile(tables []*core.Table, fp string, tp ...core.DbType) error {
	return eg.Master().DumpTablesToFile(tables, fp, tp...)
}

// DumpTables dump specify tables to io.Writer
func (eg *EngineGroup) DumpTables(tables []*core.Table, w io.Writer, tp ...core.DbType) error {
	return eg.Master().DumpTables(tables, w, tp...)
}

type CascadeArgs struct {
	trueOrFalse []bool
}

// Cascade use cascade or not
func (eg *EngineGroup) Cascade(trueOrFalse ...bool) *EGSession {
	egs := eg.NewEGSession()
	return egs.Cascade(trueOrFalse...)
}

type WhereArgs struct {
	query interface{}
	args  []interface{}
}

// Where method provide a condition query
func (eg *EngineGroup) Where(query interface{}, args ...interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.Where(query, args...)
}

type IdArgs struct {
	id interface{}
}

// Id will be deprecated, please use ID instead
func (eg *EngineGroup) Id(id interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.Id(id)
}

type IDArgs struct {
	id interface{}
}

// ID method provoide a condition as (id) = ?
func (eg *EngineGroup) ID(id interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.ID(id)
}

type BeforeArgs struct {
	closures func(interface{})
}

// Before apply before Processor, affected bean is passed to closure arg
func (eg *EngineGroup) Before(closures func(interface{})) *EGSession {
	egs := eg.NewEGSession()
	return egs.Before(closures)
}

type AfterArgs struct {
	closures func(interface{})
}

// After apply after insert Processor, affected bean is passed to closure arg
func (eg *EngineGroup) After(closures func(interface{})) *EGSession {
	egs := eg.NewEGSession()
	return egs.After(closures)
}

type CharsetArgs struct {
	charset string
}

// Charset set charset when create table, only support mysql now
func (eg *EngineGroup) Charset(charset string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Charset(charset)
}

type StoreEngineArgs struct {
	storeEngine string
}

// StoreEngine set store engine when create table, only support mysql now
func (eg *EngineGroup) StoreEngine(storeEngine string) *EGSession {
	egs := eg.NewEGSession()
	return egs.StoreEngine(storeEngine)
}

type DistinctArgs struct {
	columns []string
}

// Distinct use for distinct columns. Caution: when you are using cache,
// distinct will not be cached because cache system need id,
// but distinct will not provide id
func (eg *EngineGroup) Distinct(columns ...string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Distinct(columns...)
}

type SelectArgs struct {
	str string
}

// Select customerize your select columns or contents
func (eg *EngineGroup) Select(str string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Select(str)
}

type ColsArgs struct {
	columns []string
}

// Cols only use the parameters as select or update columns
func (eg *EngineGroup) Cols(columns ...string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Cols(columns...)
}

// AllCols indicates that all columns should be use
func (eg *EngineGroup) AllCols() *EGSession {
	egs := eg.NewEGSession()
	return egs.AllCols()
}

type MustColsArgs struct {
	columns []string
}

// MustCols specify some columns must use even if they are empty
func (eg *EngineGroup) MustCols(columns ...string) *EGSession {
	egs := eg.NewEGSession()
	return egs.MustCols(columns...)
}

type UseBoolArgs struct {
	columns []string
}

// UseBool xorm automatically retrieve condition according struct, but
// if struct has bool field, it will ignore them. So use UseBool
// to tell system to do not ignore them.
// If no parameters, it will use all the bool field of struct, or
// it will use parameters's columns
func (eg *EngineGroup) UseBool(columns ...string) *EGSession {
	egs := eg.NewEGSession()
	return egs.UseBool(columns...)
}

type OmitArgs struct {
	columns []string
}

// Omit only not use the parameters as select or update columns
func (eg *EngineGroup) Omit(columns ...string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Omit(columns...)
}

type NullableArgs struct {
	columns []string
}

// Nullable set null when column is zero-value and nullable for update
func (eg *EngineGroup) Nullable(columns ...string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Nullable(columns...)
}

type InArgs struct {
	column string
	args   []interface{}
}

// In will generate "column IN (?, ?)"
func (eg *EngineGroup) In(column string, args ...interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.In(column, args...)
}

type NotInArgs struct {
	column string
	args   []interface{}
}

// NotIn will generate "column NOT IN (?, ?)"
func (eg *EngineGroup) NotIn(column string, args ...interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.NotIn(column, args...)
}

type IncrArgs struct {
	column string
	args   []interface{}
}

// Incr provides a update string like "column = column + ?"
func (eg *EngineGroup) Incr(column string, args ...interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.Incr(column, args...)
}

type DecrArgs struct {
	column string
	args   []interface{}
}

// Decr provides a update string like "column = column - ?"
func (eg *EngineGroup) Decr(column string, args ...interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.Decr(column, args...)
}

type SetExprArgs struct {
	column     string
	expression string
}

// SetExpr provides a update string like "column = {expression}"
func (eg *EngineGroup) SetExpr(column string, expression string) *EGSession {
	egs := eg.NewEGSession()
	return egs.SetExpr(column, expression)
}

type TableArgs struct {
	tableNameOrBean interface{}
}

// Table temporarily change the Get, Find, Update's table
func (eg *EngineGroup) Table(tableNameOrBean interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.Table(tableNameOrBean)
}

type AliasArgs struct {
	alias string
}

// Alias set the table alias
func (eg *EngineGroup) Alias(alias string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Alias(alias)
}

type LimitArgs struct {
	limit int
	start []int
}

// Limit will generate "LIMIT start, limit"
func (eg *EngineGroup) Limit(limit int, start ...int) *EGSession {
	egs := eg.NewEGSession()
	return egs.Limit(limit, start...)
}

type DescArgs struct {
	colNames []string
}

// Desc will generate "ORDER BY column1 DESC, column2 DESC"
func (eg *EngineGroup) Desc(colNames ...string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Desc(colNames...)
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
func (eg *EngineGroup) Asc(colNames ...string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Asc(colNames...)
}

type OrderByArgs struct {
	order string
}

// OrderBy will generate "ORDER BY order"
func (eg *EngineGroup) OrderBy(order string) *EGSession {
	egs := eg.NewEGSession()
	return egs.OrderBy(order)
}

type JoinArgs struct {
	joinOperator string
	tablename    interface{}
	condition    string
	args         []interface{}
}

// Join the join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (eg *EngineGroup) Join(joinOperator string, tablename interface{}, condition string, args ...interface{}) *EGSession {
	egs := eg.NewEGSession()
	return egs.Join(joinOperator, tablename, condition, args...)
}

type GroupByArgs struct {
	keys string
}

// GroupBy generate group by statement
func (eg *EngineGroup) GroupBy(keys string) *EGSession {
	egs := eg.NewEGSession()
	return egs.GroupBy(keys)
}

type HavingArgs struct {
	conditions string
}

// Having generate having statement
func (eg *EngineGroup) Having(conditions string) *EGSession {
	egs := eg.NewEGSession()
	return egs.Having(conditions)
}

// IdOf get id from one struct
//
// Deprecated: use IDOf instead.
func (eg *EngineGroup) IdOf(bean interface{}) core.PK {
	return eg.Master().IdOf(bean)
}

// IDOf get id from one struct
func (eg *EngineGroup) IDOf(bean interface{}) core.PK {
	return eg.Master().IDOf(bean)
}

// IdOfV get id from one value of struct
//
// Deprecated: use IDOfV instead.
func (eg *EngineGroup) IdOfV(rv reflect.Value) core.PK {
	return eg.Master().IdOfV(rv)
}

// IDOfV get id from one value of struct
func (eg *EngineGroup) IDOfV(rv reflect.Value) core.PK {
	return eg.Master().IDOfV(rv)
}

// CreateIndexes create indexes
func (eg *EngineGroup) CreateIndexes(bean interface{}) error {
	return eg.Master().CreateIndexes(bean)
}

// CreateUniques create uniques
func (eg *EngineGroup) CreateUniques(bean interface{}) error {
	return eg.Master().CreateUniques(bean)
}

// Sync the new struct changes to database, this method will automatically add
// table, column, index, unique. but will not delete or change anything.
// If you change some field, you should change the database manually.
func (eg *EngineGroup) Sync(beans ...interface{}) error {
	return eg.Master().Sync(beans...)
}

// Sync2 synchronize structs to database tables
func (eg *EngineGroup) Sync2(beans ...interface{}) error {
	return eg.Master().Sync2(beans...)
}

// CreateTables create tabls according bean
func (eg *EngineGroup) CreateTables(beans ...interface{}) error {
	return eg.Master().CreateTables(beans...)
}

// DropTables drop specify tables
func (eg *EngineGroup) DropTables(beans ...interface{}) error {
	return eg.Master().DropTables(beans...)
}

// DropIndexes drop indexes of a table
func (eg *EngineGroup) DropIndexes(bean interface{}) error {
	return eg.Master().DropIndexes(bean)
}

func (eg *EngineGroup) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return eg.Master().Exec(sql, args...)
}

// Query a raw sql and return records as []map[string][]byte
func (eg *EngineGroup) Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	return eg.Slave().Query(sql, paramStr...)
}

// QueryString runs a raw sql and return records as []map[string]string
func (eg *EngineGroup) QueryString(sqlStr string, args ...interface{}) ([]map[string]string, error) {
	return eg.Slave().QueryString(sqlStr, args...)
}

// QueryInterface runs a raw sql and return records as []map[string]interface{}
func (eg *EngineGroup) QueryInterface(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	return eg.Slave().QueryInterface(sqlStr, args...)
}

// Insert one or more records
func (eg *EngineGroup) Insert(beans ...interface{}) (int64, error) {
	return eg.Master().Insert(beans...)
}

// InsertOne insert only one record
func (eg *EngineGroup) InsertOne(bean interface{}) (int64, error) {
	return eg.Master().InsertOne(bean)
}

// IsTableEmpty if a table has any reocrd
func (eg *EngineGroup) IsTableEmpty(bean interface{}) (bool, error) {
	return eg.Master().IsTableEmpty(bean)
}

// IsTableExist if a table is exist
func (eg *EngineGroup) IsTableExist(beanOrTableName interface{}) (bool, error) {
	return eg.Master().IsTableExist(beanOrTableName)
}

func (eg *EngineGroup) Update(bean interface{}, condiBeans ...interface{}) (int64, error) {
	return eg.Master().Update(bean, condiBeans...)
}

// Delete records, bean's non-empty fields are conditions
func (eg *EngineGroup) Delete(bean interface{}) (int64, error) {
	return eg.Master().Delete(bean)
}

// Get retrieve one record from table, bean's non-empty fields
// are conditions
func (eg *EngineGroup) Get(bean interface{}) (bool, error) {
	return eg.Slave().Get(bean)
}

// Exist returns true if the record exist otherwise return false
func (eg *EngineGroup) Exist(bean ...interface{}) (bool, error) {
	return eg.Slave().Exist(bean...)
}

// Iterate record by record handle records from table, bean's non-empty fields
// are conditions.
func (eg *EngineGroup) Iterate(bean interface{}, fun IterFunc) error {
	return eg.Master().Iterate(bean, fun)
}

func (eg *EngineGroup) Find(beans interface{}, condiBeans ...interface{}) error {
	return eg.Slave().Find(beans, condiBeans...)
}

// Rows return sql.Rows compatible Rows obj, as a forward Iterator object for iterating record by record, bean's non-empty fields
// are conditions.
func (eg *EngineGroup) Rows(bean interface{}) (*Rows, error) {
	return eg.Slave().Rows(bean)
}

// Count counts the records. bean's non-empty fields are conditions.
func (eg *EngineGroup) Count(bean ...interface{}) (int64, error) {
	return eg.Slave().Count(bean...)
}

// Sum sum the records by some column. bean's non-empty fields are conditions.
func (eg *EngineGroup) Sum(bean interface{}, colName string) (float64, error) {
	return eg.Slave().Sum(bean, colName)
}

// SumInt sum the records by some column. bean's non-empty fields are conditions.
func (eg *EngineGroup) SumInt(bean interface{}, colName string) (int64, error) {
	return eg.Slave().SumInt(bean, colName)
}

// Sums sum the records by some columns. bean's non-empty fields are conditions.
func (eg *EngineGroup) Sums(bean interface{}, colNames ...string) ([]float64, error) {
	return eg.Slave().Sums(bean, colNames...)
}

// SumsInt like Sums but return slice of int64 instead of float64.
func (eg *EngineGroup) SumsInt(bean interface{}, colNames ...string) ([]int64, error) {
	return eg.Slave().SumsInt(bean, colNames...)
}

// ImportFile SQL DDL file
func (eg *EngineGroup) ImportFile(ddlPath string) ([]sql.Result, error) {
	return eg.Master().ImportFile(ddlPath)
}

// Import SQL DDL from io.Reader
func (eg *EngineGroup) Import(r io.Reader) ([]sql.Result, error) {
	return eg.Master().Import(r)
}

// NowTime2 return current time
func (eg *EngineGroup) NowTime2(sqlTypeName string) (interface{}, time.Time) {
	return eg.Master().NowTime2(sqlTypeName)
}

// Unscoped always disable struct tag "deleted"
func (eg *EngineGroup) Unscoped() *EGSession {
	egs := eg.NewEGSession()
	return egs.Unscoped()
}

// CondDeleted returns the conditions whether a record is soft deleted.
func (eg *EngineGroup) CondDeleted(colName string) builder.Cond {
	return eg.Master().CondDeleted(colName)
}

type BufferSizeArgs struct {
	size int
}

// BufferSize sets buffer size for iterate
func (eg *EngineGroup) BufferSize(size int) *EGSession {
	egs := eg.NewEGSession()
	return egs.BufferSize(size)
}
