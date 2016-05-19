// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-xorm/core"
)

type inParam struct {
	colName string
	args    []interface{}
}

type incrParam struct {
	colName string
	arg     interface{}
}

type decrParam struct {
	colName string
	arg     interface{}
}

type exprParam struct {
	colName string
	expr    string
}

// Statement save all the sql info for executing SQL
type Statement struct {
	RefTable        *core.Table
	OutTable        *core.Table
	Engine          *Engine
	Start           int
	LimitN          int
	WhereStr        string
	IdParam         *core.PK
	Params          []interface{}
	OrderStr        string
	JoinStr         string
	joinArgs        []interface{}
	GroupByStr      string
	HavingStr       string
	ColumnStr       string
	selectStr       string
	columnMap       map[string]bool
	tableMap        map[string][]string
	useAllCols      bool
	OmitStr         string
	ConditionStr    string
	AltTableName    string
	RawSQL          string
	RawParams       []interface{}
	UseCascade      bool
	UseAutoJoin     bool
	AllowAmbiguous  bool
	StoreEngine     string
	Charset         string
	BeanArgs        []interface{}
	UseCache        bool
	UseAutoTime     bool
	noAutoCondition bool
	IsDistinct      bool
	IsForUpdate     bool
	TableAlias      string
	allUseBool      bool
	checkVersion    bool
	unscoped        bool
	mustColumnMap   map[string]bool
	nullableMap     map[string]bool
	inColumns       map[string]*inParam
	incrColumns     map[string]incrParam
	decrColumns     map[string]decrParam
	exprColumns     map[string]exprParam
}

// Init reset all the statment's fields
func (statement *Statement) Init() {
	statement.RefTable = nil
	statement.OutTable = nil
	statement.Start = 0
	statement.LimitN = 0
	statement.WhereStr = ""
	statement.Params = make([]interface{}, 0)
	statement.OrderStr = ""
	statement.UseCascade = true
	statement.JoinStr = ""
	statement.joinArgs = make([]interface{}, 0)
	statement.GroupByStr = ""
	statement.HavingStr = ""
	statement.ColumnStr = ""
	statement.OmitStr = ""
	statement.columnMap = make(map[string]bool)
	statement.tableMap = make(map[string][]string)
	statement.ConditionStr = ""
	statement.AltTableName = ""
	statement.IdParam = nil
	statement.RawSQL = ""
	statement.RawParams = make([]interface{}, 0)
	statement.BeanArgs = make([]interface{}, 0)
	statement.UseCache = true
	statement.UseAutoTime = true
	statement.AllowAmbiguous = true
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
	statement.inColumns = make(map[string]*inParam)
	statement.incrColumns = make(map[string]incrParam)
	statement.decrColumns = make(map[string]decrParam)
	statement.exprColumns = make(map[string]exprParam)
}

// NoAutoCondition if you do not want convert bean's field as query condition, then use this function
func (statement *Statement) NoAutoCondition(no ...bool) *Statement {
	statement.noAutoCondition = true
	if len(no) > 0 {
		statement.noAutoCondition = no[0]
	}
	return statement
}

// Sql add the raw sql statement
func (statement *Statement) Sql(querystring string, args ...interface{}) *Statement {
	statement.RawSQL = querystring
	statement.RawParams = args
	return statement
}

// Alias set the table alias
func (statement *Statement) Alias(alias string) *Statement {
	if statement.TableAlias != "" {
		statement.tableMapDelete(statement.TableAlias)
	}
	statement.TableAlias = alias
	statement.tableMapAdd(statement.TableName(), statement.TableAlias)
	statement.tableMapAdd(statement.TableAlias, statement.TableAlias)
	return statement
}

// Where add Where statment
func (statement *Statement) Where(querystring string, args ...interface{}) *Statement {
	// The second where will be triggered as And
	if len(statement.WhereStr) > 0 {
		return statement.And(querystring, args...)
	}

	if !strings.Contains(querystring, statement.Engine.dialect.EqStr()) {
		querystring = strings.Replace(querystring, "=", statement.Engine.dialect.EqStr(), -1)
	}
	statement.WhereStr = querystring
	statement.Params = args
	return statement
}

// And add Where & and statment
func (statement *Statement) And(querystring string, args ...interface{}) *Statement {
	if len(statement.WhereStr) > 0 {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "(%v) %s (%v)", statement.WhereStr,
			statement.Engine.dialect.AndStr(), querystring)
		statement.WhereStr = buf.String()
	} else {
		statement.WhereStr = querystring
	}
	statement.Params = append(statement.Params, args...)
	return statement
}

// Or add Where & Or statment
func (statement *Statement) Or(querystring string, args ...interface{}) *Statement {
	if len(statement.WhereStr) > 0 {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "(%v) %s (%v)", statement.WhereStr,
			statement.Engine.dialect.OrStr(), querystring)
		statement.WhereStr = buf.String()
	} else {
		statement.WhereStr = querystring
	}
	statement.Params = append(statement.Params, args...)
	return statement
}

// Table tempororily set table name, the parameter could be a string or a pointer of struct
func (statement *Statement) Table(tableNameOrBean interface{}) *Statement {
	if statement.TableName() != "" {
		statement.tableMapDelete(statement.TableName())
	}
	v := rValue(tableNameOrBean)
	t := v.Type()
	if t.Kind() == reflect.String {
		statement.AltTableName = tableNameOrBean.(string)
	} else if t.Kind() == reflect.Struct {
		statement.RefTable = statement.Engine.autoMapType(v)
	}
	if statement.TableAlias != "" {
		statement.tableMapAdd(statement.TableName(), statement.TableAlias)
	} else {
		statement.tableMapAdd(statement.TableName(), statement.TableName())
	}
	return statement
}

// Auto generating update columnes and values according a struct
func buildUpdates(engine *Engine, table *core.Table, bean interface{},
	includeVersion bool, includeUpdated bool, includeNil bool,
	includeAutoIncr bool, allUseBool bool, useAllCols bool,
	mustColumnMap map[string]bool, nullableMap map[string]bool,
	columnMap map[string]bool, update, unscoped bool) ([]string, []interface{}) {

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
		if use, ok := columnMap[col.Name]; ok && !use {
			continue
		}

		fieldValuePtr, err := col.ValueOf(bean)
		if err != nil {
			engine.logger.Error(err)
			continue
		}

		fieldValue := *fieldValuePtr
		fieldType := reflect.TypeOf(fieldValue.Interface())

		requiredField := useAllCols
		includeNil := useAllCols
		lColName := strings.ToLower(col.Name)

		if b, ok := mustColumnMap[lColName]; ok {
			if b {
				requiredField = true
			} else {
				continue
			}
		}

		// !evalphobia! set fieldValue as nil when column is nullable and zero-value
		if b, ok := nullableMap[lColName]; ok {
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
				val = engine.FormatTime(col.SQLType.Name, t)
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
							//TODO: how to handler?
							panic("not supported")
						}
					} else {
						val = fieldValue.Interface()
					}
				} else {
					// Blank struct could not be as update data
					if requiredField || !isStructZero(fieldValue) {
						bytes, err := json.Marshal(fieldValue.Interface())
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
				if fieldValue.IsNil() || !fieldValue.IsValid() || fieldValue.Len() == 0 {
					continue
				}
			}

			if col.SQLType.IsText() {
				bytes, err := json.Marshal(fieldValue.Interface())
				if err != nil {
					engine.logger.Error(err)
					continue
				}
				val = string(bytes)
			} else if col.SQLType.IsBlob() {
				var bytes []byte
				var err error
				if (fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice) &&
					fieldType.Elem().Kind() == reflect.Uint8 {
					if fieldValue.Len() > 0 {
						val = fieldValue.Bytes()
					} else {
						continue
					}
				} else {
					bytes, err = json.Marshal(fieldValue.Interface())
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
		//fmt.Println("==", col.Name, "==", fmt.Sprintf("%v", val))
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

func (statement *Statement) tableMapAdd(table, alias string) {
	key := statement.Engine.Quote(strings.ToLower(table))
	statement.tableMap[key] = append(statement.tableMap[key], alias)
}

func (statement *Statement) tableMapDelete(table string) {
	key := statement.Engine.Quote(strings.ToLower(table))
	delete(statement.tableMap, key)
}

func (statement *Statement) isKnownUnambiguousTable(table string) string {
	key := statement.Engine.Quote(strings.ToLower(table))
	arr := statement.tableMap[key]
	if len(arr) != 1 {
		return ""
	}
	return arr[0]
}

func (statement *Statement) detectTableName(col *core.Column) string {
	for _, t := range col.TableNames {
		if name := statement.isKnownUnambiguousTable(t); name != "" {
			return name
		}
	}
	if name := statement.isKnownUnambiguousTable(statement.outTableName()); name != "" {
		return name
	}
	return ""
}

func (statement *Statement) colName(col *core.Column, defaultTable string) (string, bool) {
	var colTable string
	if statement.needTableName() {
		colTable = statement.detectTableName(col)
		if colTable == "" {
			colTable = defaultTable
		}
	}
	if colTable != "" {
		ret := statement.Engine.Quote(colTable) + "." + statement.Engine.Quote(col.Name)
		return ret, true
	} else {
		ret := statement.Engine.Quote(col.Name)
		return ret, false
	}
}

// Auto generating conditions according a struct
func (statement *Statement) buildConditions(
	table *core.Table, bean interface{},
	includeVersion bool, includeUpdated bool, includeNil bool, includeAutoIncr bool,
	addedTableName bool) ([]string, []interface{}) {
	engine := statement.Engine
	var colNames []string
	var args = make([]interface{}, 0)
	for _, col := range table.Columns() {
		if !includeVersion && col.IsVersion {
			continue
		}
		if !includeUpdated && col.IsUpdated {
			continue
		}
		if !includeAutoIncr && col.IsAutoIncrement {
			continue
		}

		if engine.dialect.DBType() == core.MSSQL && col.SQLType.Name == core.Text {
			continue
		}
		if col.SQLType.IsJson() {
			continue
		}

		colName, _ := statement.colName(col, statement.aliasedTableName())

		fieldValuePtr, err := col.ValueOf(bean)
		if err != nil {
			engine.logger.Error(err)
			continue
		}

		if col.IsDeleted && !statement.unscoped { // tag "deleted" is enabled
			colNames = append(colNames, fmt.Sprintf(
				"(%v IS NULL OR %v = '0001-01-01 00:00:00')", colName, colName))
		}

		fieldValue := *fieldValuePtr
		if fieldValue.Interface() == nil {
			continue
		}

		fieldType := reflect.TypeOf(fieldValue.Interface())
		requiredField := statement.useAllCols
		if b, ok := statement.mustColumnMap[strings.ToLower(col.Name)]; ok {
			if b {
				requiredField = true
			} else {
				continue
			}
		}

		if fieldType.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				if includeNil {
					args = append(args, nil)
					colNames = append(colNames, fmt.Sprintf("%v %s ?", colName, engine.dialect.EqStr()))
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

		var val interface{}
		switch fieldType.Kind() {
		case reflect.Bool:
			if statement.allUseBool || requiredField {
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
				val = engine.FormatTime(col.SQLType.Name, t)
			} else if _, ok := reflect.New(fieldType).Interface().(core.Conversion); ok {
				continue
			} else if valNul, ok := fieldValue.Interface().(driver.Valuer); ok {
				val, _ = valNul.Value()
				if val == nil {
					continue
				}
			} else {
				if col.SQLType.IsJson() {
					if col.SQLType.IsText() {
						bytes, err := json.Marshal(fieldValue.Interface())
						if err != nil {
							engine.logger.Error(err)
							continue
						}
						val = string(bytes)
					} else if col.SQLType.IsBlob() {
						var bytes []byte
						var err error
						bytes, err = json.Marshal(fieldValue.Interface())
						if err != nil {
							engine.logger.Error(err)
							continue
						}
						val = bytes
					}
				} else {
					engine.autoMapType(fieldValue)
					if table, ok := engine.Tables[fieldValue.Type()]; ok {
						if len(table.PrimaryKeys) == 1 {
							pkField := reflect.Indirect(fieldValue).FieldByName(table.PKColumns()[0].FieldName)
							// fix non-int pk issues
							//if pkField.Int() != 0 {
							if pkField.IsValid() && !isZero(pkField.Interface()) {
								val = pkField.Interface()
							} else {
								continue
							}
						} else {
							//TODO: how to handler?
							panic(fmt.Sprintln("not supported", fieldValue.Interface(), "as", table.PrimaryKeys))
						}
					} else {
						val = fieldValue.Interface()
					}
				}
			}
		case reflect.Array, reflect.Slice, reflect.Map:
			if fieldValue == reflect.Zero(fieldType) {
				continue
			}
			if fieldValue.IsNil() || !fieldValue.IsValid() || fieldValue.Len() == 0 {
				continue
			}

			if col.SQLType.IsText() {
				bytes, err := json.Marshal(fieldValue.Interface())
				if err != nil {
					engine.logger.Error(err)
					continue
				}
				val = string(bytes)
			} else if col.SQLType.IsBlob() {
				var bytes []byte
				var err error
				if (fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice) &&
					fieldType.Elem().Kind() == reflect.Uint8 {
					if fieldValue.Len() > 0 {
						val = fieldValue.Bytes()
					} else {
						continue
					}
				} else {
					bytes, err = json.Marshal(fieldValue.Interface())
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

		args = append(args, val)
		var condi string
		if col.IsPrimaryKey && engine.dialect.DBType() == "ql" {
			condi = "id() == ?"
		} else {
			condi = fmt.Sprintf("%v %s ?", colName, engine.dialect.EqStr())
		}
		colNames = append(colNames, condi)
	}

	return colNames, args
}

// TableName return current tableName
func (statement *Statement) TableName() string {
	if statement.AltTableName != "" {
		return statement.AltTableName
	}

	if statement.RefTable != nil {
		/*schema := statement.Engine.dialect.URI().Schema
		if len(schema) > 0 {
			return schema + "." + statement.RefTable.Name
		}*/
		return statement.RefTable.Name
	}
	return ""
}

func (statement *Statement) aliasedTableName() string {
	if statement.TableAlias != "" {
		return statement.TableAlias
	}
	return statement.TableName()
}

func (statement *Statement) outTableName() string {
	if statement.OutTable != nil {
		return statement.OutTable.Name
	}
	return ""
}

// Id generate "where id = ? " statment or for composite key "where key1 = ? and key2 = ?"
func (statement *Statement) Id(id interface{}) *Statement {
	idValue := reflect.ValueOf(id)
	idType := reflect.TypeOf(idValue.Interface())

	switch idType {
	case ptrPkType:
		if pkPtr, ok := (id).(*core.PK); ok {
			statement.IdParam = pkPtr
			return statement
		}
	case pkType:
		if pk, ok := (id).(core.PK); ok {
			statement.IdParam = &pk
			return statement
		}
	}

	switch idType.Kind() {
	case reflect.String:
		statement.IdParam = &core.PK{idValue.Convert(reflect.TypeOf("")).Interface()}
		return statement
	}

	statement.IdParam = &core.PK{id}
	return statement
}

// Incr Generate  "Update ... Set column = column + arg" statment
func (statement *Statement) Incr(column string, arg ...interface{}) *Statement {
	k := strings.ToLower(column)
	if len(arg) > 0 {
		statement.incrColumns[k] = incrParam{column, arg[0]}
	} else {
		statement.incrColumns[k] = incrParam{column, 1}
	}
	return statement
}

// Decr Generate  "Update ... Set column = column - arg" statment
func (statement *Statement) Decr(column string, arg ...interface{}) *Statement {
	k := strings.ToLower(column)
	if len(arg) > 0 {
		statement.decrColumns[k] = decrParam{column, arg[0]}
	} else {
		statement.decrColumns[k] = decrParam{column, 1}
	}
	return statement
}

// SetExpr Generate  "Update ... Set column = {expression}" statment
func (statement *Statement) SetExpr(column string, expression string) *Statement {
	k := strings.ToLower(column)
	statement.exprColumns[k] = exprParam{column, expression}
	return statement
}

// Generate  "Update ... Set column = column + arg" statment
func (statement *Statement) getInc() map[string]incrParam {
	return statement.incrColumns
}

// Generate  "Update ... Set column = column - arg" statment
func (statement *Statement) getDec() map[string]decrParam {
	return statement.decrColumns
}

// Generate  "Update ... Set column = {expression}" statment
func (statement *Statement) getExpr() map[string]exprParam {
	return statement.exprColumns
}

// In generate "Where column IN (?) " statment
func (statement *Statement) In(column string, args ...interface{}) *Statement {
	length := len(args)
	if length == 0 {
		return statement
	}

	k := strings.ToLower(column)
	var newargs []interface{}
	if length == 1 &&
		reflect.TypeOf(args[0]).Kind() == reflect.Slice {
		newargs = make([]interface{}, 0)
		v := reflect.ValueOf(args[0])
		for i := 0; i < v.Len(); i++ {
			newargs = append(newargs, v.Index(i).Interface())
		}
	} else {
		newargs = args
	}

	if _, ok := statement.inColumns[k]; ok {
		statement.inColumns[k].args = append(statement.inColumns[k].args, newargs...)
	} else {
		statement.inColumns[k] = &inParam{column, newargs}
	}
	return statement
}

func (statement *Statement) genInSql() (string, []interface{}) {
	if len(statement.inColumns) == 0 {
		return "", []interface{}{}
	}

	inStrs := make([]string, len(statement.inColumns), len(statement.inColumns))
	args := make([]interface{}, 0, len(statement.inColumns))
	var buf bytes.Buffer
	var i int
	for _, params := range statement.inColumns {
		buf.Reset()
		fmt.Fprintf(&buf, "(%v IN (%v))",
			statement.Engine.quoteColumn(params.colName),
			strings.Join(makeArray("?", len(params.args)), ","))
		inStrs[i] = buf.String()
		i++
		args = append(args, params.args...)
	}

	if len(statement.inColumns) == 1 {
		return inStrs[0], args
	}
	return fmt.Sprintf("(%v)", strings.Join(inStrs, " "+statement.Engine.dialect.AndStr()+" ")), args
}

func (statement *Statement) attachInSql() {
	inSql, inArgs := statement.genInSql()
	if len(inSql) > 0 {
		if len(statement.ConditionStr) > 0 {
			statement.ConditionStr += " " + statement.Engine.dialect.AndStr() + " "
		}
		statement.ConditionStr += inSql
		statement.Params = append(statement.Params, inArgs...)
	}
}

func (statement *Statement) col2NewColsWithQuote(columns ...string) []string {
	newColumns := make([]string, 0)
	for _, col := range columns {
		col = strings.Replace(col, "`", "", -1)
		col = strings.Replace(col, statement.Engine.QuoteStr(), "", -1)
		ccols := strings.Split(col, ",")
		for _, c := range ccols {
			fields := strings.Split(strings.TrimSpace(c), ".")
			if len(fields) == 1 {
				newColumns = append(newColumns, statement.Engine.quote(fields[0]))
			} else if len(fields) == 2 {
				newColumns = append(newColumns, statement.Engine.quote(fields[0])+"."+
					statement.Engine.quote(fields[1]))
			} else {
				panic(errors.New("unwanted colnames"))
			}
		}
	}
	return newColumns
}

// Generate "Distince col1, col2 " statment
func (statement *Statement) Distinct(columns ...string) *Statement {
	statement.IsDistinct = true
	statement.Cols(columns...)
	return statement
}

// Generate "SELECT ... FOR UPDATE" statment
func (statement *Statement) ForUpdate() *Statement {
	statement.IsForUpdate = true
	return statement
}

// Select replace select
func (s *Statement) Select(str string) *Statement {
	s.selectStr = str
	return s
}

// Cols generate "col1, col2" statement
func (statement *Statement) Cols(columns ...string) *Statement {
	cols := col2NewCols(columns...)
	for _, nc := range cols {
		statement.columnMap[strings.ToLower(nc)] = true
	}

	newColumns := statement.col2NewColsWithQuote(columns...)
	//fmt.Println("=====", columns, newColumns, cols)
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
		statement.columnMap[strings.ToLower(nc)] = false
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
	var buf bytes.Buffer
	fmt.Fprintf(&buf, statement.OrderStr)
	if len(statement.OrderStr) > 0 {
		fmt.Fprint(&buf, ", ")
	}
	newColNames := statement.col2NewColsWithQuote(colNames...)
	fmt.Fprintf(&buf, "%v DESC", strings.Join(newColNames, " DESC, "))
	statement.OrderStr = buf.String()
	return statement
}

// Asc provide asc order by query condition, the input parameters are columns.
func (statement *Statement) Asc(colNames ...string) *Statement {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, statement.OrderStr)
	if len(statement.OrderStr) > 0 {
		fmt.Fprint(&buf, ", ")
	}
	newColNames := statement.col2NewColsWithQuote(colNames...)
	fmt.Fprintf(&buf, "%v ASC", strings.Join(newColNames, " ASC, "))
	statement.OrderStr = buf.String()
	return statement
}

// Join The joinOP should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (statement *Statement) Join(joinOP string, tablearg interface{}, condition string, args ...interface{}) *Statement {
	var buf bytes.Buffer
	if len(statement.JoinStr) > 0 {
		fmt.Fprintf(&buf, "%v %v JOIN ", statement.JoinStr, joinOP)
	} else {
		fmt.Fprintf(&buf, "%v JOIN ", joinOP)
	}

	var tableName, aliasName string
	switch tablearg.(type) {
	case []string:
		t := tablearg.([]string)
		if len(t) > 1 {
			tableName = t[0]
			aliasName = t[1]
			fmt.Fprintf(&buf, "%v AS %v",
				statement.Engine.Quote(tableName), statement.Engine.Quote(aliasName))
		} else if len(t) == 1 {
			tableName = t[0]
			fmt.Fprintf(&buf, statement.Engine.Quote(tableName))
		}
	case []interface{}:
		t := tablearg.([]interface{})
		l := len(t)
		if l > 0 {
			f := t[0]
			v := rValue(f)
			t := v.Type()
			if t.Kind() == reflect.String {
				tableName = f.(string)
			} else if t.Kind() == reflect.Struct {
				r := statement.Engine.autoMapType(v)
				tableName = r.Name
			}
		}
		if l > 1 {
			aliasName = fmt.Sprintf("%v", t[1])
			fmt.Fprintf(&buf, "%v AS %v",
				statement.Engine.Quote(tableName), statement.Engine.Quote(aliasName))
		} else if l == 1 {
			fmt.Fprintf(&buf, statement.Engine.Quote(tableName))
		}
	default:
		tableName = fmt.Sprintf("%v", tablearg)
		fmt.Fprintf(&buf, statement.Engine.Quote(tableName))
	}

	fmt.Fprintf(&buf, " ON %v", condition)
	statement.JoinStr = buf.String()
	statement.joinArgs = append(statement.joinArgs, args...)
	if aliasName != "" {
		statement.tableMapAdd(tableName, aliasName)
		statement.tableMapAdd(aliasName, aliasName)
	} else {
		statement.tableMapAdd(tableName, tableName)
	}
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
	if len(statement.selectStr) > 0 {
		return statement.selectStr
	}

	if len(statement.ColumnStr) > 0 {
		return statement.ColumnStr
	}

	if len(statement.GroupByStr) > 0 {
		return statement.Engine.Quote(
			strings.Replace(statement.GroupByStr, ",", statement.Engine.Quote(","), -1))
	}

	table := statement.RefTable
	if statement.OutTable != nil {
		table = statement.OutTable
	}
	colNames := make([]string, 0)
	for _, col := range table.Columns() {
		if statement.OmitStr != "" {
			if _, ok := statement.columnMap[strings.ToLower(col.Name)]; ok {
				continue
			}
		}
		if col.MapType == core.ONLYTODB {
			continue
		}

		name, qualified := statement.colName(col, "")
		if statement.needTableName() && !qualified {
			if statement.AllowAmbiguous {
				statement.Engine.logger.Warnf(
					"can't detect table name for ambiguous field '%v', "+
						"falling back to 'SELECT *'", col.Name)
				return "*"
			} else {
				panic("can't detect table name for ambiguous field '"+col.Name+"'")
			}
		}

		if col.IsPrimaryKey && statement.Engine.Dialect().DBType() == "ql" {
			colNames = append(colNames, "id() AS "+name)
		} else {
			colNames = append(colNames, name)
		}
	}
	return strings.Join(colNames, ", ")
}

func (statement *Statement) genCreateTableSQL() string {
	return statement.Engine.dialect.CreateTableSql(statement.RefTable, statement.AltTableName,
		statement.StoreEngine, statement.Charset)
}

func (s *Statement) genIndexSQL() []string {
	var sqls []string = make([]string, 0)
	tbName := s.TableName()
	quote := s.Engine.Quote
	for idxName, index := range s.RefTable.Indexes {
		if index.Type == core.IndexType {
			sql := fmt.Sprintf("CREATE INDEX %v ON %v (%v);", quote(indexName(s.RefTable.Name, idxName)),
				quote(tbName), quote(strings.Join(index.Cols, quote(","))))
			sqls = append(sqls, sql)
		}
	}
	return sqls
}

func uniqueName(tableName, uqeName string) string {
	return fmt.Sprintf("UQE_%v_%v", tableName, uqeName)
}

func (s *Statement) genUniqueSQL() []string {
	var sqls []string = make([]string, 0)
	for _, index := range s.RefTable.Indexes {
		if index.Type == core.UniqueType {
			sql := s.Engine.dialect.CreateIndexSql(s.RefTable.Name, index)
			sqls = append(sqls, sql)
		}
	}
	return sqls
}

func (s *Statement) genDelIndexSQL() []string {
	var sqls []string = make([]string, 0)
	for idxName, index := range s.RefTable.Indexes {
		var rIdxName string
		if index.Type == core.UniqueType {
			rIdxName = uniqueName(s.RefTable.Name, idxName)
		} else if index.Type == core.IndexType {
			rIdxName = indexName(s.RefTable.Name, idxName)
		}
		sql := fmt.Sprintf("DROP INDEX %v", s.Engine.Quote(rIdxName))
		if s.Engine.dialect.IndexOnTable() {
			sql += fmt.Sprintf(" ON %v", s.Engine.Quote(s.TableName()))
		}
		sqls = append(sqls, sql)
	}
	return sqls
}

func (statement *Statement) genGetSql(bean interface{}) (string, []interface{}) {
	var table *core.Table
	if statement.RefTable == nil {
		table = statement.Engine.TableInfo(bean)
		statement.RefTable = table
	} else {
		table = statement.RefTable
	}

	if !statement.noAutoCondition {
		colNames, args := statement.buildConditions(
			table, bean, true, true, false, true, statement.needTableName())

		statement.ConditionStr = strings.Join(colNames, " "+statement.Engine.dialect.AndStr()+" ")
		statement.BeanArgs = args
	}

	columnStr := statement.genColumnStr()

	statement.attachInSql() // !admpub!  fix bug:Iterate func missing "... IN (...)"
	return statement.genSelectSQL(columnStr), append(append(statement.joinArgs, statement.Params...), statement.BeanArgs...)
}

func (s *Statement) genAddColumnStr(col *core.Column) (string, []interface{}) {
	quote := s.Engine.Quote
	sql := fmt.Sprintf("ALTER TABLE %v ADD %v;", quote(s.TableName()),
		col.String(s.Engine.dialect))
	return sql, []interface{}{}
}

/*func (s *Statement) genAddIndexStr(idxName string, cols []string) (string, []interface{}) {
	quote := s.Engine.Quote
	colstr := quote(strings.Join(cols, quote(", ")))
	sql := fmt.Sprintf("CREATE INDEX %v ON %v (%v);", quote(idxName), quote(s.TableName()), colstr)
	return sql, []interface{}{}
}

func (s *Statement) genAddUniqueStr(uqeName string, cols []string) (string, []interface{}) {
	quote := s.Engine.Quote
	colstr := quote(strings.Join(cols, quote(", ")))
	sql := fmt.Sprintf("CREATE UNIQUE INDEX %v ON %v (%v);", quote(uqeName), quote(s.TableName()), colstr)
	return sql, []interface{}{}
}*/

func (statement *Statement) genCountSql(bean interface{}) (string, []interface{}) {
	table := statement.Engine.TableInfo(bean)
	statement.RefTable = table

	if !statement.noAutoCondition {
		colNames, args := statement.buildConditions(
			table, bean, true, true, false, true, statement.needTableName())

		statement.ConditionStr = strings.Join(colNames, " "+statement.Engine.Dialect().AndStr()+" ")

		statement.BeanArgs = args
	}

	// count(index fieldname) > count(0) > count(*)
	var id = "*"
	if statement.Engine.Dialect().DBType() == "ql" {
		id = ""
	}
	statement.attachInSql()
	return statement.genSelectSQL(fmt.Sprintf("count(%v)", id)), append(append(statement.joinArgs, statement.Params...), statement.BeanArgs...)
}

func (statement *Statement) genSelectSQL(columnStr string) (a string) {
	var distinct string
	if statement.IsDistinct {
		distinct = "DISTINCT "
	}

	var dialect = statement.Engine.Dialect()
	var quote = statement.Engine.Quote
	var top string
	var mssqlCondi string

	statement.processIdParam()

	var buf bytes.Buffer
	if len(statement.WhereStr) > 0 {
		if len(statement.ConditionStr) > 0 {
			fmt.Fprintf(&buf, " WHERE (%v)", statement.WhereStr)
		} else {
			fmt.Fprintf(&buf, " WHERE %v", statement.WhereStr)
		}
		if statement.ConditionStr != "" {
			fmt.Fprintf(&buf, " %s (%v)", dialect.AndStr(), statement.ConditionStr)
		}
	} else if len(statement.ConditionStr) > 0 {
		fmt.Fprintf(&buf, " WHERE %v", statement.ConditionStr)
	}
	var whereStr = buf.String()

	var fromStr string = " FROM " + quote(statement.TableName())
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
			top = fmt.Sprintf(" TOP %d ", statement.LimitN)
		}
		if statement.Start > 0 {
			var column string = "(id)"
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
			}
			var orderStr string
			if len(statement.OrderStr) > 0 {
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

	// !nashtsai! REVIEW Sprintf is considered slowest mean of string concatnation, better to work with builder pattern
	a = fmt.Sprintf("SELECT %v%v%v%v%v", top, distinct, columnStr, fromStr, whereStr)
	if len(mssqlCondi) > 0 {
		if len(whereStr) > 0 {
			a += " AND " + mssqlCondi
		} else {
			a += " WHERE " + mssqlCondi
		}
	}

	if statement.GroupByStr != "" {
		a = fmt.Sprintf("%v GROUP BY %v", a, statement.GroupByStr)
	}
	if statement.HavingStr != "" {
		a = fmt.Sprintf("%v %v", a, statement.HavingStr)
	}
	if statement.OrderStr != "" {
		a = fmt.Sprintf("%v ORDER BY %v", a, statement.OrderStr)
	}
	if dialect.DBType() != core.MSSQL && dialect.DBType() != core.ORACLE {
		if statement.Start > 0 {
			a = fmt.Sprintf("%v LIMIT %v OFFSET %v", a, statement.LimitN, statement.Start)
		} else if statement.LimitN > 0 {
			a = fmt.Sprintf("%v LIMIT %v", a, statement.LimitN)
		}
	} else if dialect.DBType() == core.ORACLE {
		if statement.Start != 0 || statement.LimitN != 0 {
			a = fmt.Sprintf("SELECT %v FROM (SELECT %v,ROWNUM RN FROM (%v) at WHERE ROWNUM <= %d) aat WHERE RN > %d", columnStr, columnStr, a, statement.Start+statement.LimitN, statement.Start)
		}
	}
	if statement.IsForUpdate {
		a = dialect.ForUpdateSql(a)
	}

	return
}

func (statement *Statement) processIdParam() {
	if statement.IdParam != nil {
		if statement.Engine.dialect.DBType() != "ql" {
			for i, col := range statement.RefTable.PKColumns() {
				colName, _ := statement.colName(col, statement.aliasedTableName())
				if i < len(*(statement.IdParam)) {
					statement.And(fmt.Sprintf("%v %s ?", colName,
						statement.Engine.dialect.EqStr()), (*(statement.IdParam))[i])
				} else {
					statement.And(fmt.Sprintf("%v %s ?", colName,
						statement.Engine.dialect.EqStr()), "")
				}
			}
		} else {
			if len(*(statement.IdParam)) <= 1 {
				statement.And("id() == ?", (*(statement.IdParam))[0])
			}
		}
	}
}

func (statement *Statement) JoinColumns(cols []*core.Column, includeTableName bool) string {
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

func (statement *Statement) convertIdSql(sqlStr string) string {
	if statement.RefTable != nil {
		cols := statement.RefTable.PKColumns()
		if len(cols) == 0 {
			return ""
		}

		colstrs := statement.JoinColumns(cols, false)
		sqls := splitNNoCase(sqlStr, " from ", 2)
		if len(sqls) != 2 {
			return ""
		}
		if statement.Engine.dialect.DBType() == "ql" {
			return fmt.Sprintf("SELECT id() FROM %v", sqls[1])
		}
		return fmt.Sprintf("SELECT %s FROM %v", colstrs, sqls[1])
	}
	return ""
}

func (statement *Statement) convertUpdateSQL(sqlStr string) (string, string) {
	if statement.RefTable == nil || len(statement.RefTable.PrimaryKeys) != 1 {
		return "", ""
	}

	colstrs := statement.JoinColumns(statement.RefTable.PKColumns(), true)
	sqls := splitNNoCase(sqlStr, "where", 2)
	if len(sqls) != 2 {
		if len(sqls) == 1 {
			return sqls[0], fmt.Sprintf("SELECT %v FROM %v",
				colstrs, statement.Engine.Quote(statement.TableName()))
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
		colstrs, statement.Engine.Quote(statement.TableName()),
		whereStr)
}
