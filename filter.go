package xorm

import (
    "fmt"
    "strings"
)

// Filter is an interface to filter SQL
type Filter interface {
    Do(sql string, session *Session) string
}

// PgSeqFilter filter SQL replace ?, ? ... to $1, $2 ...
type PgSeqFilter struct {
}

func (s *PgSeqFilter) Do(sql string, session *Session) string {
    segs := strings.Split(sql, "?")
    size := len(segs)
    res := ""
    for i, c := range segs {
        if i < size-1 {
            res += c + fmt.Sprintf("$%v", i+1)
        }
    }
    res += segs[size-1]
    return res
}

// QuoteFilter filter SQL replace ` to database's own quote character
type QuoteFilter struct {
}

func (s *QuoteFilter) Do(sql string, session *Session) string {
    return strings.Replace(sql, "`", session.Engine.QuoteStr(), -1)
}

// IdFilter filter SQL replace (id) to primary key column name
type IdFilter struct {
}

func (i *IdFilter) Do(sql string, session *Session) string {
    if session.Statement.RefTable != nil && session.Statement.RefTable.PrimaryKey != "" {
        sql = strings.Replace(sql, "`(id)`", session.Engine.Quote(session.Statement.RefTable.PrimaryKey), -1)
        sql = strings.Replace(sql, session.Engine.Quote("(id)"), session.Engine.Quote(session.Statement.RefTable.PrimaryKey), -1)
        return strings.Replace(sql, "(id)", session.Engine.Quote(session.Statement.RefTable.PrimaryKey), -1)
    }
    return sql
}
