package xorm

import (
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

var se Engine

func TestSqlite(t *testing.T) {
	os.Remove("./test.db")
	se = Create("sqlite3", "./test.db")
	se.ShowSQL = true
}

func TestSqliteCreateTable(t *testing.T) {
	directCreateTable(&se, t)
}

func TestSqliteMapper(t *testing.T) {
	mapper(&se, t)
}

func TestSqliteInsert(t *testing.T) {
	insert(&se, t)
}

func TestSqliteQuery(t *testing.T) {
	query(&se, t)
}

func TestSqliteExec(t *testing.T) {
	exec(&se, t)
}

func TestSqliteInsertAutoIncr(t *testing.T) {
	insertAutoIncr(&se, t)
}

type sss struct {
}

func (s sss) TestInsertMulti(t *testing.T) {
	insertMulti(&se, t)
}

func TestSqliteInsertMulti(t *testing.T) {
	insertMulti(&se, t)

	insertTwoTable(&se, t)
	update(&se, t)
	testdelete(&se, t)
	get(&se, t)
	cascadeGet(&se, t)
	find(&se, t)
	findMap(&se, t)
	count(&se, t)
	where(&se, t)
	in(&se, t)
	limit(&se, t)
	order(&se, t)
	join(&se, t)
	having(&se, t)
	transaction(&se, t)
	combineTransaction(&se, t)
	table(&se, t)
	createMultiTables(&se, t)
	tableOp(&se, t)
}
