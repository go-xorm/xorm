// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"xorm.io/core"
)

type db2 struct {
	core.Base
}

func (db *db2) Init(d *core.DB, uri *core.Uri, drivername, dataSourceName string) error {
	err := db.Base.Init(d, db, uri, drivername, dataSourceName)
	if err != nil {
		return err
	}
	return nil
}

func (db *db2) SqlType(c *core.Column) string {
	var res string
	switch t := c.SQLType.Name; t {
	case core.TinyInt:
		res = core.SmallInt
		return res
	case core.Bit:
		res = core.Boolean
		return res
	case core.MediumInt, core.Int, core.Integer:
		if c.IsAutoIncrement {
			return core.Serial
		}
		return core.Integer
	case core.BigInt:
		if c.IsAutoIncrement {
			return core.BigSerial
		}
		return core.BigInt
	case core.Serial, core.BigSerial:
		c.IsAutoIncrement = true
		c.Nullable = false
		res = t
	case core.Binary, core.VarBinary:
		return core.Bytea
	case core.DateTime:
		res = core.TimeStamp
	case core.TimeStampz:
		return "timestamp with time zone"
	case core.Float:
		res = core.Real
	case core.TinyText, core.MediumText, core.LongText:
		res = core.Text
	case core.NVarchar:
		res = core.Varchar
	case core.Uuid:
		return core.Uuid
	case core.Blob, core.TinyBlob, core.MediumBlob, core.LongBlob:
		return core.Bytea
	case core.Double:
		return "DOUBLE PRECISION"
	default:
		if c.IsAutoIncrement {
			return core.Serial
		}
		res = t
	}

	if strings.EqualFold(res, "bool") {
		// for bool, we don't need length information
		return res
	}
	hasLen1 := (c.Length > 0)
	hasLen2 := (c.Length2 > 0)

	if hasLen2 {
		res += "(" + strconv.Itoa(c.Length) + "," + strconv.Itoa(c.Length2) + ")"
	} else if hasLen1 {
		res += "(" + strconv.Itoa(c.Length) + ")"
	}
	return res
}

func (db *db2) SupportInsertMany() bool {
	return true
}

func (db *db2) IsReserved(name string) bool {
	_, ok := postgresReservedWords[name]
	return ok
}

func (db *db2) Quote(name string) string {
	name = strings.Replace(name, ".", `"."`, -1)
	return "\"" + name + "\""
}

func (db *db2) AutoIncrStr() string {
	return ""
}

func (db *db2) SupportEngine() bool {
	return false
}

func (db *db2) SupportCharset() bool {
	return false
}

func (db *db2) IndexOnTable() bool {
	return false
}

func (db *db2) IndexCheckSql(tableName, idxName string) (string, []interface{}) {
	if len(db.Schema) == 0 {
		args := []interface{}{tableName, idxName}
		return `SELECT indexname FROM pg_indexes WHERE tablename = ? AND indexname = ?`, args
	}

	args := []interface{}{db.Schema, tableName, idxName}
	return `SELECT indexname FROM pg_indexes ` +
		`WHERE schemaname = ? AND tablename = ? AND indexname = ?`, args
}

func (db *db2) TableCheckSql(tableName string) (string, []interface{}) {
	if len(db.Schema) == 0 {
		args := []interface{}{tableName}
		return `SELECT tablename FROM pg_tables WHERE tablename = ?`, args
	}

	args := []interface{}{db.Schema, tableName}
	return `SELECT tablename FROM pg_tables WHERE schemaname = ? AND tablename = ?`, args
}

func (db *db2) ModifyColumnSql(tableName string, col *core.Column) string {
	if len(db.Schema) == 0 {
		return fmt.Sprintf("alter table %s ALTER COLUMN %s TYPE %s",
			tableName, col.Name, db.SqlType(col))
	}
	return fmt.Sprintf("alter table %s.%s ALTER COLUMN %s TYPE %s",
		db.Schema, tableName, col.Name, db.SqlType(col))
}

func (db *db2) DropIndexSql(tableName string, index *core.Index) string {
	quote := db.Quote
	idxName := index.Name

	tableName = strings.Replace(tableName, `"`, "", -1)
	tableName = strings.Replace(tableName, `.`, "_", -1)

	if !strings.HasPrefix(idxName, "UQE_") &&
		!strings.HasPrefix(idxName, "IDX_") {
		if index.Type == core.UniqueType {
			idxName = fmt.Sprintf("UQE_%v_%v", tableName, index.Name)
		} else {
			idxName = fmt.Sprintf("IDX_%v_%v", tableName, index.Name)
		}
	}
	if db.Uri.Schema != "" {
		idxName = db.Uri.Schema + "." + idxName
	}
	return fmt.Sprintf("DROP INDEX %v", quote(idxName))
}

func (db *db2) IsColumnExist(tableName, colName string) (bool, error) {
	args := []interface{}{db.Schema, tableName, colName}
	query := "SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE table_schema = $1 AND table_name = $2" +
		" AND column_name = $3"
	if len(db.Schema) == 0 {
		args = []interface{}{tableName, colName}
		query = "SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = $1" +
			" AND column_name = $2"
	}
	db.LogSQL(query, args)

	rows, err := db.DB().Query(query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (db *db2) GetColumns(tableName string) ([]string, map[string]*core.Column, error) {
	args := []interface{}{tableName}
	s := `SELECT column_name, column_default, is_nullable, data_type, character_maximum_length, numeric_precision, numeric_precision_radix ,
    CASE WHEN p.contype = 'p' THEN true ELSE false END AS primarykey,
    CASE WHEN p.contype = 'u' THEN true ELSE false END AS uniquekey
FROM pg_attribute f
    JOIN pg_class c ON c.oid = f.attrelid JOIN pg_type t ON t.oid = f.atttypid
    LEFT JOIN pg_attrdef d ON d.adrelid = c.oid AND d.adnum = f.attnum
    LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
    LEFT JOIN pg_constraint p ON p.conrelid = c.oid AND f.attnum = ANY (p.conkey)
    LEFT JOIN pg_class AS g ON p.confrelid = g.oid
    LEFT JOIN INFORMATION_SCHEMA.COLUMNS s ON s.column_name=f.attname AND c.relname=s.table_name
WHERE c.relkind = 'r'::char AND c.relname = $1%s AND f.attnum > 0 ORDER BY f.attnum;`

	var f string
	if len(db.Schema) != 0 {
		args = append(args, db.Schema)
		f = " AND s.table_schema = $2"
	}
	s = fmt.Sprintf(s, f)

	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols := make(map[string]*core.Column)
	colSeq := make([]string, 0)

	for rows.Next() {
		col := new(core.Column)
		col.Indexes = make(map[string]int)

		var colName, isNullable, dataType string
		var maxLenStr, colDefault, numPrecision, numRadix *string
		var isPK, isUnique bool
		err = rows.Scan(&colName, &colDefault, &isNullable, &dataType, &maxLenStr, &numPrecision, &numRadix, &isPK, &isUnique)
		if err != nil {
			return nil, nil, err
		}

		// fmt.Println(args, colName, isNullable, dataType, maxLenStr, colDefault, numPrecision, numRadix, isPK, isUnique)
		var maxLen int
		if maxLenStr != nil {
			maxLen, err = strconv.Atoi(*maxLenStr)
			if err != nil {
				return nil, nil, err
			}
		}

		col.Name = strings.Trim(colName, `" `)

		if colDefault != nil || isPK {
			if isPK {
				col.IsPrimaryKey = true
			} else {
				col.Default = *colDefault
			}
		}

		if colDefault != nil && strings.HasPrefix(*colDefault, "nextval(") {
			col.IsAutoIncrement = true
		}

		col.Nullable = (isNullable == "YES")

		switch dataType {
		case "character varying", "character":
			col.SQLType = core.SQLType{Name: core.Varchar, DefaultLength: 0, DefaultLength2: 0}
		case "timestamp without time zone":
			col.SQLType = core.SQLType{Name: core.DateTime, DefaultLength: 0, DefaultLength2: 0}
		case "timestamp with time zone":
			col.SQLType = core.SQLType{Name: core.TimeStampz, DefaultLength: 0, DefaultLength2: 0}
		case "double precision":
			col.SQLType = core.SQLType{Name: core.Double, DefaultLength: 0, DefaultLength2: 0}
		case "boolean":
			col.SQLType = core.SQLType{Name: core.Bool, DefaultLength: 0, DefaultLength2: 0}
		case "time without time zone":
			col.SQLType = core.SQLType{Name: core.Time, DefaultLength: 0, DefaultLength2: 0}
		case "oid":
			col.SQLType = core.SQLType{Name: core.BigInt, DefaultLength: 0, DefaultLength2: 0}
		default:
			col.SQLType = core.SQLType{Name: strings.ToUpper(dataType), DefaultLength: 0, DefaultLength2: 0}
		}
		if _, ok := core.SqlTypes[col.SQLType.Name]; !ok {
			return nil, nil, fmt.Errorf("Unknown colType: %v", dataType)
		}

		col.Length = maxLen

		if col.SQLType.IsText() || col.SQLType.IsTime() {
			if col.Default != "" {
				col.Default = "'" + col.Default + "'"
			} else {
				if col.DefaultIsEmpty {
					col.Default = "''"
				}
			}
		}
		cols[col.Name] = col
		colSeq = append(colSeq, col.Name)
	}

	return colSeq, cols, nil
}

func (db *db2) GetTables() ([]*core.Table, error) {
	args := []interface{}{}
	s := "SELECT tablename FROM pg_tables"
	if len(db.Schema) != 0 {
		args = append(args, db.Schema)
		s = s + " WHERE schemaname = $1"
	}

	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*core.Table, 0)
	for rows.Next() {
		table := core.NewEmptyTable()
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		table.Name = name
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *db2) GetIndexes(tableName string) (map[string]*core.Index, error) {
	args := []interface{}{tableName}
	s := fmt.Sprintf("SELECT indexname, indexdef FROM pg_indexes WHERE tablename=$1")
	if len(db.Schema) != 0 {
		args = append(args, db.Schema)
		s = s + " AND schemaname=$2"
	}
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*core.Index, 0)
	for rows.Next() {
		var indexType int
		var indexName, indexdef string
		var colNames []string
		err = rows.Scan(&indexName, &indexdef)
		if err != nil {
			return nil, err
		}
		indexName = strings.Trim(indexName, `" `)
		if strings.HasSuffix(indexName, "_pkey") {
			continue
		}
		if strings.HasPrefix(indexdef, "CREATE UNIQUE INDEX") {
			indexType = core.UniqueType
		} else {
			indexType = core.IndexType
		}
		colNames = getIndexColName(indexdef)
		var isRegular bool
		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "UQE_"+tableName) {
			newIdxName := indexName[5+len(tableName):]
			isRegular = true
			if newIdxName != "" {
				indexName = newIdxName
			}
		}

		index := &core.Index{Name: indexName, Type: indexType, Cols: make([]string, 0)}
		for _, colName := range colNames {
			index.Cols = append(index.Cols, strings.Trim(colName, `" `))
		}
		index.IsRegular = isRegular
		indexes[index.Name] = index
	}
	return indexes, nil
}

func (db *db2) Filters() []core.Filter {
	return []core.Filter{&core.QuoteFilter{}}
}

type db2Driver struct{}

func (p *db2Driver) Parse(driverName, dataSourceName string) (*core.Uri, error) {
	var dbName string

	kv := strings.Split(dataSourceName, ";")
	for _, c := range kv {
		vv := strings.Split(strings.TrimSpace(c), "=")
		if len(vv) == 2 {
			switch strings.ToLower(vv[0]) {
			case "database":
				dbName = vv[1]
			}
		}
	}

	if dbName == "" {
		return nil, errors.New("no db name provided")
	}
	return &core.Uri{DbName: dbName, DbType: core.MSSQL}, nil
}
