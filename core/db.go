package core

import (
	"database/sql"
	"errors"
	"reflect"
)

type DB struct {
	*sql.DB
	Mapper IMapper
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	return &DB{db, &SnakeMapper{}}, err
}

func (db *DB) Query(query string, args ...interface{}) (*Rows, error) {
	rows, err := db.DB.Query(query, args...)
	return &Rows{rows, db.Mapper}, err
}

type Rows struct {
	*sql.Rows
	Mapper IMapper
}

// scan data to a struct's pointer according field index
func (rs *Rows) ScanStruct(dest interface{}) error {
	vv := reflect.ValueOf(dest)
	if vv.Kind() != reflect.Ptr || vv.Elem().Kind() != reflect.Struct {
		return errors.New("dest should be a struct's pointer")
	}

	vvv := vv.Elem()
	newDest := make([]interface{}, vvv.NumField())

	for j := 0; j < vvv.NumField(); j++ {
		newDest[j] = vvv.Field(j).Addr().Interface()
	}

	return rs.Rows.Scan(newDest...)
}

// scan data to a struct's pointer according field name
func (rs *Rows) ScanStruct2(dest interface{}) error {
	vv := reflect.ValueOf(dest)
	if vv.Kind() != reflect.Ptr || vv.Elem().Kind() != reflect.Struct {
		return errors.New("dest should be a struct's pointer")
	}

	cols, err := rs.Columns()
	if err != nil {
		return err
	}

	vvv := vv.Elem()
	newDest := make([]interface{}, len(cols))

	for j, name := range cols {
		f := vvv.FieldByName(rs.Mapper.Table2Obj(name))
		if f.IsValid() {
			newDest[j] = f.Addr().Interface()
		} else {
			var v interface{}
			newDest[j] = &v
		}
	}

	return rs.Rows.Scan(newDest...)
}

// scan data to a slice's pointer, slice's length should equal to columns' number
func (rs *Rows) ScanSlice(dest interface{}) error {
	vv := reflect.ValueOf(dest)
	if vv.Kind() != reflect.Ptr || vv.Elem().Kind() != reflect.Slice {
		return errors.New("dest should be a slice's pointer")
	}

	vvv := vv.Elem()
	cols, err := rs.Columns()
	if err != nil {
		return err
	}

	newDest := make([]interface{}, len(cols))

	for j := 0; j < len(cols); j++ {
		if j >= vvv.Len() {
			newDest[j] = reflect.New(vvv.Type().Elem()).Interface()
		} else {
			newDest[j] = vvv.Index(j).Addr().Interface()
		}
	}

	err = rs.Rows.Scan(newDest...)
	if err != nil {
		return err
	}

	for i, _ := range cols {
		vvv = reflect.Append(vvv, reflect.ValueOf(newDest[i]).Elem())
	}
	return nil
}

// scan data to a map's pointer
func (rs *Rows) ScanMap(dest interface{}) error {
	vv := reflect.ValueOf(dest)
	if vv.Kind() != reflect.Ptr || vv.Elem().Kind() != reflect.Map {
		return errors.New("dest should be a map's pointer")
	}

	cols, err := rs.Columns()
	if err != nil {
		return err
	}

	newDest := make([]interface{}, len(cols))
	vvv := vv.Elem()

	for i, _ := range cols {
		v := reflect.New(vvv.Type().Elem())
		newDest[i] = v.Interface()
	}

	err = rs.Rows.Scan(newDest...)
	if err != nil {
		return err
	}

	for i, name := range cols {
		vname := reflect.ValueOf(name)
		vvv.SetMapIndex(vname, reflect.ValueOf(newDest[i]).Elem())
	}

	return nil
}
