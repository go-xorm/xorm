package xorm

import (
	"database/sql"

	"github.com/go-xorm/builder"
)

type GESession struct {
	ge        *GroupEngine
	operation []string
	args      map[string]interface{}
	err       error
}

func (ges *GESession) operate(session *Session) *Session {
	for _, v := range ges.operation {
		switch v {
		case "Before":
			args := ges.args["Before"].(BeforeArgs)
			session = session.Before(args.closures)
		case "After":
			args := ges.args["After"].(AfterArgs)
			session = session.After(args.closures)
		case "Table":
			args := ges.args["Table"].(TableArgs)
			session = session.Table(args.tableNameOrBean)
		case "Alias":
			args := ges.args["Alias"].(AliasArgs)
			session = session.Alias(args.alias)
		case "NoCascade":
			session = session.NoCascade()
		case "ForUpdate":
			session = session.ForUpdate()
		case "NoAutoCondition":
			args := ges.args["NoAutoCondition"].(NoAutoConditionArgs)
			session = session.NoAutoCondition(args.no...)
		case "Limit":
			args := ges.args["Limit"].(LimitArgs)
			session = session.Limit(args.limit, args.start...)
		case "OrderBy":
			args := ges.args["OrderBy"].(OrderByArgs)
			session = session.OrderBy(args.order)
		case "Desc":
			args := ges.args["Desc"].(DescArgs)
			session = session.Desc(args.colNames...)
		case "Asc":
			args := ges.args["Asc"].(AscArgs)
			session = session.Asc(args.colNames...)
		case "StoreEngine":
			args := ges.args["StoreEngine"].(StoreEngineArgs)
			session = session.StoreEngine(args.storeEngine)
		case "Charset":
			args := ges.args["Charset"].(CharsetArgs)
			session = session.Charset(args.charset)
		case "Cascade":
			args := ges.args["Cascade"].(CascadeArgs)
			session = session.Cascade(args.trueOrFalse...)
		case "NoCache":
			session = session.NoCache()
		case "Join":
			args := ges.args["Join"].(JoinArgs)
			session = session.Join(args.joinOperator, args.tablename, args.condition, args.args...)
		case "GroupBy":
			args := ges.args["GroupBy"].(GroupByArgs)
			session = session.GroupBy(args.keys)
		case "Having":
			args := ges.args["Having"].(HavingArgs)
			session = session.Having(args.conditions)
		case "Unscoped":
			session = session.Unscoped()
		case "Incr":
			args := ges.args["Incr"].(IncrArgs)
			session = session.Incr(args.column, args.args...)
		case "Decr":
			args := ges.args["Decr"].(DecrArgs)
			session = session.Decr(args.column, args.args...)
		case "SetExpr":
			args := ges.args["SetExpr"].(SetExprArgs)
			session = session.SetExpr(args.column, args.expression)
		case "Select":
			args := ges.args["Select"].(SelectArgs)
			session = session.Select(args.str)
		case "Cols":
			args := ges.args["Cols"].(ColsArgs)
			session = session.Cols(args.columns...)
		case "AllCols":
			session = session.AllCols()
		case "MustCols":
			args := ges.args["MustCols"].(MustColsArgs)
			session = session.MustCols(args.columns...)
		case "UseBool":
			args := ges.args["UseBool"].(UseBoolArgs)
			session = session.UseBool(args.columns...)
		case "Distinct":
			args := ges.args["Distinct"].(DistinctArgs)
			session = session.Distinct(args.columns...)
		case "Omit":
			args := ges.args["Omit"].(OmitArgs)
			session = session.Omit(args.columns...)
		case "Nullable":
			args := ges.args["Nullable"].(NullableArgs)
			session = session.Nullable(args.columns...)
		case "NoAutoTime":
			session = session.NoAutoTime()
		case "Sql":
			args := ges.args["Sql"].(SqlArgs)
			session = session.Sql(args.query, args.args...)
		case "SQL":
			args := ges.args["SQL"].(SqlArgs)
			session = session.SQL(args.query, args.args...)
		case "Where":
			args := ges.args["Where"].(WhereArgs)
			session = session.Where(args.query, args.args...)
		case "And":
			args := ges.args["And"].(AndArgs)
			session = session.And(args.query, args.args...)
		case "Or":
			args := ges.args["Or"].(OrArgs)
			session = session.Or(args.query, args.args...)
		case "Id":
			args := ges.args["Id"].(IdArgs)
			session = session.Id(args.id)
		case "ID":
			args := ges.args["ID"].(IDArgs)
			session = session.ID(args.id)
		case "In":
			args := ges.args["In"].(InArgs)
			session = session.In(args.column, args.args...)
		case "NotIn":
			args := ges.args["NotIn"].(NotInArgs)
			session = session.NotIn(args.column, args.args...)
		case "BufferSize":
			args := ges.args["BufferSize"].(BufferSizeArgs)
			session = session.BufferSize(args.size)
		}
	}
	return session
}

// Before Apply before Processor, affected bean is passed to closure arg
func (ges *GESession) Before(closures func(interface{})) *GESession {
	ges.operation = append(ges.operation, "Before")
	args := BeforeArgs{
		closures: closures,
	}
	ges.args["Before"] = args
	return ges
}

// After Apply after Processor, affected bean is passed to closure arg
func (ges *GESession) After(closures func(interface{})) *GESession {
	ges.operation = append(ges.operation, "After")
	args := AfterArgs{
		closures: closures,
	}
	ges.args["After"] = args
	return ges
}

// Table can input a string or pointer to struct for special a table to operate.
func (ges *GESession) Table(tableNameOrBean interface{}) *GESession {
	ges.operation = append(ges.operation, "Table")
	args := TableArgs{
		tableNameOrBean: tableNameOrBean,
	}
	ges.args["Table"] = args
	return ges
}

// Alias set the table alias
func (ges *GESession) Alias(alias string) *GESession {
	ges.operation = append(ges.operation, "Alias")
	args := AliasArgs{
		alias: alias,
	}
	ges.args["Alias"] = args
	return ges
}

// NoCascade indicate that no cascade load child object
func (ges *GESession) NoCascade() *GESession {
	ges.operation = append(ges.operation, "NoCascade")
	return ges
}

// ForUpdate Set Read/Write locking for UPDATE
func (ges *GESession) ForUpdate() *GESession {
	ges.operation = append(ges.operation, "ForUpdate")
	return ges
}

// NoAutoCondition disable generate SQL condition from beans
func (ges *GESession) NoAutoCondition(no ...bool) *GESession {
	ges.operation = append(ges.operation, "NoAutoCondition")
	args := NoAutoConditionArgs{
		no: no,
	}
	ges.args["NoAutoCondition"] = args
	return ges
}

// Limit provide limit and offset query condition
func (ges *GESession) Limit(limit int, start ...int) *GESession {
	ges.operation = append(ges.operation, "Limit")
	args := LimitArgs{
		limit: limit,
		start: start,
	}
	ges.args["Limit"] = args
	return ges
}

// OrderBy provide order by query condition, the input parameter is the content
// after order by on a sql statement.
func (ges *GESession) OrderBy(order string) *GESession {
	ges.operation = append(ges.operation, "OrderBy")
	args := OrderByArgs{
		order: order,
	}
	ges.args["OrderBy"] = args
	return ges
}

// Desc provide desc order by query condition, the input parameters are columns.
func (ges *GESession) Desc(colNames ...string) *GESession {
	ges.operation = append(ges.operation, "Desc")
	args := DescArgs{
		colNames: colNames,
	}
	ges.args["Desc"] = args
	return ges
}

// Asc provide asc order by query condition, the input parameters are columns.
func (ges *GESession) Asc(colNames ...string) *GESession {
	ges.operation = append(ges.operation, "Asc")
	args := AscArgs{
		colNames: colNames,
	}
	ges.args["Asc"] = args
	return ges
}

// StoreEngine is only avialble mysql dialect currently
func (ges *GESession) StoreEngine(storeEngine string) *GESession {
	ges.operation = append(ges.operation, "StoreEngine")
	args := StoreEngineArgs{
		storeEngine: storeEngine,
	}
	ges.args["StoreEngine"] = args
	return ges
}

// Charset is only avialble mysql dialect currently
func (ges *GESession) Charset(charset string) *GESession {
	ges.operation = append(ges.operation, "Charset")
	args := CharsetArgs{
		charset: charset,
	}
	ges.args["Charset"] = args
	return ges
}

// Cascade indicates if loading sub Struct
func (ges *GESession) Cascade(trueOrFalse ...bool) *GESession {
	ges.operation = append(ges.operation, "Cascade")
	args := CascadeArgs{
		trueOrFalse: trueOrFalse,
	}
	ges.args["Cascade"] = args
	return ges
}

// NoCache ask this session do not retrieve data from cache system and
// get data from database directly.
func (ges *GESession) NoCache() *GESession {
	ges.operation = append(ges.operation, "NoCache")
	return ges
}

// Join join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (ges *GESession) Join(joinOperator string, tablename interface{}, condition string, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "Join")
	joinArgs := JoinArgs{
		joinOperator: joinOperator,
		tablename:    tablename,
		condition:    condition,
		args:         args,
	}
	ges.args["Join"] = joinArgs
	return ges
}

// GroupBy Generate Group By statement
func (ges *GESession) GroupBy(keys string) *GESession {
	ges.operation = append(ges.operation, "GroupBy")
	args := GroupByArgs{
		keys: keys,
	}
	ges.args["GroupBy"] = args
	return ges
}

// Having Generate Having statement
func (ges *GESession) Having(conditions string) *GESession {
	ges.operation = append(ges.operation, "Having")
	args := HavingArgs{
		conditions: conditions,
	}
	ges.args["Having"] = args
	return ges
}

// Unscoped always disable struct tag "deleted"
func (ges *GESession) Unscoped() *GESession {
	ges.operation = append(ges.operation, "Unscoped")
	return ges
}

// Incr provides a query string like "count = count + 1"
func (ges *GESession) Incr(column string, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "Incr")
	incrArgs := IncrArgs{
		column: column,
		args:   args,
	}
	ges.args["Incr"] = incrArgs
	return ges
}

// Decr provides a query string like "count = count - 1"
func (ges *GESession) Decr(column string, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "Decr")
	decrArgs := DecrArgs{
		column: column,
		args:   args,
	}
	ges.args["Decr"] = decrArgs
	return ges
}

// SetExpr provides a query string like "column = {expression}"
func (ges *GESession) SetExpr(column string, expression string) *GESession {
	ges.operation = append(ges.operation, "SetExpr")
	args := SetExprArgs{
		column:     column,
		expression: expression,
	}
	ges.args["SetExpr"] = args
	return ges
}

// Select provides some columns to special
func (ges *GESession) Select(str string) *GESession {
	ges.operation = append(ges.operation, "Select")
	args := SelectArgs{
		str: str,
	}
	ges.args["Select"] = args
	return ges
}

// Cols provides some columns to special
func (ges *GESession) Cols(columns ...string) *GESession {
	ges.operation = append(ges.operation, "Cols")
	args := ColsArgs{
		columns: columns,
	}
	ges.args["Cols"] = args
	return ges
}

// AllCols ask all columns
func (ges *GESession) AllCols() *GESession {
	ges.operation = append(ges.operation, "AllCols")
	return ges
}

// MustCols specify some columns must use even if they are empty
func (ges *GESession) MustCols(columns ...string) *GESession {
	ges.operation = append(ges.operation, "MustCols")
	args := MustColsArgs{
		columns: columns,
	}
	ges.args["MustCols"] = args
	return ges
}

// UseBool automatically retrieve condition according struct, but
// if struct has bool field, it will ignore them. So use UseBool
// to tell system to do not ignore them.
// If no parameters, it will use all the bool field of struct, or
// it will use parameters's columns
func (ges *GESession) UseBool(columns ...string) *GESession {
	ges.operation = append(ges.operation, "UseBool")
	args := UseBoolArgs{
		columns: columns,
	}
	ges.args["UseBool"] = args
	return ges
}

// Distinct use for distinct columns. Caution: when you are using cache,
// distinct will not be cached because cache system need id,
// but distinct will not provide id
func (ges *GESession) Distinct(columns ...string) *GESession {
	ges.operation = append(ges.operation, "Distinct")
	args := DistinctArgs{
		columns: columns,
	}
	ges.args["Distinct"] = args
	return ges
}

// Omit Only not use the parameters as select or update columns
func (ges *GESession) Omit(columns ...string) *GESession {
	ges.operation = append(ges.operation, "Omit")
	args := OmitArgs{
		columns: columns,
	}
	ges.args["Omit"] = args
	return ges
}

// Nullable Set null when column is zero-value and nullable for update
func (ges *GESession) Nullable(columns ...string) *GESession {
	ges.operation = append(ges.operation, "Nullable")
	args := NullableArgs{
		columns: columns,
	}
	ges.args["Nullable"] = args
	return ges
}

// NoAutoTime means do not automatically give created field and updated field
// the current time on the current session temporarily
func (ges *GESession) NoAutoTime() *GESession {
	ges.operation = append(ges.operation, "NoAutoTime")
	return ges
}

// Sql provides raw sql input parameter. When you have a complex SQL statement
// and cannot use Where, Id, In and etc. Methods to describe, you can use SQL.
//
// Deprecated: use SQL instead.
func (ges *GESession) Sql(query interface{}, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "Sql")
	sqlArgs := SqlArgs{
		query: query,
		args:  args,
	}
	ges.args["Sql"] = sqlArgs
	return ges
}

// SQL provides raw sql input parameter. When you have a complex SQL statement
// and cannot use Where, Id, In and etc. Methods to describe, you can use SQL.
func (ges *GESession) SQL(query interface{}, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "SQL")
	sqlArgs := SqlArgs{
		query: query,
		args:  args,
	}
	ges.args["SQL"] = sqlArgs
	return ges
}

// Where provides custom query condition.
func (ges *GESession) Where(query interface{}, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "Where")
	whereArgs := WhereArgs{
		query: query,
		args:  args,
	}
	ges.args["Where"] = whereArgs
	return ges
}

type AndArgs struct {
	query interface{}
	args  []interface{}
}

// And provides custom query condition.
func (ges *GESession) And(query interface{}, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "And")
	andArgs := AndArgs{
		query: query,
		args:  args,
	}
	ges.args["And"] = andArgs
	return ges
}

type OrArgs struct {
	query interface{}
	args  []interface{}
}

// Or provides custom query condition.
func (ges *GESession) Or(query interface{}, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "Or")
	orArgs := OrArgs{
		query: query,
		args:  args,
	}
	ges.args["Or"] = orArgs
	return ges
}

// Id provides converting id as a query condition
//
// Deprecated: use ID instead
func (ges *GESession) Id(id interface{}) *GESession {
	ges.operation = append(ges.operation, "Id")
	args := IdArgs{
		id: id,
	}
	ges.args["Id"] = args
	return ges
}

// ID provides converting id as a query condition
func (ges *GESession) ID(id interface{}) *GESession {
	ges.operation = append(ges.operation, "ID")
	args := IDArgs{
		id: id,
	}
	ges.args["ID"] = args
	return ges
}

// In provides a query string like "id in (1, 2, 3)"
func (ges *GESession) In(column string, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "In")
	inArgs := InArgs{
		column: column,
		args:   args,
	}
	ges.args["In"] = inArgs
	return ges
}

// NotIn provides a query string like "id in (1, 2, 3)"
func (ges *GESession) NotIn(column string, args ...interface{}) *GESession {
	ges.operation = append(ges.operation, "NotIn")
	notInArgs := NotInArgs{
		column: column,
		args:   args,
	}
	ges.args["NotIn"] = notInArgs
	return ges
}

//TODO 还需要分析如何实现
// Conds returns session query conditions except auto bean conditions
func (ges *GESession) Conds() builder.Cond {
	return ges.ge.Master().NewSession().Conds()
}

//TODO 缺少前置session操作链
// Delete records, bean's non-empty fields are conditions
func (ges *GESession) Delete(bean interface{}) (int64, error) {
	session := ges.ge.Master().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Delete(bean)
}

//TODO 缺少前置session操作链
// Exist returns true if the record exist otherwise return false
func (ges *GESession) Exist(bean ...interface{}) (bool, error) {
	session := ges.ge.Master().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Exist(bean...)
}

//TODO 缺少前置session操作链
// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct
func (ges *GESession) Find(rowsSlicePtr interface{}, condiBean ...interface{}) error {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Find(rowsSlicePtr, condiBean...)
}

//TODO 缺少前置session操作链
// Get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (ges *GESession) Get(bean interface{}) (bool, error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Get(bean)
}

//TODO 缺少前置session操作链
// Insert insert one or more beans
func (ges *GESession) Insert(beans ...interface{}) (int64, error) {
	session := ges.ge.Master().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Insert(beans...)
}

// Rows return sql.Rows compatible Rows obj, as a forward Iterator object for iterating record by record, bean's non-empty fields
// are conditions.
func (ges *GESession) Rows(bean interface{}) (*Rows, error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Rows(bean)
}

// Iterate record by record handle records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct
func (ges *GESession) Iterate(bean interface{}, fun IterFunc) error {
	return ges.ge.Slave().Iterate(bean, fun)
}

// BufferSize sets the buffersize for iterate
func (ges *GESession) BufferSize(size int) *GESession {
	ges.operation = append(ges.operation, "BufferSize")
	args := BufferSizeArgs{
		size: size,
	}
	ges.args["BufferSize"] = args
	return ges
}

//TODO 缺少前置session操作链
// Query runs a raw sql and return records as []map[string][]byte
func (ges *GESession) Query(sqlStr string, args ...interface{}) ([]map[string][]byte, error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Query(sqlStr, args...)
}

//TODO 缺少前置session操作链
// QueryString runs a raw sql and return records as []map[string]string
func (ges *GESession) QueryString(sqlStr string, args ...interface{}) ([]map[string]string, error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.QueryString(sqlStr, args...)
}

// QueryInterface runs a raw sql and return records as []map[string]interface{}
func (ges *GESession) QueryInterface(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.QueryInterface(sqlStr, args...)
}

// Exec raw sql
func (ges *GESession) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	session := ges.ge.Master().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Exec(sqlStr, args...)
}

// CreateTable create a table according a bean
func (ges *GESession) CreateTable(bean interface{}) error {
	session := ges.ge.Master().NewSession()
	defer session.Close()
	return session.CreateTable(bean)
}

// CreateIndexes create indexes
func (ges *GESession) CreateIndexes(bean interface{}) error {
	return ges.ge.Master().CreateIndexes(bean)
}

// CreateUniques create uniques
func (ges *GESession) CreateUniques(bean interface{}) error {
	return ges.ge.Master().CreateUniques(bean)
}

// DropIndexes drop indexes
func (ges *GESession) DropIndexes(bean interface{}) error {
	return ges.ge.Master().DropIndexes(bean)
}

// DropTable drop table will drop table if exist, if drop failed, it will return error
func (ges *GESession) DropTable(beanOrTableName interface{}) error {
	session := ges.ge.Master().NewSession()
	defer session.Close()
	return session.DropTable(beanOrTableName)
}

// IsTableExist if a table is exist
func (ges *GESession) IsTableExist(beanOrTableName interface{}) (bool, error) {
	return ges.ge.Master().IsTableExist(beanOrTableName)
}

// IsTableEmpty if table have any records
func (ges *GESession) IsTableEmpty(bean interface{}) (bool, error) {
	return ges.ge.Master().IsTableEmpty(bean)
}

// Sync2 synchronize structs to database tables
func (ges *GESession) Sync2(beans ...interface{}) error {
	return ges.ge.Master().Sync2(beans...)
}

//TODO 缺少前置session操作链
// Count counts the records. bean's non-empty fields
// are conditions.
func (ges *GESession) Count(bean ...interface{}) (int64, error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Count(bean...)
}

//TODO 缺少前置session操作链
// sum call sum some column. bean's non-empty fields are conditions.
func (ges *GESession) Sum(bean interface{}, columnName string) (res float64, err error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Sum(bean, columnName)
}

//TODO 缺少前置session操作链
// SumInt call sum some column. bean's non-empty fields are conditions.
func (ges *GESession) SumInt(bean interface{}, columnName string) (res int64, err error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.SumInt(bean, columnName)
}

//TODO 缺少前置session操作链
// Sums call sum some columns. bean's non-empty fields are conditions.
func (ges *GESession) Sums(bean interface{}, columnNames ...string) ([]float64, error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Sums(bean, columnNames...)
}

//TODO 缺少前置session操作链
// SumsInt sum specify columns and return as []int64 instead of []float64
func (ges *GESession) SumsInt(bean interface{}, columnNames ...string) ([]int64, error) {
	session := ges.ge.Slave().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.SumsInt(bean, columnNames...)
}

// Update records, bean's non-empty fields are updated contents,
// condiBean' non-empty filds are conditions
// CAUTION:
//        1.bool will defaultly be updated content nor conditions
//         You should call UseBool if you have bool to use.
//        2.float32 & float64 may be not inexact as conditions
func (ges *GESession) Update(bean interface{}, condiBean ...interface{}) (int64, error) {
	session := ges.ge.Master().NewSession()
	defer session.Close()
	session = ges.operate(session)
	return session.Update(bean, condiBean...)

}
