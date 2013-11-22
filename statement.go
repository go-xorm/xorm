package xorm

import (
	"fmt"
	"reflect"
	//"strconv"
	"encoding/json"
	"strings"
	"time"
)

// statement save all the sql info for executing SQL
type Statement struct {
	RefTable      *Table
	Engine        *Engine
	Start         int
	LimitN        int
	WhereStr      string
	Params        []interface{}
	OrderStr      string
	JoinStr       string
	GroupByStr    string
	HavingStr     string
	ColumnStr     string
	columnMap     map[string]bool
	OmitStr       string
	ConditionStr  string
	AltTableName  string
	RawSQL        string
	RawParams     []interface{}
	UseCascade    bool
	UseAutoJoin   bool
	StoreEngine   string
	Charset       string
	BeanArgs      []interface{}
	UseCache      bool
	UseAutoTime   bool
	IsDistinct    bool
	allUseBool    bool
	boolColumnMap map[string]bool
}

// init
func (statement *Statement) Init() {
	statement.RefTable = nil
	statement.Start = 0
	statement.LimitN = 0
	statement.WhereStr = ""
	statement.Params = make([]interface{}, 0)
	statement.OrderStr = ""
	statement.UseCascade = true
	statement.JoinStr = ""
	statement.GroupByStr = ""
	statement.HavingStr = ""
	statement.ColumnStr = ""
	statement.OmitStr = ""
	statement.columnMap = make(map[string]bool)
	statement.ConditionStr = ""
	statement.AltTableName = ""
	statement.RawSQL = ""
	statement.RawParams = make([]interface{}, 0)
	statement.BeanArgs = make([]interface{}, 0)
	statement.UseCache = statement.Engine.UseCache
	statement.UseAutoTime = true
	statement.IsDistinct = false
	statement.allUseBool = false
	statement.boolColumnMap = make(map[string]bool)
}

// add the raw sql statement
func (statement *Statement) Sql(querystring string, args ...interface{}) *Statement {
	statement.RawSQL = querystring
	statement.RawParams = args
	return statement
}

// add Where statment
func (statement *Statement) Where(querystring string, args ...interface{}) *Statement {
	statement.WhereStr = querystring
	statement.Params = args
	return statement
}

// add Where & and statment
func (statement *Statement) And(querystring string, args ...interface{}) *Statement {
	if statement.WhereStr != "" {
		statement.WhereStr = fmt.Sprintf("(%v) AND (%v)", statement.WhereStr, querystring)
	} else {
		statement.WhereStr = querystring
	}
	statement.Params = append(statement.Params, args...)
	return statement
}

// add Where & Or statment
func (statement *Statement) Or(querystring string, args ...interface{}) *Statement {
	if statement.WhereStr != "" {
		statement.WhereStr = fmt.Sprintf("(%v) OR (%v)", statement.WhereStr, querystring)
	} else {
		statement.WhereStr = querystring
	}
	statement.Params = append(statement.Params, args...)
	return statement
}

// tempororily set table name
func (statement *Statement) Table(tableNameOrBean interface{}) *Statement {
	t := rType(tableNameOrBean)
	if t.Kind() == reflect.String {
		statement.AltTableName = tableNameOrBean.(string)
	} else if t.Kind() == reflect.Struct {
		statement.RefTable = statement.Engine.autoMapType(t)
	}
	return statement
}

// Auto generating conditions according a struct
func buildConditions(engine *Engine, table *Table, bean interface{}, includeVersion bool, allUseBool bool, boolColumnMap map[string]bool) ([]string, []interface{}) {
	colNames := make([]string, 0)
	var args = make([]interface{}, 0)
	for _, col := range table.Columns {
		if !includeVersion && col.IsVersion {
			continue
		}
		fieldValue := col.ValueOf(bean)
		fieldType := reflect.TypeOf(fieldValue.Interface())
		var val interface{}
		switch fieldType.Kind() {
		case reflect.Bool:
			if allUseBool {
				val = fieldValue.Interface()
			} else if _, ok := boolColumnMap[col.Name]; ok {
				val = fieldValue.Interface()
			} else {
				// if a bool in a struct, it will not be as a condition because it default is false,
				// please use Where() instead
				continue
			}
		case reflect.String:
			if fieldValue.String() == "" {
				continue
			}
			// for MyString, should convert to string or panic
			if fieldType.String() != reflect.String.String() {
				val = fieldValue.String()
			} else {
				val = fieldValue.Interface()
			}
		case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
			if fieldValue.Int() == 0 {
				continue
			}
			val = fieldValue.Interface()
		case reflect.Float32, reflect.Float64:
			if fieldValue.Float() == 0.0 {
				continue
			}
			val = fieldValue.Interface()
		case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
			if fieldValue.Uint() == 0 {
				continue
			}
			val = fieldValue.Interface()
		case reflect.Struct:
			if fieldType == reflect.TypeOf(time.Now()) {
				t := fieldValue.Interface().(time.Time)
				if t.IsZero() {
					continue
				}
				var str string
				if col.SQLType.Name == Time {
					s := t.UTC().Format("2006-01-02 15:04:05")
					val = s[11:19]
				} else if col.SQLType.Name == Date {
					str = t.Format("2006-01-02")
					val = str
				} else {
					val = t
				}
			} else {
				engine.autoMapType(fieldValue.Type())
				if table, ok := engine.Tables[fieldValue.Type()]; ok {
					pkField := reflect.Indirect(fieldValue).FieldByName(table.PKColumn().FieldName)
					if pkField.Int() != 0 {
						val = pkField.Interface()
					} else {
						continue
					}
				} else {
					val = fieldValue.Interface()
				}
			}
		case reflect.Array, reflect.Slice, reflect.Map:
			if fieldValue == reflect.Zero(fieldType) {
				continue
			}
			if fieldValue.IsNil() || !fieldValue.IsValid() {
				continue
			}

			if col.SQLType.IsText() {
				bytes, err := json.Marshal(fieldValue.Interface())
				if err != nil {
					engine.LogSQL(err)
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
						engine.LogSQL(err)
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
		colNames = append(colNames, fmt.Sprintf("%v = ?", engine.Quote(col.Name)))
	}

	return colNames, args
}

// return current tableName
func (statement *Statement) TableName() string {
	if statement.AltTableName != "" {
		return statement.AltTableName
	}

	if statement.RefTable != nil {
		return statement.RefTable.Name
	}
	return ""
}

// Generate "Where id = ? " statment
func (statement *Statement) Id(id int64) *Statement {
	if statement.WhereStr == "" {
		statement.WhereStr = "(id)=?"
		statement.Params = []interface{}{id}
	} else {
		statement.WhereStr = statement.WhereStr + " AND (id)=?"
		statement.Params = append(statement.Params, id)
	}
	return statement
}

// Generate "Where column IN (?) " statment
func (statement *Statement) In(column string, args ...interface{}) *Statement {
	inStr := fmt.Sprintf("%v IN (%v)", column, strings.Join(makeArray("?", len(args)), ","))
	if statement.WhereStr == "" {
		statement.WhereStr = inStr
		statement.Params = args
	} else {
		statement.WhereStr = statement.WhereStr + " AND " + inStr
		statement.Params = append(statement.Params, args...)
	}
	return statement
}

func col2NewCols(columns ...string) []string {
	newColumns := make([]string, 0)
	for _, col := range columns {
		strings.Replace(col, "`", "", -1)
		strings.Replace(col, `"`, "", -1)
		ccols := strings.Split(col, ",")
		for _, c := range ccols {
			newColumns = append(newColumns, strings.TrimSpace(c))
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

// Generate "col1, col2" statement
func (statement *Statement) Cols(columns ...string) *Statement {
	newColumns := col2NewCols(columns...)
	for _, nc := range newColumns {
		statement.columnMap[nc] = true
	}
	statement.ColumnStr = statement.Engine.Quote(strings.Join(newColumns, statement.Engine.Quote(", ")))
	return statement
}

// indicates that use bool fields as update contents and query contiditions
func (statement *Statement) UseBool(columns ...string) *Statement {
	if len(columns) > 0 {
		newColumns := col2NewCols(columns...)
		for _, nc := range newColumns {
			statement.boolColumnMap[nc] = true
		}
	} else {
		statement.allUseBool = true
	}
	return statement
}

// do not use the columns
func (statement *Statement) Omit(columns ...string) {
	newColumns := col2NewCols(columns...)
	for _, nc := range newColumns {
		statement.columnMap[nc] = false
	}
	statement.OmitStr = statement.Engine.Quote(strings.Join(newColumns, statement.Engine.Quote(", ")))
}

// Generate LIMIT limit statement
func (statement *Statement) Top(limit int) *Statement {
	statement.Limit(limit)
	return statement
}

// Generate LIMIT start, limit statement
func (statement *Statement) Limit(limit int, start ...int) *Statement {
	statement.LimitN = limit
	if len(start) > 0 {
		statement.Start = start[0]
	}
	return statement
}

// Generate "Order By order" statement
func (statement *Statement) OrderBy(order string) *Statement {
	statement.OrderStr = order
	return statement
}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (statement *Statement) Join(join_operator, tablename, condition string) *Statement {
	if statement.JoinStr != "" {
		statement.JoinStr = statement.JoinStr + fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	} else {
		statement.JoinStr = fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	}
	return statement
}

// Generate "Group By keys" statement
func (statement *Statement) GroupBy(keys string) *Statement {
	statement.GroupByStr = keys
	return statement
}

// Generate "Having conditions" statement
func (statement *Statement) Having(conditions string) *Statement {
	statement.HavingStr = fmt.Sprintf("HAVING %v", conditions)
	return statement
}

func (statement *Statement) genColumnStr() string {
	table := statement.RefTable
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		if statement.OmitStr != "" {
			if _, ok := statement.columnMap[col.Name]; ok {
				continue
			}
		}
		if col.MapType == ONLYTODB {
			continue
		}
		colNames = append(colNames, statement.Engine.Quote(statement.TableName())+"."+statement.Engine.Quote(col.Name))
	}
	return strings.Join(colNames, ", ")
}

func (statement *Statement) genCreateSQL() string {
	sql := "CREATE TABLE IF NOT EXISTS " + statement.Engine.Quote(statement.TableName()) + " ("
	for _, colName := range statement.RefTable.ColumnsSeq {
		col := statement.RefTable.Columns[colName]
		sql += col.String(statement.Engine.dialect)
		sql = strings.TrimSpace(sql)
		sql += ", "
	}
	sql = sql[:len(sql)-2] + ")"
	if statement.Engine.dialect.SupportEngine() && statement.StoreEngine != "" {
		sql += " ENGINE=" + statement.StoreEngine
	}
	if statement.Engine.dialect.SupportCharset() && statement.Charset != "" {
		sql += " DEFAULT CHARSET " + statement.Charset
	}
	sql += ";"
	return sql
}

func indexName(tableName, idxName string) string {
	return fmt.Sprintf("IDX_%v_%v", tableName, idxName)
}

func (s *Statement) genIndexSQL() []string {
	var sqls []string = make([]string, 0)
	tbName := s.TableName()
	quote := s.Engine.Quote
	for idxName, index := range s.RefTable.Indexes {
		if index.Type == IndexType {
			sql := fmt.Sprintf("CREATE INDEX %v ON %v (%v);", quote(indexName(tbName, idxName)),
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
	tbName := s.TableName()
	quote := s.Engine.Quote
	for idxName, unique := range s.RefTable.Indexes {
		if unique.Type == UniqueType {
			sql := fmt.Sprintf("CREATE UNIQUE INDEX %v ON %v (%v);", quote(uniqueName(tbName, idxName)),
				quote(tbName), quote(strings.Join(unique.Cols, quote(","))))
			sqls = append(sqls, sql)
		}
	}
	return sqls
}

func (s *Statement) genDelIndexSQL() []string {
	var sqls []string = make([]string, 0)
	for idxName, index := range s.RefTable.Indexes {
		var rIdxName string
		if index.Type == UniqueType {
			rIdxName = uniqueName(s.TableName(), idxName)
		} else if index.Type == IndexType {
			rIdxName = indexName(s.TableName(), idxName)
		}
		sql := fmt.Sprintf("DROP INDEX %v", s.Engine.Quote(rIdxName))
		if s.Engine.dialect.IndexOnTable() {
			sql += fmt.Sprintf(" ON %v", s.Engine.Quote(s.TableName()))
		}
		sqls = append(sqls, sql)
	}
	return sqls
}

func (s *Statement) genDropSQL() string {
	sql := "DROP TABLE IF EXISTS " + s.Engine.Quote(s.TableName()) + ";"
	return sql
}

func (statement Statement) genGetSql(bean interface{}) (string, []interface{}) {
	table := statement.Engine.autoMap(bean)
	statement.RefTable = table

	colNames, args := buildConditions(statement.Engine, table, bean, true,
		statement.allUseBool, statement.boolColumnMap)
	statement.ConditionStr = strings.Join(colNames, " AND ")
	statement.BeanArgs = args

	var columnStr string = statement.ColumnStr
	if columnStr == "" {
		columnStr = statement.genColumnStr()
	}

	return statement.genSelectSql(columnStr), append(statement.Params, statement.BeanArgs...)
}

func (s *Statement) genAddColumnStr(col *Column) (string, []interface{}) {
	quote := s.Engine.Quote
	sql := fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v;", quote(s.TableName()),
		col.String(s.Engine.dialect))
	return sql, []interface{}{}
}

func (s *Statement) genAddIndexStr(idxName string, cols []string) (string, []interface{}) {
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
}

func (statement Statement) genCountSql(bean interface{}) (string, []interface{}) {
	table := statement.Engine.autoMap(bean)
	statement.RefTable = table

	colNames, args := buildConditions(statement.Engine, table, bean, true, statement.allUseBool, statement.boolColumnMap)
	statement.ConditionStr = strings.Join(colNames, " AND ")
	statement.BeanArgs = args
	var id string = "*"
	if table.PrimaryKey != "" {
		id = statement.Engine.Quote(table.PrimaryKey)
	}
	return statement.genSelectSql(fmt.Sprintf("COUNT(%v) AS %v", id, statement.Engine.Quote("total"))), append(statement.Params, statement.BeanArgs...)
}

func (statement Statement) genSelectSql(columnStr string) (a string) {
	if statement.GroupByStr != "" {
		columnStr = statement.Engine.Quote(strings.Replace(statement.GroupByStr, ",", statement.Engine.Quote(","), -1))
		statement.GroupByStr = columnStr
	}
	var distinct string
	if statement.IsDistinct {
		distinct = "DISTINCT "
	}
	a = fmt.Sprintf("SELECT %v%v FROM %v", distinct, columnStr,
		statement.Engine.Quote(statement.TableName()))
	if statement.JoinStr != "" {
		a = fmt.Sprintf("%v %v", a, statement.JoinStr)
	}
	if statement.WhereStr != "" {
		a = fmt.Sprintf("%v WHERE %v", a, statement.WhereStr)
		if statement.ConditionStr != "" {
			a = fmt.Sprintf("%v and %v", a, statement.ConditionStr)
		}
	} else if statement.ConditionStr != "" {
		a = fmt.Sprintf("%v WHERE %v", a, statement.ConditionStr)
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
	if statement.Start > 0 {
		a = fmt.Sprintf("%v LIMIT %v OFFSET %v", a, statement.LimitN, statement.Start)
	} else if statement.LimitN > 0 {
		a = fmt.Sprintf("%v LIMIT %v", a, statement.LimitN)
	}
	return
}
