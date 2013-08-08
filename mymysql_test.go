package xorm

import (
	_ "github.com/ziutek/mymysql/godrv"
	"testing"
)

/*
CREATE DATABASE IF NOT EXISTS xorm_test CHARACTER SET
utf8 COLLATE utf8_general_ci;
*/

func TestMyMysql(t *testing.T) {
	engine, err := NewEngine("mymysql", "xorm_test/root/")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.ShowSQL = true

	testAll(engine, t)
}
