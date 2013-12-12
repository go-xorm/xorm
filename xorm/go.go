package main

import (
    "errors"
    "fmt"
    "github.com/lunny/xorm"
    "go/format"
    "reflect"
    "strings"
    "text/template"
)

var (
    GoLangTmpl LangTmpl = LangTmpl{
        template.FuncMap{"Mapper": mapper.Table2Obj,
            "Type":    typestring,
            "Tag":     tag,
            "UnTitle": unTitle,
            "gt":      gt,
            "getCol":  getCol,
        },
        formatGo,
        genGoImports,
    }
)

var (
    errBadComparisonType = errors.New("invalid type for comparison")
    errBadComparison     = errors.New("incompatible types for comparison")
    errNoComparison      = errors.New("missing argument for comparison")
)

type kind int

const (
    invalidKind kind = iota
    boolKind
    complexKind
    intKind
    floatKind
    integerKind
    stringKind
    uintKind
)

func basicKind(v reflect.Value) (kind, error) {
    switch v.Kind() {
    case reflect.Bool:
        return boolKind, nil
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        return intKind, nil
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
        return uintKind, nil
    case reflect.Float32, reflect.Float64:
        return floatKind, nil
    case reflect.Complex64, reflect.Complex128:
        return complexKind, nil
    case reflect.String:
        return stringKind, nil
    }
    return invalidKind, errBadComparisonType
}

// eq evaluates the comparison a == b || a == c || ...
func eq(arg1 interface{}, arg2 ...interface{}) (bool, error) {
    v1 := reflect.ValueOf(arg1)
    k1, err := basicKind(v1)
    if err != nil {
        return false, err
    }
    if len(arg2) == 0 {
        return false, errNoComparison
    }
    for _, arg := range arg2 {
        v2 := reflect.ValueOf(arg)
        k2, err := basicKind(v2)
        if err != nil {
            return false, err
        }
        if k1 != k2 {
            return false, errBadComparison
        }
        truth := false
        switch k1 {
        case boolKind:
            truth = v1.Bool() == v2.Bool()
        case complexKind:
            truth = v1.Complex() == v2.Complex()
        case floatKind:
            truth = v1.Float() == v2.Float()
        case intKind:
            truth = v1.Int() == v2.Int()
        case stringKind:
            truth = v1.String() == v2.String()
        case uintKind:
            truth = v1.Uint() == v2.Uint()
        default:
            panic("invalid kind")
        }
        if truth {
            return true, nil
        }
    }
    return false, nil
}

// lt evaluates the comparison a < b.
func lt(arg1, arg2 interface{}) (bool, error) {
    v1 := reflect.ValueOf(arg1)
    k1, err := basicKind(v1)
    if err != nil {
        return false, err
    }
    v2 := reflect.ValueOf(arg2)
    k2, err := basicKind(v2)
    if err != nil {
        return false, err
    }
    if k1 != k2 {
        return false, errBadComparison
    }
    truth := false
    switch k1 {
    case boolKind, complexKind:
        return false, errBadComparisonType
    case floatKind:
        truth = v1.Float() < v2.Float()
    case intKind:
        truth = v1.Int() < v2.Int()
    case stringKind:
        truth = v1.String() < v2.String()
    case uintKind:
        truth = v1.Uint() < v2.Uint()
    default:
        panic("invalid kind")
    }
    return truth, nil
}

// le evaluates the comparison <= b.
func le(arg1, arg2 interface{}) (bool, error) {
    // <= is < or ==.
    lessThan, err := lt(arg1, arg2)
    if lessThan || err != nil {
        return lessThan, err
    }
    return eq(arg1, arg2)
}

// gt evaluates the comparison a > b.
func gt(arg1, arg2 interface{}) (bool, error) {
    // > is the inverse of <=.
    lessOrEqual, err := le(arg1, arg2)
    if err != nil {
        return false, err
    }
    return !lessOrEqual, nil
}

func getCol(cols map[string]*xorm.Column, name string) *xorm.Column {
    return cols[name]
}

func formatGo(src string) (string, error) {
    source, err := format.Source([]byte(src))
    if err != nil {
        return "", err
    }
    return string(source), nil
}

func genGoImports(tables []*xorm.Table) map[string]string {
    imports := make(map[string]string)

    for _, table := range tables {
        for _, col := range table.Columns {
            if typestring(col) == "time.Time" {
                imports["time"] = "time"
            }
        }
    }
    return imports
}

func typestring(col *xorm.Column) string {
    st := col.SQLType
    /*if col.IsPrimaryKey {
        return "int64"
    }*/
    t := xorm.SQLType2Type(st)
    s := t.String()
    if s == "[]uint8" {
        return "[]byte"
    }
    return s
}

func tag(table *xorm.Table, col *xorm.Column) string {
    isNameId := (mapper.Table2Obj(col.Name) == "Id")
    isIdPk := isNameId && typestring(col) == "int64"

    res := make([]string, 0)
    if !col.Nullable {
        if !isIdPk {
            res = append(res, "not null")
        }
    }
    if col.IsPrimaryKey {
        if !isIdPk {
            res = append(res, "pk")
        }
    }
    if col.Default != "" {
        res = append(res, "default "+col.Default)
    }
    if col.IsAutoIncrement {
        if !isIdPk {
            res = append(res, "autoincr")
        }
    }
    if col.IsCreated {
        res = append(res, "created")
    }
    if col.IsUpdated {
        res = append(res, "updated")
    }
    for name, _ := range col.Indexes {
        index := table.Indexes[name]
        var uistr string
        if index.Type == xorm.UniqueType {
            uistr = "unique"
        } else if index.Type == xorm.IndexType {
            uistr = "index"
        }
        if len(index.Cols) > 1 {
            uistr += "(" + index.Name + ")"
        }
        res = append(res, uistr)
    }

    nstr := col.SQLType.Name
    if col.Length != 0 {
        if col.Length2 != 0 {
            nstr += fmt.Sprintf("(%v, %v)", col.Length, col.Length2)
        } else {
            nstr += fmt.Sprintf("(%v)", col.Length)
        }
    }
    res = append(res, nstr)

    var tags []string
    if genJson {
        tags = append(tags, "json:\""+col.Name+"\"")
    }
    if len(res) > 0 {
        tags = append(tags, "xorm:\""+strings.Join(res, " ")+"\"")
    }
    if len(tags) > 0 {
        return "`" + strings.Join(tags, " ") + "`"
    } else {
        return ""
    }
}
