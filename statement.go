// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"

	"xorm.io/builder"
	"xorm.io/core"
)

// Statement save all the sql info for executing SQL
type Statement struct {
	RefTable        *core.Table
	Engine          *Engine
	Start           int
	LimitN          int
	idParam         *core.PK
	OrderStr        string
	JoinStr         string
	joinArgs        []interface{}
	GroupByStr      string
	HavingStr       string
	ColumnStr       string
	selectStr       string
	useAllCols      bool
	OmitStr         string
	AltTableName    string
	tableName       string
	RawSQL          string
	RawParams       []interface{}
	UseCascade      bool
	UseAutoJoin     bool
	StoreEngine     string
	Charset         string
	UseCache        bool
	UseAutoTime     bool
	noAutoCondition bool
	IsDistinct      bool
	IsForUpdate     bool
	TableAlias      string
	allUseBool      bool
	checkVersion    bool
	unscoped        bool
	columnMap       columnMap
	omitColumnMap   columnMap
	mustColumnMap   map[string]bool
	nullableMap     map[string]bool
	incrColumns     map[string]incrParam
	decrColumns     map[string]decrParam
	exprColumns     map[string]exprParam
	cond            builder.Cond
	bufferSize      int
	context         ContextCache
	lastError       error
}

// Init reset all the statement's fields
func (statement *Statement) Init() {
	statement.RefTable = nil
	statement.Start = 0
	statement.LimitN = 0
	statement.OrderStr = ""
	statement.UseCascade = true
	statement.JoinStr = ""
	statement.joinArgs = make([]interface{}, 0)
	statement.GroupByStr = ""
	statement.HavingStr = ""
	statement.ColumnStr = ""
	statement.OmitStr = ""
	statement.columnMap = columnMap{}
	statement.omitColumnMap = columnMap{}
	statement.AltTableName = ""
	statement.tableName = ""
	statement.idParam = nil
	statement.RawSQL = ""
	statement.RawParams = make([]interface{}, 0)
	statement.UseCache = true
	statement.UseAutoTime = true
	statement.noAutoCondition = false
	statement.IsDistinct = false
	statement.IsForUpdate = false
	statement.TableAlias = ""
	statement.selectStr = ""
	statement.allUseBool = false
	statement.useAllCols = false
	statement.mustColumnMap = make(map[string]bool)
	statement.nullableMap = make(map[string]bool)
	statement.checkVersion = true
	statement.unscoped = false
	statement.incrColumns = make(map[string]incrParam)
	statement.decrColumns = make(map[string]decrParam)
	statement.exprColumns = make(map[string]exprParam)
	statement.cond = builder.NewCond()
	statement.bufferSize = 0
	statement.context = nil
	statement.lastError = nil
}

// NoAutoCondition if you do not want convert bean's field as query condition, then use this function
func (statement *Statement) NoAutoCondition(no ...bool) *Statement {
	statement.noAutoCondition = true
	if len(no) > 0 {
		statement.noAutoCondition = no[0]
	}
	return statement
}

// Alias set the table alias
func (statement *Statement) Alias(alias string) *Statement {
	statement.TableAlias = alias
	return statement
}

// SQL adds raw sql statement
func (statement *Statement) SQL(query interface{}, args ...interface{}) *Statement {
	switch query.(type) {
	case (*builder.Builder):
		var err error
		statement.RawSQL, statement.RawParams, err = query.(*builder.Builder).ToSQL()
		if err != nil {
			statement.lastError = err
		}
	case string:
		statement.RawSQL = query.(string)
		statement.RawParams = args
	default:
		statement.lastError = ErrUnSupportedSQLType
	}

	return statement
}

// Where add Where statement
func (statement *Statement) Where(query interface{}, args ...interface{}) *Statement {
	return statement.And(query, args...)
}

// And add Where & and statement
func (statement *Statement) And(query interface{}, args ...interface{}) *Statement {
	switch query.(type) {
	case string:
		cond := builder.Expr(query.(string), args...)
		statement.cond = statement.cond.And(cond)
	case map[string]interface{}:
		cond := builder.Eq(query.(map[string]interface{}))
		statement.cond = statement.cond.And(cond)
	case builder.Cond:
		cond := query.(builder.Cond)
		statement.cond = statement.cond.And(cond)
		for _, v := range args {
			if vv, ok := v.(builder.Cond); ok {
				statement.cond = statement.cond.And(vv)
			}
		}
	default:
		statement.lastError = ErrConditionType
	}

	return statement
}

// Or add Where & Or statement
func (statement *Statement) Or(query interface{}, args ...interface{}) *Statement {
	switch query.(type) {
	case string:
		cond := builder.Expr(query.(string), args...)
		statement.cond = statement.cond.Or(cond)
	case map[string]interface{}:
		cond := builder.Eq(query.(map[string]interface{}))
		statement.cond = statement.cond.Or(cond)
	case builder.Cond:
		cond := query.(builder.Cond)
		statement.cond = statement.cond.Or(cond)
		for _, v := range args {
			if vv, ok := v.(builder.Cond); ok {
				statement.cond = statement.cond.Or(vv)
			}
		}
	default:
		// TODO: not support condition type
	}
	return statement
}

// In generate "Where column IN (?) " statement
func (statement *Statement) In(column string, args ...interface{}) *Statement {
	in := builder.In(statement.Engine.Quote(column), args...)
	statement.cond = statement.cond.And(in)
	return statement
}

// NotIn generate "Where column NOT IN (?) " statement
func (statement *Statement) NotIn(column string, args ...interface{}) *Statement {
	notIn := builder.NotIn(statement.Engine.Quote(column), args...)
	statement.cond = statement.cond.And(notIn)
	return statement
}

func (statement *Statement) setRefValue(v reflect.Value) error {
	var err error
	statement.RefTable, err = statement.Engine.autoMapType(reflect.Indirect(v))
	if err != nil {
		return err
	}
	statement.tableName = statement.Engine.TableName(v, true)
	return nil
}

func (statement *Statement) setRefBean(bean interface{}) error {
	var err error
	statement.RefTable, err = statement.Engine.autoMapType(rValue(bean))
	if err != nil {
		return err
	}
	statement.tableName = statement.Engine.TableName(bean, true)
	return nil
}

// Auto generating update columnes and values according a struct
func (statement *Statement) buildUpdates(bean interface{},
	includeVersion, includeUpdated, includeNil,
	includeAutoIncr, update bool) ([]string, []interface{}) {
	engine := statement.Engine
	table := statement.RefTable
	allUseBool := statement.allUseBool
	useAllCols := statement.useAllCols
	mustColumnMap := statement.mustColumnMap
	nullableMap := statement.nullableMap
	columnMap := statement.columnMap
	omitColumnMap := statement.omitColumnMap
	unscoped := statement.unscoped

	var colNames = make([]string, 0)
	var args = make([]interface{}, 0)
	for _, col := range table.Columns() {
		if !includeVersion && col.IsVersion {
			continue
		}
		if col.IsCreated {
			continue
		}
		if !includeUpdated && col.IsUpdated {
			continue
		}
		if !includeAutoIncr && col.IsAutoIncrement {
			continue
		}
		if col.IsDeleted && !unscoped {
			continue
		}
		if omitColumnMap.contain(col.Name) {
			continue
		}
		if len(columnMap) > 0 && !columnMap.contain(col.Name) {
			continue
		}

		if col.MapType == core.ONLYFROMDB {
			continue
		}

		fieldValuePtr, err := col.ValueOf(bean)
		if err != nil {
			engine.logger.Error(err)
			continue
		}

		fieldValue := *fieldValuePtr
		fieldType := reflect.TypeOf(fieldValue.Interface())
		if fieldType == nil {
			continue
		}

		requiredField := useAllCols
		includeNil := useAllCols

		if b, ok := getFlagForColumn(mustColumnMap, col); ok {
			if b {
				requiredField = true
			} else {
				continue
			}
		}

		// !evalphobia! set fieldValue as nil when column is nullable and zero-value
		if b, ok := getFlagForColumn(nullableMap, col); ok {
			if b && col.Nullable && isZero(fieldValue.Interface()) {
				var nilValue *int
				fieldValue = reflect.ValueOf(nilValue)
				fieldType = reflect.TypeOf(fieldValue.Interface())
				includeNil = true
			}
		}

		var val interface{}

		if fieldValue.CanAddr() {
			if structConvert, ok := fieldValue.Addr().Interface().(core.Conversion); ok {
				data, err := structConvert.ToDB()
				if err != nil {
					engine.logger.Error(err)
				} else {
					val = data
				}
				goto APPEND
			}
		}

		if structConvert, ok := fieldValue.Interface().(core.Conversion); ok {
			data, err := structConvert.ToDB()
			if err != nil {
				engine.logger.Error(err)
			} else {
				val = data
			}
			goto APPEND
		}

		if fieldType.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				if includeNil {
					args = append(args, nil)
					colNames = append(colNames, fmt.Sprintf("%v=?", engine.Quote(col.Name)))
				}
				continue
			} else if !fieldValue.IsValid() {
				continue
			} else {
				// dereference ptr type to instance type
				fieldValue = fieldValue.Elem()
				fieldType = reflect.TypeOf(fieldValue.Interface())
				requiredField = true
			}
		}

		switch fieldType.Kind() {
		case reflect.Bool:
			if allUseBool || requiredField {
				val = fieldValue.Interface()
			} else {
				// if a bool in a struct, it will not be as a condition because it default is false,
				// please use Where() instead
				continue
			}
		case reflect.String:
			if !requiredField && fieldValue.String() == "" {
				continue
			}
			// for MyString, should convert to string or panic
			if fieldType.String() != reflect.String.String() {
				val = fieldValue.String()
			} else {
				val = fieldValue.Interface()
			}
		case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
			if !requiredField && fieldValue.Int() == 0 {
				continue
			}
			val = fieldValue.Interface()
		case reflect.Float32, reflect.Float64:
			if !requiredField && fieldValue.Float() == 0.0 {
				continue
			}
			val = fieldValue.Interface()
		case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
			if !requiredField && fieldValue.Uint() == 0 {
				continue
			}
			t := int64(fieldValue.Uint())
			val = reflect.ValueOf(&t).Interface()
		case reflect.Struct:
			if fieldType.ConvertibleTo(core.TimeType) {
				t := fieldValue.Convert(core.TimeType).Interface().(time.Time)
				if !requiredField && (t.IsZero() || !fieldValue.IsValid()) {
					continue
				}
				val = engine.formatColTime(col, t)
			} else if nulType, ok := fieldValue.Interface().(driver.Valuer); ok {
				val, _ = nulType.Value()
			} else {
				if !col.SQLType.IsJson() {
					engine.autoMapType(fieldValue)
					if table, ok := engine.Tables[fieldValue.Type()]; ok {
						if len(table.PrimaryKeys) == 1 {
							pkField := reflect.Indirect(fieldValue).FieldByName(table.PKColumns()[0].FieldName)
							// fix non-int pk issues
							if pkField.IsValid() && (!requiredField && !isZero(pkField.Interface())) {
								val = pkField.Interface()
							} else {
								continue
							}
						} else {
							// TODO: how to handler?
							panic("not supported")
						}
					} else {
						val = fieldValue.Interface()
					}
				} else {
					// Blank struct could not be as update data
					if requiredField || !isStructZero(fieldValue) {
						bytes, err := DefaultJSONHandler.Marshal(fieldValue.Interface())
						if err != nil {
							panic(fmt.Sprintf("mashal %v failed", fieldValue.Interface()))
						}
						if col.SQLType.IsText() {
							val = string(bytes)
						} else if col.SQLType.IsBlob() {
							val = bytes
						}
					} else {
						continue
					}
				}
			}
		case reflect.Array, reflect.Slice, reflect.Map:
			if !requiredField {
				if fieldValue == reflect.Zero(fieldType) {
					continue
				}
				if fieldType.Kind() == reflect.Array {
					if isArrayValueZero(fieldValue) {
						continue
					}
				} else if fieldValue.IsNil() || !fieldValue.IsValid() || fieldValue.Len() == 0 {
					continue
				}
			}

			if col.SQLType.IsText() {
				bytes, err := DefaultJSONHandler.Marshal(fieldValue.Interface())
				if err != nil {
					engine.logger.Error(err)
					continue
				}
				val = string(bytes)
			} else if col.SQLType.IsBlob() {
				var bytes []byte
				var err error
				if fieldType.Kind() == reflect.Slice &&
					fieldType.Elem().Kind() == reflect.Uint8 {
					if fieldValue.Len() > 0 {
						val = fieldValue.Bytes()
					} else {
						continue
					}
				} else if fieldType.Kind() == reflect.Array &&
					fieldType.Elem().Kind() == reflect.Uint8 {
					val = fieldValue.Slice(0, 0).Interface()
				} else {
					bytes, err = DefaultJSONHandler.Marshal(fieldValue.Interface())
					if err != nil {
						engine.logger.Error(err)
						continue
					}
					val = bytes
				}
			} else {
				continue
			}
		default:
			val = fieldValue.Interface()
		}

	APPEND:
		args = append(args, val)
		if col.IsPrimaryKey && engine.dialect.DBType() == "ql" {
			continue
		}
		colNames = append(colNames, fmt.Sprintf("%v = ?", engine.Quote(col.Name)))
	}

	return colNames, args
}

func (statement *Statement) needTableName() bool {
	return len(statement.JoinStr) > 0
}

func (statement *Statement) colName(col *core.Column, tableName string) string {
	if statement.needTableName() {
		var nm = tableName
		if len(statement.TableAlias) > 0 {
			nm = statement.TableAlias
		}
		return statement.Engine.Quote(nm) + "." + statement.Engine.Quote(col.Name)
	}
	return statement.Engine.Quote(col.Name)
}

// TableName return current tableName
func (statement *Statement) TableName() string {
	if statement.AltTableName != "" {
		return statement.AltTableName
	}

	return statement.tableName
}

// ID generate "where id = ? " statement or for composite key "where key1 = ? and key2 = ?"
func (statement *Statement) ID(id interface{}) *Statement {
	idValue := reflect.ValueOf(id)
	idType := reflect.TypeOf(idValue.Interface())

	switch idType {
	case ptrPkType:
		if pkPtr, ok := (id).(*core.PK); ok {
			statement.idParam = pkPtr
			return statement
		}
	case pkType:
		if pk, ok := (id).(core.PK); ok {
			statement.idParam = &pk
			return statement
		}
	}

	switch idType.Kind() {
	case reflect.String:
		statement.idParam = &core.PK{idValue.Convert(reflect.TypeOf("")).Interface()}
		return statement
	}

	statement.idParam = &core.PK{id}
	return statement
}

// Incr Generate  "Update ... Set column = column + arg" statement
func (statement *Statement) Incr(column string, arg ...interface{}) *Statement {
	k := strings.ToLower(column)
	if len(arg) > 0 {
		statement.incrColumns[k] = incrParam{column, arg[0]}
	} else {
		statement.incrColumns[k] = incrParam{column, 1}
	}
	return statement
}

// Decr Generate  "Update ... Set column = column - arg" statement
func (statement *Statement) Decr(column string, arg ...interface{}) *Statement {
	k := strings.ToLower(column)
	if len(arg) > 0 {
		statement.decrColumns[k] = decrParam{column, arg[0]}
	} else {
		statement.decrColumns[k] = decrParam{column, 1}
	}
	return statement
}

// SetExpr Generate  "Update ... Set column = {expression}" statement
func (statement *Statement) SetExpr(column string, expression string) *Statement {
	k := strings.ToLower(column)
	statement.exprColumns[k] = exprParam{column, expression}
	return statement
}

// Generate  "Update ... Set column = column + arg" statement
func (statement *Statement) getInc() map[string]incrParam {
	return statement.incrColumns
}

// Generate  "Update ... Set column = column - arg" statement
func (statement *Statement) getDec() map[string]decrParam {
	return statement.decrColumns
}

// Generate  "Update ... Set column = {expression}" statement
func (statement *Statement) getExpr() map[string]exprParam {
	return statement.exprColumns
}

func (statement *Statement) col2NewColsWithQuote(columns ...string) []string {
	newColumns := make([]string, 0)
	quotes := append(strings.Split(statement.Engine.Quote(""), ""), "`")
	for _, col := range columns {
		newColumns = append(newColumns, statement.Engine.Quote(eraseAny(col, quotes...)))
	}
	return newColumns
}

func (statement *Statement) colmap2NewColsWithQuote() []string {
	newColumns := make([]string, len(statement.columnMap), len(statement.columnMap))
	copy(newColumns, statement.columnMap)
	for i := 0; i < len(statement.columnMap); i++ {
		newColumns[i] = statement.Engine.Quote(newColumns[i])
	}
	return newColumns
}

// Distinct generates "DISTINCT col1, col2 " statement
func (statement *Statement) Distinct(columns ...string) *Statement {
	statement.IsDistinct = true
	statement.Cols(columns...)
	return statement
}

// ForUpdate generates "SELECT ... FOR UPDATE" statement
func (statement *Statement) ForUpdate() *Statement {
	statement.IsForUpdate = true
	return statement
}

// Select replace select
func (statement *Statement) Select(str string) *Statement {
	statement.selectStr = str
	return statement
}

// Cols generate "col1, col2" statement
func (statement *Statement) Cols(columns ...string) *Statement {
	cols := col2NewCols(columns...)
	for _, nc := range cols {
		statement.columnMap.add(nc)
	}

	newColumns := statement.colmap2NewColsWithQuote()

	statement.ColumnStr = strings.Join(newColumns, ", ")
	statement.ColumnStr = strings.Replace(statement.ColumnStr, statement.Engine.quote("*"), "*", -1)
	return statement
}

// AllCols update use only: update all columns
func (statement *Statement) AllCols() *Statement {
	statement.useAllCols = true
	return statement
}

// MustCols update use only: must update columns
func (statement *Statement) MustCols(columns ...string) *Statement {
	newColumns := col2NewCols(columns...)
	for _, nc := range newColumns {
		statement.mustColumnMap[strings.ToLower(nc)] = true
	}
	return statement
}

// UseBool indicates that use bool fields as update contents and query contiditions
func (statement *Statement) UseBool(columns ...string) *Statement {
	if len(columns) > 0 {
		statement.MustCols(columns...)
	} else {
		statement.allUseBool = true
	}
	return statement
}

// Omit do not use the columns
func (statement *Statement) Omit(columns ...string) {
	newColumns := col2NewCols(columns...)
	for _, nc := range newColumns {
		statement.omitColumnMap = append(statement.omitColumnMap, nc)
	}
	statement.OmitStr = statement.Engine.Quote(strings.Join(newColumns, statement.Engine.Quote(", ")))
}

// Nullable Update use only: update columns to null when value is nullable and zero-value
func (statement *Statement) Nullable(columns ...string) {
	newColumns := col2NewCols(columns...)
	for _, nc := range newColumns {
		statement.nullableMap[strings.ToLower(nc)] = true
	}
}

// Top generate LIMIT limit statement
func (statement *Statement) Top(limit int) *Statement {
	statement.Limit(limit)
	return statement
}

// Limit generate LIMIT start, limit statement
func (statement *Statement) Limit(limit int, start ...int) *Statement {
	statement.LimitN = limit
	if len(start) > 0 {
		statement.Start = start[0]
	}
	return statement
}

// OrderBy generate "Order By order" statement
func (statement *Statement) OrderBy(order string) *Statement {
	if len(statement.OrderStr) > 0 {
		statement.OrderStr += ", "
	}
	statement.OrderStr += order
	return statement
}

// Desc generate `ORDER BY xx DESC`
func (statement *Statement) Desc(colNames ...string) *Statement {
	var buf builder.StringBuilder
	if len(statement.OrderStr) > 0 {
		fmt.Fprint(&buf, statement.OrderStr, ", ")
	}
	newColNames := statement.col2NewColsWithQuote(colNames...)
	fmt.Fprintf(&buf, "%v DESC", strings.Join(newColNames, " DESC, "))
	statement.OrderStr = buf.String()
	return statement
}

// Asc provide asc order by query condition, the input parameters are columns.
func (statement *Statement) Asc(colNames ...string) *Statement {
	var buf builder.StringBuilder
	if len(statement.OrderStr) > 0 {
		fmt.Fprint(&buf, statement.OrderStr, ", ")
	}
	newColNames := statement.col2NewColsWithQuote(colNames...)
	fmt.Fprintf(&buf, "%v ASC", strings.Join(newColNames, " ASC, "))
	statement.OrderStr = buf.String()
	return statement
}

// Table tempororily set table name, the parameter could be a string or a pointer of struct
func (statement *Statement) Table(tableNameOrBean interface{}) *Statement {
	v := rValue(tableNameOrBean)
	t := v.Type()
	if t.Kind() == reflect.Struct {
		var err error
		statement.RefTable, err = statement.Engine.autoMapType(v)
		if err != nil {
			statement.Engine.logger.Error(err)
			return statement
		}
	}

	statement.AltTableName = statement.Engine.TableName(tableNameOrBean, true)
	return statement
}

// Join The joinOP should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (statement *Statement) Join(joinOP string, tablename interface{}, condition string, args ...interface{}) *Statement {
	var buf builder.StringBuilder
	if len(statement.JoinStr) > 0 {
		fmt.Fprintf(&buf, "%v %v JOIN ", statement.JoinStr, joinOP)
	} else {
		fmt.Fprintf(&buf, "%v JOIN ", joinOP)
	}

	switch tp := tablename.(type) {
	case builder.Builder:
		subSQL, subQueryArgs, err := tp.ToSQL()
		if err != nil {
			statement.lastError = err
			return statement
		}
		tbs := strings.Split(tp.TableName(), ".")
		quotes := append(strings.Split(statement.Engine.Quote(""), ""), "`")

		var aliasName = strings.Trim(tbs[len(tbs)-1], strings.Join(quotes, ""))
		fmt.Fprintf(&buf, "(%s) %s ON %v", subSQL, aliasName, condition)
		statement.joinArgs = append(statement.joinArgs, subQueryArgs...)
	case *builder.Builder:
		subSQL, subQueryArgs, err := tp.ToSQL()
		if err != nil {
			statement.lastError = err
			return statement
		}
		tbs := strings.Split(tp.TableName(), ".")
		quotes := append(strings.Split(statement.Engine.Quote(""), ""), "`")

		var aliasName = strings.Trim(tbs[len(tbs)-1], strings.Join(quotes, ""))
		fmt.Fprintf(&buf, "(%s) %s ON %v", subSQL, aliasName, condition)
		statement.joinArgs = append(statement.joinArgs, subQueryArgs...)
	default:
		tbName := statement.Engine.TableName(tablename, true)
		fmt.Fprintf(&buf, "%s ON %v", tbName, condition)
	}

	statement.JoinStr = buf.String()
	statement.joinArgs = append(statement.joinArgs, args...)
	return statement
}

// GroupBy generate "Group By keys" statement
func (statement *Statement) GroupBy(keys string) *Statement {
	statement.GroupByStr = keys
	return statement
}

// Having generate "Having conditions" statement
func (statement *Statement) Having(conditions string) *Statement {
	statement.HavingStr = fmt.Sprintf("HAVING %v", conditions)
	return statement
}

// Unscoped always disable struct tag "deleted"
func (statement *Statement) Unscoped() *Statement {
	statement.unscoped = true
	return statement
}

func (statement *Statement) genColumnStr() string {
	if statement.RefTable == nil {
		return ""
	}

	var buf builder.StringBuilder
	columns := statement.RefTable.Columns()

	for _, col := range columns {
		if statement.omitColumnMap.contain(col.Name) {
			continue
		}

		if len(statement.columnMap) > 0 && !statement.columnMap.contain(col.Name) {
			continue
		}

		if col.MapType == core.ONLYTODB {
			continue
		}

		if buf.Len() != 0 {
			buf.WriteString(", ")
		}

		if statement.JoinStr != "" {
			if statement.TableAlias != "" {
				buf.WriteString(statement.TableAlias)
			} else {
				buf.WriteString(statement.TableName())
			}

			buf.WriteString(".")
		}

		statement.Engine.QuoteTo(&buf, col.Name)
	}

	return buf.String()
}

func (statement *Statement) genCreateTableSQL() string {
	return statement.Engine.dialect.CreateTableSql(statement.RefTable, statement.TableName(),
		statement.StoreEngine, statement.Charset)
}

func (statement *Statement) genIndexSQL() []string {
	var sqls []string
	tbName := statement.TableName()
	for _, index := range statement.RefTable.Indexes {
		if index.Type == core.IndexType {
			sql := statement.Engine.dialect.CreateIndexSql(tbName, index)
			/*idxTBName := strings.Replace(tbName, ".", "_", -1)
			idxTBName = strings.Replace(idxTBName, `"`, "", -1)
			sql := fmt.Sprintf("CREATE INDEX %v ON %v (%v);", quote(indexName(idxTBName, idxName)),
				quote(tbName), quote(strings.Join(index.Cols, quote(","))))*/
			sqls = append(sqls, sql)
		}
	}
	return sqls
}

func uniqueName(tableName, uqeName string) string {
	return fmt.Sprintf("UQE_%v_%v", tableName, uqeName)
}

func (statement *Statement) genUniqueSQL() []string {
	var sqls []string
	tbName := statement.TableName()
	for _, index := range statement.RefTable.Indexes {
		if index.Type == core.UniqueType {
			sql := statement.Engine.dialect.CreateIndexSql(tbName, index)
			sqls = append(sqls, sql)
		}
	}
	return sqls
}

func (statement *Statement) genDelIndexSQL() []string {
	var sqls []string
	tbName := statement.TableName()
	idxPrefixName := strings.Replace(tbName, `"`, "", -1)
	idxPrefixName = strings.Replace(idxPrefixName, `.`, "_", -1)
	for idxName, index := range statement.RefTable.Indexes {
		var rIdxName string
		if index.Type == core.UniqueType {
			rIdxName = uniqueName(idxPrefixName, idxName)
		} else if index.Type == core.IndexType {
			rIdxName = indexName(idxPrefixName, idxName)
		}
		sql := fmt.Sprintf("DROP INDEX %v", statement.Engine.Quote(statement.Engine.TableName(rIdxName, true)))
		if statement.Engine.dialect.IndexOnTable() {
			sql += fmt.Sprintf(" ON %v", statement.Engine.Quote(tbName))
		}
		sqls = append(sqls, sql)
	}
	return sqls
}

func (statement *Statement) genAddColumnStr(col *core.Column) (string, []interface{}) {
	quote := statement.Engine.Quote
	sql := fmt.Sprintf("ALTER TABLE %v ADD %v", quote(statement.TableName()),
		col.String(statement.Engine.dialect))
	if statement.Engine.dialect.DBType() == core.MYSQL && len(col.Comment) > 0 {
		sql += " COMMENT '" + col.Comment + "'"
	}
	sql += ";"
	return sql, []interface{}{}
}

func (statement *Statement) buildConds(table *core.Table, bean interface{}, includeVersion bool, includeUpdated bool, includeNil bool, includeAutoIncr bool, addedTableName bool) (builder.Cond, error) {
	return statement.Engine.buildConds(table, bean, includeVersion, includeUpdated, includeNil, includeAutoIncr, statement.allUseBool, statement.useAllCols,
		statement.unscoped, statement.mustColumnMap, statement.TableName(), statement.TableAlias, addedTableName)
}

func (statement *Statement) mergeConds(bean interface{}) error {
	if !statement.noAutoCondition {
		var addedTableName = (len(statement.JoinStr) > 0)
		autoCond, err := statement.buildConds(statement.RefTable, bean, true, true, false, true, addedTableName)
		if err != nil {
			return err
		}
		statement.cond = statement.cond.And(autoCond)
	}

	if err := statement.processIDParam(); err != nil {
		return err
	}
	return nil
}

func (statement *Statement) genConds(bean interface{}) (string, []interface{}, error) {
	if err := statement.mergeConds(bean); err != nil {
		return "", nil, err
	}

	return builder.ToSQL(statement.cond)
}

func (statement *Statement) genGetSQL(bean interface{}) (string, []interface{}, error) {
	v := rValue(bean)
	isStruct := v.Kind() == reflect.Struct
	if isStruct {
		statement.setRefBean(bean)
	}

	var columnStr = statement.ColumnStr
	if len(statement.selectStr) > 0 {
		columnStr = statement.selectStr
	} else {
		// TODO: always generate column names, not use * even if join
		if len(statement.JoinStr) == 0 {
			if len(columnStr) == 0 {
				if len(statement.GroupByStr) > 0 {
					columnStr = statement.Engine.quoteColumns(statement.GroupByStr)
				} else {
					columnStr = statement.genColumnStr()
				}
			}
		} else {
			if len(columnStr) == 0 {
				if len(statement.GroupByStr) > 0 {
					columnStr = statement.Engine.quoteColumns(statement.GroupByStr)
				}
			}
		}
	}

	if len(columnStr) == 0 {
		columnStr = "*"
	}

	if isStruct {
		if err := statement.mergeConds(bean); err != nil {
			return "", nil, err
		}
	} else {
		if err := statement.processIDParam(); err != nil {
			return "", nil, err
		}
	}
	condSQL, condArgs, err := builder.ToSQL(statement.cond)
	if err != nil {
		return "", nil, err
	}

	sqlStr, err := statement.genSelectSQL(columnStr, condSQL, true, true)
	if err != nil {
		return "", nil, err
	}

	return sqlStr, append(statement.joinArgs, condArgs...), nil
}

func (statement *Statement) genCountSQL(beans ...interface{}) (string, []interface{}, error) {
	var condSQL string
	var condArgs []interface{}
	var err error
	if len(beans) > 0 {
		statement.setRefBean(beans[0])
		condSQL, condArgs, err = statement.genConds(beans[0])
	} else {
		condSQL, condArgs, err = builder.ToSQL(statement.cond)
	}
	if err != nil {
		return "", nil, err
	}

	var selectSQL = statement.selectStr
	if len(selectSQL) <= 0 {
		if statement.IsDistinct {
			selectSQL = fmt.Sprintf("count(DISTINCT %s)", statement.ColumnStr)
		} else {
			selectSQL = "count(*)"
		}
	}
	sqlStr, err := statement.genSelectSQL(selectSQL, condSQL, false, false)
	if err != nil {
		return "", nil, err
	}

	return sqlStr, append(statement.joinArgs, condArgs...), nil
}

func (statement *Statement) genSumSQL(bean interface{}, columns ...string) (string, []interface{}, error) {
	statement.setRefBean(bean)

	var sumStrs = make([]string, 0, len(columns))
	for _, colName := range columns {
		if !strings.Contains(colName, " ") && !strings.Contains(colName, "(") {
			colName = statement.Engine.Quote(colName)
		}
		sumStrs = append(sumStrs, fmt.Sprintf("COALESCE(sum(%s),0)", colName))
	}
	sumSelect := strings.Join(sumStrs, ", ")

	condSQL, condArgs, err := statement.genConds(bean)
	if err != nil {
		return "", nil, err
	}

	sqlStr, err := statement.genSelectSQL(sumSelect, condSQL, true, true)
	if err != nil {
		return "", nil, err
	}

	return sqlStr, append(statement.joinArgs, condArgs...), nil
}

func (statement *Statement) genSelectSQL(columnStr, condSQL string, needLimit, needOrderBy bool) (string, error) {
	var (
		distinct                  string
		dialect                   = statement.Engine.Dialect()
		quote                     = statement.Engine.Quote
		fromStr                   = " FROM "
		top, mssqlCondi, whereStr string
	)
	if statement.IsDistinct && !strings.HasPrefix(columnStr, "count") {
		distinct = "DISTINCT "
	}
	if len(condSQL) > 0 {
		whereStr = " WHERE " + condSQL
	}

	if dialect.DBType() == core.MSSQL && strings.Contains(statement.TableName(), "..") {
		fromStr += statement.TableName()
	} else {
		fromStr += quote(statement.TableName())
	}

	if statement.TableAlias != "" {
		if dialect.DBType() == core.ORACLE {
			fromStr += " " + quote(statement.TableAlias)
		} else {
			fromStr += " AS " + quote(statement.TableAlias)
		}
	}
	if statement.JoinStr != "" {
		fromStr = fmt.Sprintf("%v %v", fromStr, statement.JoinStr)
	}

	if dialect.DBType() == core.MSSQL {
		if statement.LimitN > 0 {
			top = fmt.Sprintf("TOP %d ", statement.LimitN)
		}
		if statement.Start > 0 {
			var column string
			if len(statement.RefTable.PKColumns()) == 0 {
				for _, index := range statement.RefTable.Indexes {
					if len(index.Cols) == 1 {
						column = index.Cols[0]
						break
					}
				}
				if len(column) == 0 {
					column = statement.RefTable.ColumnsSeq()[0]
				}
			} else {
				column = statement.RefTable.PKColumns()[0].Name
			}
			if statement.needTableName() {
				if len(statement.TableAlias) > 0 {
					column = statement.TableAlias + "." + column
				} else {
					column = statement.TableName() + "." + column
				}
			}

			var orderStr string
			if needOrderBy && len(statement.OrderStr) > 0 {
				orderStr = " ORDER BY " + statement.OrderStr
			}

			var groupStr string
			if len(statement.GroupByStr) > 0 {
				groupStr = " GROUP BY " + statement.GroupByStr
			}
			mssqlCondi = fmt.Sprintf("(%s NOT IN (SELECT TOP %d %s%s%s%s%s))",
				column, statement.Start, column, fromStr, whereStr, orderStr, groupStr)
		}
	}

	var buf builder.StringBuilder
	fmt.Fprintf(&buf, "SELECT %v%v%v%v%v", distinct, top, columnStr, fromStr, whereStr)
	if len(mssqlCondi) > 0 {
		if len(whereStr) > 0 {
			fmt.Fprint(&buf, " AND ", mssqlCondi)
		} else {
			fmt.Fprint(&buf, " WHERE ", mssqlCondi)
		}
	}

	if statement.GroupByStr != "" {
		fmt.Fprint(&buf, " GROUP BY ", statement.GroupByStr)
	}
	if statement.HavingStr != "" {
		fmt.Fprint(&buf, " ", statement.HavingStr)
	}
	if needOrderBy && statement.OrderStr != "" {
		fmt.Fprint(&buf, " ORDER BY ", statement.OrderStr)
	}
	if needLimit {
		if dialect.DBType() != core.MSSQL && dialect.DBType() != core.ORACLE {
			if statement.Start > 0 {
				fmt.Fprintf(&buf, " LIMIT %v OFFSET %v", statement.LimitN, statement.Start)
			} else if statement.LimitN > 0 {
				fmt.Fprint(&buf, " LIMIT ", statement.LimitN)
			}
		} else if dialect.DBType() == core.ORACLE {
			if statement.Start != 0 || statement.LimitN != 0 {
				oldString := buf.String()
				buf.Reset()
				rawColStr := columnStr
				if rawColStr == "*" {
					rawColStr = "at.*"
				}
				fmt.Fprintf(&buf, "SELECT %v FROM (SELECT %v,ROWNUM RN FROM (%v) at WHERE ROWNUM <= %d) aat WHERE RN > %d",
					columnStr, rawColStr, oldString, statement.Start+statement.LimitN, statement.Start)
			}
		}
	}
	if statement.IsForUpdate {
		return dialect.ForUpdateSql(buf.String()), nil
	}

	return buf.String(), nil
}

func (statement *Statement) processIDParam() error {
	if statement.idParam == nil || statement.RefTable == nil {
		return nil
	}

	if len(statement.RefTable.PrimaryKeys) != len(*statement.idParam) {
		return fmt.Errorf("ID condition is error, expect %d primarykeys, there are %d",
			len(statement.RefTable.PrimaryKeys),
			len(*statement.idParam),
		)
	}

	for i, col := range statement.RefTable.PKColumns() {
		var colName = statement.colName(col, statement.TableName())
		statement.cond = statement.cond.And(builder.Eq{colName: (*(statement.idParam))[i]})
	}
	return nil
}

func (statement *Statement) joinColumns(cols []*core.Column, includeTableName bool) string {
	var colnames = make([]string, len(cols))
	for i, col := range cols {
		if includeTableName {
			colnames[i] = statement.Engine.Quote(statement.TableName()) +
				"." + statement.Engine.Quote(col.Name)
		} else {
			colnames[i] = statement.Engine.Quote(col.Name)
		}
	}
	return strings.Join(colnames, ", ")
}

func (statement *Statement) convertIDSQL(sqlStr string) string {
	if statement.RefTable != nil {
		cols := statement.RefTable.PKColumns()
		if len(cols) == 0 {
			return ""
		}

		colstrs := statement.joinColumns(cols, false)
		sqls := splitNNoCase(sqlStr, " from ", 2)
		if len(sqls) != 2 {
			return ""
		}

		var top string
		if statement.LimitN > 0 && statement.Engine.dialect.DBType() == core.MSSQL {
			top = fmt.Sprintf("TOP %d ", statement.LimitN)
		}

		newsql := fmt.Sprintf("SELECT %s%s FROM %v", top, colstrs, sqls[1])
		return newsql
	}
	return ""
}

func (statement *Statement) convertUpdateSQL(sqlStr string) (string, string) {
	if statement.RefTable == nil || len(statement.RefTable.PrimaryKeys) != 1 {
		return "", ""
	}

	colstrs := statement.joinColumns(statement.RefTable.PKColumns(), true)
	sqls := splitNNoCase(sqlStr, "where", 2)
	if len(sqls) != 2 {
		if len(sqls) == 1 {
			return sqls[0], fmt.Sprintf("SELECT %v FROM %v",
				colstrs, statement.Engine.Quote(statement.TableName()))
		}
		return "", ""
	}

	var whereStr = sqls[1]

	// TODO: for postgres only, if any other database?
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
		colstrs, statement.Engine.Quote(statement.TableName()),
		whereStr)
}
