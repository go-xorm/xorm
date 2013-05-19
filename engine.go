package xorm

import (
	"database/sql"
	//"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	PQSQL   = "pqsql"
	MSSQL   = "mssql"
	SQLITE  = "sqlite3"
	MYSQL   = "mysql"
	MYMYSQL = "mymysql"
)

type Engine struct {
	Mapper          IMapper
	TagIdentifier   string
	DriverName      string
	DataSourceName  string
	Tables          map[reflect.Type]Table
	AutoIncrement   string
	ShowSQL         bool
	InsertMany      bool
	QuoteIdentifier string
	Statement       Statement
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

func (engine *Engine) MakeSession() (session Session, err error) {
	db, err := engine.OpenDB()
	if err != nil {
		return Session{}, err
	}

	session = Session{Engine: engine, Db: db}
	session.Init()
	return
}

func (engine *Engine) Where(querystring string, args ...interface{}) *Engine {
	engine.Statement.Where(querystring, args...)
	return engine
}

func (engine *Engine) Id(id int64) *Engine {
	engine.Statement.Id(id)
	return engine
}

func (engine *Engine) In(column string, args ...interface{}) *Engine {
	engine.Statement.In(column, args...)
	return engine
}

func (engine *Engine) Table(tableName string) *Engine {
	engine.Statement.Table(tableName)
	return engine
}

func (engine *Engine) Limit(limit int, start ...int) *Engine {
	engine.Statement.Limit(limit, start...)
	return engine
}

func (engine *Engine) OrderBy(order string) *Engine {
	engine.Statement.OrderBy(order)
	return engine
}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (engine *Engine) Join(join_operator, tablename, condition string) *Engine {
	engine.Statement.Join(join_operator, tablename, condition)
	return engine
}

func (engine *Engine) GroupBy(keys string) *Engine {
	engine.Statement.GroupBy(keys)
	return engine
}

func (engine *Engine) Having(conditions string) *Engine {
	engine.Statement.Having(conditions)
	return engine
}

func (engine *Engine) AutoMapType(t reflect.Type) *Table {
	table, ok := engine.Tables[t]
	if !ok {
		table = engine.MapType(t)
		engine.Tables[t] = table
	}
	return &table
}

func (engine *Engine) AutoMap(bean interface{}) *Table {
	t := Type(bean)
	return engine.AutoMapType(t)
}

func (engine *Engine) MapType(t reflect.Type) Table {
	table := Table{Name: engine.Mapper.Obj2Table(t.Name()), Type: t}
	table.Columns = make(map[string]Column)
	var pkCol *Column = nil
	var pkstr = ""

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		ormTagStr := tag.Get(engine.TagIdentifier)
		var col Column
		fieldType := t.Field(i).Type

		if ormTagStr != "" {
			col = Column{FieldName: t.Field(i).Name, Nullable: true}
			ormTagStr = strings.ToLower(ormTagStr)
			tags := strings.Split(ormTagStr, " ")
			// TODO: 
			if len(tags) > 0 {
				if tags[0] == "-" {
					continue
				}
				for j, key := range tags {
					k := strings.ToLower(key)
					switch {
					case k == "pk":
						col.IsPrimaryKey = true
						col.Nullable = false
						pkCol = &col
					case k == "null":
						col.Nullable = (tags[j-1] != "not")
					case k == "autoincr":
						col.IsAutoIncrement = true
					case k == "default":
						col.Default = tags[j+1]
					case k == "text":
						col.SQLType = Text
					case strings.HasPrefix(k, "int"):
						col.SQLType = Int
						lens := k[len("int")+1 : len(k)-1]
						col.Length, _ = strconv.Atoi(lens)
					case strings.HasPrefix(k, "varchar"):
						col.SQLType = Varchar
						lens := k[len("decimal")+1 : len(k)-1]
						col.Length, _ = strconv.Atoi(lens)
					case strings.HasPrefix(k, "decimal"):
						col.SQLType = Decimal
						lens := k[len("decimal")+1 : len(k)-1]
						twolen := strings.Split(lens, ",")
						col.Length, _ = strconv.Atoi(twolen[0])
						col.Length2, _ = strconv.Atoi(twolen[1])
					case k == "date":
						col.SQLType = Date
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
			}
		}

		if col.Name == "" {
			sqlType := Type2SQLType(fieldType)
			col = Column{engine.Mapper.Obj2Table(t.Field(i).Name), t.Field(i).Name, sqlType,
				sqlType.DefaultLength, sqlType.DefaultLength2, true, "", false, false, false}
		}
		table.Columns[col.Name] = col
		if strings.ToLower(t.Field(i).Name) == "id" {
			pkstr = col.Name
		}
	}

	if pkCol == nil {
		if pkstr != "" {
			col := table.Columns[pkstr]
			col.IsPrimaryKey = true
			col.IsAutoIncrement = true
			col.Nullable = false
			col.Length = Int.DefaultLength
			table.PrimaryKey = col.Name
		}
	} else {
		table.PrimaryKey = pkCol.Name
	}

	return table
}

func (engine *Engine) Map(beans ...interface{}) (e error) {
	for _, bean := range beans {
		t := Type(bean)
		if _, ok := engine.Tables[t]; !ok {
			engine.Tables[t] = engine.MapType(t)
		}
	}
	return
}

func (engine *Engine) UnMap(beans ...interface{}) (e error) {
	for _, bean := range beans {
		t := Type(bean)
		if _, ok := engine.Tables[t]; ok {
			delete(engine.Tables, t)
		}
	}
	return
}

func (e *Engine) DropAll() error {
	session, err := e.MakeSession()
	session.Begin()
	defer session.Close()
	if err != nil {
		return err
	}

	for _, table := range e.Tables {
		e.Statement.RefTable = &table
		sql := e.Statement.genDropSQL()
		_, err = session.Exec(sql)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return session.Commit()
}

func (e *Engine) CreateTables(beans ...interface{}) error {
	session, err := e.MakeSession()
	session.Begin()
	session.Statement = e.Statement
	defer session.Close()
	defer e.Statement.Init()
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
	session, err := e.MakeSession()
	session.Begin()
	defer session.Close()
	if err != nil {
		return err
	}

	for _, table := range e.Tables {
		e.Statement.RefTable = &table
		sql := e.Statement.genCreateSQL()
		_, err = session.Exec(sql)
		if err != nil {
			session.Rollback()
			break
		}
	}
	session.Commit()
	return err
}

func (engine *Engine) Exec(sql string, args ...interface{}) (sql.Result, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return nil, err
	}
	return session.Exec(sql, args...)
}

func (engine *Engine) Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return nil, err
	}
	return session.Query(sql, paramStr...)
}

func (engine *Engine) Insert(beans ...interface{}) (int64, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return -1, err
	}
	defer engine.Statement.Init()
	session.Statement = engine.Statement
	return session.Insert(beans...)
}

func (engine *Engine) Update(bean interface{}, condiBeans ...interface{}) (int64, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return -1, err
	}
	defer engine.Statement.Init()
	session.Statement = engine.Statement
	return session.Update(bean, condiBeans...)
}

func (engine *Engine) Delete(bean interface{}) (int64, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return -1, err
	}
	defer engine.Statement.Init()
	session.Statement = engine.Statement
	return session.Delete(bean)
}

func (engine *Engine) Get(bean interface{}) error {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return err
	}
	defer engine.Statement.Init()
	session.Statement = engine.Statement
	return session.Get(bean)
}

func (engine *Engine) Find(beans interface{}, condiBeans ...interface{}) error {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return err
	}
	defer engine.Statement.Init()
	session.Statement = engine.Statement
	return session.Find(beans, condiBeans...)
}

func (engine *Engine) Count(bean interface{}) (int64, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return 0, err
	}
	defer engine.Statement.Init()
	session.Statement = engine.Statement
	return session.Count(bean)
}
