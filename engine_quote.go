// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"strings"

	"xorm.io/core"
)

// QuotePolicy describes quote handle policy
type QuotePolicy int

// All QuotePolicies
const (
	QuoteAddAlways QuotePolicy = iota
	QuoteNoAdd
	QuoteAddReserved
)

// QuoteMode quote on which types
type QuoteMode int

// All QuoteModes
const (
	QuoteTableAndColumns QuoteMode = iota
	QuoteTableOnly
	QuoteColumnsOnly
)

// Quoter represents an object has Quote method
type Quoter interface {
	Quotes() (byte, byte)
	QuotePolicy() QuotePolicy
	QuoteMode() QuoteMode
	IsReserved(string) bool
}

type quoter struct {
	dialect     core.Dialect
	quoteMode   QuoteMode
	quotePolicy QuotePolicy
}

func newQuoter(dialect core.Dialect, quoteMode QuoteMode, quotePolicy QuotePolicy) Quoter {
	return &quoter{
		dialect:     dialect,
		quoteMode:   quoteMode,
		quotePolicy: quotePolicy,
	}
}

func (q *quoter) Quotes() (byte, byte) {
	quotes := q.dialect.Quote("")
	return quotes[0], quotes[1]
}

func (q *quoter) QuotePolicy() QuotePolicy {
	return q.quotePolicy
}

func (q *quoter) QuoteMode() QuoteMode {
	return q.quoteMode
}

func (q *quoter) IsReserved(value string) bool {
	return q.dialect.IsReserved(value)
}

func quoteColumns(quoter Quoter, columnStr string) string {
	columns := strings.Split(columnStr, ",")
	return quoteJoin(quoter, columns)
}

func quoteJoin(quoter Quoter, columns []string) string {
	for i := 0; i < len(columns); i++ {
		columns[i] = quote(quoter, columns[i], true)
	}
	return strings.Join(columns, ",")
}

// quote Use QuoteStr quote the string sql
func quote(quoter Quoter, value string, isColumn bool) string {
	buf := strings.Builder{}
	quoteTo(quoter, &buf, value, isColumn)
	return buf.String()
}

// Quote add quotes to the value
func (engine *Engine) quote(value string, isColumn bool) string {
	return quote(engine, value, isColumn)
}

// Quote add quotes to the value
func (engine *Engine) Quote(value string, isColumn bool) string {
	return engine.quote(value, isColumn)
}

// Quotes return the left quote and right quote
func (engine *Engine) Quotes() (byte, byte) {
	quotes := engine.dialect.Quote("")
	return quotes[0], quotes[1]
}

// QuoteMode returns quote mode
func (engine *Engine) QuoteMode() QuoteMode {
	return engine.quoteMode
}

// QuotePolicy returns quote policy
func (engine *Engine) QuotePolicy() QuotePolicy {
	return engine.quotePolicy
}

// IsReserved return true if the value is a reserved word of the database
func (engine *Engine) IsReserved(value string) bool {
	return engine.dialect.IsReserved(value)
}

// quoteTo quotes string and writes into the buffer
func quoteTo(quoter Quoter, buf *strings.Builder, value string, isColumn bool) {
	if isColumn {
		if quoter.QuoteMode() == QuoteTableAndColumns ||
			quoter.QuoteMode() == QuoteColumnsOnly {
			if quoter.QuotePolicy() == QuoteAddAlways {
				realQuoteTo(quoter, buf, value)
				return
			} else if quoter.QuotePolicy() == QuoteAddReserved && quoter.IsReserved(value) {
				realQuoteTo(quoter, buf, value)
				return
			}
		}
		buf.WriteString(value)
		return
	}

	if quoter.QuoteMode() == QuoteTableAndColumns ||
		quoter.QuoteMode() == QuoteTableOnly {
		if quoter.QuotePolicy() == QuoteAddAlways {
			realQuoteTo(quoter, buf, value)
			return
		} else if quoter.QuotePolicy() == QuoteAddReserved && quoter.IsReserved(value) {
			realQuoteTo(quoter, buf, value)
			return
		}
	}
	buf.WriteString(value)
	return
}

func realQuoteTo(quoter Quoter, buf *strings.Builder, value string) {
	if buf == nil {
		return
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return
	} else if value == "*" {
		buf.WriteString("*")
		return
	}

	quoteLeft, quoteRight := quoter.Quotes()

	if value[0] == '`' || value[0] == quoteLeft { // no quote
		_, _ = buf.WriteString(value)
		return
	} else {
		_ = buf.WriteByte(quoteLeft)
		for i := 0; i < len(value); i++ {
			if value[i] == '.' {
				_ = buf.WriteByte(quoteRight)
				_ = buf.WriteByte('.')
				_ = buf.WriteByte(quoteLeft)
			} else {
				_ = buf.WriteByte(value[i])
			}
		}
		_ = buf.WriteByte(quoteRight)
	}
}

func unQuote(quoter Quoter, value string) string {
	left, right := quoter.Quotes()
	return strings.Trim(value, fmt.Sprintf("%v%v`", left, right))
}

func quoteJoinFunc(cols []string, quoteFunc func(string) string, sep string) string {
	for i := range cols {
		cols[i] = quoteFunc(cols[i])
	}
	return strings.Join(cols, sep+" ")
}
