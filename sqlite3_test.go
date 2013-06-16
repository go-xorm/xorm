package xorm

import (
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

var se *Engine

func autoConn() {
	if se == nil {
		os.Remove("./test.db")
		se, _ = NewEngine("sqlite3", "./test.db")
		se.ShowSQL = true
	}
}

func TestSqliteCreateTable(t *testing.T) {
	autoConn()
	directCreateTable(se, t)
}

func TestSqliteMapper(t *testing.T) {
	autoConn()
	mapper(se, t)
}

func TestSqliteInsert(t *testing.T) {
	autoConn()
	insert(se, t)
}

func TestSqliteQuery(t *testing.T) {
	autoConn()
	query(se, t)
}

func TestSqliteExec(t *testing.T) {
	autoConn()
	exec(se, t)
}

func TestSqliteInsertAutoIncr(t *testing.T) {
	autoConn()
	insertAutoIncr(se, t)
}

func TestInsertMulti(t *testing.T) {
	autoConn()
	insertMulti(se, t)
}

func TestSqliteInsertMulti(t *testing.T) {
	autoConn()
	insertMulti(se, t)
}

func TestSqliteInsertTwoTable(t *testing.T) {
	autoConn()
	insertTwoTable(se, t)
}

func TestSqliteUpdate(t *testing.T) {
	autoConn()
	update(se, t)
}

func TestSqliteDelete(t *testing.T) {
	autoConn()
	testdelete(se, t)
}

func TestSqliteGet(t *testing.T) {
	autoConn()
	get(se, t)
}

func TestSqliteCascadeGet(t *testing.T) {
	autoConn()
	cascadeGet(se, t)
}

func TestSqliteFind(t *testing.T) {
	autoConn()
	find(se, t)
}

func TestSqliteFindMap(t *testing.T) {
	autoConn()
	findMap(se, t)
}

func TestSqliteCount(t *testing.T) {
	autoConn()
	count(se, t)
}

func TestSqliteWhere(t *testing.T) {
	autoConn()
	where(se, t)
}

func TestSqliteIn(t *testing.T) {
	autoConn()
	in(se, t)
}

func TestSqliteLimit(t *testing.T) {
	autoConn()
	limit(se, t)
}

func TestSqliteOrder(t *testing.T) {
	autoConn()
	order(se, t)
}

func TestSqliteJoin(t *testing.T) {
	autoConn()
	join(se, t)
}

func TestSqliteHaving(t *testing.T) {
	autoConn()
	having(se, t)
}

func TestSqliteTransaction(t *testing.T) {
	autoConn()
	transaction(se, t)
}

func TestSqliteCombineTransaction(t *testing.T) {
	autoConn()
	combineTransaction(se, t)
}

func TestSqliteTable(t *testing.T) {
	autoConn()
	table(se, t)
}

func TestSqliteCreateMultiTables(t *testing.T) {
	autoConn()
	createMultiTables(se, t)
}

func TestSqliteTableOp(t *testing.T) {
	autoConn()
	tableOp(se, t)
}
