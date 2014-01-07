package core

import (
	"reflect"
	"strings"
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
func (col *Column) ValueOf(bean interface{}) reflect.Value {
	var fieldValue reflect.Value
	if strings.Contains(col.FieldName, ".") {
		fields := strings.Split(col.FieldName, ".")
		if len(fields) > 2 {
			return reflect.ValueOf(nil)
		}

		fieldValue = reflect.Indirect(reflect.ValueOf(bean)).FieldByName(fields[0])
		fieldValue = fieldValue.FieldByName(fields[1])
	} else {
		fieldValue = reflect.Indirect(reflect.ValueOf(bean)).FieldByName(col.FieldName)
	}
	return fieldValue
}
