package xorm

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

var me Engine

/*
CREATE DATABASE IF NOT EXISTS xorm_test CHARACTER SET
utf8 COLLATE utf8_general_ci;
*/

func TestMysql(t *testing.T) {
	// You should drop all tables before executing this testing
	engine, err := NewEngine("mysql", "root:@/xorm_test?charset=utf8")
	if err != nil {
		t.Error(err)
		return
	}
	me = *engine
	me.ShowSQL = true

	directCreateTable(&me, t)
	mapper(&me, t)
	insert(&me, t)
	query(&me, t)
	exec(&me, t)
	insertAutoIncr(&me, t)
	insertMulti(&me, t)
	insertTwoTable(&me, t)
	update(&me, t)
	testdelete(&me, t)
	get(&me, t)
	cascadeGet(&me, t)
	find(&me, t)
	findMap(&me, t)
	count(&me, t)
	where(&me, t)
	in(&me, t)
	limit(&me, t)
	order(&me, t)
	join(&me, t)
	having(&me, t)
	transaction(&me, t)
	combineTransaction(&me, t)
	table(&me, t)
	createMultiTables(&me, t)
	tableOp(&me, t)
}
