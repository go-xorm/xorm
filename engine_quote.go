// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"strings"

	"xorm.io/builder"
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

// Quoter represents an object has Quote method
type Quoter interface {
	Quotes() (byte, byte)
	QuotePolicy() QuotePolicy
	IsReserved(string) bool
	WriteTo(w *builder.BytesWriter, value string) error
}

type quoter struct {
	dialect     core.Dialect
	quotePolicy QuotePolicy
}

func newQuoter(dialect core.Dialect, quotePolicy QuotePolicy) Quoter {
	return &quoter{
		dialect:     dialect,
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

func (q *quoter) IsReserved(value string) bool {
	return q.dialect.IsReserved(value)
}

func (q *quoter) needQuote(value string) bool {
	return q.quotePolicy == QuoteAddAlways || (q.quotePolicy == QuoteAddReserved && q.IsReserved(value))
}

func (q *quoter) WriteTo(w *builder.BytesWriter, name string) error {
	leftQuote, rightQuote := q.Quotes()
	needQuote := q.needQuote(name)
	if needQuote && name[0] != '`' {
		if err := w.WriteByte(leftQuote); err != nil {
			return err
		}
	}
	if _, err := w.WriteString(name); err != nil {
		return err
	}
	if needQuote && name[len(name)-1] != '`' {
		if err := w.WriteByte(rightQuote); err != nil {
			return err
		}
	}
	return nil
}

func quoteColumns(quoter Quoter, columnStr string) string {
	columns := strings.Split(columnStr, ",")
	return quoteJoin(quoter, columns)
}

func quoteJoin(quoter Quoter, columns []string) string {
	for i := 0; i < len(columns); i++ {
		columns[i] = quote(quoter, columns[i])
	}
	return strings.Join(columns, ",")
}

// quote Use QuoteStr quote the string sql
func quote(quoter Quoter, value string) string {
	buf := strings.Builder{}
	quoteTo(quoter, &buf, value)
	return buf.String()
}

// Quote add quotes to the value
func (engine *Engine) quote(value string, isColumn bool) string {
	if isColumn {
		return quote(engine.colQuoter, value)
	}
	return quote(engine.tableQuoter, value)
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

// IsReserved return true if the value is a reserved word of the database
func (engine *Engine) IsReserved(value string) bool {
	return engine.dialect.IsReserved(value)
}

// SetTableQuotePolicy set table quote policy
func (engine *Engine) SetTableQuotePolicy(policy QuotePolicy) {
	engine.tableQuoter = newQuoter(engine.dialect, policy)
}

// SetColumnQuotePolicy set column quote policy
func (engine *Engine) SetColumnQuotePolicy(policy QuotePolicy) {
	engine.colQuoter = newQuoter(engine.dialect, policy)
}

// quoteTo quotes string and writes into the buffer
func quoteTo(quoter Quoter, buf *strings.Builder, value string) {
	left, right := quoter.Quotes()
	if (quoter.QuotePolicy() == QuoteAddAlways) ||
		(quoter.QuotePolicy() == QuoteAddReserved && quoter.IsReserved(value)) {
		realQuoteTo(left, right, buf, value)
		return
	}
	buf.WriteString(value)
}

func realQuoteTo(quoteLeft, quoteRight byte, buf *strings.Builder, value string) {
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
