package core

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	TWOSIDES = iota + 1
	ONLYTODB
	ONLYFROMDB
)

// database column
type Column struct {
	Name            string
	FieldName       string
	SQLType         SQLType
	Length          int
	Length2         int
	Nullable        bool
	Default         string
	Indexes         map[string]bool
	IsPrimaryKey    bool
	IsAutoIncrement bool
	MapType         int
	IsCreated       bool
	IsUpdated       bool
	IsCascade       bool
	IsVersion       bool
	fieldPath       []string
}

func NewColumn(name, fieldName string, sqlType SQLType, len1, len2 int, nullable bool) *Column {
	return &Column{name, fieldName, sqlType, len1, len2, nullable, "", make(map[string]bool), false, false,
		TWOSIDES, false, false, false, false, nil}
}

// generate column description string according dialect
func (col *Column) String(d Dialect) string {
	sql := d.QuoteStr() + col.Name + d.QuoteStr() + " "

	sql += d.SqlType(col) + " "

	if col.IsPrimaryKey {
		sql += "PRIMARY KEY "
		if col.IsAutoIncrement {
			sql += d.AutoIncrStr() + " "
		}
	}

	if col.Nullable {
		sql += "NULL "
	} else {
		sql += "NOT NULL "
	}

	if col.Default != "" {
		sql += "DEFAULT " + col.Default + " "
	}

	return sql
}

func (col *Column) StringNoPk(d Dialect) string {
	sql := d.QuoteStr() + col.Name + d.QuoteStr() + " "

	sql += d.SqlType(col) + " "

	if col.Nullable {
		sql += "NULL "
	} else {
		sql += "NOT NULL "
	}

	if col.Default != "" {
		sql += "DEFAULT " + col.Default + " "
	}

	return sql
}

// return col's filed of struct's value
func (col *Column) ValueOf(bean interface{}) (*reflect.Value, error) {
	dataStruct := reflect.Indirect(reflect.ValueOf(bean))
	return col.ValueOfV(&dataStruct)
}

func (col *Column) ValueOfV(dataStruct *reflect.Value) (*reflect.Value, error) {
	var fieldValue reflect.Value
	var err error
	if col.fieldPath == nil {
		col.fieldPath = strings.Split(col.FieldName, ".")
	}

	if len(col.fieldPath) == 1 {
		fieldValue = dataStruct.FieldByName(col.FieldName)
	} else if len(col.fieldPath) == 2 {
		parentField := dataStruct.FieldByName(col.fieldPath[0])
		if parentField.IsValid() {
			fieldValue = parentField.FieldByName(col.fieldPath[1])
		} else {
			err = fmt.Errorf("field  %v is not valid", col.FieldName)
		}
	} else {
		err = fmt.Errorf("Unsupported mutliderive %v", col.FieldName)
	}
	if err != nil {
		return nil, err
	}
	return &fieldValue, nil
}
