package xorm

import (
	"database/sql"
)

type Iterator struct {
	session *Session
	startId int
	rows    *sql.Rows
}

func (iter *Iterator) IsValid() bool {
	return iter.session != nil && iter.rows != nil
}

/*
func (iter *Iterator) Next(bean interface{}) (bool, error) {
	if !iter.IsValid() {
		return errors.New("iterator is not valied.")
	}
	if iter.rows.Next() {
		iter.rows.Scan(...)
	}
}*/

// close the iterator
func (iter *Iterator) Close() {
	if iter.rows != nil {
		iter.rows.Close()
	}
	if iter.session != nil && iter.session.IsAutoClose {
		iter.session.Close()
	}
}
