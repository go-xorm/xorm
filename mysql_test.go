package xorm

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

/*
CREATE DATABASE IF NOT EXISTS xorm_test CHARACTER SET
utf8 COLLATE utf8_general_ci;
*/

func TestMysql(t *testing.T) {
	engine, err := NewEngine("mysql", "root:@/xorm_test?charset=utf8")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.ShowSQL = true

	testAll(engine, t)
	testAll2(engine, t)
}
