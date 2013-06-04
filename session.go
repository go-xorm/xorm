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
}

func (session *Session) Init() {
	session.Statement = Statement{Engine: session.Engine}
	session.IsAutoCommit = true
	session.IsCommitedOrRollbacked = false
}

func (session *Session) Close() {
	defer session.Db.Close()
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

func (session *Session) Limit(limit int, start ...int) *Session {
	session.Statement.Limit(limit, start...)
	return session
}

func (session *Session) OrderBy(order string) *Session {
	session.Statement.OrderBy(order)
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

func (session *Session) Begin() error {
	if session.IsAutoCommit {
		session.IsAutoCommit = false
		session.IsCommitedOrRollbacked = false
		tx, err := session.Db.Begin()
		session.Tx = tx
		if session.Engine.ShowSQL {
			fmt.Println("BEGIN TRANSACTION")
		}
		return err
	}
	return nil
}

func (session *Session) Rollback() error {
	if !session.IsAutoCommit && !session.IsCommitedOrRollbacked {
		if session.Engine.ShowSQL {
			fmt.Println("ROLL BACK")
		}
		session.IsCommitedOrRollbacked = true
		return session.Tx.Rollback()
	}
	return nil
}

func (session *Session) Commit() error {
	if !session.IsAutoCommit && !session.IsCommitedOrRollbacked {
		if session.Engine.ShowSQL {
			fmt.Println("COMMIT")
		}
		session.IsCommitedOrRollbacked = true
		return session.Tx.Commit()
	}
	return nil
}

func (session *Session) scanMapIntoStruct(obj interface{}, objMap map[string][]byte) error {
	dataStruct := reflect.Indirect(reflect.ValueOf(obj))
	if dataStruct.Kind() != reflect.Struct {
		return errors.New("expected a pointer to a struct")
	}

	table := session.Engine.Tables[Type(obj)]

	for key, data := range objMap {
		structField := dataStruct.FieldByName(table.Columns[key].FieldName)
		if !structField.CanSet() {
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
			if structField.Type().String() == "time.Time" {
				x, err := time.Parse("2006-01-02 15:04:05", string(data))
				if err != nil {
					x, err = time.Parse("2006-01-02 15:04:05.000 -0700", string(data))

					if err != nil {
						return errors.New("unsupported time format: " + string(data))
					}
				}

				v = x
			} else if session.Statement.UseCascade {
				session.Engine.AutoMapType(structField.Type())
				if _, ok := session.Engine.Tables[structField.Type()]; ok {
					x, err := strconv.ParseInt(string(data), 10, 64)
					if err != nil {
						return errors.New("arg " + key + " as int: " + err.Error())
					}

					if x != 0 {
						structInter := reflect.New(structField.Type())
						session.Statement.Init()
						err = session.Id(x).Get(structInter.Interface())
						if err != nil {
							return err
						}

						v = structInter.Elem().Interface()
					} else {
						//fmt.Println("zero value of struct type " + structField.Type().String())
						continue
					}

				} else {
					fmt.Println("unsupported struct type in Scan: " + structField.Type().String())
					continue
				}
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
	if session.Statement.RefTable != nil && session.Statement.RefTable.PrimaryKey != "" {
		sql = strings.Replace(sql, "(id)", session.Statement.RefTable.PrimaryKey, -1)
	}
	if session.Engine.ShowSQL {
		fmt.Println(sql)
		fmt.Println(args)
	}
	if session.IsAutoCommit {
		return session.innerExec(sql, args...)
	}
	return session.Tx.Exec(sql, args...)
}

func (session *Session) CreateTable(bean interface{}) error {
	statement := session.Statement
	defer statement.Init()
	statement.RefTable = session.Engine.AutoMap(bean)
	sql := statement.genCreateSQL()
	_, err := session.Exec(sql)
	return err
}

func (session *Session) Get(bean interface{}) error {
	statement := session.Statement
	defer statement.Init()
	statement.Limit(1)

	fmt.Println(bean)

	sql, args := statement.genGetSql(bean)
	resultsSlice, err := session.Query(sql, args...)

	if err != nil {
		return err
	}
	if len(resultsSlice) == 0 {
		return nil
	} else if len(resultsSlice) == 1 {
		results := resultsSlice[0]
		err := session.scanMapIntoStruct(bean, results)
		if err != nil {
			return err
		}
	} else {
		return errors.New("More than one record")
	}
	return nil
}

func (session *Session) Count(bean interface{}) (int64, error) {
	statement := session.Statement
	defer session.Statement.Init()
	sql, args := statement.genCountSql(bean)

	resultsSlice, err := session.Query(sql, args...)
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

func (session *Session) Find(rowsSlicePtr interface{}, condiBean ...interface{}) error {
	statement := session.Statement
	defer session.Statement.Init()
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice && sliceValue.Kind() != reflect.Map {
		return errors.New("needs a pointer to a slice or a map")
	}

	sliceElementType := sliceValue.Type().Elem()
	table := session.Engine.AutoMapType(sliceElementType)
	statement.RefTable = table

	if len(condiBean) > 0 {
		colNames, args := BuildConditions(session.Engine, table, condiBean[0])
		statement.ColumnStr = strings.Join(colNames, " and ")
		statement.BeanArgs = args
	}

	sql := statement.generateSql()
	resultsSlice, err := session.Query(sql, append(statement.Params, statement.BeanArgs...)...)

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

func (session *Session) Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	if session.Statement.RefTable != nil && session.Statement.RefTable.PrimaryKey != "" {
		sql = strings.Replace(sql, "(id)", session.Statement.RefTable.PrimaryKey, -1)
	}
	if session.Engine.ShowSQL {
		fmt.Println(sql)
		fmt.Println(paramStr)
	}
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
		session.Begin()
	}

	for _, bean := range beans {
		sliceValue := reflect.Indirect(reflect.ValueOf(bean))
		if sliceValue.Kind() == reflect.Slice {
			if session.Engine.InsertMany {
				lastId, err = session.InsertMulti(bean)
				if err != nil {
					if !isInTransaction {
						session.Rollback()
					}
					return lastId, err
				}
			} else {
				size := sliceValue.Len()
				for i := 0; i < size; i++ {
					lastId, err = session.InsertOne(sliceValue.Index(i).Interface())
					if err != nil {
						if !isInTransaction {
							session.Rollback()
						}
						return lastId, err
					}
				}
			}
		} else {
			lastId, err = session.InsertOne(bean)
			if err != nil {
				if !isInTransaction {
					session.Rollback()
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

func (session *Session) InsertMulti(rowsSlicePtr interface{}) (int64, error) {
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
	cols := make([]Column, 0)

	for i := 0; i < size; i++ {
		elemValue := sliceValue.Index(i).Interface()
		colPlaces := make([]string, 0)

		if i == 0 {
			for _, col := range table.Columns {
				fieldValue := reflect.Indirect(reflect.ValueOf(elemValue)).FieldByName(col.FieldName)
				val := fieldValue.Interface()
				if col.IsAutoIncrement && fieldValue.Int() == 0 {
					continue
				}
				if table, ok := session.Engine.Tables[fieldValue.Type()]; ok {
					pkField := reflect.Indirect(fieldValue).FieldByName(table.PKColumn().FieldName)
					fmt.Println(pkField.Interface())
					args = append(args, pkField.Interface())
				} else {
					args = append(args, val)
				}
				colNames = append(colNames, col.Name)
				cols = append(cols, col)
				colPlaces = append(colPlaces, "?")
			}
		} else {
			for _, col := range cols {
				fieldValue := reflect.Indirect(reflect.ValueOf(elemValue)).FieldByName(col.FieldName)
				val := fieldValue.Interface()
				if col.IsAutoIncrement && fieldValue.Int() == 0 {
					continue
				}
				if table, ok := session.Engine.Tables[fieldValue.Type()]; ok {
					pkField := reflect.Indirect(fieldValue).FieldByName(table.PKColumn().FieldName)
					args = append(args, pkField.Interface())
				} else {
					args = append(args, val)
				}
				colPlaces = append(colPlaces, "?")
			}
		}
		colMultiPlaces = append(colMultiPlaces, strings.Join(colPlaces, ", "))
	}

	statement := fmt.Sprintf("INSERT INTO %v%v%v (%v) VALUES (%v)",
		session.Engine.QuoteIdentifier,
		session.Statement.TableName(),
		session.Engine.QuoteIdentifier,
		strings.Join(colNames, ", "),
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

func (session *Session) InsertOne(bean interface{}) (int64, error) {
	table := session.Engine.AutoMap(bean)
	session.Statement.RefTable = table
	colNames := make([]string, 0)
	colPlaces := make([]string, 0)
	var args = make([]interface{}, 0)
	for _, col := range table.Columns {
		fieldValue := reflect.Indirect(reflect.ValueOf(bean)).FieldByName(col.FieldName)
		val := fieldValue.Interface()
		if col.IsAutoIncrement && fieldValue.Int() == 0 {
			continue
		}
		if table, ok := session.Engine.Tables[fieldValue.Type()]; ok {
			pkField := reflect.Indirect(fieldValue).FieldByName(table.PKColumn().FieldName)
			args = append(args, pkField.Interface())
		} else {
			args = append(args, val)
		}
		colNames = append(colNames, col.Name)
		colPlaces = append(colPlaces, "?")
	}

	sql := fmt.Sprintf("INSERT INTO %v%v%v (%v) VALUES (%v)",
		session.Engine.QuoteIdentifier,
		session.Statement.TableName(),
		session.Engine.QuoteIdentifier,
		strings.Join(colNames, ", "),
		strings.Join(colPlaces, ", "))

	res, err := session.Exec(sql, args...)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return id, nil
}

func (session *Session) Update(bean interface{}, condiBean ...interface{}) (int64, error) {
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

	statement := fmt.Sprintf("UPDATE %v%v%v SET %v %v",
		session.Engine.QuoteIdentifier,
		session.Statement.TableName(),
		session.Engine.QuoteIdentifier,
		strings.Join(colNames, ", "),
		condition)

	eargs := append(append(args, st.Params...), condiArgs...)
	res, err := session.Exec(statement, eargs...)
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
		session.Engine.QuoteIdentifier,
		session.Statement.TableName(),
		session.Engine.QuoteIdentifier,
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
