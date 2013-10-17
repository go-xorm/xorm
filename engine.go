package xorm

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	POSTGRES = "postgres"
	SQLITE   = "sqlite3"
	MYSQL    = "mysql"
	MYMYSQL  = "mymysql"
)

// a dialect is a driver's wrapper
type dialect interface {
	Init(DriverName, DataSourceName string) error
	SqlType(t *Column) string
	SupportInsertMany() bool
	QuoteStr() string
	AutoIncrStr() string
	SupportEngine() bool
	SupportCharset() bool
	IndexOnTable() bool
	IndexCheckSql(tableName, idxName string) (string, []interface{})
	TableCheckSql(tableName string) (string, []interface{})
	ColumnCheckSql(tableName, colName string) (string, []interface{})

	GetColumns(tableName string) (map[string]*Column, error)
	GetTables() ([]*Table, error)
	GetIndexes(tableName string) (map[string]*Index, error)
}

type Engine struct {
	Mapper         IMapper
	TagIdentifier  string
	DriverName     string
	DataSourceName string
	dialect        dialect
	Tables         map[reflect.Type]*Table
	mutex          *sync.Mutex
	ShowSQL        bool
	ShowErr        bool
	ShowDebug      bool
	ShowWarn       bool
	Pool           IConnectPool
	Filters        []Filter
	Logger         io.Writer
	Cacher         Cacher
	UseCache       bool
}

// If engine's database support batch insert records like
// "insert into user values (name, age), (name, age)".
// When the return is ture, then engine.Insert(&users) will
// generate batch sql and exeute.
func (engine *Engine) SupportInsertMany() bool {
	return engine.dialect.SupportInsertMany()
}

// Engine's database use which charactor as quote.
// mysql, sqlite use ` and postgres use "
func (engine *Engine) QuoteStr() string {
	return engine.dialect.QuoteStr()
}

// Use QuoteStr quote the string sql
func (engine *Engine) Quote(sql string) string {
	return engine.dialect.QuoteStr() + sql + engine.dialect.QuoteStr()
}

// A simple wrapper to dialect's SqlType method
func (engine *Engine) SqlType(c *Column) string {
	return engine.dialect.SqlType(c)
}

// Database's autoincrement statement
func (engine *Engine) AutoIncrStr() string {
	return engine.dialect.AutoIncrStr()
}

// Set engine's pool, the pool default is Go's standard library's connection pool.
func (engine *Engine) SetPool(pool IConnectPool) error {
	engine.Pool = pool
	return engine.Pool.Init(engine)
}

// SetMaxConns is only available for go 1.2+
func (engine *Engine) SetMaxConns(conns int) {
	engine.Pool.SetMaxConns(conns)
}

// SetDefaltCacher set the default cacher. Xorm's default not enable cacher.
func (engine *Engine) SetDefaultCacher(cacher Cacher) {
	if cacher == nil {
		engine.UseCache = false
	} else {
		engine.UseCache = true
		engine.Cacher = cacher
	}
}

// If you has set default cacher, and you want temporilly stop use cache,
// you can use NoCache()
func (engine *Engine) NoCache() *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.NoCache()
}

// Set a table use a special cacher
func (engine *Engine) MapCacher(bean interface{}, cacher Cacher) {
	t := rType(bean)
	engine.AutoMapType(t)
	engine.Tables[t].Cacher = cacher
}

// OpenDB provides a interface to operate database directly.
func (engine *Engine) OpenDB() (*sql.DB, error) {
	return sql.Open(engine.DriverName, engine.DataSourceName)
}

// New a session
func (engine *Engine) NewSession() *Session {
	session := &Session{Engine: engine}
	session.Init()
	return session
}

// Close the engine
func (engine *Engine) Close() error {
	return engine.Pool.Close(engine)
}

// Test if database is alive.
func (engine *Engine) Test() error {
	session := engine.NewSession()
	defer session.Close()
	engine.LogSQL("PING DATABASE", engine.DriverName)
	return session.Ping()
}

func (engine *Engine) LogSQL(contents ...interface{}) {
	if engine.ShowSQL {
		io.WriteString(engine.Logger, fmt.Sprintln(contents...))
	}
}

func (engine *Engine) LogError(contents ...interface{}) {
	if engine.ShowErr {
		io.WriteString(engine.Logger, fmt.Sprintln(contents...))
	}
}

func (engine *Engine) LogDebug(contents ...interface{}) {
	if engine.ShowDebug {
		io.WriteString(engine.Logger, fmt.Sprintln(contents...))
	}
}

func (engine *Engine) LogWarn(contents ...interface{}) {
	if engine.ShowWarn {
		io.WriteString(engine.Logger, fmt.Sprintln(contents...))
	}
}

func (engine *Engine) Sql(querystring string, args ...interface{}) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Sql(querystring, args...)
}

func (engine *Engine) NoAutoTime() *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.NoAutoTime()
}

func (engine *Engine) DBMetas() ([]*Table, error) {
	tables, err := engine.dialect.GetTables()
	if err != nil {
		return nil, err
	}

	for _, table := range tables {
		cols, err := engine.dialect.GetColumns(table.Name)
		if err != nil {
			return nil, err
		}
		table.Columns = cols

		indexes, err := engine.dialect.GetIndexes(table.Name)
		if err != nil {
			return nil, err
		}
		table.Indexes = indexes

		for _, index := range indexes {
			for _, name := range index.Cols {
				if col, ok := table.Columns[name]; ok {
					col.Indexes[index.Name] = true
				} else {
					return nil, errors.New("Unkonwn col " + name + " in indexes")
				}
			}
		}
	}
	return tables, nil
}

func (engine *Engine) Cascade(trueOrFalse ...bool) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Cascade(trueOrFalse...)
}

func (engine *Engine) Where(querystring string, args ...interface{}) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Where(querystring, args...)
}

func (engine *Engine) Id(id int64) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Id(id)
}

func (engine *Engine) Charset(charset string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Charset(charset)
}

func (engine *Engine) StoreEngine(storeEngine string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.StoreEngine(storeEngine)
}

func (engine *Engine) Cols(columns ...string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Cols(columns...)
}

func (engine *Engine) Omit(columns ...string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Omit(columns...)
}

/*func (engine *Engine) Trans(t string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Trans(t)
}*/

func (engine *Engine) In(column string, args ...interface{}) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.In(column, args...)
}

func (engine *Engine) Table(tableNameOrBean interface{}) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Table(tableNameOrBean)
}

func (engine *Engine) Limit(limit int, start ...int) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Limit(limit, start...)
}

func (engine *Engine) Desc(colNames ...string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Desc(colNames...)
}

func (engine *Engine) Asc(colNames ...string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Asc(colNames...)
}

func (engine *Engine) OrderBy(order string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.OrderBy(order)
}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (engine *Engine) Join(join_operator, tablename, condition string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Join(join_operator, tablename, condition)
}

func (engine *Engine) GroupBy(keys string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.GroupBy(keys)
}

func (engine *Engine) Having(conditions string) *Session {
	session := engine.NewSession()
	session.IsAutoClose = true
	return session.Having(conditions)
}

//
func (engine *Engine) AutoMapType(t reflect.Type) *Table {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	table, ok := engine.Tables[t]
	if !ok {
		table = engine.MapType(t)
		engine.Tables[t] = table
	}
	return table
}

func (engine *Engine) AutoMap(bean interface{}) *Table {
	t := rType(bean)
	return engine.AutoMapType(t)
}

func (engine *Engine) newTable() *Table {
	table := &Table{}
	table.Indexes = make(map[string]*Index)
	table.Columns = make(map[string]*Column)
	table.ColumnsSeq = make([]string, 0)
	table.Cacher = engine.Cacher
	return table
}

func (engine *Engine) MapType(t reflect.Type) *Table {
	table := engine.newTable()
	table.Name = engine.Mapper.Obj2Table(t.Name())
	table.Type = t

	var idFieldColName string

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		ormTagStr := tag.Get(engine.TagIdentifier)
		var col *Column
		fieldType := t.Field(i).Type

		if ormTagStr != "" {
			col = &Column{FieldName: t.Field(i).Name, Nullable: true, IsPrimaryKey: false,
				IsAutoIncrement: false, MapType: TWOSIDES, Indexes: make(map[string]bool)}
			tags := strings.Split(ormTagStr, " ")

			if len(tags) > 0 {
				if tags[0] == "-" {
					continue
				}
				if (strings.ToUpper(tags[0]) == "EXTENDS") &&
					(fieldType.Kind() == reflect.Struct) {
					parentTable := engine.MapType(fieldType)
					for name, col := range parentTable.Columns {
						col.FieldName = fmt.Sprintf("%v.%v", fieldType.Name(), col.FieldName)
						table.Columns[name] = col
						table.ColumnsSeq = append(table.ColumnsSeq, name)
					}

					table.PrimaryKey = parentTable.PrimaryKey
					continue
				}
				var indexType int
				var indexName string
				for j, key := range tags {
					k := strings.ToUpper(key)
					switch {
					case k == "<-":
						col.MapType = ONLYFROMDB
					case k == "->":
						col.MapType = ONLYTODB
					case k == "PK":
						col.IsPrimaryKey = true
						col.Nullable = false
					case k == "NULL":
						col.Nullable = (strings.ToUpper(tags[j-1]) != "NOT")
					case k == "AUTOINCR":
						col.IsAutoIncrement = true
					case k == "DEFAULT":
						col.Default = tags[j+1]
					case k == "CREATED":
						col.IsCreated = true
					case k == "UPDATED":
						col.IsUpdated = true
					/*case strings.HasPrefix(k, "--"):
					col.Comment = k[2:len(k)]*/
					case strings.HasPrefix(k, "INDEX(") && strings.HasSuffix(k, ")"):
						indexType = IndexType
						indexName = k[len("INDEX")+1 : len(k)-1]
					case k == "INDEX":
						indexType = IndexType
					case strings.HasPrefix(k, "UNIQUE(") && strings.HasSuffix(k, ")"):
						indexName = k[len("UNIQUE")+1 : len(k)-1]
						indexType = UniqueType
					case k == "UNIQUE":
						indexType = UniqueType
					case k == "NOTNULL":
						col.Nullable = false
					case k == "NOT":
					default:
						if strings.HasPrefix(k, "'") && strings.HasSuffix(k, "'") {
							if key != col.Default {
								col.Name = key[1 : len(key)-1]
							}
						} else if strings.Contains(k, "(") && strings.HasSuffix(k, ")") {
							fs := strings.Split(k, "(")
							if _, ok := sqlTypes[fs[0]]; !ok {
								continue
							}
							col.SQLType = SQLType{fs[0], 0, 0}
							fs2 := strings.Split(fs[1][0:len(fs[1])-1], ",")
							if len(fs2) == 2 {
								col.Length, _ = strconv.Atoi(fs2[0])
								col.Length2, _ = strconv.Atoi(fs2[1])
							} else if len(fs2) == 1 {
								col.Length, _ = strconv.Atoi(fs2[0])
							}
						} else {
							if _, ok := sqlTypes[k]; ok {
								col.SQLType = SQLType{k, 0, 0}
							} else if key != col.Default {
								col.Name = key
							}
						}
						engine.SqlType(col)
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
				if indexType == IndexType {
					if indexName == "" {
						indexName = col.Name
					}
					if index, ok := table.Indexes[indexName]; ok {
						index.AddColumn(col.Name)
						col.Indexes[index.Name] = true
					} else {
						index := NewIndex(indexName, IndexType)
						index.AddColumn(col.Name)
						table.AddIndex(index)
						col.Indexes[index.Name] = true
					}
				} else if indexType == UniqueType {
					if indexName == "" {
						indexName = col.Name
					}
					if index, ok := table.Indexes[indexName]; ok {
						index.AddColumn(col.Name)
						col.Indexes[index.Name] = true
					} else {
						index := NewIndex(indexName, UniqueType)
						index.AddColumn(col.Name)
						table.AddIndex(index)
						col.Indexes[index.Name] = true
					}
				}
			}
		} else {
			sqlType := Type2SQLType(fieldType)
			col = &Column{engine.Mapper.Obj2Table(t.Field(i).Name), t.Field(i).Name, sqlType,
				sqlType.DefaultLength, sqlType.DefaultLength2, true, "", make(map[string]bool), false, false,
				TWOSIDES, false, false, false}
		}
		if col.IsAutoIncrement {
			col.Nullable = false
		}

		table.AddColumn(col)

		if col.FieldName == "Id" || strings.HasSuffix(col.FieldName, ".Id") {
			idFieldColName = col.Name
		}
	}

	if idFieldColName != "" && table.PrimaryKey == "" {
		col := table.Columns[idFieldColName]
		col.IsPrimaryKey = true
		col.IsAutoIncrement = true
		col.Nullable = false
		table.PrimaryKey = col.Name
	}

	return table
}

// Map a struct to a table
func (engine *Engine) Map(beans ...interface{}) (e error) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	for _, bean := range beans {
		t := rType(bean)
		engine.Tables[t] = engine.MapType(t)
	}
	return
}

// Is a table has any reocrd
func (engine *Engine) IsTableEmpty(bean interface{}) (bool, error) {
	t := rType(bean)
	if t.Kind() != reflect.Struct {
		return false, errors.New("bean should be a struct or struct's point")
	}
	engine.AutoMapType(t)
	session := engine.NewSession()
	defer session.Close()
	rows, err := session.Count(bean)
	return rows > 0, err
}

// Is a table is exist
func (engine *Engine) IsTableExist(bean interface{}) (bool, error) {
	t := rType(bean)
	if t.Kind() != reflect.Struct {
		return false, errors.New("bean should be a struct or struct's point")
	}
	table := engine.AutoMapType(t)
	session := engine.NewSession()
	defer session.Close()
	has, err := session.isTableExist(table.Name)
	return has, err
}

// create indexes
func (engine *Engine) CreateIndexes(bean interface{}) error {
	session := engine.NewSession()
	defer session.Close()
	return session.CreateIndexes(bean)
}

// create uniques
func (engine *Engine) CreateUniques(bean interface{}) error {
	session := engine.NewSession()
	defer session.Close()
	return session.CreateUniques(bean)
}

// If enabled cache, clear the cache bean
func (engine *Engine) ClearCacheBean(bean interface{}, id int64) error {
	t := rType(bean)
	if t.Kind() != reflect.Struct {
		return errors.New("error params")
	}
	table := engine.AutoMap(bean)
	if table.Cacher != nil {
		table.Cacher.ClearIds(table.Name)
		table.Cacher.DelBean(table.Name, id)
	}
	return nil
}

// If enabled cache, clear some tables' cache
func (engine *Engine) ClearCache(beans ...interface{}) error {
	for _, bean := range beans {
		t := rType(bean)
		if t.Kind() != reflect.Struct {
			return errors.New("error params")
		}
		table := engine.AutoMap(bean)
		if table.Cacher != nil {
			table.Cacher.ClearIds(table.Name)
			table.Cacher.ClearBeans(table.Name)
		}
	}
	return nil
}

// Sync the new struct change to database, this method will automatically add
// table, column, index, unique. but will not delete or change anything.
// If you change some field, you should change the database manually.
func (engine *Engine) Sync(beans ...interface{}) error {
	for _, bean := range beans {
		table := engine.AutoMap(bean)

		s := engine.NewSession()
		defer s.Close()
		isExist, err := s.Table(bean).isTableExist(table.Name)
		if err != nil {
			return err
		}
		if !isExist {
			err = engine.CreateTables(bean)
			if err != nil {
				return err
			}
		}
		/*isEmpty, err := engine.IsEmptyTable(bean)
		if err != nil {
			return err
		}*/
		var isEmpty bool = false
		if isEmpty {
			err = engine.DropTables(bean)
			if err != nil {
				return err
			}
			err = engine.CreateTables(bean)
			if err != nil {
				return err
			}
		} else {
			for _, col := range table.Columns {
				session := engine.NewSession()
				session.Statement.RefTable = table
				defer session.Close()
				isExist, err := session.isColumnExist(table.Name, col.Name)
				if err != nil {
					return err
				}
				if !isExist {
					session := engine.NewSession()
					session.Statement.RefTable = table
					defer session.Close()
					err = session.addColumn(col.Name)
					if err != nil {
						return err
					}
				}
			}

			for name, index := range table.Indexes {
				session := engine.NewSession()
				session.Statement.RefTable = table
				defer session.Close()
				if index.Type == UniqueType {
					isExist, err := session.isIndexExist(table.Name, name, true)
					if err != nil {
						return err
					}
					if !isExist {
						session := engine.NewSession()
						session.Statement.RefTable = table
						defer session.Close()
						err = session.addUnique(table.Name, name)
						if err != nil {
							return err
						}
					}
				} else if index.Type == IndexType {
					isExist, err := session.isIndexExist(table.Name, name, false)
					if err != nil {
						return err
					}
					if !isExist {
						session := engine.NewSession()
						session.Statement.RefTable = table
						defer session.Close()
						err = session.addIndex(table.Name, name)
						if err != nil {
							return err
						}
					}
				} else {
					return errors.New("unknow index type")
				}
			}
		}
	}
	return nil
}

func (engine *Engine) UnMap(beans ...interface{}) (e error) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	for _, bean := range beans {
		t := rType(bean)
		if _, ok := engine.Tables[t]; ok {
			delete(engine.Tables, t)
		}
	}
	return
}

// Drop all mapped table
func (engine *Engine) DropAll() error {
	session := engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}
	err = session.DropAll()
	if err != nil {
		session.Rollback()
		return err
	}
	return session.Commit()
}

// CreateTables create tabls according bean
func (engine *Engine) CreateTables(beans ...interface{}) error {
	session := engine.NewSession()
	err := session.Begin()
	defer session.Close()
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

func (engine *Engine) DropTables(beans ...interface{}) error {
	session := engine.NewSession()
	err := session.Begin()
	defer session.Close()
	if err != nil {
		return err
	}

	for _, bean := range beans {
		err = session.DropTable(bean)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return session.Commit()
}

func (engine *Engine) CreateAll() error {
	session := engine.NewSession()
	defer session.Close()
	return session.CreateAll()
}

func (engine *Engine) Exec(sql string, args ...interface{}) (sql.Result, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Exec(sql, args...)
}

func (engine *Engine) Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Query(sql, paramStr...)
}

func (engine *Engine) Insert(beans ...interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Insert(beans...)
}

func (engine *Engine) InsertOne(bean interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.InsertOne(bean)
}

func (engine *Engine) Update(bean interface{}, condiBeans ...interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Update(bean, condiBeans...)
}

func (engine *Engine) Delete(bean interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Delete(bean)
}

func (engine *Engine) Get(bean interface{}) (bool, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Get(bean)
}

func (engine *Engine) Find(beans interface{}, condiBeans ...interface{}) error {
	session := engine.NewSession()
	defer session.Close()
	return session.Find(beans, condiBeans...)
}

func (engine *Engine) Iterate(bean interface{}, fun IterFunc) error {
	session := engine.NewSession()
	defer session.Close()
	return session.Iterate(bean, fun)
}

func (engine *Engine) Count(bean interface{}) (int64, error) {
	session := engine.NewSession()
	defer session.Close()
	return session.Count(bean)
}
