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

func getTypeName(obj interface{}) (typestr string) {
	typ := reflect.TypeOf(obj)
	typestr = typ.String()

	lastDotIndex := strings.LastIndex(typestr, ".")
	if lastDotIndex != -1 {
		typestr = typestr[lastDotIndex+1:]
	}

	return
}

func StructName(s interface{}) string {
	v := reflect.TypeOf(s)
	return Type2StructName(v)
}

func Type2StructName(v reflect.Type) string {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Name()
}

type Session struct {
	Db              *sql.DB
	Engine          *Engine
	Statements      []Statement
	Mapper          IMapper
	AutoCommit      bool
	ParamIteration  int
	CurStatementIdx int
}

func (session *Session) Init() {
	session.Statements = make([]Statement, 0)
	/*session.Statement.TableName = ""
	session.Statement.LimitStr = 0
	session.Statement.OffsetStr = 0
	session.Statement.WhereStr = ""
	session.Statement.ParamStr = make([]interface{}, 0)
	session.Statement.OrderStr = ""
	session.Statement.JoinStr = ""
	session.Statement.GroupByStr = ""
	session.Statement.HavingStr = ""*/
	session.CurStatementIdx = -1

	session.ParamIteration = 1
}

func (session *Session) Close() {
	defer session.Db.Close()
}

func (session *Session) CurrentStatement() *Statement {
	if session.CurStatementIdx > -1 {
		return &session.Statements[session.CurStatementIdx]
	}
	return nil
}

func (session *Session) AutoStatement() *Statement {
	if session.CurStatementIdx == -1 {
		session.newStatement()
	}
	return session.CurrentStatement()
}

//Execute sql
func (session *Session) Exec(finalQueryString string, args ...interface{}) (sql.Result, error) {
	rs, err := session.Db.Prepare(finalQueryString)
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

func (session *Session) Where(querystring interface{}, args ...interface{}) *Session {
	statement := session.AutoStatement()
	switch querystring := querystring.(type) {
	case string:
		statement.WhereStr = querystring
	case int:
		if session.Engine.Protocol == "pgsql" {
			statement.WhereStr = fmt.Sprintf("%v%v%v = $%v", session.Engine.QuoteIdentifier, statement.Table.PKColumn().Name, session.Engine.QuoteIdentifier, session.ParamIteration)
		} else {
			statement.WhereStr = fmt.Sprintf("%v%v%v = ?", session.Engine.QuoteIdentifier, statement.Table.PKColumn().Name, session.Engine.QuoteIdentifier)
			session.ParamIteration++
		}
		args = append(args, querystring)
	}
	statement.ParamStr = args
	return session
}

func (session *Session) Limit(start int, size ...int) *Session {
	session.AutoStatement().LimitStr = start
	if len(size) > 0 {
		session.CurrentStatement().OffsetStr = size[0]
	}
	return session
}

func (session *Session) Offset(offset int) *Session {
	session.AutoStatement().OffsetStr = offset
	return session
}

func (session *Session) OrderBy(order string) *Session {
	session.AutoStatement().OrderStr = order
	return session
}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (session *Session) Join(join_operator, tablename, condition string) *Session {
	if session.AutoStatement().JoinStr != "" {
		session.CurrentStatement().JoinStr = session.CurrentStatement().JoinStr + fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	} else {
		session.CurrentStatement().JoinStr = fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	}

	return session
}

func (session *Session) GroupBy(keys string) *Session {
	session.AutoStatement().GroupByStr = fmt.Sprintf("GROUP BY %v", keys)
	return session
}

func (session *Session) Having(conditions string) *Session {
	session.AutoStatement().HavingStr = fmt.Sprintf("HAVING %v", conditions)
	return session
}

func (session *Session) Begin() {

}

func (session *Session) Rollback() {

}

func (session *Session) Commit() {
	for _, statement := range session.Statements {
		sql := statement.generateSql()
		session.Exec(sql)
	}
}

func (session *Session) TableName(bean interface{}) string {
	return session.Mapper.Obj2Table(StructName(bean))
}

func (session *Session) newStatement() {
	state := Statement{}
	state.Session = session
	session.Statements = append(session.Statements, state)
	session.CurStatementIdx = len(session.Statements) - 1
}

func (session *Session) scanMapIntoStruct(obj interface{}, objMap map[string][]byte) error {
	dataStruct := reflect.Indirect(reflect.ValueOf(obj))
	if dataStruct.Kind() != reflect.Struct {
		return errors.New("expected a pointer to a struct")
	}

	tablName := session.TableName(obj)
	table := session.Engine.Tables[tablName]

	for key, data := range objMap {
		structField := dataStruct.FieldByName(table.Columns[key].FieldName)
		if !structField.CanSet() {
			continue
		}

		var v interface{}

		switch structField.Type().Kind() {
		case reflect.Slice:
			v = data
		case reflect.String:
			v = string(data)
		case reflect.Bool:
			v = string(data) == "1"
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
			if structField.Type().String() != "time.Time" {
				return errors.New("unsupported struct type in Scan: " + structField.Type().String())
			}

			x, err := time.Parse("2006-01-02 15:04:05", string(data))
			if err != nil {
				x, err = time.Parse("2006-01-02 15:04:05.000 -0700", string(data))

				if err != nil {
					return errors.New("unsupported time format: " + string(data))
				}
			}

			v = x
		default:
			return errors.New("unsupported type in Scan: " + reflect.TypeOf(v).String())
		}

		structField.Set(reflect.ValueOf(v))
	}

	return nil
}

func (session *Session) Get(output interface{}) error {
	statement := session.AutoStatement()
	session.Limit(1)
	tableName := session.TableName(output)
	table := session.Engine.Tables[tableName]
	statement.Table = &table

	resultsSlice, err := session.FindMap(statement)
	if err != nil {
		return err
	}
	if len(resultsSlice) == 0 {
		return nil
	} else if len(resultsSlice) == 1 {
		results := resultsSlice[0]
		err := session.scanMapIntoStruct(output, results)
		if err != nil {
			return err
		}
	} else {
		return errors.New("More than one record")
	}
	return nil
}

func (session *Session) Count(bean interface{}) (int64, error) {
	statement := session.AutoStatement()
	session.Limit(1)
	tableName := session.TableName(bean)
	table := session.Engine.Tables[tableName]
	statement.Table = &table

	resultsSlice, err := session.SQL2Map(statement.genCountSql(), statement.ParamStr)
	if err != nil {
		return 0, err
	}

	var total int64 = 0
	for _, results := range resultsSlice {
		total, err = strconv.ParseInt(string(results["total"]), 10, 64)
		break
	}

	return int64(total), err
}

func (session *Session) Find(rowsSlicePtr interface{}) error {
	statement := session.AutoStatement()

	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return errors.New("needs a pointer to a slice")
	}

	sliceElementType := sliceValue.Type().Elem()

	tableName := session.Mapper.Obj2Table(Type2StructName(sliceElementType))
	table := session.Engine.Tables[tableName]
	statement.Table = &table

	resultsSlice, err := session.FindMap(statement)
	if err != nil {
		return err
	}

	for _, results := range resultsSlice {
		newValue := reflect.New(sliceElementType)
		err := session.scanMapIntoStruct(newValue.Interface(), results)
		if err != nil {
			return err
		}
		sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
	}
	return nil
}

func (session *Session) SQL2Map(sqls string, paramStr []interface{}) (resultsSlice []map[string][]byte, err error) {
	if session.Engine.ShowSQL {
		fmt.Println(sqls)
	}
	s, err := session.Db.Prepare(sqls)
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
				str = rawValue.Interface().(time.Time).Format("2006-01-02 15:04:05.000 -0700")
				result[key] = []byte(str)
			}

		}
		resultsSlice = append(resultsSlice, result)
	}
	return resultsSlice, nil
}

func (session *Session) FindMap(statement *Statement) (resultsSlice []map[string][]byte, err error) {
	sqls := statement.generateSql()
	return session.SQL2Map(sqls, statement.ParamStr)
}

func (session *Session) Insert(beans ...interface{}) (int64, error) {
	var lastId int64 = -1
	for _, bean := range beans {
		lastId, err := session.InsertOne(bean)
		if err != nil {
			return lastId, err
		}
	}
	return lastId, nil
}

func (session *Session) InsertOne(bean interface{}) (int64, error) {
	tableName := session.TableName(bean)
	table := session.Engine.Tables[tableName]

	colNames := make([]string, 0)
	colPlaces := make([]string, 0)
	var args = make([]interface{}, 0)
	for _, col := range table.Columns {
		fieldValue := reflect.Indirect(reflect.ValueOf(bean)).FieldByName(col.FieldName)
		val := fieldValue.Interface()
		if col.AutoIncrement {
			if fieldValue.Int() == 0 {
				continue
			}
		}
		args = append(args, val)
		colNames = append(colNames, col.Name)
		colPlaces = append(colPlaces, "?")
	}

	statement := fmt.Sprintf("INSERT INTO %v%v%v (%v) VALUES (%v)",
		session.Engine.QuoteIdentifier,
		tableName,
		session.Engine.QuoteIdentifier,
		strings.Join(colNames, ", "),
		strings.Join(colPlaces, ", "))

	if session.Engine.ShowSQL {
		fmt.Println(statement)
	}

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

func (session *Session) Update(bean interface{}) (int64, error) {
	tableName := session.TableName(bean)
	table := session.Engine.Tables[tableName]

	colNames := make([]string, 0)
	var args = make([]interface{}, 0)
	for _, col := range table.Columns {
		if col.Name == "" {
			continue
		}
		fieldValue := reflect.Indirect(reflect.ValueOf(bean)).FieldByName(col.FieldName)
		fieldType := reflect.TypeOf(fieldValue.Interface())
		val := fieldValue.Interface()
		switch fieldType.Kind() {
		case reflect.String:
			if fieldValue.String() == "" {
				continue
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			if fieldValue.Int() == 0 {
				continue
			}
		case reflect.Struct:
			if fieldType == reflect.TypeOf(time.Now()) {
				t := fieldValue.Interface().(time.Time)
				if t.IsZero() {
					continue
				}
			}
		default:
			continue
		}
		args = append(args, val)
		colNames = append(colNames, session.Engine.QuoteIdentifier+col.Name+session.Engine.QuoteIdentifier+"=?")
	}

	var condition = ""
	st := session.CurrentStatement()
	if st != nil && st.WhereStr != "" {
		condition = fmt.Sprintf("WHERE %v", st.WhereStr)
	}

	if condition == "" {
		fieldValue := reflect.Indirect(reflect.ValueOf(bean)).FieldByName(table.PKColumn().FieldName)
		if fieldValue.Int() != 0 {
			condition = fmt.Sprintf("WHERE %v = ?", table.PKColumn().Name)
			args = append(args, fieldValue.Interface())
		}
	}

	statement := fmt.Sprintf("UPDATE %v%v%v SET %v %v",
		session.Engine.QuoteIdentifier,
		tableName,
		session.Engine.QuoteIdentifier,
		strings.Join(colNames, ", "),
		condition)

	if session.Engine.ShowSQL {
		fmt.Println(statement)
	}

	res, err := session.Exec(statement, args...)
	if err != nil {
		return -1, err
	}

	id, err := res.RowsAffected()

	if err != nil {
		return -1, err
	}
	return id, nil
}

func (session *Session) Delete(bean interface{}) (int64, error) {
	tableName := session.TableName(bean)
	table := session.Engine.Tables[tableName]

	colNames := make([]string, 0)
	var args = make([]interface{}, 0)
	for _, col := range table.Columns {
		if col.Name == "" {
			continue
		}
		fieldValue := reflect.Indirect(reflect.ValueOf(bean)).FieldByName(col.FieldName)
		fieldType := reflect.TypeOf(fieldValue.Interface())
		val := fieldValue.Interface()
		switch fieldType.Kind() {
		case reflect.String:
			if fieldValue.String() == "" {
				continue
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			if fieldValue.Int() == 0 {
				continue
			}
		case reflect.Struct:
			if fieldType == reflect.TypeOf(time.Now()) {
				t := fieldValue.Interface().(time.Time)
				if t.IsZero() {
					continue
				}
			}
		default:
			continue
		}
		args = append(args, val)
		colNames = append(colNames, session.Engine.QuoteIdentifier+col.Name+session.Engine.QuoteIdentifier+"=?")
	}

	var condition = ""
	st := session.CurrentStatement()
	if st != nil && st.WhereStr != "" {
		condition = fmt.Sprintf("WHERE %v", st.WhereStr)
		if len(colNames) > 0 {
			condition += " and "
			condition += strings.Join(colNames, " and ")
		}
	} else {
		condition = "WHERE " + strings.Join(colNames, " and ")
	}

	statement := fmt.Sprintf("DELETE FROM %v%v%v %v",
		session.Engine.QuoteIdentifier,
		tableName,
		session.Engine.QuoteIdentifier,
		condition)

	if session.Engine.ShowSQL {
		fmt.Println(statement)
	}

	res, err := session.Exec(statement, args...)
	if err != nil {
		return -1, err
	}

	id, err := res.RowsAffected()

	if err != nil {
		return -1, err
	}
	return id, nil
}
