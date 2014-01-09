package core

import (
	"database/sql"
	"reflect"
)

type DB struct {
	*sql.DB
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	return &DB{db}, err
}

func (db *DB) Query(query string, args ...interface{}) (*Rows, error) {
	rows, err := db.DB.Query(query, args...)
	return &Rows{rows}, err
}

type Rows struct {
	*sql.Rows
}

func (rs *Rows) Scan(dest ...interface{}) error {
	newDest := make([]interface{}, 0)
	for _, s := range dest {
		vv := reflect.ValueOf(s)
		switch vv.Kind() {
		case reflect.Ptr:
			vvv := vv.Elem()
			if vvv.Kind() == reflect.Struct {
				for j := 0; j < vvv.NumField(); j++ {
					newDest = append(newDest, vvv.FieldByIndex([]int{j}).Addr().Interface())
				}
			} else {
				newDest = append(newDest, s)
			}
		}
	}

	return rs.Rows.Scan(newDest...)
}
