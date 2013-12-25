package xorm

import (
	"database/sql"
	"reflect"
)

type Iterator struct {
	session  *Session
	stmt     *sql.Stmt
	rows     *sql.Rows
	fields   []string
	beanType reflect.Type
}

func newIterator(session *Session, bean interface{}) (*Iterator, error) {
	iterator := new(Iterator)
	iterator.session = session
	iterator.beanType = reflect.Indirect(reflect.ValueOf(bean)).Type()

	err := iterator.session.newDb()
	if err != nil {
		return nil, err
	}

	defer iterator.session.Statement.Init()

	var sql string
	var args []interface{}
	iterator.session.Statement.RefTable = iterator.session.Engine.autoMap(bean)
	if iterator.session.Statement.RawSQL == "" {
		sql, args = iterator.session.Statement.genGetSql(bean)
	} else {
		sql = iterator.session.Statement.RawSQL
		args = iterator.session.Statement.RawParams
	}

	for _, filter := range iterator.session.Engine.Filters {
		sql = filter.Do(sql, session)
	}

	iterator.session.Engine.LogSQL(sql)
	iterator.session.Engine.LogSQL(args)

	iterator.stmt, err = iterator.session.Db.Prepare(sql)
	if err != nil {
		defer iterator.Close()
		return nil, err
	}

	iterator.rows, err = iterator.stmt.Query(args...)
	if err != nil {
		defer iterator.Close()
		return nil, err
	}

	iterator.fields, err = iterator.rows.Columns()
	if err != nil {
		defer iterator.Close()
		return nil, err
	}

	return iterator, nil
}

// iterate to next record and reuse passed bean obj
func (iterator *Iterator) NextReuse(bean interface{}) (interface{}, error) {
	if iterator.rows != nil && iterator.rows.Next() {
		result, err := row2map(iterator.rows, iterator.fields) // !nashtsai! TODO remove row2map then scanMapIntoStruct conversation for better performance
		if err == nil {
			err = iterator.session.scanMapIntoStruct(bean, result)
		}
		if err == nil {
			return bean, nil
		} else {
			return nil, err
		}
	}
	return nil, sql.ErrNoRows
}

// iterate to next record
func (iterator *Iterator) Next() (interface{}, error) {
	b := reflect.New(iterator.beanType).Interface()
	return iterator.NextReuse(b)
}

// close session if session.IsAutoClose is true, and claimed any opened resources
func (iterator *Iterator) Close() {
	if iterator.session.IsAutoClose {
		defer iterator.session.Close()
	}
	if iterator.stmt != nil {
		defer iterator.stmt.Close()
	}
	if iterator.rows != nil {
		defer iterator.rows.Close()
	}
}
