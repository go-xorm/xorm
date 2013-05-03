package xorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type SQLType struct {
	Name          string
	DefaultLength int
}

var (
	Int     = SQLType{"int", 11}
	Char    = SQLType{"char", 1}
	Varchar = SQLType{"varchar", 50}
	Date    = SQLType{"date", 24}
	Decimal = SQLType{"decimal", 26}
	Float   = SQLType{"float", 31}
	Double  = SQLType{"double", 31}
)

const (
	PQSQL   = "pqsql"
	MSSQL   = "mssql"
	SQLITE  = "sqlite"
	MYSQL   = "mysql"
	MYMYSQL = "mymysql"
)

type Column struct {
	Name          string
	FieldName     string
	SQLType       SQLType
	Length        int
	Nullable      bool
	Default       string
	IsUnique      bool
	IsPrimaryKey  bool
	AutoIncrement bool
}

type Table struct {
	Name       string
	Type       reflect.Type
	Columns    map[string]Column
	PrimaryKey string
}

func (table *Table) ColumnStr() string {
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		if col.Name == "" {
			continue
		}
		colNames = append(colNames, col.Name)
	}
	return strings.Join(colNames, ", ")
}

func (table *Table) PlaceHolders() string {
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		if col.Name == "" {
			continue
		}
		colNames = append(colNames, "?")
	}
	return strings.Join(colNames, ", ")
}

func (table *Table) PKColumn() Column {
	return table.Columns[table.PrimaryKey]
}

type Engine struct {
	Mapper          IMapper
	Protocol        string
	UserName        string
	Password        string
	Host            string
	Port            int
	DBName          string
	Charset         string
	Others          string
	Tables          map[string]Table
	AutoIncrement   string
	ShowSQL         bool
	QuoteIdentifier string
}

func (e *Engine) OpenDB() (db *sql.DB, err error) {
	db = nil
	err = nil
	if e.Protocol == "sqlite" {
		// 'sqlite:///foo.db'
		db, err = sql.Open("sqlite3", e.Others)
		// 'sqlite:///:memory:'
	} else if e.Protocol == "mysql" {
		// 'mysql://<username>:<passwd>@<host>/<dbname>?charset=<encoding>'
		connstr := strings.Join([]string{e.UserName, ":",
			e.Password, "@tcp(", e.Host, ":3306)/", e.DBName, "?charset=", e.Charset}, "")
		db, err = sql.Open(e.Protocol, connstr)
	} else if e.Protocol == "mymysql" {
		//   DBNAME/USER/PASSWD
		connstr := strings.Join([]string{e.DBName, e.UserName, e.Password}, "/")
		db, err = sql.Open(e.Protocol, connstr)
		//   unix:SOCKPATH*DBNAME/USER/PASSWD
		//   unix:SOCKPATH,OPTIONS*DBNAME/USER/PASSWD
		//   tcp:ADDR*DBNAME/USER/PASSWD
		//   tcp:ADDR,OPTIONS*DBNAME/USER/PASSWD
	}

	return
}

func (engine *Engine) MakeSession() (session Session, err error) {
	db, err := engine.OpenDB()
	if err != nil {
		return Session{}, err
	}
	if engine.Protocol == "pgsql" {
		engine.QuoteIdentifier = "\""
		session = Session{Engine: engine, Db: db, ParamIteration: 1}
	} else if engine.Protocol == "mssql" {
		engine.QuoteIdentifier = ""
		session = Session{Engine: engine, Db: db, ParamIteration: 1}
	} else {
		engine.QuoteIdentifier = "`"
		session = Session{Engine: engine, Db: db, ParamIteration: 1}
	}
	session.Mapper = engine.Mapper
	session.Init()
	return
}

func (sqlType SQLType) genSQL(length int) string {
	if sqlType == Date {
		return " datetime "
	}
	return sqlType.Name + "(" + strconv.Itoa(length) + ")"
}

func (e *Engine) genCreateSQL(table *Table) string {
	sql := "CREATE TABLE IF NOT EXISTS `" + table.Name + "` ("
	//fmt.Println(session.Mapper.Obj2Table(session.PrimaryKey))
	for _, col := range table.Columns {
		if col.Name != "" {
			sql += "`" + col.Name + "` " + col.SQLType.genSQL(col.Length) + " "
			if col.Nullable {
				sql += " NULL "
			} else {
				sql += " NOT NULL "
			}
			//fmt.Println(key)
			if col.IsPrimaryKey {
				sql += "PRIMARY KEY "
			}
			if col.AutoIncrement {
				sql += e.AutoIncrement + " "
			}
			if col.IsUnique {
				sql += "Unique "
			}
			sql += ","
		}
	}
	sql = sql[:len(sql)-2] + ");"
	if e.ShowSQL {
		fmt.Println(sql)
	}
	return sql
}

func (e *Engine) genDropSQL(table *Table) string {
	sql := "DROP TABLE IF EXISTS `" + table.Name + "`;"
	if e.ShowSQL {
		fmt.Println(sql)
	}
	return sql
}

func Type(bean interface{}) reflect.Type {
	sliceValue := reflect.Indirect(reflect.ValueOf(bean))
	return reflect.TypeOf(sliceValue.Interface())
}

func Type2SQLType(t reflect.Type) (st SQLType) {
	switch k := t.Kind(); k {
	case reflect.Int, reflect.Int32, reflect.Int64:
		st = Int
	case reflect.String:
		st = Varchar
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			st = Date
		}
	default:
		st = Varchar
	}
	return
}

/*
map an object into a table object
*/
func (engine *Engine) MapOne(bean interface{}) Table {
	t := Type(bean)
	return engine.MapType(t)
}

func (engine *Engine) MapType(t reflect.Type) Table {
	table := Table{Name: engine.Mapper.Obj2Table(t.Name()), Type: t}
	table.Columns = make(map[string]Column)
	var pkCol *Column = nil
	var pkstr = ""

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		ormTagStr := tag.Get("xorm")
		var col Column
		fieldType := t.Field(i).Type

		if ormTagStr != "" {
			col = Column{FieldName: t.Field(i).Name}
			ormTagStr = strings.ToLower(ormTagStr)
			tags := strings.Split(ormTagStr, " ")
			// TODO: 
			if len(tags) > 0 {
				if tags[0] == "-" {
					continue
				}
				for j, key := range tags {
					switch k := strings.ToLower(key); k {
					case "pk":
						col.IsPrimaryKey = true
						pkCol = &col
					case "null":
						col.Nullable = (tags[j-1] != "not")
					case "autoincr":
						col.AutoIncrement = true
					case "default":
						col.Default = tags[j+1]
					case "int":
						col.SQLType = Int
					case "not":
					default:
						col.Name = k
					}
				}
				if col.SQLType.Name == "" {
					col.SQLType = Type2SQLType(fieldType)
					col.Length = col.SQLType.DefaultLength
				}

				if col.Name == "" {
					col.Name = engine.Mapper.Obj2Table(t.Field(i).Name)
				}
			}
		}

		if col.Name == "" {
			sqlType := Type2SQLType(fieldType)
			col = Column{engine.Mapper.Obj2Table(t.Field(i).Name), t.Field(i).Name, sqlType,
				sqlType.DefaultLength, true, "", false, false, false}
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
			col.AutoIncrement = true
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
		//t := getBeanType(bean)
		tableName := engine.Mapper.Obj2Table(StructName(bean))
		if _, ok := engine.Tables[tableName]; !ok {
			table := engine.MapOne(bean)
			engine.Tables[table.Name] = table
		}
	}
	return
}

func (engine *Engine) UnMap(beans ...interface{}) (e error) {
	for _, bean := range beans {
		//t := getBeanType(bean)
		tableName := engine.Mapper.Obj2Table(StructName(bean))
		if _, ok := engine.Tables[tableName]; ok {
			delete(engine.Tables, tableName)
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
		sql := e.genDropSQL(&table)
		_, err = session.Exec(sql)
		if err != nil {
			session.Rollback()
			break
		}
	}
	session.Commit()
	return err
}

func (e *Engine) CreateTables(beans ...interface{}) error {
	session, err := e.MakeSession()
	session.Begin()
	defer session.Close()
	if err != nil {
		return err
	}
	for _, bean := range beans {
		table := e.MapOne(bean)
		e.Tables[table.Name] = table
		sql := e.genCreateSQL(&table)
		_, err = session.Exec(sql)
		if err != nil {
			session.Rollback()
			break
		}
	}
	session.Commit()
	return err
}

func (e *Engine) CreateAll() error {
	session, err := e.MakeSession()
	session.Begin()
	defer session.Close()
	if err != nil {
		return err
	}

	for _, table := range e.Tables {
		sql := e.genCreateSQL(&table)
		_, err = session.Exec(sql)
		if err != nil {
			session.Rollback()
			break
		}
	}
	session.Commit()
	return err
}

func (engine *Engine) Insert(beans ...interface{}) (int64, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return -1, err
	}

	return session.Insert(beans...)
}

func (engine *Engine) Update(bean interface{}) (int64, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return -1, err
	}

	return session.Update(bean)
}

func (engine *Engine) Delete(bean interface{}) (int64, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return -1, err
	}

	return session.Delete(bean)
}

func (engine *Engine) Get(bean interface{}) error {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return err
	}

	return session.Get(bean)
}

func (engine *Engine) Find(beans interface{}) error {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return err
	}

	return session.Find(beans)
}

func (engine *Engine) Count(bean interface{}) (int64, error) {
	session, err := engine.MakeSession()
	defer session.Close()
	if err != nil {
		return 0, err
	}

	return session.Count(bean)
}
