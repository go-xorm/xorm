package xorm

import (
	"database/sql"

	"github.com/go-xorm/builder"
)

type EGSession struct {
	eg        *EngineGroup
	operation []string
	args      map[string]interface{}
	err       error
}

func (egs *EGSession) operate(session *Session) *Session {
	for _, v := range egs.operation {
		switch v {
		case "Before":
			args := egs.args["Before"].(BeforeArgs)
			session = session.Before(args.closures)
		case "After":
			args := egs.args["After"].(AfterArgs)
			session = session.After(args.closures)
		case "Table":
			args := egs.args["Table"].(TableArgs)
			session = session.Table(args.tableNameOrBean)
		case "Alias":
			args := egs.args["Alias"].(AliasArgs)
			session = session.Alias(args.alias)
		case "NoCascade":
			session = session.NoCascade()
		case "ForUpdate":
			session = session.ForUpdate()
		case "NoAutoCondition":
			args := egs.args["NoAutoCondition"].(NoAutoConditionArgs)
			session = session.NoAutoCondition(args.no...)
		case "Limit":
			args := egs.args["Limit"].(LimitArgs)
			session = session.Limit(args.limit, args.start...)
		case "OrderBy":
			args := egs.args["OrderBy"].(OrderByArgs)
			session = session.OrderBy(args.order)
		case "Desc":
			args := egs.args["Desc"].(DescArgs)
			session = session.Desc(args.colNames...)
		case "Asc":
			args := egs.args["Asc"].(AscArgs)
			session = session.Asc(args.colNames...)
		case "StoreEngine":
			args := egs.args["StoreEngine"].(StoreEngineArgs)
			session = session.StoreEngine(args.storeEngine)
		case "Charset":
			args := egs.args["Charset"].(CharsetArgs)
			session = session.Charset(args.charset)
		case "Cascade":
			args := egs.args["Cascade"].(CascadeArgs)
			session = session.Cascade(args.trueOrFalse...)
		case "NoCache":
			session = session.NoCache()
		case "Join":
			args := egs.args["Join"].(JoinArgs)
			session = session.Join(args.joinOperator, args.tablename, args.condition, args.args...)
		case "GroupBy":
			args := egs.args["GroupBy"].(GroupByArgs)
			session = session.GroupBy(args.keys)
		case "Having":
			args := egs.args["Having"].(HavingArgs)
			session = session.Having(args.conditions)
		case "Unscoped":
			session = session.Unscoped()
		case "Incr":
			args := egs.args["Incr"].(IncrArgs)
			session = session.Incr(args.column, args.args...)
		case "Decr":
			args := egs.args["Decr"].(DecrArgs)
			session = session.Decr(args.column, args.args...)
		case "SetExpr":
			args := egs.args["SetExpr"].(SetExprArgs)
			session = session.SetExpr(args.column, args.expression)
		case "Select":
			args := egs.args["Select"].(SelectArgs)
			session = session.Select(args.str)
		case "Cols":
			args := egs.args["Cols"].(ColsArgs)
			session = session.Cols(args.columns...)
		case "AllCols":
			session = session.AllCols()
		case "MustCols":
			args := egs.args["MustCols"].(MustColsArgs)
			session = session.MustCols(args.columns...)
		case "UseBool":
			args := egs.args["UseBool"].(UseBoolArgs)
			session = session.UseBool(args.columns...)
		case "Distinct":
			args := egs.args["Distinct"].(DistinctArgs)
			session = session.Distinct(args.columns...)
		case "Omit":
			args := egs.args["Omit"].(OmitArgs)
			session = session.Omit(args.columns...)
		case "Nullable":
			args := egs.args["Nullable"].(NullableArgs)
			session = session.Nullable(args.columns...)
		case "NoAutoTime":
			session = session.NoAutoTime()
		case "Sql":
			args := egs.args["Sql"].(SqlArgs)
			session = session.Sql(args.query, args.args...)
		case "SQL":
			args := egs.args["SQL"].(SqlArgs)
			session = session.SQL(args.query, args.args...)
		case "Where":
			args := egs.args["Where"].(WhereArgs)
			session = session.Where(args.query, args.args...)
		case "And":
			args := egs.args["And"].(AndArgs)
			session = session.And(args.query, args.args...)
		case "Or":
			args := egs.args["Or"].(OrArgs)
			session = session.Or(args.query, args.args...)
		case "Id":
			args := egs.args["Id"].(IdArgs)
			session = session.Id(args.id)
		case "ID":
			args := egs.args["ID"].(IDArgs)
			session = session.ID(args.id)
		case "In":
			args := egs.args["In"].(InArgs)
			session = session.In(args.column, args.args...)
		case "NotIn":
			args := egs.args["NotIn"].(NotInArgs)
			session = session.NotIn(args.column, args.args...)
		case "BufferSize":
			args := egs.args["BufferSize"].(BufferSizeArgs)
			session = session.BufferSize(args.size)
		}
	}
	return session
}

// Before Apply before Processor, affected bean is passed to closure arg
func (egs *EGSession) Before(closures func(interface{})) *EGSession {
	egs.operation = append(egs.operation, "Before")
	args := BeforeArgs{
		closures: closures,
	}
	egs.args["Before"] = args
	return egs
}

// After Apply after Processor, affected bean is passed to closure arg
func (egs *EGSession) After(closures func(interface{})) *EGSession {
	egs.operation = append(egs.operation, "After")
	args := AfterArgs{
		closures: closures,
	}
	egs.args["After"] = args
	return egs
}

// Table can input a string or pointer to struct for special a table to operate.
func (egs *EGSession) Table(tableNameOrBean interface{}) *EGSession {
	egs.operation = append(egs.operation, "Table")
	args := TableArgs{
		tableNameOrBean: tableNameOrBean,
	}
	egs.args["Table"] = args
	return egs
}

// Alias set the table alias
func (egs *EGSession) Alias(alias string) *EGSession {
	egs.operation = append(egs.operation, "Alias")
	args := AliasArgs{
		alias: alias,
	}
	egs.args["Alias"] = args
	return egs
}

// NoCascade indicate that no cascade load child object
func (egs *EGSession) NoCascade() *EGSession {
	egs.operation = append(egs.operation, "NoCascade")
	return egs
}

// ForUpdate Set Read/Write locking for UPDATE
func (egs *EGSession) ForUpdate() *EGSession {
	egs.operation = append(egs.operation, "ForUpdate")
	return egs
}

// NoAutoCondition disable generate SQL condition from beans
func (egs *EGSession) NoAutoCondition(no ...bool) *EGSession {
	egs.operation = append(egs.operation, "NoAutoCondition")
	args := NoAutoConditionArgs{
		no: no,
	}
	egs.args["NoAutoCondition"] = args
	return egs
}

// Limit provide limit and offset query condition
func (egs *EGSession) Limit(limit int, start ...int) *EGSession {
	egs.operation = append(egs.operation, "Limit")
	args := LimitArgs{
		limit: limit,
		start: start,
	}
	egs.args["Limit"] = args
	return egs
}

// OrderBy provide order by query condition, the input parameter is the content
// after order by on a sql statement.
func (egs *EGSession) OrderBy(order string) *EGSession {
	egs.operation = append(egs.operation, "OrderBy")
	args := OrderByArgs{
		order: order,
	}
	egs.args["OrderBy"] = args
	return egs
}

// Desc provide desc order by query condition, the input parameters are columns.
func (egs *EGSession) Desc(colNames ...string) *EGSession {
	egs.operation = append(egs.operation, "Desc")
	args := DescArgs{
		colNames: colNames,
	}
	egs.args["Desc"] = args
	return egs
}

// Asc provide asc order by query condition, the input parameters are columns.
func (egs *EGSession) Asc(colNames ...string) *EGSession {
	egs.operation = append(egs.operation, "Asc")
	args := AscArgs{
		colNames: colNames,
	}
	egs.args["Asc"] = args
	return egs
}

// StoreEngine is only avialble mysql dialect currently
func (egs *EGSession) StoreEngine(storeEngine string) *EGSession {
	egs.operation = append(egs.operation, "StoreEngine")
	args := StoreEngineArgs{
		storeEngine: storeEngine,
	}
	egs.args["StoreEngine"] = args
	return egs
}

// Charset is only avialble mysql dialect currently
func (egs *EGSession) Charset(charset string) *EGSession {
	egs.operation = append(egs.operation, "Charset")
	args := CharsetArgs{
		charset: charset,
	}
	egs.args["Charset"] = args
	return egs
}

// Cascade indicates if loading sub Struct
func (egs *EGSession) Cascade(trueOrFalse ...bool) *EGSession {
	egs.operation = append(egs.operation, "Cascade")
	args := CascadeArgs{
		trueOrFalse: trueOrFalse,
	}
	egs.args["Cascade"] = args
	return egs
}

// NoCache ask this session do not retrieve data from cache system and
// get data from database directly.
func (egs *EGSession) NoCache() *EGSession {
	egs.operation = append(egs.operation, "NoCache")
	return egs
}

// Join join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (egs *EGSession) Join(joinOperator string, tablename interface{}, condition string, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "Join")
	joinArgs := JoinArgs{
		joinOperator: joinOperator,
		tablename:    tablename,
		condition:    condition,
		args:         args,
	}
	egs.args["Join"] = joinArgs
	return egs
}

// GroupBy Generate Group By statement
func (egs *EGSession) GroupBy(keys string) *EGSession {
	egs.operation = append(egs.operation, "GroupBy")
	args := GroupByArgs{
		keys: keys,
	}
	egs.args["GroupBy"] = args
	return egs
}

// Having Generate Having statement
func (egs *EGSession) Having(conditions string) *EGSession {
	egs.operation = append(egs.operation, "Having")
	args := HavingArgs{
		conditions: conditions,
	}
	egs.args["Having"] = args
	return egs
}

// Unscoped always disable struct tag "deleted"
func (egs *EGSession) Unscoped() *EGSession {
	egs.operation = append(egs.operation, "Unscoped")
	return egs
}

// Incr provides a query string like "count = count + 1"
func (egs *EGSession) Incr(column string, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "Incr")
	incrArgs := IncrArgs{
		column: column,
		args:   args,
	}
	egs.args["Incr"] = incrArgs
	return egs
}

// Decr provides a query string like "count = count - 1"
func (egs *EGSession) Decr(column string, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "Decr")
	decrArgs := DecrArgs{
		column: column,
		args:   args,
	}
	egs.args["Decr"] = decrArgs
	return egs
}

// SetExpr provides a query string like "column = {expression}"
func (egs *EGSession) SetExpr(column string, expression string) *EGSession {
	egs.operation = append(egs.operation, "SetExpr")
	args := SetExprArgs{
		column:     column,
		expression: expression,
	}
	egs.args["SetExpr"] = args
	return egs
}

// Select provides some columns to special
func (egs *EGSession) Select(str string) *EGSession {
	egs.operation = append(egs.operation, "Select")
	args := SelectArgs{
		str: str,
	}
	egs.args["Select"] = args
	return egs
}

// Cols provides some columns to special
func (egs *EGSession) Cols(columns ...string) *EGSession {
	egs.operation = append(egs.operation, "Cols")
	args := ColsArgs{
		columns: columns,
	}
	egs.args["Cols"] = args
	return egs
}

// AllCols ask all columns
func (egs *EGSession) AllCols() *EGSession {
	egs.operation = append(egs.operation, "AllCols")
	return egs
}

// MustCols specify some columns must use even if they are empty
func (egs *EGSession) MustCols(columns ...string) *EGSession {
	egs.operation = append(egs.operation, "MustCols")
	args := MustColsArgs{
		columns: columns,
	}
	egs.args["MustCols"] = args
	return egs
}

// UseBool automatically retrieve condition according struct, but
// if struct has bool field, it will ignore them. So use UseBool
// to tell system to do not ignore them.
// If no parameters, it will use all the bool field of struct, or
// it will use parameters's columns
func (egs *EGSession) UseBool(columns ...string) *EGSession {
	egs.operation = append(egs.operation, "UseBool")
	args := UseBoolArgs{
		columns: columns,
	}
	egs.args["UseBool"] = args
	return egs
}

// Distinct use for distinct columns. Caution: when you are using cache,
// distinct will not be cached because cache system need id,
// but distinct will not provide id
func (egs *EGSession) Distinct(columns ...string) *EGSession {
	egs.operation = append(egs.operation, "Distinct")
	args := DistinctArgs{
		columns: columns,
	}
	egs.args["Distinct"] = args
	return egs
}

// Omit Only not use the parameters as select or update columns
func (egs *EGSession) Omit(columns ...string) *EGSession {
	egs.operation = append(egs.operation, "Omit")
	args := OmitArgs{
		columns: columns,
	}
	egs.args["Omit"] = args
	return egs
}

// Nullable Set null when column is zero-value and nullable for update
func (egs *EGSession) Nullable(columns ...string) *EGSession {
	egs.operation = append(egs.operation, "Nullable")
	args := NullableArgs{
		columns: columns,
	}
	egs.args["Nullable"] = args
	return egs
}

// NoAutoTime means do not automatically give created field and updated field
// the current time on the current session temporarily
func (egs *EGSession) NoAutoTime() *EGSession {
	egs.operation = append(egs.operation, "NoAutoTime")
	return egs
}

// Sql provides raw sql input parameter. When you have a complex SQL statement
// and cannot use Where, Id, In and etc. Methods to describe, you can use SQL.
//
// Deprecated: use SQL instead.
func (egs *EGSession) Sql(query string, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "Sql")
	sqlArgs := SqlArgs{
		query: query,
		args:  args,
	}
	egs.args["Sql"] = sqlArgs
	return egs
}

type SQLArgs struct {
	query interface{}
	args  []interface{}
}

// SQL provides raw sql input parameter. When you have a complex SQL statement
// and cannot use Where, Id, In and etc. Methods to describe, you can use SQL.
func (egs *EGSession) SQL(query interface{}, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "SQL")
	sqlArgs := SQLArgs{
		query: query,
		args:  args,
	}
	egs.args["SQL"] = sqlArgs
	return egs
}

// Where provides custom query condition.
func (egs *EGSession) Where(query interface{}, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "Where")
	whereArgs := WhereArgs{
		query: query,
		args:  args,
	}
	egs.args["Where"] = whereArgs
	return egs
}

type AndArgs struct {
	query interface{}
	args  []interface{}
}

// And provides custom query condition.
func (egs *EGSession) And(query interface{}, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "And")
	andArgs := AndArgs{
		query: query,
		args:  args,
	}
	egs.args["And"] = andArgs
	return egs
}

type OrArgs struct {
	query interface{}
	args  []interface{}
}

// Or provides custom query condition.
func (egs *EGSession) Or(query interface{}, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "Or")
	orArgs := OrArgs{
		query: query,
		args:  args,
	}
	egs.args["Or"] = orArgs
	return egs
}

// Id provides converting id as a query condition
//
// Deprecated: use ID instead
func (egs *EGSession) Id(id interface{}) *EGSession {
	egs.operation = append(egs.operation, "Id")
	args := IdArgs{
		id: id,
	}
	egs.args["Id"] = args
	return egs
}

// ID provides converting id as a query condition
func (egs *EGSession) ID(id interface{}) *EGSession {
	egs.operation = append(egs.operation, "ID")
	args := IDArgs{
		id: id,
	}
	egs.args["ID"] = args
	return egs
}

// In provides a query string like "id in (1, 2, 3)"
func (egs *EGSession) In(column string, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "In")
	inArgs := InArgs{
		column: column,
		args:   args,
	}
	egs.args["In"] = inArgs
	return egs
}

// NotIn provides a query string like "id in (1, 2, 3)"
func (egs *EGSession) NotIn(column string, args ...interface{}) *EGSession {
	egs.operation = append(egs.operation, "NotIn")
	notInArgs := NotInArgs{
		column: column,
		args:   args,
	}
	egs.args["NotIn"] = notInArgs
	return egs
}

// Conds returns session query conditions except auto bean conditions
func (egs *EGSession) Conds() builder.Cond {
	return egs.eg.Master().NewSession().Conds()
}

// Delete records, bean's non-empty fields are conditions
func (egs *EGSession) Delete(bean interface{}) (int64, error) {
	session := egs.eg.Master().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Delete(bean)
}

// Exist returns true if the record exist otherwise return false
func (egs *EGSession) Exist(bean ...interface{}) (bool, error) {
	session := egs.eg.Master().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Exist(bean...)
}

// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct
func (egs *EGSession) Find(rowsSlicePtr interface{}, condiBean ...interface{}) error {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Find(rowsSlicePtr, condiBean...)
}

// Get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (egs *EGSession) Get(bean interface{}) (bool, error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Get(bean)
}

// Insert insert one or more beans
func (egs *EGSession) Insert(beans ...interface{}) (int64, error) {
	session := egs.eg.Master().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Insert(beans...)
}

// Rows return sql.Rows compatible Rows obj, as a forward Iterator object for iterating record by record, bean's non-empty fields
// are conditions.
func (egs *EGSession) Rows(bean interface{}) (*Rows, error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Rows(bean)
}

// Iterate record by record handle records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct
func (egs *EGSession) Iterate(bean interface{}, fun IterFunc) error {
	return egs.eg.Slave().Iterate(bean, fun)
}

// BufferSize sets the buffersize for iterate
func (egs *EGSession) BufferSize(size int) *EGSession {
	egs.operation = append(egs.operation, "BufferSize")
	args := BufferSizeArgs{
		size: size,
	}
	egs.args["BufferSize"] = args
	return egs
}

// Query runs a raw sql and return records as []map[string][]byte
func (egs *EGSession) Query(sqlStr string, args ...interface{}) ([]map[string][]byte, error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Query(sqlStr, args...)
}

// QueryString runs a raw sql and return records as []map[string]string
func (egs *EGSession) QueryString(sqlStr string, args ...interface{}) ([]map[string]string, error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.QueryString(sqlStr, args...)
}

// QueryInterface runs a raw sql and return records as []map[string]interface{}
func (egs *EGSession) QueryInterface(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.QueryInterface(sqlStr, args...)
}

// Exec raw sql
func (egs *EGSession) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	session := egs.eg.Master().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Exec(sqlStr, args...)
}

// CreateTable create a table according a bean
func (egs *EGSession) CreateTable(bean interface{}) error {
	session := egs.eg.Master().NewSession()
	defer session.Close()
	return session.CreateTable(bean)
}

// CreateIndexes create indexes
func (egs *EGSession) CreateIndexes(bean interface{}) error {
	return egs.eg.Master().CreateIndexes(bean)
}

// CreateUniques create uniques
func (egs *EGSession) CreateUniques(bean interface{}) error {
	return egs.eg.Master().CreateUniques(bean)
}

// DropIndexes drop indexes
func (egs *EGSession) DropIndexes(bean interface{}) error {
	return egs.eg.Master().DropIndexes(bean)
}

// DropTable drop table will drop table if exist, if drop failed, it will return error
func (egs *EGSession) DropTable(beanOrTableName interface{}) error {
	session := egs.eg.Master().NewSession()
	defer session.Close()
	return session.DropTable(beanOrTableName)
}

// IsTableExist if a table is exist
func (egs *EGSession) IsTableExist(beanOrTableName interface{}) (bool, error) {
	return egs.eg.Master().IsTableExist(beanOrTableName)
}

// IsTableEmpty if table have any records
func (egs *EGSession) IsTableEmpty(bean interface{}) (bool, error) {
	return egs.eg.Master().IsTableEmpty(bean)
}

// Sync2 synchronize structs to database tables
func (egs *EGSession) Sync2(beans ...interface{}) error {
	return egs.eg.Master().Sync2(beans...)
}

// Count counts the records. bean's non-empty fields
// are conditions.
func (egs *EGSession) Count(bean ...interface{}) (int64, error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Count(bean...)
}

// sum call sum some column. bean's non-empty fields are conditions.
func (egs *EGSession) Sum(bean interface{}, columnName string) (res float64, err error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Sum(bean, columnName)
}

// SumInt call sum some column. bean's non-empty fields are conditions.
func (egs *EGSession) SumInt(bean interface{}, columnName string) (res int64, err error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.SumInt(bean, columnName)
}

// Sums call sum some columns. bean's non-empty fields are conditions.
func (egs *EGSession) Sums(bean interface{}, columnNames ...string) ([]float64, error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Sums(bean, columnNames...)
}

// SumsInt sum specify columns and return as []int64 instead of []float64
func (egs *EGSession) SumsInt(bean interface{}, columnNames ...string) ([]int64, error) {
	session := egs.eg.Slave().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.SumsInt(bean, columnNames...)
}

// Update records, bean's non-empty fields are updated contents,
// condiBean' non-empty filds are conditions
// CAUTION:
//        1.bool will defaultly be updated content nor conditions
//         You should call UseBool if you have bool to use.
//        2.float32 & float64 may be not inexact as conditions
func (egs *EGSession) Update(bean interface{}, condiBean ...interface{}) (int64, error) {
	session := egs.eg.Master().NewSession()
	defer session.Close()
	session = egs.operate(session)
	return session.Update(bean, condiBean...)

}
