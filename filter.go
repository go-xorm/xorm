package xorm

import (
	"fmt"
	"strings"
)

type Filter interface {
	Do(sql string, session *Session) string
}

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

type PgQuoteFilter struct {
}

func (s *PgQuoteFilter) Do(sql string, session *Session) string {
	return strings.Replace(sql, "`", session.Engine.QuoteStr(), -1)
}

type IdFilter struct {
}

func (i *IdFilter) Do(sql string, session *Session) string {
	if session.Statement.RefTable != nil && session.Statement.RefTable.PrimaryKey != "" {
		return strings.Replace(sql, "(id)", session.Engine.Quote(session.Statement.RefTable.PrimaryKey), -1)
	}
	return sql
}
