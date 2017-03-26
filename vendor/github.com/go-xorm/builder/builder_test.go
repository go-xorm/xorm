// Copyright 2016 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"fmt"
	"reflect"
	"testing"
)

type MyInt int

func TestBuilderCond(t *testing.T) {
	var cases = []struct {
		cond Cond
		sql  string
		args []interface{}
	}{
		{
			Eq{"a": 1}.And(Like{"b", "c"}).Or(Eq{"a": 2}.And(Like{"b", "g"})),
			"(a=? AND b LIKE ?) OR (a=? AND b LIKE ?)",
			[]interface{}{1, "%c%", 2, "%g%"},
		},
		{
			Eq{"a": 1}.Or(Like{"b", "c"}).And(Eq{"a": 2}.Or(Like{"b", "g"})),
			"(a=? OR b LIKE ?) AND (a=? OR b LIKE ?)",
			[]interface{}{1, "%c%", 2, "%g%"},
		},
		{
			Eq{"d": []string{"e", "f"}},
			"d IN (?,?)",
			[]interface{}{"e", "f"},
		},
		{
			Neq{"d": []string{"e", "f"}},
			"d NOT IN (?,?)",
			[]interface{}{"e", "f"},
		},
		{
			Lt{"d": 3},
			"d<?",
			[]interface{}{3},
		},
		{
			Lte{"d": 3},
			"d<=?",
			[]interface{}{3},
		},
		{
			Gt{"d": 3},
			"d>?",
			[]interface{}{3},
		},
		{
			Gte{"d": 3},
			"d>=?",
			[]interface{}{3},
		},
		{
			Between{"d", 0, 2},
			"d BETWEEN ? AND ?",
			[]interface{}{0, 2},
		},
		{
			IsNull{"d"},
			"d IS NULL",
			[]interface{}{},
		},
		{
			NotIn("a", 1, 2).And(NotIn("b", "c", "d")),
			"a NOT IN (?,?) AND b NOT IN (?,?)",
			[]interface{}{1, 2, "c", "d"},
		},
		{
			In("a", 1, 2).Or(In("b", "c", "d")),
			"a IN (?,?) OR b IN (?,?)",
			[]interface{}{1, 2, "c", "d"},
		},
		{
			In("a", []int{1, 2}).Or(In("b", []string{"c", "d"})),
			"a IN (?,?) OR b IN (?,?)",
			[]interface{}{1, 2, "c", "d"},
		},
		{
			In("a", Expr("select id from x where name > ?", "b")),
			"a IN (select id from x where name > ?)",
			[]interface{}{"b"},
		},
		{
			In("a", []MyInt{1, 2}).Or(In("b", []string{"c", "d"})),
			"a IN (?,?) OR b IN (?,?)",
			[]interface{}{MyInt(1), MyInt(2), "c", "d"},
		},
		{
			In("a", []int{}),
			"a IN ()",
			[]interface{}{},
		},
		{
			NotIn("a", Expr("select id from x where name > ?", "b")),
			"a NOT IN (select id from x where name > ?)",
			[]interface{}{"b"},
		},
		{
			NotIn("a", []int{}),
			"a NOT IN ()",
			[]interface{}{},
		},
		// FIXME: since map will not guarantee the sequence, this may be failed random
		/*{
			Or(Eq{"a": 1, "b": 2}, Eq{"c": 3, "d": 4}),
			"(a=? AND b=?) OR (c=? AND d=?)",
			[]interface{}{1, 2, 3, 4},
		},*/
	}

	for _, k := range cases {
		sql, args, err := ToSQL(k.cond)
		if err != nil {
			t.Error(err)
			return
		}
		if sql != k.sql {
			t.Error("want", k.sql, "get", sql)
			return
		}
		fmt.Println(sql)

		if !(len(args) == 0 && len(k.args) == 0) {
			if !reflect.DeepEqual(args, k.args) {
				t.Error("want", k.args, "get", args)
				return
			}
		}
		fmt.Println(args)
	}
}

func TestBuilderSelect(t *testing.T) {
	sql, args, err := Select("c, d").From("table1").Where(Eq{"a": 1}).ToSQL()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(sql, args)

	sql, args, err = Select("c, d").From("table1").LeftJoin("table2", Eq{"table1.id": 1}.And(Lt{"table2.id": 3})).
		RightJoin("table3", "table2.id = table3.tid").Where(Eq{"a": 1}).ToSQL()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(sql, args)
}

func TestBuilderInsert(t *testing.T) {
	sql, args, err := Insert(Eq{"c": 1, "d": 2}).Into("table1").ToSQL()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(sql, args)
}

func TestBuilderUpdate(t *testing.T) {
	sql, args, err := Update(Eq{"a": 2}).From("table1").Where(Eq{"a": 1}).ToSQL()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(sql, args)

	sql, args, err = Update(Eq{"a": 2, "b": Incr(1)}).From("table2").Where(Eq{"a": 1}).ToSQL()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(sql, args)

	sql, args, err = Update(Eq{"a": 2, "b": Incr(1), "c": Decr(1), "d": Expr("select count(*) from table2")}).From("table2").Where(Eq{"a": 1}).ToSQL()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(sql, args)
}

func TestBuilderDelete(t *testing.T) {
	sql, args, err := Delete(Eq{"a": 1}).From("table1").ToSQL()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(sql, args)
}
